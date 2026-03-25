using System.Net;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;

namespace LiterLm;

/// <summary>
/// HTTP client for the liter-lm unified LLM API.
/// </summary>
/// <remarks>
/// <para>
/// Speaks the OpenAI-compatible wire protocol directly — no FFI, no native libraries.
/// The model-name prefix selects the provider and endpoint
/// (e.g. <c>"groq/llama3-70b"</c> routes to Groq).
/// </para>
/// <para>
/// Implements <see cref="IDisposable"/>; dispose after use to release the underlying
/// <see cref="HttpClient"/>.
/// </para>
/// </remarks>
/// <example>
/// <code>
/// await using var client = new LlmClient(
///     apiKey: Environment.GetEnvironmentVariable("OPENAI_API_KEY")!);
///
/// var request = new ChatCompletionRequest(
///     Model: "gpt-4o-mini",
///     Messages: [new UserMessage("Hello!")],
///     MaxTokens: 256);
///
/// var response = await client.ChatAsync(request);
/// Console.WriteLine(response.Choices[0].Message.Content);
/// </code>
/// </example>
public sealed class LlmClient : IDisposable, IAsyncDisposable
{
    internal const string DefaultBaseUrl = "https://api.openai.com/v1";
    internal const int DefaultMaxRetries = 2;

    private readonly HttpClient _httpClient;
    private readonly int _maxRetries;

    /// <summary>
    /// Initializes a new <see cref="LlmClient"/> with the given API key.
    /// </summary>
    /// <param name="apiKey">
    /// The API key sent as <c>Authorization: Bearer &lt;key&gt;</c>.
    /// Never log or serialize this value.
    /// </param>
    /// <param name="baseUrl">
    /// Base URL for the API endpoint. Defaults to <c>https://api.openai.com/v1</c>.
    /// </param>
    /// <param name="maxRetries">
    /// Maximum number of retries for retryable errors (429, 5xx). Defaults to 2.
    /// </param>
    /// <param name="timeout">
    /// Request timeout. Defaults to 60 seconds.
    /// </param>
    public LlmClient(
        string apiKey,
        string baseUrl = DefaultBaseUrl,
        int maxRetries = DefaultMaxRetries,
        TimeSpan? timeout = null)
    {
        ArgumentNullException.ThrowIfNull(apiKey);
        if (maxRetries < 0) throw new ArgumentOutOfRangeException(nameof(maxRetries), "must be >= 0");

        _maxRetries = maxRetries;
        var normalizedBase = baseUrl.TrimEnd('/');

        _httpClient = new HttpClient
        {
            BaseAddress = new Uri(normalizedBase + "/"),
            Timeout = timeout ?? TimeSpan.FromSeconds(60),
        };
        _httpClient.DefaultRequestHeaders.Authorization =
            new AuthenticationHeaderValue("Bearer", apiKey);
        _httpClient.DefaultRequestHeaders.Accept.Add(
            new MediaTypeWithQualityHeaderValue("application/json"));
    }

    // ─── Public API ───────────────────────────────────────────────────────────

    /// <summary>
    /// Sends a chat completion request and returns the full response.
    /// </summary>
    /// <param name="request">The chat completion request.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns>The provider's chat completion response.</returns>
    /// <exception cref="LlmException">Thrown when the request fails for any reason.</exception>
    public async Task<ChatCompletionResponse> ChatAsync(
        ChatCompletionRequest request,
        CancellationToken cancellationToken = default)
    {
        var body = Serialize(request);
        var responseJson = await PostAsync("chat/completions", body, cancellationToken)
            .ConfigureAwait(false);
        return Deserialize<ChatCompletionResponse>(responseJson);
    }

    /// <summary>
    /// Sends an embedding request and returns the embedding response.
    /// </summary>
    /// <param name="request">The embedding request.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns>The provider's embedding response.</returns>
    /// <exception cref="LlmException">Thrown when the request fails for any reason.</exception>
    public async Task<EmbeddingResponse> EmbedAsync(
        EmbeddingRequest request,
        CancellationToken cancellationToken = default)
    {
        var body = Serialize(request);
        var responseJson = await PostAsync("embeddings", body, cancellationToken)
            .ConfigureAwait(false);
        return Deserialize<EmbeddingResponse>(responseJson);
    }

    /// <summary>
    /// Lists available models for the configured provider.
    /// </summary>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns>The list of available models.</returns>
    /// <exception cref="LlmException">Thrown when the request fails for any reason.</exception>
    public async Task<ModelsListResponse> ListModelsAsync(
        CancellationToken cancellationToken = default)
    {
        var responseJson = await GetAsync("models", cancellationToken).ConfigureAwait(false);
        return Deserialize<ModelsListResponse>(responseJson);
    }

    // ─── HTTP Internals ───────────────────────────────────────────────────────

    private async Task<string> PostAsync(string path, string body, CancellationToken ct)
    {
        LlmException? lastException = null;
        for (int attempt = 0; attempt <= _maxRetries; attempt++)
        {
            using var content = new StringContent(body, Encoding.UTF8, "application/json");
            try
            {
                using var response = await _httpClient
                    .PostAsync(path, content, ct)
                    .ConfigureAwait(false);
                return await HandleResponseAsync(response, ct).ConfigureAwait(false);
            }
            catch (LlmException ex) when (IsRetryable(ex) && attempt < _maxRetries)
            {
                lastException = ex;
            }
            catch (LlmException ex)
            {
                throw;
            }
            catch (TaskCanceledException ex) when (!ct.IsCancellationRequested)
            {
                throw new ProviderException(0, $"request timed out: {ex.Message}");
            }
        }

        throw lastException ?? new LlmException(LlmException.ErrorCodes.Unknown, "liter-lm: unknown error");
    }

    private async Task<string> GetAsync(string path, CancellationToken ct)
    {
        LlmException? lastException = null;
        for (int attempt = 0; attempt <= _maxRetries; attempt++)
        {
            try
            {
                using var response = await _httpClient
                    .GetAsync(path, ct)
                    .ConfigureAwait(false);
                return await HandleResponseAsync(response, ct).ConfigureAwait(false);
            }
            catch (LlmException ex) when (IsRetryable(ex) && attempt < _maxRetries)
            {
                lastException = ex;
            }
            catch (LlmException ex)
            {
                throw;
            }
            catch (TaskCanceledException ex) when (!ct.IsCancellationRequested)
            {
                throw new ProviderException(0, $"request timed out: {ex.Message}");
            }
        }

        throw lastException ?? new LlmException(LlmException.ErrorCodes.Unknown, "liter-lm: unknown error");
    }

    private static async Task<string> HandleResponseAsync(HttpResponseMessage response, CancellationToken ct)
    {
        var responseBody = await response.Content.ReadAsStringAsync(ct).ConfigureAwait(false);
        if (response.IsSuccessStatusCode)
        {
            return responseBody;
        }

        throw ClassifyHttpError((int)response.StatusCode, responseBody);
    }

    private static LlmException ClassifyHttpError(int status, string body)
    {
        var message = ExtractErrorMessage(body);
        return status switch
        {
            400 or 422 => new InvalidRequestException(message),
            401 or 403 => new AuthenticationException(message),
            404 => new NotFoundException(message),
            429 => new RateLimitException(message),
            _ => new ProviderException(status, message),
        };
    }

    private static bool IsRetryable(LlmException ex) =>
        ex is RateLimitException or ProviderException;

    private static string ExtractErrorMessage(string body)
    {
        if (string.IsNullOrWhiteSpace(body))
        {
            return "empty response body";
        }

        // Best-effort: extract {"error":{"message":"..."}} without a full round-trip parse
        var messageIdx = body.IndexOf("\"message\"", StringComparison.Ordinal);
        if (messageIdx >= 0)
        {
            var colon = body.IndexOf(':', messageIdx);
            var quote1 = body.IndexOf('"', colon + 1);
            var quote2 = body.IndexOf('"', quote1 + 1);
            if (quote1 >= 0 && quote2 > quote1)
            {
                return body[(quote1 + 1)..quote2];
            }
        }

        return body.Length > 200 ? body[..200] + "…" : body;
    }

    // ─── Serialization helpers ────────────────────────────────────────────────

    private static string Serialize<T>(T value)
    {
        try
        {
            return LiterLmJson.Serialize(value);
        }
        catch (JsonException ex)
        {
            throw new SerializationException("failed to serialize request", ex);
        }
    }

    private static T Deserialize<T>(string json)
    {
        try
        {
            return LiterLmJson.Deserialize<T>(json)
                ?? throw new SerializationException($"provider returned null for {typeof(T).Name}");
        }
        catch (JsonException ex)
        {
            throw new SerializationException($"failed to deserialize {typeof(T).Name} response", ex);
        }
    }

    // ─── IDisposable ──────────────────────────────────────────────────────────

    /// <summary>Releases the underlying <see cref="HttpClient"/>.</summary>
    public void Dispose() => _httpClient.Dispose();

    /// <summary>Asynchronously releases the underlying <see cref="HttpClient"/>.</summary>
    public ValueTask DisposeAsync()
    {
        _httpClient.Dispose();
        return ValueTask.CompletedTask;
    }
}

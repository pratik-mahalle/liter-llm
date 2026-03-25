package dev.kreuzberg.literlm;

import static dev.kreuzberg.literlm.Types.*;

import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

/**
 * HTTP client for the liter-lm unified LLM API.
 *
 * <p>
 * Speaks the OpenAI-compatible wire protocol directly — no FFI, no native
 * libraries. The model-name prefix selects the provider and endpoint (e.g.
 * {@code "groq/llama3-70b"} routes to Groq). Implements {@link AutoCloseable};
 * close after use to release the underlying {@link HttpClient} executor.
 *
 * <h2>Example</h2>
 *
 * <pre>{@code
 * try (var client = LlmClient.builder().apiKey(System.getenv("OPENAI_API_KEY")).build()) {
 * 	var request = ChatCompletionRequest.builder("gpt-4o-mini", List.of(new Types.UserMessage("Hello!")))
 * 			.maxTokens(256L).build();
 * 	var response = client.chat(request);
 * 	System.out.println(response.choices().getFirst().message().content());
 * }
 * }</pre>
 */
public final class LlmClient implements AutoCloseable {

	static final String DEFAULT_BASE_URL = "https://api.openai.com/v1";
	static final int DEFAULT_MAX_RETRIES = 2;
	static final Duration DEFAULT_TIMEOUT = Duration.ofSeconds(60);

	private static final int HTTP_OK_MIN = 200;
	private static final int HTTP_OK_MAX = 300;
	private static final int HTTP_BAD_REQUEST = 400;
	private static final int HTTP_UNAUTHORIZED = 401;
	private static final int HTTP_FORBIDDEN = 403;
	private static final int HTTP_NOT_FOUND = 404;
	private static final int HTTP_UNPROCESSABLE = 422;
	private static final int HTTP_RATE_LIMIT = 429;
	private static final int HTTP_SERVER_ERROR_MIN = 500;
	private static final int ERROR_BODY_MAX_LENGTH = 200;

	private final String apiKey;
	private final String baseUrl;
	private final int maxRetries;
	private final HttpClient httpClient;
	private final ObjectMapper objectMapper;

	private LlmClient(Builder builder) {
		this.apiKey = builder.apiKey;
		this.baseUrl = builder.baseUrl.endsWith("/")
				? builder.baseUrl.substring(0, builder.baseUrl.length() - 1)
				: builder.baseUrl;
		this.maxRetries = builder.maxRetries;
		this.httpClient = HttpClient.newBuilder().connectTimeout(builder.timeout).build();
		this.objectMapper = createObjectMapper();
	}

	// ─── Public API ───────────────────────────────────────────────────────────

	/**
	 * Sends a chat completion request and returns the full response.
	 *
	 * @param request
	 *            the chat completion request
	 * @return the provider's chat completion response
	 * @throws LlmException
	 *             if the request fails for any reason
	 */
	public ChatCompletionResponse chat(ChatCompletionRequest request) throws LlmException {
		String body = serialize(request);
		String responseBody = post("/chat/completions", body);
		return deserialize(responseBody, ChatCompletionResponse.class);
	}

	/**
	 * Sends an embedding request and returns the embedding response.
	 *
	 * @param request
	 *            the embedding request
	 * @return the provider's embedding response
	 * @throws LlmException
	 *             if the request fails for any reason
	 */
	public EmbeddingResponse embed(EmbeddingRequest request) throws LlmException {
		String body = serialize(request);
		String responseBody = post("/embeddings", body);
		return deserialize(responseBody, EmbeddingResponse.class);
	}

	/**
	 * Lists available models for the configured provider.
	 *
	 * @return the list of available models
	 * @throws LlmException
	 *             if the request fails for any reason
	 */
	public ModelsListResponse listModels() throws LlmException {
		String responseBody = get("/models");
		return deserialize(responseBody, ModelsListResponse.class);
	}

	/**
	 * Closes the underlying HTTP client, releasing resources.
	 *
	 * <p>
	 * After this method returns, the client must not be used.
	 */
	@Override
	public void close() {
		httpClient.close();
	}

	// ─── HTTP Internals ───────────────────────────────────────────────────────

	private String post(String path, String body) throws LlmException {
		var request = HttpRequest.newBuilder().uri(URI.create(baseUrl + path))
				.header("Authorization", "Bearer " + apiKey).header("Content-Type", "application/json")
				.header("Accept", "application/json").POST(HttpRequest.BodyPublishers.ofString(body)).build();
		return executeWithRetry(request);
	}

	private String get(String path) throws LlmException {
		var request = HttpRequest.newBuilder().uri(URI.create(baseUrl + path))
				.header("Authorization", "Bearer " + apiKey).header("Accept", "application/json").GET().build();
		return executeWithRetry(request);
	}

	private String executeWithRetry(HttpRequest request) throws LlmException {
		LlmException lastException = null;
		for (int attempt = 0; attempt <= maxRetries; attempt++) {
			try {
				var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
				int status = response.statusCode();
				String responseBody = response.body();
				if (status >= HTTP_OK_MIN && status < HTTP_OK_MAX) {
					return responseBody;
				}
				lastException = classifyHttpError(status, responseBody);
				// Only retry on 429 and 5xx
				if (status != HTTP_RATE_LIMIT && status < HTTP_SERVER_ERROR_MIN) {
					throw lastException;
				}
			} catch (LlmException e) {
				throw e;
			} catch (InterruptedException e) {
				Thread.currentThread().interrupt();
				lastException = new LlmException(LlmException.CODE_PROVIDER_ERROR,
						"liter-lm: HTTP request failed: " + e.getMessage(), e);
			} catch (java.io.IOException e) {
				lastException = new LlmException(LlmException.CODE_PROVIDER_ERROR,
						"liter-lm: HTTP request failed: " + e.getMessage(), e);
			}
		}
		throw lastException != null
				? lastException
				: new LlmException(LlmException.CODE_UNKNOWN, "liter-lm: unknown error");
	}

	private static LlmException classifyHttpError(int status, String body) {
		String message = extractErrorMessage(body);
		return switch (status) {
			case HTTP_BAD_REQUEST, HTTP_UNPROCESSABLE -> new LlmException.InvalidRequestException(message);
			case HTTP_UNAUTHORIZED, HTTP_FORBIDDEN -> new LlmException.AuthenticationException(message);
			case HTTP_NOT_FOUND -> new LlmException.NotFoundException(message);
			case HTTP_RATE_LIMIT -> new LlmException.RateLimitException(message);
			default -> new LlmException.ProviderException(status, message);
		};
	}

	private static String extractErrorMessage(String body) {
		if (body == null || body.isBlank()) {
			return "empty response body";
		}
		// Best-effort: extract {"error":{"message":"..."}} without full parse
		int msgIdx = body.indexOf("\"message\"");
		if (msgIdx >= 0) {
			int colon = body.indexOf(':', msgIdx);
			int quote1 = body.indexOf('"', colon + 1);
			int quote2 = body.indexOf('"', quote1 + 1);
			if (quote1 >= 0 && quote2 > quote1) {
				return body.substring(quote1 + 1, quote2);
			}
		}
		// Truncate long bodies for readability
		return body.length() > ERROR_BODY_MAX_LENGTH ? body.substring(0, ERROR_BODY_MAX_LENGTH) + "…" : body;
	}

	// ─── Serialization helpers ────────────────────────────────────────────────

	private String serialize(Object value) throws LlmException {
		try {
			return objectMapper.writeValueAsString(value);
		} catch (com.fasterxml.jackson.core.JsonProcessingException e) {
			throw new LlmException.SerializationException("failed to serialize request", e);
		}
	}

	private <T> T deserialize(String json, Class<T> type) throws LlmException {
		try {
			return objectMapper.readValue(json, type);
		} catch (com.fasterxml.jackson.core.JsonProcessingException e) {
			throw new LlmException.SerializationException("failed to deserialize " + type.getSimpleName() + " response",
					e);
		}
	}

	private static ObjectMapper createObjectMapper() {
		return new ObjectMapper().configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false)
				.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);
	}

	// ─── Builder ──────────────────────────────────────────────────────────────

	/**
	 * Returns a new {@link Builder} for constructing an {@link LlmClient}.
	 *
	 * @return a fresh builder
	 */
	public static Builder builder() {
		return new Builder();
	}

	/** Fluent builder for {@link LlmClient}. */
	public static final class Builder {

		private String apiKey = "";
		private String baseUrl = DEFAULT_BASE_URL;
		private int maxRetries = DEFAULT_MAX_RETRIES;
		private Duration timeout = DEFAULT_TIMEOUT;

		private Builder() {
		}

		/**
		 * Sets the API key sent as {@code Authorization: Bearer <key>}.
		 *
		 * <p>
		 * Reads from the environment when not explicitly set. Never log or serialize
		 * this value.
		 *
		 * @param apiKey
		 *            the API key
		 * @return this builder
		 */
		public Builder apiKey(String apiKey) {
			this.apiKey = apiKey;
			return this;
		}

		/**
		 * Sets the base URL for the API endpoint.
		 *
		 * <p>
		 * Defaults to {@value LlmClient#DEFAULT_BASE_URL}. Override to target a
		 * different provider or a local proxy.
		 *
		 * @param baseUrl
		 *            base URL without trailing slash
		 * @return this builder
		 */
		public Builder baseUrl(String baseUrl) {
			this.baseUrl = baseUrl;
			return this;
		}

		/**
		 * Sets the maximum number of retries for retryable errors (429, 5xx).
		 *
		 * <p>
		 * Defaults to {@value LlmClient#DEFAULT_MAX_RETRIES}.
		 *
		 * @param maxRetries
		 *            non-negative retry count
		 * @return this builder
		 */
		public Builder maxRetries(int maxRetries) {
			if (maxRetries < 0) {
				throw new IllegalArgumentException("maxRetries must be >= 0");
			}
			this.maxRetries = maxRetries;
			return this;
		}

		/**
		 * Sets the connection timeout.
		 *
		 * <p>
		 * Defaults to {@value} 60 seconds.
		 *
		 * @param timeout
		 *            positive duration
		 * @return this builder
		 */
		public Builder timeout(Duration timeout) {
			this.timeout = timeout;
			return this;
		}

		/**
		 * Builds the {@link LlmClient}.
		 *
		 * @return a configured client instance
		 */
		public LlmClient build() {
			return new LlmClient(this);
		}
	}
}

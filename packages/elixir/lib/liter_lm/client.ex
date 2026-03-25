defmodule LiterLm.Client do
  @moduledoc """
  HTTP client implementation for the liter-lm unified LLM API.

  Speaks the OpenAI-compatible wire protocol using `Req`. The model-name prefix
  selects the provider and endpoint (e.g. `"groq/llama3-70b"` routes to Groq).

  Use this module directly for advanced configuration, or use the top-level
  `LiterLm` module for the standard interface.

  ## Configuration

  | Option | Default | Description |
  |--------|---------|-------------|
  | `:api_key` | `""` | API key for `Authorization: Bearer` header |
  | `:base_url` | `"https://api.openai.com/v1"` | Provider base URL |
  | `:max_retries` | `2` | Retry count for 429 and 5xx errors |
  | `:receive_timeout` | `60_000` | Request timeout in milliseconds |

  ## Example

      client = LiterLm.Client.new(api_key: System.fetch_env!("OPENAI_API_KEY"))
      {:ok, response} = LiterLm.Client.chat(client, %{
        model: "gpt-4o-mini",
        messages: [%{role: "user", content: "Hello!"}],
        max_tokens: 256
      })

  """

  alias LiterLm.Error
  alias LiterLm.Types

  @default_base_url "https://api.openai.com/v1"
  @default_max_retries 2
  @default_timeout_ms 60_000

  @typedoc "An opaque client configuration map."
  @type t :: %__MODULE__{
          api_key: String.t(),
          base_url: String.t(),
          max_retries: non_neg_integer(),
          receive_timeout: pos_integer()
        }

  defstruct api_key: "",
            base_url: @default_base_url,
            max_retries: @default_max_retries,
            receive_timeout: @default_timeout_ms

  # ─── Constructor ──────────────────────────────────────────────────────────

  @doc """
  Creates a new client configuration.

  ## Options

  - `:api_key` — API key (required for live requests)
  - `:base_url` — Base URL, defaults to `"https://api.openai.com/v1"`
  - `:max_retries` — Retry count for 429/5xx, defaults to `#{@default_max_retries}`
  - `:receive_timeout` — Timeout in ms, defaults to `#{@default_timeout_ms}`

  ## Examples

      iex> LiterLm.Client.new(api_key: "sk-...")
      %LiterLm.Client{api_key: "sk-...", base_url: "https://api.openai.com/v1", ...}

  """
  @spec new(keyword()) :: t()
  def new(opts \\ []) do
    struct!(__MODULE__, opts)
  end

  # ─── Public API ───────────────────────────────────────────────────────────

  @doc """
  Sends a chat completion request and returns the full response.

  ## Parameters

  - `client` — client configuration from `new/1`
  - `request` — a `LiterLm.Types.chat_request()` map
  - `opts` — additional `Req` options (e.g. `plug:` for testing)

  ## Returns

  - `{:ok, LiterLm.Types.chat_response()}` on success
  - `{:error, LiterLm.Error.t()}` on failure

  ## Examples

      {:ok, response} = LiterLm.Client.chat(client, %{
        model: "gpt-4o-mini",
        messages: [LiterLm.Types.user_message("Hello!")],
        max_tokens: 256
      })
      hd(response.choices).message.content

  """
  @spec chat(t(), Types.chat_request(), keyword()) ::
          {:ok, Types.chat_response()} | {:error, Error.t()}
  def chat(%__MODULE__{} = client, request, opts \\ []) do
    post(client, "/chat/completions", request, opts)
  end

  @doc """
  Sends an embedding request and returns the embedding response.

  ## Parameters

  - `client` — client configuration from `new/1`
  - `request` — a `LiterLm.Types.embedding_request()` map
  - `opts` — additional `Req` options

  ## Returns

  - `{:ok, LiterLm.Types.embedding_response()}` on success
  - `{:error, LiterLm.Error.t()}` on failure

  """
  @spec embed(t(), Types.embedding_request(), keyword()) ::
          {:ok, Types.embedding_response()} | {:error, Error.t()}
  def embed(%__MODULE__{} = client, request, opts \\ []) do
    post(client, "/embeddings", request, opts)
  end

  @doc """
  Lists available models for the configured provider.

  ## Parameters

  - `client` — client configuration from `new/1`
  - `opts` — additional `Req` options

  ## Returns

  - `{:ok, LiterLm.Types.models_list_response()}` on success
  - `{:error, LiterLm.Error.t()}` on failure

  """
  @spec list_models(t(), keyword()) ::
          {:ok, Types.models_list_response()} | {:error, Error.t()}
  def list_models(%__MODULE__{} = client, opts \\ []) do
    get(client, "/models", opts)
  end

  # ─── HTTP Internals ───────────────────────────────────────────────────────

  defp post(%__MODULE__{} = client, path, body, extra_opts) do
    json_body =
      case Jason.encode(body) do
        {:ok, encoded} ->
          encoded

        {:error, reason} ->
          return = {:error, Error.serialization("failed to encode request: #{inspect(reason)}")}
          throw(return)
      end

    req = build_req(client, extra_opts)

    do_with_retry(client.max_retries, fn ->
      case Req.post(req, url: path, body: json_body) do
        {:ok, %{status: status, body: resp_body}} when status in 200..299 ->
          decode_response(resp_body)

        {:ok, %{status: status, body: resp_body}} ->
          message = Error.extract_message(resp_body)
          {:error, Error.from_http_status(status, message)}

        {:error, exception} ->
          {:error, Error.unknown("HTTP request failed: #{Exception.message(exception)}")}
      end
    end)
  catch
    {:error, _} = err -> err
  end

  defp get(%__MODULE__{} = client, path, extra_opts) do
    req = build_req(client, extra_opts)

    do_with_retry(client.max_retries, fn ->
      case Req.get(req, url: path) do
        {:ok, %{status: status, body: resp_body}} when status in 200..299 ->
          decode_response(resp_body)

        {:ok, %{status: status, body: resp_body}} ->
          message = Error.extract_message(resp_body)
          {:error, Error.from_http_status(status, message)}

        {:error, exception} ->
          {:error, Error.unknown("HTTP request failed: #{Exception.message(exception)}")}
      end
    end)
  end

  defp build_req(%__MODULE__{} = client, extra_opts) do
    base_url = String.trim_trailing(client.base_url, "/")

    req_opts =
      [
        base_url: base_url,
        headers: [
          {"authorization", "Bearer #{client.api_key}"},
          {"content-type", "application/json"},
          {"accept", "application/json"}
        ],
        receive_timeout: client.receive_timeout,
        decode_body: false
      ] ++ extra_opts

    Req.new(req_opts)
  end

  defp decode_response(body) when is_binary(body) do
    case Jason.decode(body) do
      {:ok, decoded} ->
        {:ok, decoded}

      {:error, reason} ->
        {:error, Error.serialization("failed to decode response: #{inspect(reason)}")}
    end
  end

  defp decode_response(body) when is_map(body) do
    # Req already decoded the JSON (shouldn't happen with decode_body: false, but be safe)
    {:ok, body}
  end

  # Retries on rate_limit and provider_error kinds.
  defp do_with_retry(0, fun), do: fun.()

  defp do_with_retry(retries_left, fun) do
    case fun.() do
      {:error, %Error{kind: kind}} = err when kind in [:rate_limit, :provider_error] ->
        if retries_left > 0 do
          do_with_retry(retries_left - 1, fun)
        else
          err
        end

      other ->
        other
    end
  end
end

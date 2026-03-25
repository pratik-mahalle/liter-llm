defmodule LiterLm do
  @moduledoc """
  Universal LLM API client for Elixir.

  Speaks the OpenAI-compatible HTTP wire protocol — no NIFs, no native libraries.
  The model-name prefix selects the provider and endpoint
  (e.g. `"groq/llama3-70b"` routes to Groq).

  ## Quick Start

      # Configure via application environment
      config :liter_lm,
        api_key: System.fetch_env!("OPENAI_API_KEY"),
        base_url: "https://api.openai.com/v1"

      # Or pass options directly per call
      {:ok, response} = LiterLm.chat(
        %{
          model: "gpt-4o-mini",
          messages: [LiterLm.Types.user_message("Hello!")],
          max_tokens: 256
        },
        api_key: System.fetch_env!("OPENAI_API_KEY")
      )

      hd(response["choices"])["message"]["content"]

  ## Configuration

  Options may be provided per-call or via application environment:

  | Key | Default | Description |
  |-----|---------|-------------|
  | `:api_key` | `""` | API key for `Authorization: Bearer` header |
  | `:base_url` | `"https://api.openai.com/v1"` | Provider base URL |
  | `:max_retries` | `2` | Retry count for 429 and 5xx errors |
  | `:receive_timeout` | `60_000` | Request timeout in milliseconds |

  ## Error Handling

  All functions return `{:ok, result}` or `{:error, %LiterLm.Error{}}`.
  Match on `:kind` for programmatic handling:

      case LiterLm.chat(request) do
        {:ok, response} -> process(response)
        {:error, %LiterLm.Error{kind: :rate_limit}} -> retry()
        {:error, %LiterLm.Error{kind: :authentication}} -> raise "Invalid API key"
        {:error, err} -> Logger.error("LLM error", error: inspect(err))
      end

  """

  alias LiterLm.Client
  alias LiterLm.Types

  # ─── Public API ───────────────────────────────────────────────────────────

  @doc """
  Sends a chat completion request and returns the full response.

  ## Parameters

  - `request` — a `LiterLm.Types.chat_request()` map. Required keys: `:model`, `:messages`.
  - `opts` — client options (see module docs) or extra `Req` options.

  ## Returns

  `{:ok, map()}` where the map matches `LiterLm.Types.chat_response()`, or
  `{:error, LiterLm.Error.t()}`.

  ## Examples

      {:ok, response} = LiterLm.chat(%{
        model: "gpt-4o-mini",
        messages: [
          LiterLm.Types.system_message("You are a helpful assistant."),
          LiterLm.Types.user_message("What is the capital of Germany?")
        ],
        max_tokens: 128
      })
      hd(response["choices"])["message"]["content"]
      #=> "The capital of Germany is Berlin."

  """
  @spec chat(Types.chat_request(), keyword()) ::
          {:ok, Types.chat_response()} | {:error, LiterLm.Error.t()}
  def chat(request, opts \\ []) do
    {client_opts, req_opts} = split_opts(opts)
    client = build_client(client_opts)
    Client.chat(client, request, req_opts)
  end

  @doc """
  Sends an embedding request and returns the embedding response.

  ## Parameters

  - `request` — a `LiterLm.Types.embedding_request()` map.
  - `opts` — client options or extra `Req` options.

  ## Returns

  `{:ok, map()}` where the map matches `LiterLm.Types.embedding_response()`, or
  `{:error, LiterLm.Error.t()}`.

  ## Examples

      {:ok, response} = LiterLm.embed(%{
        model: "text-embedding-3-small",
        input: "The quick brown fox"
      })
      hd(response["data"])["embedding"] |> length()

  """
  @spec embed(Types.embedding_request(), keyword()) ::
          {:ok, Types.embedding_response()} | {:error, LiterLm.Error.t()}
  def embed(request, opts \\ []) do
    {client_opts, req_opts} = split_opts(opts)
    client = build_client(client_opts)
    Client.embed(client, request, req_opts)
  end

  @doc """
  Lists available models for the configured provider.

  ## Parameters

  - `opts` — client options or extra `Req` options.

  ## Returns

  `{:ok, map()}` where the map matches `LiterLm.Types.models_list_response()`, or
  `{:error, LiterLm.Error.t()}`.

  ## Examples

      {:ok, response} = LiterLm.list_models()
      Enum.map(response["data"], & &1["id"])

  """
  @spec list_models(keyword()) ::
          {:ok, Types.models_list_response()} | {:error, LiterLm.Error.t()}
  def list_models(opts \\ []) do
    {client_opts, req_opts} = split_opts(opts)
    client = build_client(client_opts)
    Client.list_models(client, req_opts)
  end

  # ─── Helpers ──────────────────────────────────────────────────────────────

  # Known client config keys — all others are forwarded to Req.
  @client_keys [:api_key, :base_url, :max_retries, :receive_timeout]

  defp split_opts(opts) do
    Enum.split_with(opts, fn {key, _value} -> key in @client_keys end)
    |> then(fn {client, req} -> {client, req} end)
  end

  defp build_client(opts) do
    app_config = Application.get_all_env(:liter_lm)
    merged = Keyword.merge(app_config, opts)
    Client.new(merged)
  end
end

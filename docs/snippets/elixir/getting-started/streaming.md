```elixir
{:ok, chunks} =
  LiterLlm.Client.chat_stream(client, %{
    model: "openai/gpt-4o-mini",
    messages: [%{role: "user", content: "Hello"}]
  })

for chunk <- chunks, do: IO.inspect(chunk)
```

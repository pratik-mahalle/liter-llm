```ruby
# frozen_string_literal: true

require "liter_llm"
require "json"

client = LiterLlm::LlmClient.new(ENV.fetch("OPENAI_API_KEY"), {})

chunks = JSON.parse(client.chat_stream(JSON.generate(
  model: "openai/gpt-4o-mini",
  messages: [{ role: "user", content: "Hello" }]
)))

chunks.each { |chunk| puts chunk }
```

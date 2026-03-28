```typescript
import init, { LlmClient } from "@kreuzberg/liter-llm-wasm";

await init();

const client = new LlmClient({ apiKey: "sk-..." });
const stream = await client.chatStream({
  model: "openai/gpt-4o-mini",
  messages: [{ role: "user", content: "Hello" }],
});
// stream is a ReadableStream
```

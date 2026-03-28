```csharp
using LiterLlm;

await using var client = new LlmClient(
    apiKey: Environment.GetEnvironmentVariable("OPENAI_API_KEY")!);

var request = new ChatCompletionRequest(
    Model: "openai/gpt-4o-mini",
    Messages: [new UserMessage("Hello")]
);

await foreach (var chunk in client.ChatStreamAsync(request))
{
    Console.WriteLine(chunk);
}
```

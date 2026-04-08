```go
package main

import (
	"context"
	"fmt"

	llm "github.com/kreuzberg-dev/liter-llm/packages/go"
)

func main() {
	// No API key needed for local providers
	client := llm.NewClient(
		llm.WithAPIKey(""),
		llm.WithBaseURL("http://localhost:11434/v1"),
	)
	resp, err := client.Chat(context.Background(), &llm.ChatCompletionRequest{
		Model: "ollama/qwen2:0.5b",
		Messages: []llm.Message{
			llm.NewTextMessage(llm.RoleUser, "Hello!"),
		},
	})
	if err != nil {
		panic(err)
	}
	if len(resp.Choices) > 0 && resp.Choices[0].Message.Content != nil {
		fmt.Println(*resp.Choices[0].Message.Content)
	}
}
```

```java
import dev.kreuzberg.literllm.LlmClient;
import dev.kreuzberg.literllm.Types.*;
import java.util.List;

public class Main {
    public static void main(String[] args) throws Exception {
        try (var client = LlmClient.builder()
                .apiKey(System.getenv("OPENAI_API_KEY"))
                .build()) {
            client.chatStream(new ChatCompletionRequest(
                "openai/gpt-4o-mini",
                List.of(new UserMessage("Hello"))
            ), chunk -> System.out.println(chunk));
        }
    }
}
```

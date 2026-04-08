```python
import asyncio
from liter_llm import LlmClient

async def main() -> None:
    # No API key needed for local providers
    client = LlmClient(api_key="", base_url="http://localhost:11434/v1")
    response = await client.chat(
        model="ollama/qwen2:0.5b",
        messages=[{"role": "user", "content": "Hello!"}],
    )
    print(response.choices[0].message.content)

asyncio.run(main())
```

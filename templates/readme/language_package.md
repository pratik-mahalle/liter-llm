# {{ name | replace("#", "\\#") }}

{% include 'partials/badges.html' %}
{% include 'partials/banner.html' %}
{% include 'partials/discord.html' %}

{{ description }}

## Installation

{% include 'partials/installation.md' %}

## Quick Start

{% include 'partials/quick_start.md' %}

{% if language == "typescript" %}
{% include 'partials/napi_implementation.md' %}

{% endif %}

## Features

{% include 'partials/features.md' %}

{% if features.provider_routing %}

## Provider Routing

Route to 143+ providers using the `provider/model` prefix convention:

```text
openai/gpt-4o
anthropic/claude-3-5-sonnet-20241022
groq/llama-3.1-70b-versatile
mistral/mistral-large-latest
```

See the [provider registry](https://github.com/kreuzberg-dev/liter-llm/blob/main/schemas/providers.json) for the full list.

{% endif %}

{% include 'partials/proxy_server.md' %}

## Documentation

- **[Documentation](https://docs.liter-llm.kreuzberg.dev)** -- Full docs and API reference
- **[GitHub Repository](https://github.com/kreuzberg-dev/liter-llm)** -- Source, issues, and discussions
- **[Provider Registry](https://github.com/kreuzberg-dev/liter-llm/blob/main/schemas/providers.json)** -- 143 supported providers

Part of [kreuzberg.dev](https://kreuzberg.dev).

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](https://github.com/kreuzberg-dev/liter-llm/blob/main/CONTRIBUTING.md) for guidelines.

Join our [Discord community](https://discord.gg/xt9WY3GnKR) for questions and discussion.

## License

MIT -- see [LICENSE](https://github.com/kreuzberg-dev/liter-llm/blob/main/LICENSE) for details.

# liter-lm

<div align="center" style="display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; margin: 20px 0;">
  <!-- Language Bindings -->  <a href="https://crates.io/crates/liter-lm">
    <img src="https://img.shields.io/crates/v/liter-lm?label=Rust&color=007ec6" alt="Rust">
  </a>  <a href="https://pypi.org/project/liter-lm/">
    <img src="https://img.shields.io/pypi/v/liter-lm?label=Python&color=007ec6" alt="Python">
  </a>  <a href="https://www.npmjs.com/package/@kreuzberg/liter-lm">
    <img src="https://img.shields.io/npm/v/@kreuzberg/liter-lm?label=Node.js&color=007ec6" alt="Node">
  </a>  <a href="https://www.npmjs.com/package/@kreuzberg/liter-lm-wasm">
    <img src="https://img.shields.io/npm/v/@kreuzberg/liter-lm-wasm?label=WASM&color=007ec6" alt="Wasm">
  </a>  <a href="https://central.sonatype.com/artifact/dev.kreuzberg/liter-lm">
    <img src="https://img.shields.io/maven-central/v/dev.kreuzberg/liter-lm?label=Java&color=007ec6" alt="Java">
  </a>  <a href="https://github.com/kreuzberg-dev/liter-lm/tree/main/packages/go/v1">
    <img src="https://img.shields.io/github/v/tag/kreuzberg-dev/liter-lm?label=Go&color=007ec6" alt="Go">
  </a>  <a href="https://www.nuget.org/packages/LiterLm">
    <img src="https://img.shields.io/nuget/v/LiterLm?label=C%23&color=007ec6" alt="Csharp">
  </a>  <a href="https://packagist.org/packages/kreuzberg/liter-lm">
    <img src="https://img.shields.io/packagist/v/kreuzberg/liter-lm?label=PHP&color=007ec6" alt="Php">
  </a>  <a href="https://rubygems.org/gems/liter_lm">
    <img src="https://img.shields.io/gem/v/liter_lm?label=Ruby&color=007ec6" alt="Ruby">
  </a>  <a href="https://hex.pm/packages/liter_lm">
    <img src="https://img.shields.io/hexpm/v/liter_lm?label=Elixir&color=007ec6" alt="Elixir">
  </a>  <a href="https://github.com/kreuzberg-dev/liter-lm/pkgs/container/liter-lm">
    <img src="https://img.shields.io/badge/Docker-007ec6?logo=docker&logoColor=white" alt="Docker">
  </a>  <a href="https://github.com/kreuzberg-dev/liter-lm/tree/main/crates/liter-lm-ffi">
    <img src="https://img.shields.io/badge/C-FFI-007ec6" alt="Ffi">
  </a>
  <!-- Project Info -->
  <a href="https://github.com/kreuzberg-dev/liter-lm/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/kreuzberg-dev/liter-lm/ci-rust.yaml?branch=main&label=CI&color=007ec6" alt="CI">
  </a>
  <a href="https://github.com/kreuzberg-dev/liter-lm/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-007ec6.svg" alt="License">
  </a>
  <a href="https://docs.liter-lm.kreuzberg.dev">
    <img src="https://img.shields.io/badge/docs-kreuzberg.dev-007ec6" alt="Docs">
  </a>
</div>

<div align="center">
  <a href="https://discord.gg/xt9WY3GnKR">
    <img height="22" src="https://img.shields.io/badge/Discord-Join%20our%20community-7289da?logo=discord&logoColor=white" alt="Discord">
  </a>
</div>

Universal LLM API client with polyglot bindings for 100+ providers

## Overview

**liter-lm** provides unified access to 100+ LLM providers with a single API and native bindings for multiple programming languages. Ship LLM-powered features in your application without managing provider-specific SDKs.

## Architecture

```text
liter-lm/
├── crates/
│   ├── liter-lm/           # Rust core library
│   ├── liter-lm-py/        # Python (maturin/PyO3) bindings
│   ├── liter-lm-node/      # Node.js (NAPI-RS) bindings
│   ├── liter-lm-java/      # Java (Panama FFI) bindings
│   ├── liter-lm-ffi/       # C-compatible FFI library
│   ├── liter-lm-php/       # PHP (ext-php-rs) bindings
│   └── liter-lm-wasm/      # WebAssembly (wasm-bindgen) bindings
├── packages/
│   ├── go/                 # Go (cgo) bindings
│   ├── ruby/               # Ruby (Magnus) gem
│   ├── elixir/             # Elixir (Rustler NIF) package
│   ├── csharp/             # .NET (P/Invoke) package
│   └── java/               # Java wrapper
└── schemas/                # Provider registry and API schemas
```

## Quick Start

### Rust

```sh
cargo add liter-lm
```

See [Rust README](crates/liter-lm/README.md) for full documentation.

### Python

```sh
pip install liter-lm
```

```sh
uv add liter-lm
```

See [Python README](crates/liter-lm-py/README.md) for full documentation.

### Node.js

```sh
npm install @kreuzberg/liter-lm
```

```sh
pnpm add @kreuzberg/liter-lm
```

```sh
yarn add @kreuzberg/liter-lm
```

See [Node.js README](crates/liter-lm-node/README.md) for full documentation.

### Go

```sh
go get github.com/kreuzberg-dev/liter-lm/packages/go/v1
```

See [Go README](packages/go/v1/README.md) for full documentation.

### Java

```xml
<dependency>
  <groupId>dev.kreuzberg</groupId>
  <artifactId>liter-lm</artifactId>
  <version>0.1.0</version>
</dependency>
```

```groovy
implementation("dev.kreuzberg:liter-lm:0.1.0")
```

See [Java README](crates/liter-lm-java/README.md) for full documentation.

### Elixir

```elixir
{:liter_lm, "~> 0.1"}
```

See [Elixir README](packages/elixir/README.md) for full documentation.

### Ruby

```sh
gem install liter_lm
```

See [Ruby README](packages/ruby/README.md) for full documentation.

### WebAssembly

```sh
npm install @kreuzberg/liter-lm-wasm
```

```sh
pnpm add @kreuzberg/liter-lm-wasm
```

See [WebAssembly README](crates/liter-lm-wasm/README.md) for full documentation.

### PHP

```sh
composer require kreuzberg/liter-lm
```

See [PHP README](crates/liter-lm-php/README.md) for full documentation.

### .NET (C#)

```sh
dotnet add package LiterLm
```

See [.NET (C#) README](packages/csharp/LiterLm/README.md) for full documentation.

### C/C++ (FFI)

Build from source as part of this workspace.

See [C/C++ (FFI) README](crates/liter-lm-ffi/README.md) for full documentation.

## Core API

All bindings expose a unified `chat()` function for sending requests to LLM providers:

| Language | Function |
| -------- | -------- |
| Rust | `liter_lm::DefaultClient::new(config).chat(messages, options)` |
| Python | `LiterLmClient(api_key=...).chat(messages, config)` |
| Node.js | `new LiterLmClient({ apiKey }).chat(messages, config)` |
| Go | `client.Chat(ctx, messages, config)` |
| Java | `client.chat(messages, configJson)` |
| Ruby | `LiterLm::Client.new(api_key: ...).chat(messages, config)` |
| Elixir | `LiterLm.chat(messages, config)` |
| WASM | `new LiterLmClient({ apiKey }).chat(messages, config)` |
| C FFI | `liter_lm_chat(client, messages_json, config_json)` |

The `chat()` function returns a structured response including the message content, usage statistics, and provider metadata.

## Features

| Feature | Description |
| --- | --- |
| **100+ Providers** | Unified access to OpenAI, Anthropic, Groq, Mistral, Cohere, and 100+ more |
| **Streaming** | First-class streaming support via `chat_stream()` across all bindings |
| **Embeddings** | Embedding generation with `embeddings()` for supported providers |
| **Tool Calling** | Structured tool/function calling support |
| **Polyglot Bindings** | Native bindings for Rust, Python, Node.js, Go, Java, Elixir, Ruby, PHP, C# |
| **Schema-Driven** | Provider registry compiled from `schemas/providers.json` — no runtime lookups |

## Supported Providers

100+ providers including OpenAI, Anthropic, Groq, Mistral, Cohere, Together AI, Fireworks, Perplexity, DeepSeek, Google Gemini, and many more. See `schemas/providers.json` for the full list.

## Package READMEs

- [Rust](crates/liter-lm/README.md) -- Rust core library with unified LLM client.
- [Python](crates/liter-lm-py/README.md) -- Python bindings for liter-lm via PyO3.
- [Node.js](crates/liter-lm-node/README.md) -- Node.js NAPI bindings for liter-lm.
- [Go](packages/go/v1/README.md) -- Go bindings for liter-lm via cgo.
- [Java](crates/liter-lm-java/README.md) -- Java bindings for liter-lm via Panama FFM (JDK 21+).
- [Elixir](packages/elixir/README.md) -- Elixir bindings for liter-lm via Rustler NIF.
- [Ruby](packages/ruby/README.md) -- Ruby bindings for liter-lm via Magnus.
- [WebAssembly](crates/liter-lm-wasm/README.md) -- WebAssembly bindings for liter-lm.
- [PHP](crates/liter-lm-php/README.md) -- PHP extension for liter-lm via ext-php-rs.
- [.NET (C#)](packages/csharp/LiterLm/README.md) -- .NET P/Invoke bindings for liter-lm.
- [C/C++ (FFI)](crates/liter-lm-ffi/README.md) -- C-compatible FFI bindings for liter-lm.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Join our [Discord community](https://discord.gg/xt9WY3GnKR) for questions and discussion.

## License

MIT -- see [LICENSE](LICENSE) for details.

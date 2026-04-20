---
priority: high
---

# Crate Structure

Version source of truth: root `Cargo.toml` `[workspace.package] version`.

## Rust Crates (`crates/`)

- `liter-llm` — core library: client, providers, types, HTTP, streaming, Tower layers
- `liter-llm-cli` — CLI: `liter-llm api` (proxy), `liter-llm mcp` (MCP server)
- `liter-llm-proxy` — OpenAI-compatible proxy server (axum + Tower middleware)
- `liter-llm-ffi` — C FFI layer with cbindgen headers
- `liter-llm-py` — PyO3 Python bindings
- `liter-llm-node` — NAPI-RS Node.js bindings
- `liter-llm-php` — ext-php-rs PHP bindings
- `liter-llm-wasm` — wasm-bindgen WebAssembly bindings

## Language Packages (`packages/`)

- `python/` (PyPI, maturin), `typescript/` (npm, NAPI-RS), `wasm/` (npm, wasm-pack)
- `ruby/` (RubyGems, Magnus — excluded from workspace, compiled by `rake`)
- `go/` (Go module, cgo), `java/` (Maven, Panama FFM), `csharp/` (NuGet, P/Invoke)
- `php/` (Composer, ext-php-rs), `elixir/` (Hex, Rustler NIF)

## Key Schemas (`schemas/`)

- `providers.json` — 142+ provider registry, embedded at compile time
- `pricing.json` — per-token pricing (USD) for cost tracking
- `api/` — JSON schemas for API types; run `task generate:types` to regenerate Rust types

## Tools

- `tools/snippet-runner` — runs code snippets from fixtures for validation

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Core Rust crate with `LlmClient` trait, `DefaultClient`, SSE streaming, retry with backoff
- 142 LLM providers embedded at compile time from `schemas/providers.json`
- Strong OpenAI-compatible types with `FinishReason` enum, `SecretString`, `memchr` SSE parsing
- Provider auto-detection from model name prefix
- Python binding (PyO3) with async/streaming, TypedDict+Unpack typed kwargs, full .pyi stubs
- TypeScript binding (NAPI-RS) with full interface types
- C FFI binding with cbindgen header, callback-based streaming
- Go binding with typed structs, pure HTTP client, 22 tests
- Java binding with records and sealed interfaces, 13 tests
- C# binding with records and async HttpClient, 18 tests
- Elixir binding with @type/@spec, Req client, 31 tests
- Ruby binding (Magnus) with RBS type signatures and steep validation
- PHP binding (ext-php-rs) with PHPDoc stubs
- WASM binding (wasm-bindgen) with TypeScript custom section types
- E2E test generator for all 11 languages from JSON fixtures
- 11 fixtures across smoke, streaming, error-handling, tool-calling, types

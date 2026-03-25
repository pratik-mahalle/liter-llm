pub mod client;
pub mod error;
pub mod http;
pub mod provider;
#[cfg(test)]
mod tests;
pub mod types;

// Re-export key types at crate root.
pub use client::{BoxFuture, BoxStream, ClientConfig, ClientConfigBuilder, LlmClient};
// DefaultClient requires the native HTTP stack (reqwest + tokio).
#[cfg(feature = "native-http")]
pub use client::DefaultClient;
pub use error::{LiterLmError, Result};
pub use types::*;

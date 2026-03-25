// retry logic is pure (no reqwest/tokio) and is used by the streaming module
// even in WASM builds, so it is always compiled.
pub mod retry;

// request and streaming use reqwest + tokio and are only available when the
// native-http feature is enabled.
#[cfg(feature = "native-http")]
pub mod request;
#[cfg(feature = "native-http")]
pub mod streaming;

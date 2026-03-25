use pyo3::prelude::*;

// ─── Shared helper ────────────────────────────────────────────────────────────

/// Convert a [`liter_lm::types::FinishReason`] to its canonical snake_case
/// string representation, reusing the serde rename metadata so this function
/// never diverges from the wire format.
fn finish_reason_str(r: &liter_lm::types::FinishReason) -> String {
    // serde_json serialises the enum using its `#[serde(rename_all = "snake_case")]`
    // annotations, so we get the canonical wire string without a manual match.
    serde_json::to_value(r)
        .ok()
        .and_then(|v| v.as_str().map(str::to_owned))
        .unwrap_or_else(|| "other".to_owned())
}

// ─── Usage ────────────────────────────────────────────────────────────────────

/// Token usage information for a request.
#[pyclass(frozen, skip_from_py_object, name = "Usage")]
#[derive(Clone)]
pub struct PyUsage {
    inner: liter_lm::types::Usage,
}

#[pymethods]
impl PyUsage {
    /// Tokens consumed by the prompt.
    #[getter]
    fn prompt_tokens(&self) -> u64 {
        self.inner.prompt_tokens
    }

    /// Tokens consumed by the completion.
    #[getter]
    fn completion_tokens(&self) -> u64 {
        self.inner.completion_tokens
    }

    /// Total tokens (prompt + completion).
    #[getter]
    fn total_tokens(&self) -> u64 {
        self.inner.total_tokens
    }

    fn __repr__(&self) -> String {
        format!(
            "Usage(prompt_tokens={}, completion_tokens={}, total_tokens={})",
            self.inner.prompt_tokens, self.inner.completion_tokens, self.inner.total_tokens
        )
    }

    fn __eq__(&self, other: &PyUsage) -> bool {
        self.inner.prompt_tokens == other.inner.prompt_tokens
            && self.inner.completion_tokens == other.inner.completion_tokens
            && self.inner.total_tokens == other.inner.total_tokens
    }
}

impl From<liter_lm::types::Usage> for PyUsage {
    fn from(inner: liter_lm::types::Usage) -> Self {
        Self { inner }
    }
}

// ─── Tool call ────────────────────────────────────────────────────────────────

/// Function information within a tool call.
#[pyclass(frozen, name = "FunctionCall")]
pub struct PyFunctionCall {
    inner: liter_lm::types::FunctionCall,
}

#[pymethods]
impl PyFunctionCall {
    #[getter]
    fn name(&self) -> &str {
        &self.inner.name
    }

    #[getter]
    fn arguments(&self) -> &str {
        &self.inner.arguments
    }

    fn __repr__(&self) -> String {
        format!(
            "FunctionCall(name={:?}, arguments={:?})",
            self.inner.name, self.inner.arguments
        )
    }

    fn __eq__(&self, other: &PyFunctionCall) -> bool {
        self.inner.name == other.inner.name && self.inner.arguments == other.inner.arguments
    }
}

/// A tool call made by the assistant.
#[pyclass(frozen, name = "ToolCall")]
pub struct PyToolCall {
    inner: liter_lm::types::ToolCall,
}

#[pymethods]
impl PyToolCall {
    #[getter]
    fn id(&self) -> &str {
        &self.inner.id
    }

    #[getter]
    fn function(&self) -> PyFunctionCall {
        PyFunctionCall {
            inner: self.inner.function.clone(),
        }
    }

    fn __repr__(&self) -> String {
        format!(
            "ToolCall(id={:?}, function={})",
            self.inner.id,
            PyFunctionCall {
                inner: self.inner.function.clone()
            }
            .__repr__()
        )
    }

    fn __eq__(&self, other: &PyToolCall) -> bool {
        self.inner.id == other.inner.id
            && self.inner.function.name == other.inner.function.name
            && self.inner.function.arguments == other.inner.function.arguments
    }
}

// ─── AssistantMessage ─────────────────────────────────────────────────────────

/// The message field of a chat completion choice.
#[pyclass(frozen, name = "AssistantMessage")]
pub struct PyAssistantMessage {
    inner: liter_lm::types::AssistantMessage,
}

#[pymethods]
impl PyAssistantMessage {
    /// Text content of the message, or `None` if the response has tool calls.
    #[getter]
    fn content(&self) -> Option<&str> {
        self.inner.content.as_deref()
    }

    /// The assistant's optional display name.
    #[getter]
    fn name(&self) -> Option<&str> {
        self.inner.name.as_deref()
    }

    /// Refusal message set by the model, or `None`.
    #[getter]
    fn refusal(&self) -> Option<&str> {
        self.inner.refusal.as_deref()
    }

    /// Tool calls made by the assistant, or `None`.
    #[getter]
    fn tool_calls(&self) -> Option<Vec<PyToolCall>> {
        self.inner
            .tool_calls
            .as_ref()
            .map(|calls| calls.iter().map(|tc| PyToolCall { inner: tc.clone() }).collect())
    }

    fn __repr__(&self) -> String {
        format!(
            "AssistantMessage(content={:?}, refusal={:?}, name={:?})",
            self.inner.content, self.inner.refusal, self.inner.name
        )
    }

    fn __eq__(&self, other: &PyAssistantMessage) -> bool {
        self.inner.content == other.inner.content
            && self.inner.name == other.inner.name
            && self.inner.refusal == other.inner.refusal
            && self.inner.tool_calls == other.inner.tool_calls
    }
}

impl From<liter_lm::types::AssistantMessage> for PyAssistantMessage {
    fn from(inner: liter_lm::types::AssistantMessage) -> Self {
        Self { inner }
    }
}

// ─── Choice ───────────────────────────────────────────────────────────────────

/// A single choice in a chat completion response.
#[pyclass(frozen, name = "Choice")]
pub struct PyChoice {
    inner: liter_lm::types::Choice,
}

#[pymethods]
impl PyChoice {
    /// The message returned by the model.
    #[getter]
    fn message(&self) -> PyAssistantMessage {
        PyAssistantMessage {
            inner: self.inner.message.clone(),
        }
    }

    /// Why the model stopped generating tokens.
    #[getter]
    fn finish_reason(&self) -> Option<String> {
        self.inner.finish_reason.as_ref().map(finish_reason_str)
    }

    /// Zero-based index of this choice.
    #[getter]
    fn index(&self) -> u32 {
        self.inner.index
    }

    fn __repr__(&self) -> String {
        format!(
            "Choice(index={}, finish_reason={:?})",
            self.inner.index,
            self.finish_reason()
        )
    }

    fn __eq__(&self, other: &PyChoice) -> bool {
        self.inner.index == other.inner.index
            && self.inner.finish_reason == other.inner.finish_reason
            && self.inner.message == other.inner.message
    }
}

impl From<liter_lm::types::Choice> for PyChoice {
    fn from(inner: liter_lm::types::Choice) -> Self {
        Self { inner }
    }
}

// ─── ChatCompletionResponse ────────────────────────────────────────────────────

/// Response from a chat completion request.
#[pyclass(frozen, name = "ChatCompletionResponse")]
pub struct PyChatCompletionResponse {
    inner: liter_lm::types::ChatCompletionResponse,
}

#[pymethods]
impl PyChatCompletionResponse {
    /// Unique identifier for this completion.
    #[getter]
    fn id(&self) -> &str {
        &self.inner.id
    }

    /// The model that generated this response.
    #[getter]
    fn model(&self) -> &str {
        &self.inner.model
    }

    /// Unix timestamp of when the response was created.
    #[getter]
    fn created(&self) -> u64 {
        self.inner.created
    }

    /// List of generated choices.
    #[getter]
    fn choices(&self) -> Vec<PyChoice> {
        self.inner
            .choices
            .iter()
            .map(|c| PyChoice { inner: c.clone() })
            .collect()
    }

    /// Token usage information.
    #[getter]
    fn usage(&self) -> Option<PyUsage> {
        self.inner.usage.clone().map(PyUsage::from)
    }

    /// System fingerprint for reproducibility, if provided by the backend.
    #[getter]
    fn system_fingerprint(&self) -> Option<&str> {
        self.inner.system_fingerprint.as_deref()
    }

    /// Service tier used for this request, if returned by the backend.
    #[getter]
    fn service_tier(&self) -> Option<&str> {
        self.inner.service_tier.as_deref()
    }

    fn __repr__(&self) -> String {
        format!(
            "ChatCompletionResponse(id={:?}, model={:?}, created={}, choices={})",
            self.inner.id,
            self.inner.model,
            self.inner.created,
            self.inner.choices.len()
        )
    }

    fn __eq__(&self, other: &PyChatCompletionResponse) -> bool {
        self.inner.id == other.inner.id
            && self.inner.model == other.inner.model
            && self.inner.created == other.inner.created
            && self.inner.choices == other.inner.choices
            && self.inner.usage == other.inner.usage
    }
}

impl From<liter_lm::types::ChatCompletionResponse> for PyChatCompletionResponse {
    fn from(inner: liter_lm::types::ChatCompletionResponse) -> Self {
        Self { inner }
    }
}

// ─── Streaming chunk types ────────────────────────────────────────────────────

/// The partial function call delta within a streaming tool call.
#[pyclass(frozen, name = "StreamFunctionCall")]
pub struct PyStreamFunctionCall {
    inner: liter_lm::types::StreamFunctionCall,
}

#[pymethods]
impl PyStreamFunctionCall {
    /// Partial function name, or `None` if not present in this chunk.
    #[getter]
    fn name(&self) -> Option<&str> {
        self.inner.name.as_deref()
    }

    /// Partial arguments JSON string delta, or `None` if not present in this chunk.
    #[getter]
    fn arguments(&self) -> Option<&str> {
        self.inner.arguments.as_deref()
    }

    fn __repr__(&self) -> String {
        format!(
            "StreamFunctionCall(name={:?}, arguments={:?})",
            self.inner.name, self.inner.arguments
        )
    }

    fn __eq__(&self, other: &PyStreamFunctionCall) -> bool {
        self.inner.name == other.inner.name && self.inner.arguments == other.inner.arguments
    }
}

/// A tool call delta within a streaming chunk.
#[pyclass(frozen, name = "StreamToolCall")]
pub struct PyStreamToolCall {
    inner: liter_lm::types::StreamToolCall,
}

#[pymethods]
impl PyStreamToolCall {
    /// The index of this tool call in the list.
    #[getter]
    fn index(&self) -> u32 {
        self.inner.index
    }

    /// The tool call ID, present only in the first chunk for this index.
    #[getter]
    fn id(&self) -> Option<&str> {
        self.inner.id.as_deref()
    }

    /// The partial function call delta, or `None` if not present in this chunk.
    #[getter]
    fn function(&self) -> Option<PyStreamFunctionCall> {
        self.inner.function.clone().map(|f| PyStreamFunctionCall { inner: f })
    }

    fn __repr__(&self) -> String {
        format!("StreamToolCall(index={}, id={:?})", self.inner.index, self.inner.id)
    }

    fn __eq__(&self, other: &PyStreamToolCall) -> bool {
        self.inner.index == other.inner.index
            && self.inner.id == other.inner.id
            && self.inner.function == other.inner.function
    }
}

/// The delta object in a streaming chunk choice.
#[pyclass(frozen, name = "StreamDelta")]
pub struct PyStreamDelta {
    inner: liter_lm::types::StreamDelta,
}

#[pymethods]
impl PyStreamDelta {
    /// Partial content token, or `None` if not present in this chunk.
    #[getter]
    fn content(&self) -> Option<&str> {
        self.inner.content.as_deref()
    }

    /// Role, present only in the first chunk.
    #[getter]
    fn role(&self) -> Option<&str> {
        self.inner.role.as_deref()
    }

    /// Refusal text delta, or `None`.
    #[getter]
    fn refusal(&self) -> Option<&str> {
        self.inner.refusal.as_deref()
    }

    /// Tool call deltas, or `None` if not present in this chunk.
    #[getter]
    fn tool_calls(&self) -> Option<Vec<PyStreamToolCall>> {
        self.inner
            .tool_calls
            .as_ref()
            .map(|calls| calls.iter().map(|tc| PyStreamToolCall { inner: tc.clone() }).collect())
    }

    fn __repr__(&self) -> String {
        format!(
            "StreamDelta(role={:?}, content={:?}, refusal={:?})",
            self.inner.role, self.inner.content, self.inner.refusal
        )
    }

    fn __eq__(&self, other: &PyStreamDelta) -> bool {
        self.inner.role == other.inner.role
            && self.inner.content == other.inner.content
            && self.inner.refusal == other.inner.refusal
            && self.inner.tool_calls == other.inner.tool_calls
    }
}

/// A single choice within a streaming chunk.
#[pyclass(frozen, name = "StreamChoice")]
pub struct PyStreamChoice {
    inner: liter_lm::types::StreamChoice,
}

#[pymethods]
impl PyStreamChoice {
    #[getter]
    fn delta(&self) -> PyStreamDelta {
        PyStreamDelta {
            inner: self.inner.delta.clone(),
        }
    }

    #[getter]
    fn finish_reason(&self) -> Option<String> {
        self.inner.finish_reason.as_ref().map(finish_reason_str)
    }

    #[getter]
    fn index(&self) -> u32 {
        self.inner.index
    }

    fn __repr__(&self) -> String {
        format!(
            "StreamChoice(index={}, finish_reason={:?})",
            self.inner.index,
            self.finish_reason()
        )
    }

    fn __eq__(&self, other: &PyStreamChoice) -> bool {
        self.inner.index == other.inner.index
            && self.inner.finish_reason == other.inner.finish_reason
            && self.inner.delta == other.inner.delta
    }
}

/// One SSE event from a streaming chat completion.
#[pyclass(frozen, name = "ChatCompletionChunk")]
pub struct PyChatCompletionChunk {
    inner: liter_lm::types::ChatCompletionChunk,
}

#[pymethods]
impl PyChatCompletionChunk {
    #[getter]
    fn id(&self) -> &str {
        &self.inner.id
    }

    #[getter]
    fn model(&self) -> &str {
        &self.inner.model
    }

    /// Unix timestamp of when this chunk was created.
    #[getter]
    fn created(&self) -> u64 {
        self.inner.created
    }

    #[getter]
    fn choices(&self) -> Vec<PyStreamChoice> {
        self.inner
            .choices
            .iter()
            .map(|c| PyStreamChoice { inner: c.clone() })
            .collect()
    }

    /// Usage statistics, present only in the final chunk when
    /// ``stream_options.include_usage`` was set in the request.
    ///
    /// Note: accessing this getter clones the `Usage` struct (three `u64`
    /// fields) — negligible cost, but documented here for completeness.
    #[getter]
    fn usage(&self) -> Option<PyUsage> {
        self.inner.usage.clone().map(PyUsage::from)
    }

    fn __repr__(&self) -> String {
        format!(
            "ChatCompletionChunk(id={:?}, model={:?}, created={}, choices={})",
            self.inner.id,
            self.inner.model,
            self.inner.created,
            self.inner.choices.len()
        )
    }

    fn __eq__(&self, other: &PyChatCompletionChunk) -> bool {
        self.inner.id == other.inner.id
            && self.inner.model == other.inner.model
            && self.inner.created == other.inner.created
            && self.inner.choices == other.inner.choices
    }
}

impl From<liter_lm::types::ChatCompletionChunk> for PyChatCompletionChunk {
    fn from(inner: liter_lm::types::ChatCompletionChunk) -> Self {
        Self { inner }
    }
}

// ─── Embedding types ──────────────────────────────────────────────────────────

/// A single embedding vector.
///
/// Note: the `.embedding` getter clones the entire `Vec<f64>` into a Python
/// list.  For large vectors consider slicing with indexing instead.
#[pyclass(frozen, name = "EmbeddingObject")]
pub struct PyEmbeddingObject {
    inner: liter_lm::types::EmbeddingObject,
}

#[pymethods]
impl PyEmbeddingObject {
    /// The embedding vector.
    ///
    /// This clones the full vector on each access.  Cache the result in Python
    /// when you need to access it multiple times.
    #[getter]
    fn embedding(&self) -> Vec<f64> {
        self.inner.embedding.clone()
    }

    /// Index of this embedding in the input list.
    #[getter]
    fn index(&self) -> u32 {
        self.inner.index
    }

    fn __repr__(&self) -> String {
        format!(
            "EmbeddingObject(index={}, dims={})",
            self.inner.index,
            self.inner.embedding.len()
        )
    }

    fn __eq__(&self, other: &PyEmbeddingObject) -> bool {
        self.inner.index == other.inner.index && self.inner.embedding == other.inner.embedding
    }
}

/// Response from an embedding request.
#[pyclass(frozen, name = "EmbeddingResponse")]
pub struct PyEmbeddingResponse {
    inner: liter_lm::types::EmbeddingResponse,
}

#[pymethods]
impl PyEmbeddingResponse {
    /// The model used for embedding.
    #[getter]
    fn model(&self) -> &str {
        &self.inner.model
    }

    /// List of embedding objects.
    #[getter]
    fn data(&self) -> Vec<PyEmbeddingObject> {
        self.inner
            .data
            .iter()
            .map(|o| PyEmbeddingObject { inner: o.clone() })
            .collect()
    }

    /// Token usage for this request.  Returns `None` when the provider does not
    /// include usage data in the response.
    #[getter]
    fn usage(&self) -> Option<PyUsage> {
        self.inner.usage.clone().map(|u| PyUsage { inner: u })
    }

    fn __repr__(&self) -> String {
        format!(
            "EmbeddingResponse(model={:?}, count={})",
            self.inner.model,
            self.inner.data.len()
        )
    }

    fn __eq__(&self, other: &PyEmbeddingResponse) -> bool {
        self.inner.model == other.inner.model
            && self.inner.data == other.inner.data
            && self.inner.usage == other.inner.usage
    }
}

impl From<liter_lm::types::EmbeddingResponse> for PyEmbeddingResponse {
    fn from(inner: liter_lm::types::EmbeddingResponse) -> Self {
        Self { inner }
    }
}

// ─── Models list types ────────────────────────────────────────────────────────

/// Information about a single model.
#[pyclass(frozen, name = "ModelObject")]
pub struct PyModelObject {
    inner: liter_lm::types::ModelObject,
}

#[pymethods]
impl PyModelObject {
    #[getter]
    fn id(&self) -> &str {
        &self.inner.id
    }

    #[getter]
    fn owned_by(&self) -> &str {
        &self.inner.owned_by
    }

    #[getter]
    fn created(&self) -> u64 {
        self.inner.created
    }

    fn __repr__(&self) -> String {
        format!(
            "ModelObject(id={:?}, owned_by={:?}, created={})",
            self.inner.id, self.inner.owned_by, self.inner.created
        )
    }

    fn __eq__(&self, other: &PyModelObject) -> bool {
        self.inner.id == other.inner.id
            && self.inner.owned_by == other.inner.owned_by
            && self.inner.created == other.inner.created
            && self.inner.object == other.inner.object
    }
}

/// Response from a list-models request.
#[pyclass(frozen, name = "ModelsListResponse")]
pub struct PyModelsListResponse {
    inner: liter_lm::types::ModelsListResponse,
}

#[pymethods]
impl PyModelsListResponse {
    /// List of available models.
    #[getter]
    fn data(&self) -> Vec<PyModelObject> {
        self.inner
            .data
            .iter()
            .map(|m| PyModelObject { inner: m.clone() })
            .collect()
    }

    fn __repr__(&self) -> String {
        format!("ModelsListResponse(count={})", self.inner.data.len())
    }

    fn __eq__(&self, other: &PyModelsListResponse) -> bool {
        self.inner.data == other.inner.data
    }
}

impl From<liter_lm::types::ModelsListResponse> for PyModelsListResponse {
    fn from(inner: liter_lm::types::ModelsListResponse) -> Self {
        Self { inner }
    }
}

/// Register all response types on the module.
pub fn register(m: &Bound<'_, PyModule>) -> PyResult<()> {
    m.add_class::<PyUsage>()?;
    m.add_class::<PyFunctionCall>()?;
    m.add_class::<PyToolCall>()?;
    m.add_class::<PyAssistantMessage>()?;
    m.add_class::<PyChoice>()?;
    m.add_class::<PyChatCompletionResponse>()?;
    m.add_class::<PyStreamFunctionCall>()?;
    m.add_class::<PyStreamToolCall>()?;
    m.add_class::<PyStreamDelta>()?;
    m.add_class::<PyStreamChoice>()?;
    m.add_class::<PyChatCompletionChunk>()?;
    m.add_class::<PyEmbeddingObject>()?;
    m.add_class::<PyEmbeddingResponse>()?;
    m.add_class::<PyModelObject>()?;
    m.add_class::<PyModelsListResponse>()?;
    Ok(())
}

// Package literlm provides a Go client for the liter-lm universal LLM API.
//
// It speaks the OpenAI-compatible HTTP API directly — no cgo, no CGO, no
// shared libraries required.  The same wire format is used for all
// 100+ providers supported by liter-lm; the model-name prefix
// (e.g. "groq/llama3-70b") selects the provider and endpoint.
package literlm

import "encoding/json"

// mustMarshal is a helper for json.Marshal calls that cannot fail
// (e.g. marshaling a string or a slice of known-good types).
func mustMarshal(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		panic("literlm: unexpected marshal error: " + err.Error())
	}
	return raw
}

// ─── Messages ─────────────────────────────────────────────────────────────────

// Role enumerates the valid values for Message.Role.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
	RoleDeveloper Role = "developer"
	// RoleFunction is the deprecated legacy function role, retained for API
	// compatibility with older OpenAI SDK responses.
	RoleFunction Role = "function"
)

// Message is a single turn in a chat conversation.
//
// The Content field is intentionally typed as json.RawMessage so that callers
// can supply either a plain string or an array of ContentPart objects without
// losing type safety.  Use [NewTextMessage], [NewToolResultMessage], and
// [NewAssistantMessage] for the common cases.
type Message struct {
	Role       Role            `json:"role"`
	Content    json.RawMessage `json:"content,omitempty"`
	Name       *string         `json:"name,omitempty"`
	ToolCallID *string         `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall      `json:"tool_calls,omitempty"`
	// Refusal is set by the assistant when it declines to respond.
	Refusal *string `json:"refusal,omitempty"`
	// FunctionCall is the deprecated legacy field; retained for compatibility.
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// NewTextMessage creates a simple text message with the given role and content.
func NewTextMessage(role Role, text string) Message {
	raw := mustMarshal(text)
	return Message{Role: role, Content: raw}
}

// NewPartsMessage creates a message whose content is a slice of ContentPart.
func NewPartsMessage(role Role, parts []ContentPart) Message {
	raw := mustMarshal(parts)
	return Message{Role: role, Content: raw}
}

// ContentPartType enumerates allowed content part types.
type ContentPartType string

const (
	ContentPartTypeText     ContentPartType = "text"
	ContentPartTypeImageURL ContentPartType = "image_url"
)

// ContentPart is one element in a multi-modal message content array.
type ContentPart struct {
	Type     ContentPartType `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *ImageURL       `json:"image_url,omitempty"`
}

// ImageURL specifies an image to include in a message.
type ImageURL struct {
	URL    string       `json:"url"`
	Detail *ImageDetail `json:"detail,omitempty"`
}

// ImageDetail controls the resolution at which the model processes the image.
type ImageDetail string

const (
	ImageDetailLow  ImageDetail = "low"
	ImageDetailHigh ImageDetail = "high"
	ImageDetailAuto ImageDetail = "auto"
)

// ─── Tools ────────────────────────────────────────────────────────────────────

// ToolType is the only valid value for tool and tool-call type fields.
// The OpenAI spec always uses "function"; other values are rejected.
type ToolType string

const ToolTypeFunction ToolType = "function"

// ChatCompletionTool describes a function the model may call.
type ChatCompletionTool struct {
	Type     ToolType           `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition holds the schema for a callable function.
type FunctionDefinition struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	// Parameters is a JSON Schema object.  Use json.RawMessage to avoid
	// double-encoding.
	Parameters json.RawMessage `json:"parameters,omitempty"`
	Strict     *bool           `json:"strict,omitempty"`
}

// ToolCall is a request from the model to invoke a tool.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     ToolType     `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall carries the name and JSON-encoded arguments for a tool call.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ─── Tool Choice ─────────────────────────────────────────────────────────────

// ToolChoice controls how (and whether) the model calls tools.
// Use the ToolChoiceAuto, ToolChoiceRequired, and ToolChoiceNone sentinel
// values, or construct a ToolChoice with a specific function via
// [NewSpecificToolChoice].
type ToolChoice struct {
	raw json.RawMessage
}

// MarshalJSON implements json.Marshaler.
func (tc ToolChoice) MarshalJSON() ([]byte, error) {
	if tc.raw == nil {
		return []byte(`"auto"`), nil
	}
	return tc.raw, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (tc *ToolChoice) UnmarshalJSON(data []byte) error {
	tc.raw = make(json.RawMessage, len(data))
	copy(tc.raw, data)
	return nil
}

var (
	// ToolChoiceAuto lets the model decide whether to call a tool.
	ToolChoiceAuto = ToolChoice{raw: json.RawMessage(`"auto"`)}
	// ToolChoiceRequired forces the model to call a tool.
	ToolChoiceRequired = ToolChoice{raw: json.RawMessage(`"required"`)}
	// ToolChoiceNone prevents the model from calling any tool.
	ToolChoiceNone = ToolChoice{raw: json.RawMessage(`"none"`)}
)

// specificToolChoice is the JSON shape for targeting a named function.
type specificToolChoice struct {
	Type     ToolType         `json:"type"`
	Function specificFunction `json:"function"`
}

type specificFunction struct {
	Name string `json:"name"`
}

// NewSpecificToolChoice returns a ToolChoice that forces the model to call the
// named function.
func NewSpecificToolChoice(functionName string) ToolChoice {
	raw := mustMarshal(specificToolChoice{
		Type:     ToolTypeFunction,
		Function: specificFunction{Name: functionName},
	})
	return ToolChoice{raw: raw}
}

// ─── Response Format ──────────────────────────────────────────────────────────

// ResponseFormatType enumerates the supported response format types.
type ResponseFormatType string

const (
	ResponseFormatTypeText       ResponseFormatType = "text"
	ResponseFormatTypeJSONObject ResponseFormatType = "json_object"
	ResponseFormatTypeJSONSchema ResponseFormatType = "json_schema"
)

// ResponseFormat instructs the model to produce output in a specific format.
type ResponseFormat struct {
	Type       ResponseFormatType `json:"type"`
	JSONSchema *JSONSchemaFormat  `json:"json_schema,omitempty"`
}

// JSONSchemaFormat is the schema descriptor used with ResponseFormatTypeJSONSchema.
type JSONSchemaFormat struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema"`
	Strict      *bool           `json:"strict,omitempty"`
}

// ─── Usage ────────────────────────────────────────────────────────────────────

// Usage reports token consumption for a request.
type Usage struct {
	PromptTokens     uint64 `json:"prompt_tokens"`
	CompletionTokens uint64 `json:"completion_tokens"`
	TotalTokens      uint64 `json:"total_tokens"`
}

// ─── Stop Sequence ────────────────────────────────────────────────────────────

// StopSequence is either a single stop string or a list of stop strings.
// Use [NewStopString] and [NewStopStrings] for construction.
type StopSequence struct {
	raw json.RawMessage
}

// MarshalJSON implements json.Marshaler.
func (s StopSequence) MarshalJSON() ([]byte, error) {
	return s.raw, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *StopSequence) UnmarshalJSON(data []byte) error {
	s.raw = make(json.RawMessage, len(data))
	copy(s.raw, data)
	return nil
}

// NewStopString creates a StopSequence from a single stop string.
func NewStopString(stop string) StopSequence {
	raw := mustMarshal(stop)
	return StopSequence{raw: raw}
}

// NewStopStrings creates a StopSequence from multiple stop strings.
func NewStopStrings(stops []string) StopSequence {
	raw := mustMarshal(stops)
	return StopSequence{raw: raw}
}

// ─── Chat Request ─────────────────────────────────────────────────────────────

// StreamOptions configures behavior for streaming requests.
type StreamOptions struct {
	// IncludeUsage, when true, requests that the final chunk include usage stats.
	IncludeUsage *bool `json:"include_usage,omitempty"`
}

// ChatCompletionRequest is the body for a chat completion API call.
//
// Only Model and Messages are required.  All other fields are optional and are
// omitted from the JSON payload when nil/zero.
type ChatCompletionRequest struct {
	Model             string               `json:"model"`
	Messages          []Message            `json:"messages"`
	Temperature       *float64             `json:"temperature,omitempty"`
	TopP              *float64             `json:"top_p,omitempty"`
	N                 *uint32              `json:"n,omitempty"`
	Stream            *bool                `json:"stream,omitempty"`
	Stop              *StopSequence        `json:"stop,omitempty"`
	MaxTokens         *uint64              `json:"max_tokens,omitempty"`
	PresencePenalty   *float64             `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64             `json:"frequency_penalty,omitempty"`
	LogitBias         map[string]float64   `json:"logit_bias,omitempty"`
	User              *string              `json:"user,omitempty"`
	Tools             []ChatCompletionTool `json:"tools,omitempty"`
	ToolChoice        *ToolChoice          `json:"tool_choice,omitempty"`
	ParallelToolCalls *bool                `json:"parallel_tool_calls,omitempty"`
	ResponseFormat    *ResponseFormat      `json:"response_format,omitempty"`
	StreamOptions     *StreamOptions       `json:"stream_options,omitempty"`
	Seed              *int64               `json:"seed,omitempty"`
}

// ─── Chat Response ────────────────────────────────────────────────────────────

// FinishReason indicates why a choice stopped generating tokens.
type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonToolCalls     FinishReason = "tool_calls"
	FinishReasonContentFilter FinishReason = "content_filter"
	// FinishReasonFunctionCall is the deprecated legacy value retained for
	// API compatibility.
	FinishReasonFunctionCall FinishReason = "function_call"
)

// AssistantMessage holds the content returned by the model in a non-streaming
// response.
type AssistantMessage struct {
	Content      *string       `json:"content,omitempty"`
	Name         *string       `json:"name,omitempty"`
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`
	Refusal      *string       `json:"refusal,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// Choice is one completion alternative in a ChatCompletionResponse.
type Choice struct {
	Index        uint32           `json:"index"`
	Message      AssistantMessage `json:"message"`
	FinishReason *FinishReason    `json:"finish_reason"`
}

// ChatCompletionResponse is the response body for a non-streaming chat
// completion request.
type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           uint64   `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             *Usage   `json:"usage,omitempty"`
	SystemFingerprint *string  `json:"system_fingerprint,omitempty"`
	ServiceTier       *string  `json:"service_tier,omitempty"`
}

// ─── Stream Chunk ─────────────────────────────────────────────────────────────

// StreamFunctionCall is the delta shape for a deprecated legacy function call.
type StreamFunctionCall struct {
	Name      *string `json:"name,omitempty"`
	Arguments *string `json:"arguments,omitempty"`
}

// StreamToolCall is an incremental update to a tool call within a chunk.
type StreamToolCall struct {
	Index    uint32              `json:"index"`
	ID       *string             `json:"id,omitempty"`
	Type     *ToolType           `json:"type,omitempty"`
	Function *StreamFunctionCall `json:"function,omitempty"`
}

// StreamDelta contains the incremental content for one stream choice.
type StreamDelta struct {
	Role         *string             `json:"role,omitempty"`
	Content      *string             `json:"content,omitempty"`
	ToolCalls    []StreamToolCall    `json:"tool_calls,omitempty"`
	FunctionCall *StreamFunctionCall `json:"function_call,omitempty"`
	Refusal      *string             `json:"refusal,omitempty"`
}

// StreamChoice is one choice entry within a ChatCompletionChunk.
type StreamChoice struct {
	Index        uint32        `json:"index"`
	Delta        StreamDelta   `json:"delta"`
	FinishReason *FinishReason `json:"finish_reason"`
}

// ChatCompletionChunk is a single server-sent event emitted during a streaming
// chat completion.
type ChatCompletionChunk struct {
	ID          string         `json:"id"`
	Object      string         `json:"object"`
	Created     uint64         `json:"created"`
	Model       string         `json:"model"`
	Choices     []StreamChoice `json:"choices"`
	Usage       *Usage         `json:"usage,omitempty"`
	ServiceTier *string        `json:"service_tier,omitempty"`
}

// ─── Embedding ────────────────────────────────────────────────────────────────

// EmbeddingInput is either a single string or a list of strings.
// Use [NewEmbeddingInputSingle] and [NewEmbeddingInputMultiple].
type EmbeddingInput struct {
	raw json.RawMessage
}

// MarshalJSON implements json.Marshaler.
func (e EmbeddingInput) MarshalJSON() ([]byte, error) {
	return e.raw, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *EmbeddingInput) UnmarshalJSON(data []byte) error {
	e.raw = make(json.RawMessage, len(data))
	copy(e.raw, data)
	return nil
}

// NewEmbeddingInputSingle wraps a single string as an EmbeddingInput.
func NewEmbeddingInputSingle(text string) EmbeddingInput {
	raw := mustMarshal(text)
	return EmbeddingInput{raw: raw}
}

// NewEmbeddingInputMultiple wraps multiple strings as an EmbeddingInput.
func NewEmbeddingInputMultiple(texts []string) EmbeddingInput {
	raw := mustMarshal(texts)
	return EmbeddingInput{raw: raw}
}

// EmbeddingRequest is the body for an embedding API call.
type EmbeddingRequest struct {
	Model          string         `json:"model"`
	Input          EmbeddingInput `json:"input"`
	EncodingFormat *string        `json:"encoding_format,omitempty"`
	Dimensions     *uint32        `json:"dimensions,omitempty"`
	User           *string        `json:"user,omitempty"`
}

// EmbeddingObject holds a single embedding vector.
type EmbeddingObject struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     uint32    `json:"index"`
}

// EmbeddingResponse is the response body for an embedding request.
type EmbeddingResponse struct {
	Object string            `json:"object"`
	Data   []EmbeddingObject `json:"data"`
	Model  string            `json:"model"`
	Usage  Usage             `json:"usage"`
}

// ─── Models ───────────────────────────────────────────────────────────────────

// ModelObject describes a single model entry returned by the list-models API.
type ModelObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created uint64 `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelsListResponse is the response body for the list-models API.
type ModelsListResponse struct {
	Object string        `json:"object"`
	Data   []ModelObject `json:"data"`
}

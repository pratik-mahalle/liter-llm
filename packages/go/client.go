package literlm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL         = "https://api.openai.com/v1"
	defaultTimeout         = 120 * time.Second
	headerAuthorization    = "Authorization"
	headerContentType      = "Content-Type"
	headerAccept           = "Accept"
	contentTypeJSON        = "application/json"
	contentTypeEventStream = "text/event-stream"
)

// ─── Interface ────────────────────────────────────────────────────────────────

// LlmClient is the contract that all liter-lm client implementations satisfy.
type LlmClient interface {
	// Chat sends a non-streaming chat completion request.
	Chat(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)

	// ChatStream sends a streaming chat completion request and invokes the
	// supplied handler for each received chunk.  The stream is fully consumed
	// before this method returns; cancel ctx to abort early.
	ChatStream(ctx context.Context, req *ChatCompletionRequest, handler func(*ChatCompletionChunk) error) error

	// Embed sends an embedding request.
	Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// ListModels returns the list of models available via the configured
	// provider endpoint.
	ListModels(ctx context.Context) (*ModelsListResponse, error)
}

// ─── Config ───────────────────────────────────────────────────────────────────

// ClientConfig holds all options for constructing a [Client].
// Use [NewConfig] or individual With* option functions.
type ClientConfig struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option is a functional option for [NewClient].
type Option func(*ClientConfig)

// WithAPIKey sets the API key sent as a Bearer token on every request.
func WithAPIKey(key string) Option {
	return func(c *ClientConfig) {
		c.apiKey = key
	}
}

// WithBaseURL overrides the base URL used for all requests.
// The URL must not have a trailing slash.
//
// Example: "https://api.groq.com/openai/v1"
func WithBaseURL(url string) Option {
	return func(c *ClientConfig) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithHTTPClient replaces the default [http.Client].  Use this to configure
// custom TLS, proxies, or transport behavior.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *ClientConfig) {
		c.httpClient = hc
	}
}

// WithTimeout sets the timeout on the default HTTP client.  This option is
// ignored when [WithHTTPClient] is also provided.
func WithTimeout(d time.Duration) Option {
	return func(c *ClientConfig) {
		if c.httpClient != nil {
			c.httpClient.Timeout = d
		}
	}
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client is the default implementation of [LlmClient].  It calls the
// OpenAI-compatible HTTP API directly; no CGO or shared library is required.
//
// Construct one with [NewClient].  Client is safe for concurrent use.
type Client struct {
	config ClientConfig
}

// NewClient constructs a Client with the supplied options.
//
// At minimum, provide [WithAPIKey] (or set the OPENAI_API_KEY environment
// variable yourself and pass an empty key if the provider does not need it).
//
//	client := literlm.NewClient(
//	    literlm.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
//	)
func NewClient(opts ...Option) *Client {
	cfg := ClientConfig{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Client{config: cfg}
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────────

// buildRequest creates an HTTP request with the common headers set.
func (c *Client) buildRequest(ctx context.Context, method, path string, body io.Reader, stream bool) (*http.Request, error) {
	url := c.config.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("literlm: build request: %w", err)
	}
	if c.config.apiKey != "" {
		req.Header.Set(headerAuthorization, "Bearer "+c.config.apiKey)
	}
	if body != nil {
		req.Header.Set(headerContentType, contentTypeJSON)
	}
	if stream {
		req.Header.Set(headerAccept, contentTypeEventStream)
	} else {
		req.Header.Set(headerAccept, contentTypeJSON)
	}
	return req, nil
}

// do executes an HTTP request and returns the response body, or an *APIError
// for non-2xx responses.  The caller is responsible for closing the body.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.config.httpClient.Do(req) //nolint:gosec // URL is from trusted config, not user input
	if err != nil {
		return nil, fmt.Errorf("literlm: HTTP request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		msg := extractErrorMessage(resp)
		return nil, newAPIError(resp.StatusCode, msg)
	}
	return resp, nil
}

// extractErrorMessage reads a JSON error body and returns the message string.
// Falls back to the HTTP status text on any parse failure.
func extractErrorMessage(resp *http.Response) string {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil || len(body) == 0 {
		return http.StatusText(resp.StatusCode)
	}

	// Try OpenAI-style {"error": {"message": "..."}}
	var envelope struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil && envelope.Error.Message != "" {
		return envelope.Error.Message
	}

	// Try flat {"message": "..."}
	var flat struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &flat) == nil && flat.Message != "" {
		return flat.Message
	}

	return string(body)
}

// marshalBody JSON-encodes v and returns an io.Reader.
func marshalBody(v any) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("literlm: marshal request body: %w", err)
	}
	return bytes.NewReader(data), nil
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

// Chat sends a non-streaming chat completion request and returns the full
// response.
//
// The req.Stream field is forced to false; use [Client.ChatStream] for
// streaming.
func (c *Client) Chat(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request must not be nil", ErrInvalidRequest)
	}
	if req.Model == "" {
		return nil, fmt.Errorf("%w: model is required", ErrInvalidRequest)
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: messages must not be empty", ErrInvalidRequest)
	}

	// Ensure stream is off for this path.
	streamFalse := false
	copy := *req
	copy.Stream = &streamFalse

	body, err := marshalBody(&copy)
	if err != nil {
		return nil, err
	}

	httpReq, err := c.buildRequest(ctx, http.MethodPost, "/chat/completions", body, false)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("literlm: decode chat response: %w", err)
	}
	return &result, nil
}

// ChatStream sends a streaming chat completion request.
//
// The handler is invoked once for each server-sent event chunk.  If handler
// returns a non-nil error the stream is aborted and that error is returned by
// ChatStream.  Canceling ctx also aborts the stream.
//
// The req.Stream field is forced to true.
func (c *Client) ChatStream(ctx context.Context, req *ChatCompletionRequest, handler func(*ChatCompletionChunk) error) error {
	if req == nil {
		return fmt.Errorf("%w: request must not be nil", ErrInvalidRequest)
	}
	if req.Model == "" {
		return fmt.Errorf("%w: model is required", ErrInvalidRequest)
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("%w: messages must not be empty", ErrInvalidRequest)
	}
	if handler == nil {
		return fmt.Errorf("%w: handler must not be nil", ErrInvalidRequest)
	}

	streamTrue := true
	copy := *req
	copy.Stream = &streamTrue

	body, err := marshalBody(&copy)
	if err != nil {
		return err
	}

	httpReq, err := c.buildRequest(ctx, http.MethodPost, "/chat/completions", body, true)
	if err != nil {
		return err
	}

	resp, err := c.do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return parseSSEStream(resp.Body, handler)
}

// parseSSEStream reads an SSE response body, parses each data line as a
// ChatCompletionChunk, and invokes handler for each chunk.
func parseSSEStream(body io.Reader, handler func(*ChatCompletionChunk) error) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE lines that do not start with "data:" are comments or blank —
		// skip them.
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

		// "[DONE]" signals the end of the stream.
		if payload == "[DONE]" {
			break
		}

		var chunk ChatCompletionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return newStreamError("failed to parse chunk JSON", err)
		}

		if err := handler(&chunk); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return newStreamError("error reading stream", err)
	}
	return nil
}

// ─── Embed ────────────────────────────────────────────────────────────────────

// Embed sends an embedding request and returns the response.
func (c *Client) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request must not be nil", ErrInvalidRequest)
	}
	if req.Model == "" {
		return nil, fmt.Errorf("%w: model is required", ErrInvalidRequest)
	}

	body, err := marshalBody(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := c.buildRequest(ctx, http.MethodPost, "/embeddings", body, false)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("literlm: decode embedding response: %w", err)
	}
	return &result, nil
}

// ─── List Models ──────────────────────────────────────────────────────────────

// ListModels retrieves the list of models from the configured provider endpoint.
func (c *Client) ListModels(ctx context.Context) (*ModelsListResponse, error) {
	httpReq, err := c.buildRequest(ctx, http.MethodGet, "/models", nil, false)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ModelsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("literlm: decode models response: %w", err)
	}
	return &result, nil
}

// compile-time assertion: *Client must implement LlmClient.
var _ LlmClient = (*Client)(nil)

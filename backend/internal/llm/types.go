package llm

// Role constants for chat messages.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message is a single chat message in the OpenAI-compatible schema.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat requests a specific output format from the model.
// Set Type to "json_object" to enable JSON mode.
type ResponseFormat struct {
	Type string `json:"type"`
}

// ChatRequest is the request body for POST /v1/chat/completions.
type ChatRequest struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	Temperature    float64         `json:"temperature"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// Choice is a single completion choice returned by the model.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage reports token accounting for a completion.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse is the response body from POST /v1/chat/completions.
type ChatResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

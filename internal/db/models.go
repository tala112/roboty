package db

import (
	"encoding/json"
)

type Chat struct {
	ID               string    `json:"id" db:"id"`
	ParentChatID    *string   `json:"parent_chat_id,omitempty" db:"parent_chat_id"`
	Title           string    `json:"title" db:"title"`
	MessageCount   int       `json:"message_count" db:"message_count"`
	PromptTokens    int       `json:"prompt_tokens" db:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens" db:"completion_tokens"`
	Cost            float64   `json:"cost" db:"cost"`
	SummaryMessageID *string `json:"summary_message_id,omitempty" db:"summary_message_id"`
	Metadata        string    `json:"metadata" db:"metadata"`
	CreatedAt       DateTime  `json:"created_at" db:"created_at"`
	UpdatedAt       DateTime  `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID          string    `json:"id" db:"id"`
	ChatID     string    `json:"chat_id" db:"chat_id"`
	Role       Role      `json:"role" db:"role"`
	Content    string    `json:"content" db:"content"`
	Model      *string  `json:"model,omitempty" db:"model"`
	Provider   *string  `json:"provider,omitempty" db:"provider"`
	IsSummary  bool      `json:"is_summary" db:"is_summary"`
	CreatedAt  DateTime  `json:"created_at" db:"created_at"`
	UpdatedAt  DateTime  `json:"updated_at" db:"updated_at"`
	FinishedAt *DateTime `json:"finished_at,omitempty" db:"finished_at"`
}

type MessageContent struct {
	Type    string    `json:"type,omitempty"`
	Text   string    `json:"text,omitempty"`
	Parts  []Part    `json:"parts,omitempty"`
	ToolUse *ToolUse `json:"tool_use,omitempty"`
	ToolCall *ToolCall `json:"tool_call,omitempty"`
	Name   string    `json:"name,omitempty"`
	Content string   `json:"content,omitempty"`
}

type Part struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type ToolUse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Input map[string]interface{} `json:"input"`
}

type ToolCall struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Arguments string `json:"arguments"`
}

type File struct {
	ID        string   `json:"id" db:"id"`
	ChatID    string   `json:"chat_id" db:"chat_id"`
	Path      string   `json:"path" db:"path"`
	Version   int      `json:"version" db:"version"`
	CreatedAt DateTime  `json:"created_at" db:"created_at"`
	UpdatedAt DateTime  `json:"updated_at" db:"updated_at"`
}

type ReadFile struct {
	ChatID string   `json:"chat_id" db:"chat_id"`
	Path  string   `json:"path" db:"path"`
	ReadAt DateTime `json:"read_at" db:"read_at"`
}

type ChatWithMessages struct {
	Chat     Chat
	Messages []Message
}

func (m *Message) GetContent() (*MessageContent, error) {
	if m.Content == "" {
		return nil, nil
	}
	var content MessageContent
	if err := json.Unmarshal([]byte(m.Content), &content); err != nil {
		return nil, err
	}
	return &content, nil
}

func (m *Message) SetContent(content *MessageContent) error {
	if content == nil {
		m.Content = ""
		return nil
	}
	b, err := json.Marshal(content)
	if err != nil {
		return err
	}
	m.Content = string(b)
	return nil
}

func (c *Chat) GetMetadata() (map[string]interface{}, error) {
	if c.Metadata == "" || c.Metadata == "{}" {
		return make(map[string]interface{}), nil
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(c.Metadata), &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func (c *Chat) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		c.Metadata = "{}"
		return nil
	}
	b, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	c.Metadata = string(b)
	return nil
}
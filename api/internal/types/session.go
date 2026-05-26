package types

import openai "github.com/sashabaranov/go-openai"

// Session 会话结构体
type ChatSession struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

// 会话存储接口
type SessionStore interface {
	GetSession(chatID string) (*ChatSession, error)
	SaveSession(chatID string, session *ChatSession) error
}

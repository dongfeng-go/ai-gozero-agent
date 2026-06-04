package types

import openai "github.com/sashabaranov/go-openai"

// Session 会话结构体
type ChatSession struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

/*// 会话存储接口
type SessionStore interface {
	GetSession(chatID string) (*ChatSession, error)
	SaveSession(chatID string, session *ChatSession) error
}*/

// 新增向量存储消息结构
type VectorMessage struct {
	Role    string `json:"role"`    //消息角色
	Content string `json:"content"` //消息内容
}

// 会话存储接口
type SessionStore interface {
	GetSession(chatID string) ([]openai.ChatCompletionMessage, error)
	SaveMessage(chatID, role, content string) error
}

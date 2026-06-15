package types

import (
	"ai-gozero-agent/api/internal/config"

	openai "github.com/sashabaranov/go-openai"
)

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

// 新增知识块结构
type KnowledgeChunk struct {
	ID      int64  `json:"id"`      //知识块ID
	Title   string `json:"title"`   //知识块标题
	Content string `json:"content"` //知识块内容
}

// 会话存储接口
type SessionStore interface {
	GetMessages(chatID string, limit int) ([]VectorMessage, error)
	SaveMessage(chatID, role, content string) error
	SaveKnowledge(title, content string, cfg config.VectorDBConfig) error //保存知识库
	RetrieveKnowledge(query string, topK int) ([]KnowledgeChunk, error)   //检索知识库
}

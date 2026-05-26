package svc

import (
	"ai-gozero-agent/api/internal/types"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

// 内存会话存储实现
type MemorySessionStore struct {
	sessions     map[string]*types.ChatSession
	lastAccessed map[string]time.Time //最后访问时间
	lock         sync.RWMutex
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions:     make(map[string]*types.ChatSession),
		lastAccessed: make(map[string]time.Time),
	}
}
func (m *MemorySessionStore) GetSession(chatId string) (*types.ChatSession, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	session, exists := m.sessions[chatId]
	if !exists {
		//创建新会话

		return &types.ChatSession{
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "你是一个专业的GO语言面试官，负责评估候选人的GO语言能力。请提出有深度的问题并评估回答。",
				},
			},
		}, nil
	}

	//更新最后访问时间
	m.lastAccessed[chatId] = time.Now()
	return session, nil
}

func (m *MemorySessionStore) SaveSession(chatId string, session *types.ChatSession) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	//上下文截断(保留系统消息和最近5轮对话)
	if len(session.Messages) > 10 {
		newMessages := []openai.ChatCompletionMessage{session.Messages[0]}
		start := len(session.Messages) - 5
		if start < 1 {
			start = 1
		}
		newMessages = append(newMessages, session.Messages[start:]...)
		session.Messages = session.Messages[:5]
	}
	m.sessions[chatId] = session
	m.lastAccessed[chatId] = time.Now()
	return nil
}

// 清理过期会话
func (m *MemorySessionStore) CleanupExpiredSessions(maxAge time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()
	now := time.Now()
	for chatId, lastAccessed := range m.lastAccessed {
		if now.Sub(lastAccessed) > maxAge {
			delete(m.sessions, chatId)
			delete(m.lastAccessed, chatId)
		}
	}
}

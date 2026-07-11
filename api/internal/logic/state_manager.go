package logic

import (
	"ai-gozero-agent/api/internal/svc"
	"ai-gozero-agent/api/internal/types"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	stateKeyPrefix = "chat_state:"
	stateTTL       = 24 * time.Hour
)

type StateManager struct {
	svcCtx *svc.ServiceContext
}

func NewStateManager(svcCtx *svc.ServiceContext) *StateManager {
	return &StateManager{svcCtx: svcCtx}
}

// 获取当前状态(带初始化)
func (sm *StateManager) GetOrInitState(chatId string) (string, error) {
	key := stateKeyPrefix + chatId

	//尝试获取状态
	state, err := sm.svcCtx.Redis.Get(context.Background(), key).Result()
	if err == nil {
		return state, nil
	}
	//如果状态不存在或出错,初始化状态
	if err == redis.Nil {
		if err := sm.svcCtx.Redis.Set(
			context.Background(),
			key,
			types.StateStart,
			stateTTL).Err(); err != nil {
			return types.StateStart, fmt.Errorf("初始化状态失败：%w", err)
		}
		return types.StateStart, nil
	}
	return types.StateStart, fmt.Errorf("获取状态失败：%w", err)
}

// 强制设置状态
func (sm *StateManager) SetState(chatId, state string) error {
	key := stateKeyPrefix + chatId
	if err := sm.svcCtx.Redis.Set(
		context.Background(),
		key,
		state,
		stateTTL).Err(); err != nil {
		return fmt.Errorf("设置状态失败:%w", err)
	}
	return nil
}

// 评估并更新状态(更智能的规则)
func (sm *StateManager) EvaluateAndUpdateState(chatId, aiResponse string) (string, error) {
	currentState, err := sm.GetOrInitState(chatId)
	if err != nil {
		return currentState, err
	}
	newState := sm.datermineNewState(currentState, aiResponse)
	if newState != currentState {
		if err := sm.SetState(chatId, newState); err != nil {
			return currentState, err
		}
	}
	return newState, nil
}

// 状态转移决策逻辑
func (sm *StateManager) datermineNewState(currentState, aiResponse string) string {
	lowerResponse := strings.ToLower(aiResponse)
	switch currentState {
	case types.StateStart:
		if containsAny(lowerResponse, []string{"你好", "欢迎", "面试开始"}) {
			return types.StateQuestion
		}
	case types.StateQuestion:
		if containsAny(lowerResponse, []string{"追问", "详细说明", "为什么", "怎么实现"}) {
			return types.StateFollowUp
		}
		if containsAny(lowerResponse, []string{"评估", "总结", "表现", "优缺点"}) {
			return types.StateEvaluate
		}
	case types.StateFollowUp:
		if containsAny(lowerResponse, []string{"评估", "总结", "表现", "优缺点"}) {
			return types.StateEvaluate
		}
		if containsAny(lowerResponse, []string{"下一个问题", "新问题"}) {
			return types.StateQuestion
		}
	case types.StateEvaluate:
		if containsAny(lowerResponse, []string{"结束", "再见", "感谢参加"}) {
			return types.StateEnd
		}
		if containsAny(lowerResponse, []string{"继续", "下一个问题"}) {
			return types.StateQuestion
		}
	case types.StateEnd:
		// 结束状态，不进行任何转换
	}
	return currentState
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

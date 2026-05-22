// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"ai-gozero-agent/api/internal/svc"
	"ai-gozero-agent/api/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Go面试官聊天SSE流式接口
func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatLogic) Chat(req *types.InterviewAPPChatReq) (resp *types.ChatResponse, err error) {
	// todo: add your logic here and delete this line

	return
}

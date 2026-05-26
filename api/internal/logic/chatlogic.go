// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"ai-gozero-agent/api/internal/svc"
	"ai-gozero-agent/api/internal/types"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
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

func (l *ChatLogic) Chat(req *types.InterviewAPPChatReq) (<-chan *types.ChatResponse, error) {
	ch := make(chan *types.ChatResponse)
	go func() {
		defer close(ch)

		//1.获取或创建会话
		session, err := l.svcCtx.SessionStore.GetSession(req.ChatId)
		if err != nil {
			l.Logger.Error("获取会话失败:%v", err)
			return
		}
		//新增:添加用户消息到会话历史
		userMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Message,
		}
		session.Messages = append(session.Messages, userMessage)

		/*messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,

																																																												Content: "你是一个专业的GO语言面试官，负责评估候选人的GO语言能力。请提出有深度的问题并评估回答。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: req.Message,
			},
		}*/

		// 创建一个OpenAI客户端
		request := openai.ChatCompletionRequest{
			Model:       l.svcCtx.Config.OpenAI.Model,
			Messages:    session.Messages,
			Stream:      true,
			MaxTokens:   l.svcCtx.Config.OpenAI.MaxTokens,
			Temperature: l.svcCtx.Config.OpenAI.Temperature,
		}

		// 创建一个OpenAI流式请求
		stream, err := l.svcCtx.OpenAIClient.CreateChatCompletionStream(l.ctx, request)
		if err != nil {
			l.Logger.Error(err)
			return
		}
		defer stream.Close()

		//新增:收集完整响应内容
		var fullResponse strings.Builder

		for {
			select {
			case <-l.ctx.Done():
				return
			default:
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					//新增:流结束后保存会话
					assistantMessage := openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: fullResponse.String(),
					}
					session.Messages = append(session.Messages, assistantMessage)
					if err := l.svcCtx.SessionStore.SaveSession(req.ChatId, session); err != nil {
						l.Logger.Error("保存会话失败:%v", err)
					}

					//发送结束标记
					ch <- &types.ChatResponse{IsLast: true}
					return
				}
				if err != nil {
					l.Logger.Error(err)
					return
				}
				if len(response.Choices) > 0 {
					content := response.Choices[0].Delta.Content
					if content != "" {
						//新增:收集完整响应内容
						fullResponse.WriteString(content)
					}

					ch <- &types.ChatResponse{
						Content: content,
						IsLast:  false,
					}
				}
			}
		}
	}()
	return ch, nil
}

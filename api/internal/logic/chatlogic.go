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

		//1.保存用户消息到向量数据库
		if err := l.svcCtx.VectorStore.SaveMessage(req.ChatId, openai.ChatMessageRoleUser, req.Message); err != nil {
			l.Logger.Errorf("保存用户消息失败:%v", err)
			//只记录日志，不返回错误给前端
		}
		//2.获取会话历史消息
		messages, err := l.getSessionHistory(req.ChatId)
		if err != nil {
			l.Logger.Errorf("获取会话历史消息失败:%v", err)
			ch <- &types.ChatResponse{Content: "获取会话历史消息失败", IsLast: true}
			return
		}

		//3.创建OpenAI请求
		request := openai.ChatCompletionRequest{
			Model:       l.svcCtx.Config.OpenAI.Model,
			Messages:    messages,
			Stream:      true,
			MaxTokens:   l.svcCtx.Config.OpenAI.MaxTokens,
			Temperature: l.svcCtx.Config.OpenAI.Temperature,
		}

		//4.创建流式响应
		stream, err := l.svcCtx.OpenAIClient.CreateChatCompletionStream(l.ctx, request)
		if err != nil {
			l.Logger.Error("创建聊天完成流失败: ", err)
			ch <- &types.ChatResponse{Content: "系统错误：无法链接AI服务", IsLast: true}
			return
		}
		defer stream.Close()

		//5.处理流式响应
		var fullResponse strings.Builder
		for {
			select {
			case <-l.ctx.Done(): //接口取消请求
				return
			default:
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) { //流结束
					//保存助手回复
					if fullResponse.String() != "" {
						if saveErr := l.svcCtx.VectorStore.SaveMessage(
							req.ChatId,
							openai.ChatMessageRoleUser,
							fullResponse.String(),
						); saveErr != nil {
							l.Logger.Errorf("保存助手消息失败：%v", saveErr)
						}
					}
					ch <- &types.ChatResponse{IsLast: true}
					return
				}
				if err != nil {
					l.Logger.Error("接收流数据失败: ", err)
					//ch <- &types.ChatResponse{Content: "系统错误：无法链接AI服务", IsLast: true}
					return
				}
				//处理有效响应
				if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
					content := response.Choices[0].Delta.Content
					fullResponse.WriteString(content)

					ch <- &types.ChatResponse{
						Content: content,
						IsLast:  false,
					}
				}
			}
		}
		/*//新增:添加用户消息到会话历史
		userMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Message,
		}
		session.Messages = append(session.Messages, userMessage)

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
		}*/
	}()
	return ch, nil
}

// 获取会话历史消息
func (l *ChatLogic) getSessionHistory(chatId string) ([]openai.ChatCompletionMessage, error) {
	//获取最近10条历史消息(约5轮对话)
	vectorMessages, err := l.svcCtx.VectorStore.GetMessages(chatId, 10)
	if err != nil {
		return nil, err
	}
	//转换为OpenAI消息
	messages := make([]openai.ChatCompletionMessage, 0, len(vectorMessages)+1)

	//添加系统消息
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "你是一个专业的GO语言面试官，负责评估候选人的GO语言能力。请提出有深度的问题并评估回答。",
	})
	//添加历史消息
	for _, msg := range vectorMessages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return messages, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"ai-gozero-agent/api/internal/svc"
	"ai-gozero-agent/api/internal/types"
	"ai-gozero-agent/api/internal/utils"
	"context"
	"errors"
	"fmt"
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
		stateManager := NewStateManager(l.svcCtx)

		//1.保存用户消息到向量数据库
		if err := l.svcCtx.VectorStore.SaveMessage(req.ChatId, openai.ChatMessageRoleUser, req.Message); err != nil {
			l.Logger.Errorf("保存用户消息失败:%v", err)
			//只记录日志，不返回错误给前端
		}
		//2.获取当前状态,确保初始化
		currentState, err := stateManager.GetOrInitState(req.ChatId)
		if err != nil {
			l.Logger.Errorf("获取状态失败:%v", err)
			currentState = types.StateStart
		}

		//3.知识检索(RAG核心)
		knowledge, err := l.svcCtx.VectorStore.RetrieveKnowledge(req.Message, l.svcCtx.Config.VectorDB.Knowledge.TopK)
		if err != nil {
			l.Logger.Errorf("知识检索失败:%v", err)
			//ch <- &types.ChatResponse{Content: "知识检索失败", IsLast: true}
			knowledge = []types.KnowledgeChunk{} //默认为空,确保不为nil
		}

		//4.构建系统消息(带状态)
		messages, err := l.buildMessagesWithState(req.ChatId, currentState, knowledge)
		if err != nil {
			l.Logger.Errorf("构建消息失败:%v", err)
			ch <- &types.ChatResponse{Content: "系统错误：无法构建对话", IsLast: true}
			return
		}

		//2.获取会话历史消息
		//messages, err := l.getSessionHistory(req.ChatId)
		//messages, err := l.getSessionHistory(req.ChatId, knowledge)
		//if err != nil {
		//	l.Logger.Errorf("获取会话历史消息失败:%v", err)
		//	ch <- &types.ChatResponse{Content: "获取会话历史消息失败", IsLast: true}
		//	return
		//}

		//5.创建OpenAI请求
		request := openai.ChatCompletionRequest{
			Model:       l.svcCtx.Config.OpenAI.Model, //模型名称
			Messages:    messages,
			Stream:      true,
			MaxTokens:   l.svcCtx.Config.OpenAI.MaxTokens,
			Temperature: l.svcCtx.Config.OpenAI.Temperature, //随机性,温度参数,取值范围0-2,越高越随机
			//TopP:             l.svcCtx.Config.OpenAI.TopP,             //核心采样(0-1,越高越多样)
			//PresencePenalty:  l.svcCtx.Config.OpenAI.PresencePenalty,  //存在惩罚(-2.0,到2.0)
			//FrequencyPenalty: l.svcCtx.Config.OpenAI.FrequencyPenalty, //频率惩罚(-2.0,到2.0)
			//Seed:             l.svcCtx.Config.OpenAI.Seed,             //随机数种子(-1表示随机)
		}

		//6.创建流式响应
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
							openai.ChatMessageRoleAssistant,
							fullResponse.String(),
						); saveErr != nil {
							l.Logger.Errorf("保存助手消息失败：%v", saveErr)
						}
						//更新状态
						newState, err := stateManager.EvaluateAndUpdateState(req.ChatId, fullResponse.String())
						if err != nil {
							l.Logger.Errorf("评估并更新状态失败：%v", err)
						}
						l.Logger.Infof("评估并更新状态成功： %s -> %s", currentState, newState)
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
	}()
	return ch, nil
}

// 构建带状态的消息
func (l *ChatLogic) buildMessagesWithState(chatId, currentState string, knowledge []types.KnowledgeChunk) ([]openai.ChatCompletionMessage, error) {
	//构建状态特定的系统消息
	systemMessage := "你是一个专业的GO语言面试官，负责评估候选人的GO语言能力。"
	systemMessage += "\n\n当前状态：" + currentState
	switch currentState {
	case types.StateStart:
		systemMessage += "\n目标：欢迎候选人并开始面试流程"
	case types.StateQuestion:
		systemMessage += "\n目标：提出有深度的问题考察GO语言核心概念"
	case types.StateFollowUp:
		systemMessage += "\n目标：基于候选人的回答进行追问,深入考察理解深度"
	case types.StateEvaluate:
		systemMessage += "\n目标：全面评估候选人的技术能力"
	case types.StateEnd:
		systemMessage += "\n目标：结束面试并提供反馈"
	}
	//注入知识库
	if len(knowledge) > 0 {
		systemMessage += "\n\n相关背景知识："
		for i, k := range knowledge {
			//限制知识片段长度
			truncatedContent := utils.TruncateText(k.Content, 500)
			//truncatedContent := utils.TruncateText(k.Content, l.svcCtx.Config.VectorDB.Knowledge.MaxContextLength)
			systemMessage += fmt.Sprintf("\n[知识片段%d]%s：%s", i+1, k.Title, truncatedContent)
		}
	}

	//获取历史消息
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemMessage,
		},
	}
	history, err := l.svcCtx.VectorStore.GetMessages(chatId, 10)
	if err != nil {
		return nil, err
	}
	for _, msg := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return messages, nil
}

// 获取会话历史消息
func (l *ChatLogic) getSessionHistory(chatId string, knowledge []types.KnowledgeChunk) ([]openai.ChatCompletionMessage, error) {
	//获取最近10条历史消息(约5轮对话)
	vectorMessages, err := l.svcCtx.VectorStore.GetMessages(chatId, 10)
	if err != nil {
		return nil, err
	}
	// 构建系统消息 - 注入知识
	systemMessage := "你是一个专业的GO语言面试官，负责评估候选人的GO语言能力。请提出有深度的问题并评估回答。"
	//systemMessage := "你是一个AI智能问答助手。"
	if len(knowledge) > 0 {
		systemMessage += "\n\n相关背景知识："
		for i, k := range knowledge {
			//限制知识片段长度
			truncatedContent := utils.TruncateText(k.Content, 500)
			//truncatedContent := utils.TruncateText(k.Content, l.svcCtx.Config.VectorDB.Knowledge.MaxContextLength)
			systemMessage += fmt.Sprintf("\n[知识片段%d]%s：%s", i+1, k.Title, truncatedContent)
		}
	}
	fmt.Println("检索的数据:", systemMessage)
	//转换为OpenAI消息
	messages := make([]openai.ChatCompletionMessage, 0, len(vectorMessages)+1)

	//添加系统消息
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemMessage,
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

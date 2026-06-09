// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"ai-gozero-agent/api/internal/config"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

type ServiceContext struct {
	Config       config.Config
	OpenAIClient *openai.Client
	//SessionStore types.SessionStore //新增会话存储
	VectorStore *VectorStore //消息存储
}

func NewServiceContext(c config.Config) *ServiceContext {
	//创建OpenAI客户端
	//openaiConfig := openai.DefaultConfig(c.OpenAI.ApiKey)
	openaiConfig := openai.DefaultConfig("") //ollama不需要apikey
	openaiConfig.BaseURL = c.OpenAI.BaseURL
	openAIClient := openai.NewClientWithConfig(openaiConfig)

	//初始化向量存储
	vectorStore, err := NewVectorStore(c.VectorDB, openAIClient)
	if err != nil {
		log.Fatalf("初始化向量数据库失败: %v", err)
	}

	if err := vectorStore.TestConnection(); err != nil {
		log.Fatalf("向量数据库连接失败: %v", err)
	} else {
		log.Println("向量数据库连接成功")
	}

	return &ServiceContext{
		Config:       c,
		OpenAIClient: openAIClient,
		VectorStore:  vectorStore, //新增向量存储
	}
}

/*func NewServiceContext(c config.Config) *ServiceContext {
	conf := openai.DefaultConfig(c.OpenAI.ApiKey)
	conf.BaseURL = c.OpenAI.BaseURL

	return &ServiceContext{
		Config:       c,
		OpenAIClient: openai.NewClientWithConfig(conf),
		SessionStore: NewMemorySessionStore(), //新增内存会话存储
	}
}*/

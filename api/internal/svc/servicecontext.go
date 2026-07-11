// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"ai-gozero-agent/api/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	openai "github.com/sashabaranov/go-openai"
	//"github.com/unidoc/unipdf/v3/common/license"
)

type ServiceContext struct {
	Config       config.Config
	OpenAIClient *openai.Client
	//SessionStore types.SessionStore //新增会话存储
	VectorStore *VectorStore //消息存储
	PdfClient   *PdfClient
	Redis       *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	//创建OpenAI客户端
	openaiConfig := openai.DefaultConfig(c.OpenAI.ApiKey)
	//openaiConfig := openai.DefaultConfig("") //ollama不需要apikey
	openaiConfig.BaseURL = c.OpenAI.BaseURL
	openAIClient := openai.NewClientWithConfig(openaiConfig)

	//初始化Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port),
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})
	// 测试Redis连接
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	} else {
		log.Println("Redis连接成功")
	}

	//初始化向量存储
	vectorStore, err := NewVectorStore(c.VectorDB, openAIClient)
	if err != nil {
		log.Fatalf("初始化向量数据库失败: %v", err)
	}
	// 测试向量数据库连接
	if err := vectorStore.TestConnection(); err != nil {
		log.Fatalf("向量数据库连接失败: %v", err)
	} else {
		log.Println("向量数据库连接成功")
	}

	// 设置UniPDF key
	/*err = license.SetMeteredKey(c.UniPDFLicense)
	if err != nil {
		log.Fatalf("设置 UniPDF 许可证失败: %v", err)
		//如果没有授权,UniPDF会加水印
	}*/

	return &ServiceContext{
		Config:       c,
		OpenAIClient: openAIClient,
		VectorStore:  vectorStore, //新增向量存储
		PdfClient:    NewPdfClient(c.MCP.Endpoint),
		Redis:        rdb,
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

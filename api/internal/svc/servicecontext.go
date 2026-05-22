// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"ai-gozero-agent/api/internal/config"

	openai "github.com/sashabaranov/go-openai"
)

type ServiceContext struct {
	Config       config.Config
	OpenAIClient *openai.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	conf := openai.DefaultConfig(c.OpenAI.ApiKey)
	conf.BaseURL = c.OpenAI.BaseURL

	return &ServiceContext{
		Config:       c,
		OpenAIClient: openai.NewClientWithConfig(conf),
	}
}

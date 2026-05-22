// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	OpenAI struct {
		ApiKey      string
		BaseURL     string
		Model       string
		MaxTokens   int
		Temperature float32
	}
}

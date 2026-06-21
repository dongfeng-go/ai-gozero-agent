// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	MCP struct {
		Endpoint string
	}
	OpenAI struct {
		ApiKey  string `json:"apiKey"`  //api密钥
		BaseURL string `json:"baseUrl"` //api地址
		Model   string `json:"model"`   //模型名称
		//EmbeddingModel string `json:"embeddingModel"` //嵌入模型名称

		//核心生成参数
		MaxTokens   int     `json:"maxTokens"`   //生成最大长度
		Temperature float32 `json:"temperature"` //随机性,温度参数,取值范围0-2,越高越随机
		//TopP        float32 `json:"topP"`        //核心采样(0-1,越高越多样)
		////StopSequences   []string `json:"stop_sequences"`   //停止生成序列(如输入"\n","###"则生成结束)
		//PresencePenalty  float32 `json:"presencePenalty"`  //存在惩罚(-2.0,到2.0)
		//FrequencyPenalty float32 `json:"frequencyPenalty"` //频率惩罚(-2.0,到2.0)
		//Seed             *int    `json:"seed"`             //随机数种子(-1表示随机)
	}
	VectorDB VectorDBConfig //向量数据库配置

	UniPDFLicense string //UniPDF商业版许可证密钥
}

// 向量数据库配置
type VectorDBConfig struct {
	Host           string
	Port           int
	DBName         string
	User           string
	Password       string
	Table          string
	MaxConn        int
	EmbeddingModel string
	Knowledge      Knowledge
}

type Knowledge struct {
	MaxChunkSize     int
	TopK             int
	MaxContextLength int
}

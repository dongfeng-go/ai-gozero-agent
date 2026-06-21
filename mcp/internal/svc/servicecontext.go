package svc

import (
	"ai-gozero-agent/mcp/internal/config"
	"log"

	"github.com/unidoc/unipdf/v3/common/license"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 设置UniPDF key
	err := license.SetMeteredKey(c.UniPDFLicense)
	if err != nil {
		log.Fatalf("设置 UniPDF 许可证失败: %v", err)
		//如果没有授权,UniPDF会加水印
	}
	return &ServiceContext{
		Config: c,
	}
}

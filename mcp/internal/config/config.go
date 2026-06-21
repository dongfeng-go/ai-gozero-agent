package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	UniPDFLicense string //UniPDF商业版许可证密钥
}

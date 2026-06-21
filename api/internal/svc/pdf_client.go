package svc

import (
	"ai-gozero-agent/mcp/types/mcp"
	"context"
	"errors"
	"io"
	"mime/multipart"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type PdfClient struct {
	client mcp.PdfProcessorClient
}

func NewPdfClient(endpoint string) *PdfClient {
	// 创建gRPC连接
	conn := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{endpoint},
		NonBlock:  true,
	})
	return &PdfClient{
		client: mcp.NewPdfProcessorClient(conn.Conn()),
	}
}

// ExtractText 简化接口:传入文件流和文件名，返回解析后的文本
func (c *PdfClient) ExtractText(file multipart.File, fileName string) (string, error) {
	//创建grpc流
	stream, err := c.client.ExtractText(context.Background())
	if err != nil {
		logx.Errorf("grpc链接失败: %v", err)
		return "", err
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logx.Errorf("关闭流失败: %v", err)
		}
	}()
	//发送元数据
	if err := stream.Send(&mcp.PdfRequest{
		Data: &mcp.PdfRequest_Metadata{
			Metadata: &mcp.Metadata{
				Filename: fileName,
				MimeType: "application/pdf",
			},
		},
	}); err != nil {
		logx.Errorf("发送元数据失败: %v", err)
		return "", err
	}
	//一次性发送整个文件
	fileDate, err := io.ReadAll(file)
	if err != nil {
		logx.Errorf("读取文件失败: %v", err)
		return "", err
	}

	if err := stream.Send(&mcp.PdfRequest{
		Data: &mcp.PdfRequest_Chunk{
			Chunk: fileDate,
		},
	}); err != nil {
		logx.Errorf("发送文件失败: %v", err)
		return "", err
	}
	logx.Debugf("发送文件成功")
	//关闭发送并接受响应
	err = stream.CloseSend()
	if err != nil {
		logx.Errorf("关闭发送失败: %v", err)
		return "", err
	}
	resp, err := stream.Recv()
	if err != nil {
		logx.Errorf("关闭发送并接受响应失败: %v", err)
		return "", err
	}
	if resp.Error != "" {
		logx.Errorf("PDF解析失败: %s", resp.Error)
		return "", errors.New(resp.Error)
	}
	return resp.GetContent(), nil
}

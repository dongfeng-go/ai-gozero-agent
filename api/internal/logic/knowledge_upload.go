package logic

import (
	"ai-gozero-agent/api/internal/svc"
	"ai-gozero-agent/api/internal/types"
	"ai-gozero-agent/api/internal/utils"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowledgeUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowledgeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeUploadLogic {
	return &KnowledgeUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KnowledgeUploadLogic) KnowledgeUpload(req *types.KnowledgeUploadReq) (*types.KnowledgeUploadResp, error) {
	fmt.Println("进入Logic处理！！：", req.Title)
	//分块处理知识内容 todo
	chunks := utils.SplitText(req.Content, l.svcCtx.Config.VectorDB.Knowledge.MaxChunkSize)
	fmt.Println("准备分块！！：")
	for _, chunk := range chunks {
		fmt.Println("分块处理内容！！：", chunk)
		//保存知识块
		err := l.svcCtx.VectorStore.SaveKnowledge(req.Title, chunk, l.svcCtx.Config.VectorDB)
		if err != nil {
			fmt.Println("保存知识块失败！！：", err)
			return nil, err
		}
	}
	fmt.Println("分块保存结束！！：")
	return &types.KnowledgeUploadResp{
		Chunks: len(chunks),
		Msg:    "上传成功",
	}, nil
}

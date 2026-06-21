package handler

import (
	"ai-gozero-agent/api/internal/logic"
	"ai-gozero-agent/api/internal/svc"
	"ai-gozero-agent/api/internal/types"
	"errors"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func KnowledgeUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//设置SSE响应头
		setSSEHeaders(w)
		fmt.Println("进入上传知识库！！")
		//获取文件
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.Error(w, err)
			return
		}
		defer file.Close()

		//验证PDF
		if header.Header.Get("Content-Type") != "application/pdf" {
			httpx.Error(w, errors.New("仅支持PDF文件"))
			return
		}

		//提取文本
		content, err := svcCtx.PdfClient.ExtractText(file, header.Filename)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		//获取标题
		//title := r.FormValue("title")
		title := header.Filename
		fmt.Println("上传文件标题：", title)
		l := logic.NewKnowledgeLogic(r.Context(), svcCtx)
		resp, err := l.KnowledgeUpload(&types.KnowledgeUploadReq{
			Title:   title,
			Content: content,
		})
		if err != nil {
			httpx.Error(w, err)
			return
		} else {
			httpx.OkJson(w, resp)
		}
	}
}

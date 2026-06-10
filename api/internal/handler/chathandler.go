// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"ai-gozero-agent/api/internal/logic"
	"ai-gozero-agent/api/internal/svc"
	"ai-gozero-agent/api/internal/types"
	"ai-gozero-agent/api/internal/utils"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Go面试官聊天SSE流式接口
func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头（与 OPTIONS 保持一致）
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Cache-Control, Connection")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 设置SSE响应头
		setSSEHeaders(w)
		flusher, _ := w.(http.Flusher)
		// 立即刷新，确保响应头被发送给客户端
		//flusher.Flush()

		// 处理请求
		var req types.InterviewAPPChatReq
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// 解析 multipart 表单，32MB 内存限制
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				sendSSEError(w, flusher, "解析表单失败: "+err.Error())
				return
			}
			req.Message = r.FormValue("message")
			req.ChatId = r.FormValue("chatId")

			// 处理 PDF 文件
			if file, header, err := r.FormFile("file"); err == nil {
				defer file.Close()
				if header.Header.Get("Content-Type") != "application/pdf" {
					sendSSEError(w, flusher, "仅支持PDF文件")
					return
				}
				//使用UniPDF提取文本
				pdfContent, err := utils.ExtractPDFText(file)
				if err != nil {
					sendSSEError(w, flusher, "PDF提取失败: "+err.Error())
					return
				}
				req.Message = utils.CombineMessages(req.Message, pdfContent)
			}
		} else {
			// 原有 JSON 解析逻辑
			//httpx.Parse(r, &req)
			//if err := httpx.Parse(r, &req); err != nil {
			if err := httpx.ParseJsonBody(r, &req); err != nil {
				sendSSEError(w, flusher, err.Error())
				return
			}
			// 注意：JSON 模式下文件上传不支持，保持原有逻辑
		}

		//处理PDF文件
		/*var pdfContent string
		if file, header, err := r.FormFile("file"); err == nil {
			defer file.Close()

			if header.Header.Get("Content-Type") != "application/pdf" {//验证文件类型
		        http.Error(w, "仅支持PDF文件", http.StatusBadRequest)
				return
			}

			if content, err := utils.ExtractPDFText(file); err == nil {//使用UniPDF提取文本
				pdfContent = content
			} else {

						logx.Errorf("PDF提取失败：%v", err)

						http.Error(w, "PDF提取失败", http.StatusInternalServerError)
				return
			}
		}
		req.Message = utils.CombineMessages(req.Message, pdfContent)*/
		logx.Infof("用户问题+++++++666666：%s", req.Message)

		//创建取消上下文
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel() //释放资源

		l := logic.NewChatLogic(ctx, svcCtx)
		respChan, err := l.Chat(&req)
		if err != nil {
			sendSSEError(w, flusher, err.Error())
			return
		}

		//处理流式响应
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-respChan:
				if !ok {
					//fmt.Fprint(w, "event: end\ndata: {}\n\n") //结束事件
					fmt.Fprint(w, "event: end\ndata: [[DONE]]\n\n") //结束事件
					flusher.Flush()
					return
				}
				//加上这个内容处理,符合前端markdown格式
				safeContent := strings.ReplaceAll(resp.Content, "\n", "\\n")
				safeContent = strings.ReplaceAll(safeContent, "\r", "\\r")
				//直接输出内容,不加JSON包装
				fmt.Fprintf(w, "data: %s\n\n", safeContent)
				//直接输出内容,不加JSON包装
				//fmt.Fprintf(w, "data: %s\n\n", resp.Content)
				flusher.Flush()

				if resp.IsLast {
					// 停止循环
					//fmt.Fprint(w, "data: [[DONE]]\n\n")
					//flusher.Flush()
					return
				}
			}
		}

	}
}

// setSSEHeaders 设置服务器推送时间(SSE)的响应头
func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	//w.Header().Set("Access-Control-Allow-Origin", "Origin")
	w.Header().Set("X-Accel-Buffering", "no")
	//w.Header().Set("Transfer-Encoding", "chunked")
}

func sendSSEError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
	_, fprintf := fmt.Fprintf(w, "event: error\ndata: {\"error\":\"%s\"}\n\n", errMsg)
	if fprintf != nil {
		return
	}
	flusher.Flush()
}

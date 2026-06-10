package utils

import (
	"bytes"
	"io"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// ExtractPDFText 使用Uni提取PDF文本
func ExtractPDFText(file io.Reader) (string, error) {
	//创建内存缓冲区避免重复读取
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return "", err
	}

	//创建PDF阅读器
	pdfReader, err := model.NewPdfReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", err
	}

	//提取文贝
	var textBuilder strings.Builder
	if numPages, err := pdfReader.GetNumPages(); err == nil {
		for i := 1; i <= numPages; i++ {
			if page, err := pdfReader.GetPage(i); err == nil {
				if ex, err := extractor.New(page); err == nil {
					if pageText, err := ex.ExtractText(); err == nil {
						textBuilder.WriteString(strings.TrimSpace(pageText))
					}
				}
			}
		}
	}
	return textBuilder.String(), nil
}

// 简单拼接用户消息和PDF内容
func CombineMessages(userMsg, pdfContent string) string {
	if pdfContent == "" {
		return userMsg
	}
	return userMsg + "\n[PDF内容开始]" + pdfContent + "[PDF内容结束]"
}

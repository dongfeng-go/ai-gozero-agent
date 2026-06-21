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
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}
	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return "", err
		}
		ex, err := extractor.New(page)
		if err != nil {
			return "", err
		}
		pageText, err := ex.ExtractText()
		if err != nil {
			return "", err
		}
		textBuilder.WriteString(strings.TrimSpace(pageText))
		textBuilder.WriteString("\n\n")
	}
	return textBuilder.String(), nil
}

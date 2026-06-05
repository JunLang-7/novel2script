package text

import (
	"bytes"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// DetectAndReadFile 自动检测文件编码并返回 UTF-8 内容。
// 支持 UTF-8、GBK/GB2312 编码。
func DetectAndReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	if utf8.Valid(data) {
		return string(data), nil
	}

	// 尝试 GBK → UTF-8 转换
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

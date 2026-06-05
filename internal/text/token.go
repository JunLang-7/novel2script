package text

// EstimateTokens 估算中文文本的 token 数量。
// 中文约 1.5-2 字/token（对 Claude），混合文本取保守估值 1.5 字/token。
func EstimateTokens(text string) int {
	runes := len([]rune(text))
	// 保守估算：1.5 个中文字符 ≈ 1 个 token
	return runes * 2 / 3
}

// FormatCharCount 将字符数格式化为可读形式。
func FormatCharCount(count int) string {
	if count < 1000 {
		return itoa(count) + "字"
	}
	if count < 10000 {
		return itoa(count/1000) + "千字"
	}
	return itoa(count/10000) + "万字"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

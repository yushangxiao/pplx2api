package utils

import (
	"fmt"
	"pplx2api/config"
	"strings"
)

// cleanURL 对 URL 进行简单的转义处理，防止破坏 Markdown 结构
// cleanURL 强力清洗 URL
func cleanURL(rawURL string) string {
	// 1. 去除首尾所有空白字符 (空格、换行、Tab)
	rawURL = strings.TrimSpace(rawURL)

	// 2. 移除可能存在的双重协议头 (防止 https://https://)
	if strings.HasPrefix(rawURL, "https://https://") {
		rawURL = strings.Replace(rawURL, "https://https://", "https://", 1)
	}

	// 3. 替换 URL 中间的空格为 %20
	rawURL = strings.ReplaceAll(rawURL, " ", "%20")
	// 4. 转义括号
	rawURL = strings.ReplaceAll(rawURL, "(", "%28")
	rawURL = strings.ReplaceAll(rawURL, ")", "%29")

	return rawURL
}

// cleanTitle 清洗标题，防止破坏 Markdown 结构
func cleanTitle(title string) string {
	// 移除换行符
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	// 转义方括号
	title = strings.ReplaceAll(title, "[", "\\[")
	title = strings.ReplaceAll(title, "]", "\\]")
	return strings.TrimSpace(title)
}

func searchShowDetails(index int, title, url, snippet string) string {
	// 强制清洗
	safeURL := cleanURL(url)
	if safeURL == "" {
		return ""
	}
	// 确保 ]( 后面紧跟 URL，没有任何空格
	return fmt.Sprintf("[%d] [%s](%s)", index, cleanTitle(title), safeURL)
}

func searchShowCompatible(index int, title, url, snippet string) string {
	if len([]rune(snippet)) > 50 {
		runeSnippet := []rune(snippet)
		snippet = string(runeSnippet[:50]) + "..."
	}
	snippet = strings.ReplaceAll(snippet, "\n", " ")

	// 强制清洗
	safeURL := cleanURL(url)
	if safeURL == "" {
		return ""
	}

	// 确保 ]( 后面紧跟 URL
	return fmt.Sprintf("%d. [%s](%s) - %s", index, cleanTitle(title), safeURL, snippet)
}

func SearchShow(index int, title, url, snippet string) string {
	index++
	// 这里的 TrimSpace 其实在 cleanURL 里也做了，双重保险
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}

	if config.ConfigInstance.SearchResultCompatible {
		return searchShowCompatible(index, title, url, snippet)
	}
	return searchShowDetails(index, title, url, snippet)
}

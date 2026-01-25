package config

var ModelReverseMap = map[string]string{}
var ModelMap = map[string]string{
	// Perplexity 原生模型
	"sonar": "turbo", // 基础模型

	// Perplexity 映射的高级模型 (根据你的需求保留)
	"claude-4.5-sonnet":       "claude45sonnet",
	"claude-4.5-sonnet-think": "claude45sonnetthinking",
	"gemini-3-pro":            "gemini30pro",
	"gpt-5.2":                 "gpt52",

	// 修复 Search/Research (将 sonar-pro 映射到网页版的高级模型)
	"sonar-pro": "gpt52",

	// 修复 Reasoning (将 sonar-reasoning-pro 映射到网页版的思考模型)
	"sonar-reasoning-pro": "claude45sonnetthinking",
	"sonar-reasoning":     "claude45sonnetthinking",
}
var MaxModelMap = map[string]string{}

// Get returns the value for the given key from the ModelMap.
// If the key doesn't exist, it returns the provided default value.
func ModelMapGet(key string, defaultValue string) string {
	if value, exists := ModelMap[key]; exists {
		return value
	}
	return defaultValue
}

// GetReverse returns the value for the given key from the ModelReverseMap.
// If the key doesn't exist, it returns the provided default value.
func ModelReverseMapGet(key string, defaultValue string) string {
	if value, exists := ModelReverseMap[key]; exists {
		return value
	}
	return defaultValue
}

var ResponseModels []map[string]string

func init() {
	// 构建反向映射
	for k, v := range ModelMap {
		ModelReverseMap[v] = k
	}
	buildResponseModels()
}

// buildResponseModels 构建响应模型列表
func buildResponseModels() {
	ResponseModels = make([]map[string]string, 0, len(ModelMap)*2)

	for modelID := range ModelMap {
		// 添加普通模型
		ResponseModels = append(ResponseModels, map[string]string{
			"id": modelID,
		})

		// 添加搜索模型
		ResponseModels = append(ResponseModels, map[string]string{
			"id": modelID + "-search",
		})
	}
}

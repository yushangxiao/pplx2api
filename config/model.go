package config

var ModelReverseMap = map[string]string{}
var ModelMap = map[string]string{
	"claude-4.0-sonnet":       "claude2",
	"claude-4.0-sonnet-think": "claude37sonnetthinking",
	"deepseek-r1":             "r1",
	"o4-mini":                 "o4mini",
	"gpt-4o":                  "gpt4o",
	"gemini-2.5-pro-06-05":    "gemini2flash",
	"grok4":                   "grok4",
	"gpt-4.1":                 "gpt41",
	// "claude-4.0-opus":         "claude40opus",
	"claude-4.0-opus-think": "claude40opusthinking",
	"o3":                    "o3",
	"o3-pro":                "o3pro",
}
var MaxModelMap = map[string]string{
	"o3-pro":                "o3pro",
	"claude-4.0-opus-think": "claude40opusthinking",
}

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
		// 如果不是最大订阅用户，跳过最大模型
		if !ConfigInstance.IsMaxSubscribe {
			if _, isMaxModel := MaxModelMap[modelID]; isMaxModel {
				continue
			}
		}

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

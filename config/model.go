package config

var ModelReverseMap = map[string]string{}
var ModelMap = map[string]string{
	"claude-4-6-sonnet":       "claude46sonnet",
	"claude-4.6-sonnet-think": "claude46sonnetthinking",
	"gemini-3.1-pro":          "gemini31pro_high",
	"gpt-5.2":                 "gpt52",
	"gpt-5.2-think":           "gpt52_thinking",
}
var MaxModelMap = map[string]string{
	"claude-4.6-opus":       "claude46opus",
	"claude-4.6-opus-think": "claude46opusthinking",
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
	//给modelmap添加max模型
	for k, v := range MaxModelMap {
		ModelMap[k] = v
	}
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

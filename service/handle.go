package service

import (
	"fmt"
	"net/http"
	"pplx2api/config"
	"pplx2api/core"
	"pplx2api/logger"
	"pplx2api/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type ChatCompletionRequest struct {
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
	Stream   bool                     `json:"stream"`
	Tools    []map[string]interface{} `json:"tools,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthCheckHandler handles the health check endpoint
func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// ChatCompletionsHandler handles the chat completions endpoint
func ChatCompletionsHandler(c *gin.Context) {

	// Parse request body
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}
	// logger.Info(fmt.Sprintf("Received request: %v", req))
	// Validate request
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "No messages provided",
		})
		return
	}

	// Get model or use default
	model := req.Model
	if model == "" {
		model = "claude-3.7-sonnet"
	}
	openSearch := false
	if strings.HasSuffix(model, "-search") {
		openSearch = true
		model = strings.TrimSuffix(model, "-search")
	}
	model = config.ModelMapGet(model, model) // 获取模型名称
	var prompt strings.Builder
	img_data_list := []string{}
	// Format messages into a single prompt
	for _, msg := range req.Messages {
		role, roleOk := msg["role"].(string)
		if !roleOk {
			continue // 忽略无效格式
		}

		content, exists := msg["content"]
		if !exists {
			continue
		}

		prompt.WriteString(utils.GetRolePrefix(role)) // 获取角色前缀
		switch v := content.(type) {
		case string: // 如果 content 直接是 string
			prompt.WriteString(v + "\n\n")
		case []interface{}: // 如果 content 是 []interface{} 类型的数组
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if itemType, ok := itemMap["type"].(string); ok {
						if itemType == "text" {
							if text, ok := itemMap["text"].(string); ok {
								prompt.WriteString(text + "\n\n")
							}
						} else if itemType == "image_url" {
							if imageUrl, ok := itemMap["image_url"].(map[string]interface{}); ok {
								if url, ok := imageUrl["url"].(string); ok {
									if len(url) > 50 {
										logger.Info(fmt.Sprintf("Image URL: %s ……", url[:50]))
									}
									if strings.HasPrefix(url, "data:image/") {
										// 保留 base64 编码的图片数据
										url = strings.Split(url, ",")[1]
									}
									img_data_list = append(img_data_list, url) // 收集图片数据
								}
							}
						}
					}
				}
			}
		}
	}
	fmt.Println(prompt.String())                             // 输出最终构造的内容
	fmt.Println("img_data_list_length:", len(img_data_list)) // 输出图片数据列表长度
	var rootPrompt strings.Builder
	rootPrompt.WriteString(prompt.String())
	// 切号重试机制
	var pplxClient *core.Client
	index := config.Sr.NextIndex()
	for i := 0; i < config.ConfigInstance.RetryCount; i++ {
		if i > 0 {
			prompt.Reset()
			prompt.WriteString(rootPrompt.String())
		}
		index = (index + 1) % len(config.ConfigInstance.Sessions)
		session, err := config.ConfigInstance.GetSessionForModel(index)
		logger.Info(fmt.Sprintf("Using session for model %s: %s", model, session.SessionKey))
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get session for model %s: %v", model, err))
			logger.Info("Retrying another session")
			continue
		}
		// Initialize the Claude client
		pplxClient = core.NewClient(session.SessionKey, config.ConfigInstance.Proxy, model, openSearch)
		if len(img_data_list) > 0 {
			err := pplxClient.UploadImage(img_data_list)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to upload file: %v", err))
				logger.Info("Retrying another session")

				continue
			}
		}
		if prompt.Len() > config.ConfigInstance.MaxChatHistoryLength {
			err := pplxClient.UploadText(prompt.String())
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to upload text: %v", err))
				logger.Info("Retrying another session")

				continue
			}
			prompt.Reset()
			prompt.WriteString(config.ConfigInstance.PromptForFile)
		}
		if _, err := pplxClient.SendMessage(prompt.String(), req.Stream, config.ConfigInstance.IsIncognito, c); err != nil {
			logger.Error(fmt.Sprintf("Failed to send message: %v", err))
			logger.Info("Retrying another session")

			continue // Retry on error
		}

		return

	}
	logger.Error("Failed for all retries")
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "Failed to process request after multiple attempts"})
}

func ModelsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": config.ResponseModels,
	})
}

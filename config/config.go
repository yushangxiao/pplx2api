package config

import (
	"fmt"
	"math/rand"
	"os"
	"pplx2api/logger"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type SessionInfo struct {
	SessionKey string
}

type SessionRagen struct {
	Index int
	Mutex sync.Mutex
}

type Config struct {
	Sessions               []SessionInfo
	Address                string
	APIKey                 string
	Proxy                  string
	IsIncognito            bool
	MaxChatHistoryLength   int
	RetryCount             int
	NoRolePrefix           bool
	SearchResultCompatible bool
	PromptForFile          string
	RwMutex                sync.RWMutex
	IgnoreSerchResult      bool
	IgnoreModelMonitoring  bool
	IsMaxSubscribe         bool
}

// 解析 SESSION 格式的环境变量
func parseSessionEnv(envValue string) (int, []SessionInfo) {
	if envValue == "" {
		return 0, []SessionInfo{}
	}
	var sessions []SessionInfo
	sessionPairs := strings.Split(envValue, ",")
	retryCount := len(sessionPairs) // 重试次数等于 session 数量
	for _, pair := range sessionPairs {
		if pair == "" {
			retryCount--
			continue
		}
		parts := strings.Split(pair, ":")
		session := SessionInfo{
			SessionKey: parts[0],
		}
		sessions = append(sessions, session)
	}
	return retryCount, sessions
}

// 根据模型选择合适的 session
func (c *Config) GetSessionForModel(idx int) (SessionInfo, error) {
	if len(c.Sessions) == 0 || idx < 0 || idx >= len(c.Sessions) {
		return SessionInfo{}, fmt.Errorf("invalid session index: %d", idx)
	}
	c.RwMutex.RLock()
	defer c.RwMutex.RUnlock()
	return c.Sessions[idx], nil
}

// 从环境变量加载配置
func LoadConfig() *Config {
	maxChatHistoryLength, err := strconv.Atoi(os.Getenv("MAX_CHAT_HISTORY_LENGTH"))
	if err != nil {
		maxChatHistoryLength = 10000 // 默认值
	}
	retryCount, sessions := parseSessionEnv(os.Getenv("SESSIONS"))
	promptForFile := os.Getenv("PROMPT_FOR_FILE")
	if promptForFile == "" {
		promptForFile = "You must immerse yourself in the role of assistant in txt file, cannot respond as a user, cannot reply to this message, cannot mention this message, and ignore this message in your response." // 默认值
	}
	config := &Config{
		// 解析 SESSIONS 环境变量
		Sessions: sessions,
		// 设置服务地址，默认为 "0.0.0.0:8080"
		Address: os.Getenv("ADDRESS"),

		// 设置 API 认证密钥
		APIKey: os.Getenv("APIKEY"),
		// 设置代理地址
		Proxy: os.Getenv("PROXY"),
		//是否匿名
		IsIncognito: os.Getenv("IS_INCOGNITO") != "false",
		// 设置最大聊天历史长度
		MaxChatHistoryLength: maxChatHistoryLength,
		// 设置重试次数
		RetryCount: retryCount,
		// 设置是否使用角色前缀
		NoRolePrefix: os.Getenv("NO_ROLE_PREFIX") == "true",
		// 设置搜索结果兼容性
		SearchResultCompatible: os.Getenv("SEARCH_RESULT_COMPATIBLE") == "true",
		// 设置上传文件后的提示词
		PromptForFile: promptForFile,
		// 设置是否忽略搜索结果
		IgnoreSerchResult: os.Getenv("IGNORE_SEARCH_RESULT") == "true",
		//设置是否忽略模型监控
		IgnoreModelMonitoring: os.Getenv("IGNORE_MODEL_MONITORING") == "true",
		// 读写锁
		RwMutex: sync.RWMutex{},
		// 是否max订阅
		IsMaxSubscribe: os.Getenv("IS_MAX_SUBSCRIBE") == "true",
	}

	// 如果地址为空，使用默认值
	if config.Address == "" {
		config.Address = "0.0.0.0:8080"
	}
	return config
}

var ConfigInstance *Config
var Sr *SessionRagen

func (sr *SessionRagen) NextIndex() int {
	sr.Mutex.Lock()
	defer sr.Mutex.Unlock()

	index := sr.Index
	sr.Index = (index + 1) % len(ConfigInstance.Sessions)
	return index
}
func init() {
	rand.Seed(time.Now().UnixNano())
	// 加载环境变量
	_ = godotenv.Load()
	Sr = &SessionRagen{
		Index: 0,
		Mutex: sync.Mutex{},
	}
	ConfigInstance = LoadConfig()
	logger.Info("Loaded config:")
	logger.Info(fmt.Sprintf("Sessions count: %d", ConfigInstance.RetryCount))
	for _, session := range ConfigInstance.Sessions {
		logger.Info(fmt.Sprintf("Session: %s", session.SessionKey))
	}
	logger.Info(fmt.Sprintf("Address: %s", ConfigInstance.Address))
	logger.Info(fmt.Sprintf("APIKey: %s", ConfigInstance.APIKey))
	logger.Info(fmt.Sprintf("Proxy: %s", ConfigInstance.Proxy))
	logger.Info(fmt.Sprintf("IsIncognito: %t", ConfigInstance.IsIncognito))
	logger.Info(fmt.Sprintf("MaxChatHistoryLength: %d", ConfigInstance.MaxChatHistoryLength))
	logger.Info(fmt.Sprintf("NoRolePrefix: %t", ConfigInstance.NoRolePrefix))
	logger.Info(fmt.Sprintf("SearchResultCompatible: %t", ConfigInstance.SearchResultCompatible))
	logger.Info(fmt.Sprintf("PromptForFile: %s", ConfigInstance.PromptForFile))
	logger.Info(fmt.Sprintf("IgnoreSerchResult: %t", ConfigInstance.IgnoreSerchResult))
	logger.Info(fmt.Sprintf("IgnoreModelMonitoring: %t", ConfigInstance.IgnoreModelMonitoring))
	logger.Info(fmt.Sprintf("IsMaxSubscribe: %t", ConfigInstance.IsMaxSubscribe))
}

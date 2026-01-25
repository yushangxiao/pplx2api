package job

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"pplx2api/config"
	"pplx2api/core"
)

const (
	// ConfigFileName is the name of the file to store sessions
	ConfigFileName = "sessions.json"
)

var (
	sessionUpdaterInstance *SessionUpdater
	sessionUpdaterOnce     sync.Once
)

// SessionConfig represents the structure to be saved to file
type SessionConfig struct {
	Sessions []config.SessionInfo `json:"sessions"`
}

// SessionUpdater 管理 Perplexity 会话的定时更新
type SessionUpdater struct {
	interval    time.Duration
	stopChan    chan struct{}
	isRunning   bool
	runningLock sync.Mutex
	configPath  string
}

// NewSessionUpdater 创建一个新的会话更新器
// interval: 更新间隔时间
func GetSessionUpdater(interval time.Duration) *SessionUpdater {
	sessionUpdaterOnce.Do(func() {
		// 使用当前文件夹下的配置文件
		configPath := ConfigFileName

		sessionUpdaterInstance = &SessionUpdater{
			interval:   interval,
			stopChan:   make(chan struct{}),
			isRunning:  false,
			configPath: configPath,
		}
		// 初始化时从文件加载会话
		sessionUpdaterInstance.loadSessionsFromFile()
	})
	return sessionUpdaterInstance
}

// loadSessionsFromFile loads sessions from the config file if it exists
func (su *SessionUpdater) loadSessionsFromFile() {
	// Check if file exists
	if _, err := os.Stat(su.configPath); os.IsNotExist(err) {
		log.Println("No sessions config file found, will create on first update")
		return
	}

	// Read the file
	data, err := ioutil.ReadFile(su.configPath)
	if err != nil {
		log.Printf("Failed to read sessions config file: %v", err)
		return
	}

	// Parse the JSON
	var sessionConfig SessionConfig
	if err := json.Unmarshal(data, &sessionConfig); err != nil {
		log.Printf("Failed to parse sessions config file: %v", err)
		return
	}

	// Update the config with loaded sessions
	config.ConfigInstance.RwMutex.Lock()
	config.ConfigInstance.Sessions = sessionConfig.Sessions
	config.ConfigInstance.RwMutex.Unlock()

	log.Printf("Loaded %d sessions from config file", len(sessionConfig.Sessions))
}

// saveSessionsToFile saves the current sessions to the config file
func (su *SessionUpdater) saveSessionsToFile() error {
	// Get current sessions
	config.ConfigInstance.RwMutex.RLock()
	sessionsCopy := make([]config.SessionInfo, len(config.ConfigInstance.Sessions))
	copy(sessionsCopy, config.ConfigInstance.Sessions)
	config.ConfigInstance.RwMutex.RUnlock()

	// Create config structure
	sessionConfig := SessionConfig{
		Sessions: sessionsCopy,
	}

	// Convert to JSON
	data, err := json.MarshalIndent(sessionConfig, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	err = ioutil.WriteFile(su.configPath, data, 0644)
	if err != nil {
		return err
	}

	log.Printf("Saved %d sessions to sessions.json file", len(sessionsCopy))
	return nil
}

// Start 启动定时更新任务
func (su *SessionUpdater) Start() {
	su.runningLock.Lock()
	defer su.runningLock.Unlock()
	if su.isRunning {
		log.Println("Session updater is already running")
		return
	}
	su.isRunning = true
	su.stopChan = make(chan struct{})
	go su.runUpdateLoop()
	log.Println("Session updater started with interval:", su.interval)
}

// Stop 停止定时更新任务
func (su *SessionUpdater) Stop() {
	su.runningLock.Lock()
	defer su.runningLock.Unlock()
	if !su.isRunning {
		log.Println("Session updater is not running")
		return
	}
	close(su.stopChan)
	su.isRunning = false
	log.Println("Session updater stopped")
}

// runUpdateLoop 运行更新循环
func (su *SessionUpdater) runUpdateLoop() {
	ticker := time.NewTicker(su.interval)
	defer ticker.Stop()
	// 立即执行一次更新
	// su.updateAllSessions()
	for {
		select {
		case <-ticker.C:
			su.updateAllSessions()
		case <-su.stopChan:
			log.Println("Update loop terminated")
			return
		}
	}
}

// updateAllSessions 更新所有会话
func (su *SessionUpdater) updateAllSessions() {
	log.Println("Starting session update for all sessions...")
	// 复制当前会话列表，避免长时间持有锁
	config.ConfigInstance.RwMutex.RLock()
	sessionsCopy := make([]config.SessionInfo, len(config.ConfigInstance.Sessions))
	copy(sessionsCopy, config.ConfigInstance.Sessions)
	proxy := config.ConfigInstance.Proxy
	config.ConfigInstance.RwMutex.RUnlock()
	// 如果没有会话需要更新，直接返回
	if len(sessionsCopy) == 0 {
		log.Println("No sessions to update")
		return
	}
	// 创建更新后的会话切片
	updatedSessions := make([]config.SessionInfo, len(sessionsCopy))
	var wg sync.WaitGroup
	// 对每个会话执行更新
	for i, session := range sessionsCopy {
		wg.Add(1)
		go func(index int, origSession config.SessionInfo) {
			defer wg.Done()
			// 创建客户端并更新 cookie
			// 写死 model 和 openSearch 参数
			client := core.NewClient(origSession.SessionKey, proxy, "claude-3-opus-20240229", false)
			newCookie, err := client.GetNewCookie()
			if err != nil {
				log.Printf("Failed to update session %d: %v", index, err)
				// 如果更新失败，保留原始会话
				updatedSessions[index] = origSession
				return
			}
			// 创建更新后的会话对象
			updatedSessions[index] = config.SessionInfo{
				SessionKey: newCookie,
			}
		}(i, session)
	}
	// 等待所有更新完成
	wg.Wait()
	// 一次性替换所有会话
	config.ConfigInstance.RwMutex.Lock()
	config.ConfigInstance.Sessions = updatedSessions
	config.ConfigInstance.RwMutex.Unlock()
	log.Printf("All %d sessions have been updated", len(updatedSessions))

	// 保存更新后的配置到文件
	if err := su.saveSessionsToFile(); err != nil {
		log.Printf("Failed to save updated config: %v", err)
	}
}

# Render 部署指南

本文档提供了在 Render 云平台部署 pplx2api 的详细指南和优化配置。

## 🚀 快速开始

### 1. 环境变量配置

在 Render 仪表板中设置以下环境变量：

```bash
SESSIONS=your_session_token_1,your_session_token_2  # 多个session用逗号分隔
ADDRESS=0.0.0.0:8080
APIKEY=your_api_key_here  # 用于API认证
IS_INCOGNITO=true  # 启用隐私模式
MAX_CHAT_HISTORY_LENGTH=5000  # Render环境建议减小历史长度
NO_ROLE_PREFIX=false
SEARCH_RESULT_COMPATIBLE=false
IGNORE_SEARCH_RESULT=false
RENDER=true  # 标识Render环境，启用优化配置
```

### 2. 部署配置

**Build Command:**
```bash
go mod download && go build -o main ./main.go
```

**Start Command:**
```bash
./main
```

### 3. 实例类型推荐

- **Free Tier**: 适合测试和小规模使用
- **Starter Plan**: 推荐用于生产环境（$7/月）
- **Standard Plan**: 高流量场景（$25/月）

## ⚙️ Render 优化配置

项目已针对 Render 环境进行了以下优化：

### 超时设置优化
- API 调用超时从 10分钟 减少到 2分钟
- 响应头超时保持 30秒
- 适应 Render 的 10分钟请求超时限制

### 内存优化
- 缓冲区大小从 1MB 减少到 256KB/512KB
- 减少大文件处理时的内存占用

### 会话管理优化
- 在 Render 环境中自动禁用会话自动更新
- 避免后台任务导致的超时问题

### 健康检查优化
- 快速响应的健康检查端点
- 包含服务状态和时间戳信息

## 🔧 故障排除

### 常见问题

1. **超时错误**
   - 检查 `MAX_CHAT_HISTORY_LENGTH` 是否设置过大
   - 确认网络连接正常，特别是代理设置

2. **内存不足**
   - 升级到更大的实例类型
   - 减少并发请求数量

3. **会话失效**
   - 确保 SESSIONS 环境变量配置正确
   - 检查 Perplexity 账户状态

### 性能监控

建议在 Render 仪表板中监控：
- CPU 使用率
- 内存使用情况
- 响应时间
- 错误率

## 📊 最佳实践

1. **使用多个会话**：在 SESSIONS 中配置多个 token 以提高可用性
2. **合理设置超时**：根据实际需求调整超时时间
3. **监控日志**：定期检查应用日志以发现潜在问题
4. **定期更新**：保持项目依赖项更新到最新版本

## 🌐 网络配置

如果需要在 Render 上使用代理：

```bash
PROXY=http://your-proxy-server:port
```

## 📝 支持

如果遇到部署问题，请检查：
1. 环境变量配置是否正确
2. Render 实例资源是否充足
3. 网络连接是否正常

更多帮助请参考项目主 README 或提交 Issue。
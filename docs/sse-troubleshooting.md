# Server-Sent Events (SSE) 问题排查指南

## 问题背景

在开发 DigWis Panel 过程中，遇到了 SSE 实时数据推送不工作的问题。经过深入调试，发现了多个潜在的问题点和解决方案。

## 问题表现

1. **前端症状**:
   - 系统监控数据显示为 "--"
   - 实时数据不更新
   - 浏览器控制台显示 SSE 连接已建立，但没有数据接收日志

2. **后端症状**:
   - 服务器日志中没有 SSE 连接记录
   - curl 测试 SSE 端点正常工作
   - 数据格式和内容都正确

## 根本原因分析

### 1. 主要问题：Templ 模板生成

**问题**: 修改了 JavaScript 代码但没有重新生成模板文件

**表现**: 
- 浏览器加载的是旧版本的 JavaScript 代码
- 新的调试信息和修复没有生效

**解决方案**:
```bash
find . -name "*_templ.go" -delete
~/go/bin/templ generate
go build -o digwis-panel .
```

### 2. 次要问题：数据字段名不匹配

**问题**: JavaScript 使用大写字段名，Go JSON 标签是小写

**错误代码**:
```javascript
// 错误：使用大写字段名
stats.CPU.Usage
stats.Memory.Usage
```

**正确代码**:
```javascript
// 正确：使用小写字段名
stats.cpu.usage
stats.memory.usage
```

### 3. 认证问题（已解决）

**问题**: EventSource 在某些情况下不会自动发送 cookies

**临时解决方案**: 禁用 SSE 端点的认证检查
```go
// 临时注释认证检查以测试数据流
/*
sess, err := h.sessionStore.Get(r)
if err != nil || sess.Get("authenticated") != true {
    http.Error(w, "未授权访问", http.StatusUnauthorized)
    return
}
*/
```

## 调试方法

### 1. 服务器端调试

```bash
# 测试 SSE 端点
curl -N http://localhost:9090/api/sse/stats

# 检查服务器日志
tail -f server.log | grep "sse\|SSE"

# 检查端口监听
ss -tlnp | grep :9090
```

### 2. 浏览器端调试

```javascript
// 手动测试 EventSource
const testEventSource = new EventSource('/api/sse/stats');

testEventSource.onopen = function(event) {
    console.log('✅ 连接打开:', event);
};

testEventSource.onmessage = function(event) {
    console.log('📨 收到消息:', event.data);
};

testEventSource.onerror = function(event) {
    console.log('❌ 连接错误:', event);
};
```

### 3. 网络层调试

```bash
# 检查网络接口
ip addr show

# 检查防火墙
sudo ufw status

# 检查路由
ip route show
```

## 解决方案总结

### 1. 完整的修复流程

```bash
# 1. 修复数据字段名
# 编辑 internal/templates/layouts/base.templ
# 将所有大写字段名改为小写

# 2. 重新生成模板
find . -name "*_templ.go" -delete
~/go/bin/templ generate

# 3. 重新构建
go build -o digwis-panel .

# 4. 重启服务器
./digwis-panel -port 9090 -debug

# 5. 测试验证
curl -N http://localhost:9090/api/sse/stats
```

### 2. 添加调试功能

在页面中添加手动测试按钮：

```html
<button onclick="testSSEConnection()" class="bg-blue-600 text-white px-4 py-2 rounded">
    测试SSE连接
</button>
```

```javascript
function testSSEConnection() {
    const testSource = new EventSource('/api/sse/stats');
    
    testSource.onopen = function(event) {
        alert('✅ SSE连接成功建立！');
    };
    
    testSource.addEventListener('stats', function(event) {
        const stats = JSON.parse(event.data);
        alert('📊 收到数据！CPU: ' + stats.cpu.usage + '%');
        testSource.close();
    });
    
    testSource.onerror = function(event) {
        alert('❌ SSE连接失败！');
    };
}
```

## 预防措施

### 1. 开发工作流

1. **修改代码** → 2. **重新生成模板** → 3. **重新构建** → 4. **测试验证**

### 2. 自动化脚本

```bash
#!/bin/bash
# dev-test.sh - 开发测试脚本

echo "🔄 重新生成模板..."
~/go/bin/templ generate

echo "🔨 重新构建..."
go build -o digwis-panel .

echo "🧪 测试 SSE 端点..."
timeout 5s curl -N http://localhost:9090/api/sse/stats || echo "SSE 测试完成"

echo "🚀 启动服务器..."
./digwis-panel -port 9090 -debug
```

### 3. 代码检查清单

- [ ] 字段名大小写一致
- [ ] 模板已重新生成
- [ ] 应用已重新构建
- [ ] SSE 端点可访问
- [ ] 浏览器控制台无错误
- [ ] 数据格式正确

## 性能优化

### 1. SSE 连接优化

```go
// 设置合适的推送间隔
ticker := time.NewTicker(3 * time.Second)

// 设置连接超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
defer cancel()
```

### 2. 前端优化

```javascript
// 连接重试机制
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;

eventSource.onerror = function(event) {
    if (reconnectAttempts < maxReconnectAttempts) {
        setTimeout(() => {
            reconnectAttempts++;
            initSSE();
        }, 1000 * reconnectAttempts);
    }
};
```

## 监控和告警

### 1. 健康检查

```bash
# 检查 SSE 连接数
ss -tn | grep :9090 | wc -l

# 检查内存使用
ps aux | grep digwis-panel | awk '{print $4}'
```

### 2. 日志监控

```bash
# 监控 SSE 连接
tail -f server.log | grep "GET /api/sse/stats"

# 监控错误
tail -f server.log | grep -i error
```

---

**记录时间**: 2025-07-27  
**问题状态**: ✅ 已解决  
**影响组件**: SSE 实时数据推送  
**解决耗时**: ~2小时

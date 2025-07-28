# DigWis Panel 开发工作流程

## 项目技术栈

- **后端**: Go + Chi Router
- **前端**: Tailwind CSS + HTMX + Alpine.js
- **模板引擎**: Templ
- **实时数据**: Server-Sent Events (SSE)

## 开发环境设置

### 1. 安装依赖工具

```bash
# 安装 Templ 模板生成器
go install github.com/a-h/templ/cmd/templ@latest

# 验证安装
~/go/bin/templ version
```

### 2. 项目结构

```
digwis-panel/
├── assets/                 # 静态资源
│   ├── css/
│   └── js/
├── docs/                   # 项目文档
├── internal/
│   ├── handlers/          # HTTP 处理器
│   ├── router/            # 路由器
│   ├── server/            # 服务器配置
│   ├── session/           # 会话管理
│   ├── system/            # 系统监控
│   └── templates/         # 模板文件
│       ├── components/    # 组件模板
│       ├── layouts/       # 布局模板
│       └── pages/         # 页面模板
├── main.go
└── go.mod
```

## 标准开发流程

### 1. 修改模板文件

当需要修改前端界面时：

```bash
# 编辑模板文件
vim internal/templates/pages/dashboard.templ
```

### 2. 重新生成模板

**⚠️ 重要：每次修改 .templ 文件后必须执行此步骤**

```bash
# 重新生成所有模板
~/go/bin/templ generate

# 或者强制重新生成（推荐）
find . -name "*_templ.go" -delete && ~/go/bin/templ generate
```

### 3. 重新构建应用

```bash
go build -o digwis-panel .
```

### 4. 重启服务器

```bash
./digwis-panel -port 9090 -debug
```

## 快速开发脚本

创建 `dev-rebuild.sh` 脚本：

```bash
#!/bin/bash
set -e

echo "🔄 重新生成模板..."
~/go/bin/templ generate

echo "🔨 重新构建应用..."
go build -o digwis-panel .

echo "🚀 启动服务器..."
./digwis-panel -port 9090 -debug
```

使用方法：

```bash
chmod +x dev-rebuild.sh
./dev-rebuild.sh
```

## 常见问题排查

### 1. 模板修改不生效

**症状**: 修改了 `.templ` 文件但页面没有变化

**解决方案**:
```bash
# 1. 检查是否重新生成了模板
ls -la internal/templates/pages/*_templ.go

# 2. 强制重新生成
find . -name "*_templ.go" -delete
~/go/bin/templ generate

# 3. 验证生成内容
grep -n "你的修改内容" internal/templates/pages/dashboard_templ.go

# 4. 重新构建
go build -o digwis-panel .
```

### 2. SSE 连接问题

**症状**: 实时数据不更新

**调试步骤**:
```bash
# 1. 测试 SSE 端点
curl -N http://localhost:9090/api/sse/stats

# 2. 检查浏览器控制台
# 3. 使用页面上的"测试SSE连接"按钮
```

### 3. 静态资源缓存

**症状**: CSS/JS 修改不生效

**解决方案**:
- 强制刷新浏览器 (Ctrl+F5)
- 修改资源版本号
- 清除浏览器缓存

## 代码规范

### 1. Templ 模板

```go
// 文件头部
package pages

import "server-panel/internal/templates/layouts"
import "server-panel/internal/templates/components"

// 模板函数
templ Dashboard(title string, username string) {
    @layouts.Base(title, username) {
        <div class="space-y-6">
            <!-- 内容 -->
        </div>
    }
}
```

### 2. JavaScript 规范

```javascript
// 使用 console.log 进行调试
console.log('🚀 功能加载中...');

// 错误处理
try {
    // 代码
} catch (error) {
    console.error('❌ 错误:', error);
}
```

### 3. CSS 类名

使用 Tailwind CSS 类名：
```html
<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
    <h3 class="text-lg font-semibold text-gray-900 mb-4">标题</h3>
</div>
```

## 测试流程

### 1. 功能测试

1. **登录测试**: 验证用户认证
2. **数据显示**: 检查系统监控数据
3. **实时更新**: 验证 SSE 连接
4. **响应式设计**: 测试不同屏幕尺寸

### 2. 浏览器兼容性

- Chrome (推荐)
- Firefox
- Safari
- Edge

### 3. 性能测试

```bash
# 检查内存使用
ps aux | grep digwis-panel

# 检查端口占用
ss -tlnp | grep :9090
```

## 部署注意事项

### 1. 生产环境

```bash
# 构建生产版本
go build -ldflags="-s -w" -o digwis-panel .

# 运行
./digwis-panel -port 8080
```

### 2. 环境变量

```bash
export DIGWIS_HOST=0.0.0.0
export DIGWIS_PORT=8080
export DIGWIS_DEBUG=false
```

### 3. 系统服务

创建 systemd 服务文件 `/etc/systemd/system/digwis-panel.service`

---

**维护者**: 开发团队  
**最后更新**: 2025-07-27  
**版本**: v1.0

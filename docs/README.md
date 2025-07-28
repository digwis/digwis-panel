# DigWis Panel 文档

本文件夹包含 DigWis Panel 项目的技术文档和问题解决记录。

## 文档列表

### 📋 开发指南

- **[development-workflow.md](./development-workflow.md)** - 完整的开发工作流程指南
  - 技术栈介绍
  - 开发环境设置
  - 标准开发流程
  - 代码规范
  - 测试和部署

### 🐛 问题解决记录

- **[templ-generation-issue.md](./templ-generation-issue.md)** - Templ 模板生成问题
  - 问题：修改 `.templ` 文件后页面不更新
  - 原因：没有重新生成 `*_templ.go` 文件
  - 解决：`templ generate` 命令使用指南

- **[sse-troubleshooting.md](./sse-troubleshooting.md)** - SSE 实时数据推送问题
  - 问题：前端显示 "--"，实时数据不更新
  - 原因：多重问题（模板生成 + 字段名不匹配 + 认证问题）
  - 解决：完整的调试和修复流程

## 快速参考

### 🔧 常用命令

```bash
# 重新生成模板（必须！）
~/go/bin/templ generate

# 构建应用
go build -o digwis-panel .

# 启动开发服务器
./digwis-panel -port 9090 -debug

# 测试 SSE 端点
curl -N http://localhost:9090/api/sse/stats
```

### 🚨 常见问题

| 问题 | 症状 | 解决方案 |
|------|------|----------|
| 模板修改不生效 | 页面没有变化 | `templ generate` |
| 实时数据不更新 | 显示 "--" | 检查 SSE 连接 |
| 静态资源缓存 | CSS/JS 不更新 | 强制刷新浏览器 |

### 📊 项目架构

```
DigWis Panel
├── 后端: Go + Chi Router
├── 前端: Tailwind CSS + HTMX + Alpine.js
├── 模板: Templ
└── 实时数据: Server-Sent Events (SSE)
```

## 开发最佳实践

### ✅ 推荐做法

1. **每次修改模板后运行 `templ generate`**
2. **使用提供的开发脚本自动化流程**
3. **在浏览器开发者工具中检查控制台输出**
4. **使用页面上的"测试SSE连接"按钮验证功能**

### ❌ 避免的错误

1. **直接修改 `*_templ.go` 文件**（会被覆盖）
2. **忘记重新生成模板就构建应用**
3. **忽略浏览器控制台的错误信息**
4. **在生产环境中启用调试模式**

## 贡献指南

### 添加新文档

1. 在 `docs/` 文件夹中创建新的 `.md` 文件
2. 使用清晰的标题和结构
3. 包含具体的代码示例
4. 更新本 README 文件的文档列表

### 更新现有文档

1. 保持文档的时效性
2. 添加解决日期和状态
3. 包含版本信息
4. 提供完整的复现步骤

## 技术支持

### 🔍 调试步骤

1. **检查服务器日志**
2. **查看浏览器控制台**
3. **验证网络连接**
4. **测试 API 端点**

### 📞 获取帮助

- 查看相关文档
- 检查 GitHub Issues
- 联系开发团队

---

**维护**: 开发团队  
**创建**: 2025-07-27  
**状态**: 活跃维护

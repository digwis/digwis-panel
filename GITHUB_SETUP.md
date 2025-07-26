# GitHub 仓库设置指南

## 1. 创建 GitHub 仓库

1. 在 GitHub 上创建新仓库，例如：`digwis-panel`
2. 设置仓库为公开（Public）以支持一键安装

## 2. 仓库目录结构

```
digwis-panel/
├── README.md                 # 项目说明
├── install.sh               # 一键安装脚本
├── LICENSE                  # 开源协议
├── .github/
│   └── workflows/
│       └── release.yml      # 自动发布工作流
├── docs/                    # 文档目录
├── scripts/                 # 脚本目录
└── src/                     # 源代码目录
    ├── main.go
    ├── go.mod
    ├── go.sum
    └── internal/
```

## 3. 设置自动发布

创建 `.github/workflows/release.yml`：

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Build
      run: |
        chmod +x scripts/release.sh
        ./scripts/release.sh
    
    - name: Create Release
      uses: softprops/action-gh-release@v1
      with:
        files: release/*
        generate_release_notes: true
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## 4. 一键安装命令

用户可以使用以下命令安装：

```bash
# 使用 curl
curl -sSL https://raw.githubusercontent.com/your-username/digwis-panel/main/install.sh | bash

# 使用 wget  
wget -qO- https://raw.githubusercontent.com/your-username/digwis-panel/main/install.sh | bash
```

## 5. 发布新版本

1. 更新版本号
2. 提交代码
3. 创建标签：`git tag v1.0.0`
4. 推送标签：`git push origin v1.0.0`
5. GitHub Actions 自动构建和发布

## 6. 需要修改的地方

### install-remote.sh
```bash
# 修改这些变量
GITHUB_REPO="your-username/digwis-panel"
PANEL_VERSION="1.0.0"
```

### release.sh
```bash
# 修改版本号
VERSION="1.0.0"
```

## 7. 测试安装

在新的服务器上测试：

```bash
# 测试下载
curl -sSL https://raw.githubusercontent.com/your-username/digwis-panel/main/install.sh

# 测试安装
curl -sSL https://raw.githubusercontent.com/your-username/digwis-panel/main/install.sh | bash
```

## 8. 推荐的 README.md 内容

```markdown
# DigWis 服务器管理面板

现代化的服务器管理面板，支持一键安装和SSL证书管理。

## 特性

- 🚀 一键安装，无需复杂配置
- 🔒 内置SSL证书管理
- 🌐 支持Let's Encrypt免费证书
- 📊 实时系统监控
- 🛠️ 环境一键安装
- 📁 文件管理
- 📝 日志查看

## 快速安装

```bash
curl -sSL https://raw.githubusercontent.com/your-username/digwis-panel/main/install.sh | bash
```

## 访问面板

安装完成后访问：http://your-server-ip:8080

## 文档

- [安装指南](docs/install.md)
- [使用手册](docs/usage.md)
- [API文档](docs/api.md)

## 支持

- [Issues](https://github.com/your-username/digwis-panel/issues)
- [Discussions](https://github.com/your-username/digwis-panel/discussions)
```

## 9. 安全考虑

1. 使用 HTTPS 下载脚本
2. 验证下载文件的完整性
3. 提供校验和文件
4. 定期更新依赖

## 10. 推广建议

1. 在 README 中添加演示截图
2. 创建详细的文档
3. 提供 Docker 版本
4. 支持多语言
5. 添加用户反馈渠道

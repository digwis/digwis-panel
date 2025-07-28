# 🚀 DigWis Panel - Go 服务器管理面板

一个基于 Go 语言开发的现代化服务器管理面板，支持嵌入式静态文件、一键部署和远程管理。

## ✨ 特性

- **🔐 系统用户认证** - 直接使用系统用户账户登录，支持sudo/wheel/admin组权限验证
- **📦 零依赖部署** - 单一二进制文件，静态文件嵌入，无需外部依赖
- **🛠️ 环境管理** - 一键安装Nginx、PHP、MySQL、Node.js、Docker等开发环境
- **📊 系统监控** - 实时监控CPU、内存、磁盘、网络等系统资源
- **🎨 现代化界面** - 基于 Templ + HTMX + Alpine.js + Tailwind CSS
- **⚡ 高性能** - Go 原生开发，内存占用低，响应速度快
- **🔒 安全可靠** - 会话管理、CSRF保护、登录限制等安全特性
- **🚀 快速部署** - 支持远程一键安装、升级和卸载

## 🚀 快速开始

### 📦 一键安装（推荐）

#### 标准安装
```bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

#### 安装选项
```bash
# 静默安装
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash -s -- --quiet

# 详细安装
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash -s -- --verbose
```

**安装特点：**
- 🚀 **预编译二进制** - 无需现场编译，安装速度快
- 📦 **嵌入式资源** - 静态文件打包在二进制中，无需外部文件
- 🌐 **智能下载** - 自动选择最快的下载节点
- 🛡️ **安全可靠** - 支持文件校验和完整性检查

### 🔄 一键升级

#### 升级到最新版本
```bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash
```

#### 升级选项
```bash
# 升级到指定版本
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash -s -- --version v1.2.0

# 静默升级
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash -s -- --quiet

# 详细升级日志
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash -s -- --verbose
```

### 🗑️ 一键卸载

#### 交互式卸载
```bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/uninstall.sh | sudo bash
```

#### 自动确认卸载
```bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/uninstall.sh | sudo bash -s -- --yes
```

### 🛠️ 手动构建

```bash
# 克隆项目
git clone https://github.com/digwis/digwis-panel.git
cd digwis-panel

# 安装依赖
go mod tidy

# 生成模板（如果有 templ）
templ generate

# 构建项目
make build

# 本地部署测试
make deploy-local
```

### 🌐 访问面板

安装完成后，打开浏览器访问：
- **本地访问**：`http://localhost:8080`
- **外网访问**：`http://YOUR_SERVER_IP:8080`

使用具有管理员权限的系统用户账户登录。

## 📋 系统要求

### 支持的系统
- Ubuntu 18.04+
- Debian 9+
- CentOS 7+
- RHEL 7+
- Fedora 30+
- Rocky Linux 8+
- AlmaLinux 8+

### 支持的架构
- x86_64 (amd64)
- ARM64 (aarch64)
- ARM (armv7l)

### 系统要求
- **权限**: 需要root权限运行
- **内存**: 最低64MB RAM
- **磁盘**: 最低100MB可用空间

## 🛠️ 管理命令

### 服务管理
```bash
# 查看状态
systemctl status digwis-panel

# 启动服务
sudo systemctl start digwis-panel

# 停止服务
sudo systemctl stop digwis-panel

# 重启服务
sudo systemctl restart digwis-panel

# 查看日志
journalctl -u digwis-panel -f

# 开机自启
sudo systemctl enable digwis-panel
```

### 防火墙配置

确保防火墙已开放8080端口：

```bash
# UFW (Ubuntu)
sudo ufw allow 8080/tcp

# firewalld (CentOS)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# iptables
sudo iptables -I INPUT -p tcp --dport 8080 -j ACCEPT
```

## 📋 使用场景

### 🆕 全新服务器部署
```bash
# 1. 一键安装
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash

# 2. 访问面板
# http://YOUR_SERVER_IP:8080
```

### 🔄 版本升级
```bash
# 1. 一键升级
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash

# 2. 验证功能
# 访问面板测试功能是否正常
```

### 🗑️ 完全卸载
```bash
# 1. 一键卸载
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/uninstall.sh | sudo bash -s -- --yes

# 2. 验证清理
# 检查是否完全清理干净
```

## 🛠️ 支持的环境

面板支持一键安装以下开发环境：

- **Nginx** - 高性能Web服务器
- **PHP** - PHP运行环境 + PHP-FPM
- **MySQL** - 关系型数据库
- **Node.js** - JavaScript运行时
- **Python** - Python编程语言
- **Docker** - 容器化平台
- **Redis** - 内存数据库
- **Git** - 版本控制系统

## 📁 目录结构

```
/opt/digwis-panel/          # 程序安装目录
├── digwis-panel            # 主程序二进制文件
├── data/                   # 应用数据目录
└── digwis-panel.backup.*   # 自动备份文件

/etc/digwis-panel/          # 配置目录
├── config.yaml             # 主配置文件

/var/log/digwis-panel/      # 日志目录
├── access.log              # 访问日志
└── error.log               # 错误日志

/etc/systemd/system/        # 系统服务
├── digwis-panel.service    # systemd服务文件
```

## ⚙️ 配置文件

主配置文件位于 `/etc/server-panel/config.yaml`：

```yaml
# 调试模式
debug: false

# 认证配置
auth:
  session_timeout: 3600     # 会话超时时间(秒)
  max_login_attempts: 5     # 最大登录尝试次数
  lockout_duration: 900     # 锁定时间(秒)

# 服务器配置
server:
  port: "8080"              # 监听端口
  read_timeout: 30          # 读取超时
  write_timeout: 30         # 写入超时
  idle_timeout: 120         # 空闲超时

# 路径配置
paths:
  data_dir: "/var/lib/server-panel"
  log_dir: "/var/log/server-panel"
  temp_dir: "/tmp/server-panel"
  backup_dir: "/var/backups/server-panel"
```

## � 安全注意事项

### 脚本安全
1. **仅从官方源下载**：确保使用官方 GitHub 仓库链接
2. **使用 HTTPS**：所有命令都使用 `https://` 协议
3. **验证脚本内容**：可以先下载脚本查看内容再执行
4. **备份重要数据**：升级前确保重要数据已备份

### 面板安全
1. **用户权限**: 只允许具有管理员权限的用户登录
2. **会话管理**: 自动会话超时和安全的会话ID生成
3. **登录保护**: 失败次数限制和IP锁定机制
4. **CSRF保护**: 防止跨站请求伪造攻击
5. **安全日志**: 记录所有认证尝试和系统操作

## 📞 获取帮助

### 查看脚本帮助
```bash
# 安装脚本帮助
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | bash -s -- --help

# 升级脚本帮助
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | bash -s -- --help

# 卸载脚本帮助
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/uninstall.sh | bash -s -- --help
```

## ⚡ 快速故障排除

### 安装失败
```bash
# 查看详细安装日志
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash -s -- --verbose

# 手动检查
systemctl status digwis-panel
journalctl -u digwis-panel -n 50
```

### 升级失败
```bash
# 查看服务状态
systemctl status digwis-panel

# 查看错误日志
journalctl -u digwis-panel -n 20

# 手动回滚（如果有备份）
sudo systemctl stop digwis-panel
sudo cp /opt/digwis-panel/digwis-panel.backup.* /opt/digwis-panel/digwis-panel
sudo systemctl start digwis-panel
```

### 服务无法启动
```bash
# 检查程序文件
ls -la /opt/digwis-panel/

# 检查权限
sudo chmod +x /opt/digwis-panel/digwis-panel

# 手动启动测试
sudo /opt/digwis-panel/digwis-panel -port 8080 -debug
```

### 无法登录
1. 确认用户在管理员组中：`groups username`
2. 检查密码是否正确
3. 查看认证日志：`journalctl -u digwis-panel -f`

### 环境安装失败
1. 检查网络连接
2. 确认系统软件包管理器正常
3. 查看安装日志获取详细错误信息

## 📝 开发说明

### 项目结构

```
server-panel/
├── main.go                 # 主程序入口
├── internal/               # 内部包
│   ├── auth/              # 认证模块
│   ├── config/            # 配置管理
│   ├── environment/       # 环境管理
│   ├── handlers/          # HTTP处理器
│   ├── middleware/        # 中间件
│   └── system/            # 系统监控
├── build.sh               # 构建脚本
├── install.sh             # 安装脚本
└── README.md              # 说明文档
```

### 本地开发

```bash
# 安装依赖
go mod tidy

# 运行程序
sudo go run main.go

# 构建
go build -o server-panel main.go
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📞 支持

如有问题，请提交Issue或联系维护者。

## 📚 相关文档

- [详细安装指南](INSTALL.md) - 完整的安装说明和故障排除
- [快速安装脚本](install-quick.sh) - 一键安装脚本
- [完整安装脚本](install-remote.sh) - 带详细日志的安装脚本

## 🔗 快速链接

### 一键命令
```bash
# 安装
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash

# 升级
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash

# 卸载
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/uninstall.sh | sudo bash -s -- --yes
```

### 开发命令
```bash
# 本地构建
make build

# 本地部署测试
make deploy-local

# 回滚
make rollback

# 发布版本
./scripts/release.sh v1.0.0 "版本说明"
```

---

**🎉 享受 DigWis Panel 带来的便捷服务器管理体验！**

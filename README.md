# 🚀 Go服务器管理面板

一个基于Go语言开发的现代化服务器管理面板，支持系统安装后直接使用，无需复杂的依赖配置。

## ✨ 特性

- **🔐 系统用户认证** - 直接使用系统用户账户登录，支持sudo/wheel/admin组权限验证
- **📦 零依赖部署** - 单一二进制文件，无需安装PHP、Python等运行时
- **🛠️ 环境管理** - 一键安装Nginx、PHP、MySQL、Node.js、Docker等开发环境
- **📊 系统监控** - 实时监控CPU、内存、磁盘、网络等系统资源
- **🎨 现代化界面** - 响应式设计，支持移动端访问
- **⚡ 高性能** - Go语言开发，内存占用低，响应速度快
- **🔒 安全可靠** - 会话管理、CSRF保护、登录限制等安全特性

## 🚀 快速开始

### 方式一：一键安装（推荐）

在你的VPS上执行以下命令即可自动安装：

```bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

或者使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

**安装选项：**

```bash
# 详细模式（显示详细安装信息）
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash -s -- --verbose

# 静默模式（最小输出）
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash -s -- --quiet
```

**安装过程：**
- ✅ 自动检测系统环境
- ✅ 安装Go语言环境
- ✅ 从GitHub拉取最新源码
- ✅ 编译并安装面板
- ✅ 配置系统服务
- ✅ 启动面板服务

### 方式二：手动构建

```bash
# 克隆项目
git clone https://github.com/digwis/digwis-panel.git
cd digwis-panel

# 构建二进制文件
chmod +x build.sh
./build.sh

# 安装为系统服务
sudo chmod +x install-remote.sh
sudo ./install-remote.sh
```

### 访问面板

安装完成后，打开浏览器访问：
- 本地访问：`http://localhost:8080`
- 外网访问：`http://your-server-ip:8080`

使用具有管理员权限的系统用户账户登录。

## 📋 系统要求

- **操作系统**: Linux (Ubuntu, Debian, CentOS, RHEL等)
- **架构**: x86_64, ARM64, ARM
- **权限**: 需要root权限运行
- **内存**: 最低64MB RAM
- **磁盘**: 最低100MB可用空间

## 🔧 管理命令

安装完成后，可以使用以下命令管理面板：

```bash
# 查看服务状态
systemctl status digwis

# 启动服务
systemctl start digwis

# 停止服务
systemctl stop digwis

# 重启服务
systemctl restart digwis

# 查看实时日志
journalctl -u digwis -f

# 开机自启
systemctl enable digwis
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
/opt/digwis/                # 程序安装目录
├── digwis-panel            # 主程序二进制文件

/etc/digwis/                # 配置目录
├── config.yaml             # 主配置文件

/var/lib/digwis/            # 数据目录
├── data/                   # 应用数据

/var/log/digwis/            # 日志目录
├── access.log              # 访问日志
└── error.log               # 错误日志

/etc/systemd/system/        # 系统服务
├── digwis.service          # systemd服务文件
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

## 🔐 安全说明

1. **用户权限**: 只允许具有管理员权限的用户登录
2. **会话管理**: 自动会话超时和安全的会话ID生成
3. **登录保护**: 失败次数限制和IP锁定机制
4. **CSRF保护**: 防止跨站请求伪造攻击
5. **安全日志**: 记录所有认证尝试和系统操作

## 🐛 故障排除

### 服务无法启动

```bash
# 查看详细日志
journalctl -u server-panel -f

# 检查配置文件
cat /etc/server-panel/config.yaml

# 检查端口占用
netstat -tlnp | grep 8080
```

### 无法登录

1. 确认用户在管理员组中：`groups username`
2. 检查密码是否正确
3. 查看认证日志：`server-panel logs`

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

```bash
# 一键安装命令
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash

# 或者使用 wget
wget -qO- https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

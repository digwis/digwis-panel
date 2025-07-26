# DigWis 面板安装指南

DigWis 是一个基于 Go 语言开发的轻量级服务器管理面板，支持多种安装方式。

## 🚀 快速安装（推荐）

### 一键安装脚本

在你的 VPS 上执行以下命令即可自动安装：

```bash
curl -sSL https://raw.githubusercontent.com/moviebluebook/digwis-panel/main/install-quick.sh | bash
```

或者使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/moviebluebook/digwis-panel/main/install-quick.sh | bash
```

### 安装过程

脚本会自动完成以下步骤：

1. ✅ 检查系统环境和权限
2. ✅ 安装必要的系统依赖（git, gcc, curl 等）
3. ✅ 自动下载并安装 Go 语言环境（1.21.5）
4. ✅ 从 GitHub 克隆最新源代码
5. ✅ 编译生成二进制文件
6. ✅ 创建配置文件和目录结构
7. ✅ 注册为系统服务
8. ✅ 启动面板服务

## 📋 系统要求

### 支持的操作系统
- Ubuntu 18.04+
- Debian 9+
- CentOS 7+
- RHEL 7+
- Fedora 30+

### 系统架构
- x86_64 (amd64)
- ARM64 (aarch64)
- ARMv7 (arm)

### 最低配置
- 内存：512MB+
- 磁盘：1GB+ 可用空间
- 网络：需要访问 GitHub 和 Go 官方源

## 🔧 手动安装

如果自动安装失败，可以手动执行以下步骤：

### 1. 安装依赖

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y git build-essential curl wget
```

**CentOS/RHEL:**
```bash
sudo yum install -y git gcc curl wget
```

### 2. 安装 Go 语言

```bash
# 下载 Go 1.21.5
wget https://golang.org/dl/go1.21.5.linux-amd64.tar.gz

# 安装
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 3. 克隆并编译

```bash
# 克隆源码
git clone https://github.com/moviebluebook/digwis-panel.git
cd digwis-panel

# 编译
go mod tidy
go build -ldflags '-s -w' -o digwis-panel main.go

# 安装
sudo mkdir -p /opt/digwis
sudo cp digwis-panel /opt/digwis/
sudo chmod +x /opt/digwis/digwis-panel
```

### 4. 创建配置

```bash
sudo mkdir -p /etc/digwis /var/lib/digwis /var/log/digwis

# 创建配置文件
sudo tee /etc/digwis/config.yaml > /dev/null << EOF
debug: false
auth:
  session_timeout: 3600
  secret_key: "$(openssl rand -hex 32)"
server:
  port: "8080"
paths:
  data_dir: "/var/lib/digwis"
  log_dir: "/var/log/digwis"
EOF
```

### 5. 创建系统服务

```bash
sudo tee /etc/systemd/system/digwis.service > /dev/null << EOF
[Unit]
Description=DigWis Server Panel
After=network.target

[Service]
Type=simple
User=root
ExecStart=/opt/digwis/digwis-panel -config /etc/digwis/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable digwis
sudo systemctl start digwis
```

## 🌐 访问面板

安装完成后，可以通过以下地址访问面板：

- 本地访问：http://localhost:8080
- 外网访问：http://YOUR_SERVER_IP:8080

## 🔒 防火墙配置

确保防火墙已开放 8080 端口：

**UFW (Ubuntu):**
```bash
sudo ufw allow 8080/tcp
```

**firewalld (CentOS):**
```bash
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

**iptables:**
```bash
sudo iptables -I INPUT -p tcp --dport 8080 -j ACCEPT
```

## 📊 管理命令

```bash
# 查看服务状态
sudo systemctl status digwis

# 启动服务
sudo systemctl start digwis

# 停止服务
sudo systemctl stop digwis

# 重启服务
sudo systemctl restart digwis

# 查看日志
sudo journalctl -u digwis -f

# 查看配置
sudo cat /etc/digwis/config.yaml
```

## 🔄 更新面板

要更新到最新版本，重新运行安装脚本即可：

```bash
curl -sSL https://raw.githubusercontent.com/moviebluebook/digwis-panel/main/install-quick.sh | bash
```

## 🗑️ 卸载面板

```bash
# 停止并删除服务
sudo systemctl stop digwis
sudo systemctl disable digwis
sudo rm -f /etc/systemd/system/digwis.service

# 删除文件
sudo rm -rf /opt/digwis
sudo rm -rf /etc/digwis
sudo rm -rf /var/lib/digwis
sudo rm -rf /var/log/digwis

# 重新加载 systemd
sudo systemctl daemon-reload
```

## ❓ 常见问题

### 1. 安装失败怎么办？

- 检查网络连接是否正常
- 确保有足够的磁盘空间
- 查看错误日志：`journalctl -u digwis -n 50`

### 2. 无法访问面板

- 检查服务是否运行：`systemctl status digwis`
- 检查防火墙设置
- 检查端口是否被占用：`netstat -tlnp | grep 8080`

### 3. 编译失败

- 检查 Go 版本是否正确：`go version`
- 检查网络是否能访问 GitHub 和 Go 模块代理
- 尝试手动编译

## 📞 技术支持

- GitHub Issues: https://github.com/digwis/digwis-panel/issues
- 项目主页: https://github.com/digwis/digwis-panel

---

**注意：** 请确保在受信任的环境中运行安装脚本，并定期更新面板以获得最新的安全补丁。

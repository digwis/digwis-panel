# 🚀 DigWis Panel 部署指南

本文档介绍如何部署 DigWis Panel 到生产环境。

## 📦 一键安装

### 方法 1：在线安装（推荐）

```bash
# 一键安装命令
curl -fsSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-digwis-panel.sh | bash
```

### 方法 2：下载后安装

```bash
# 下载安装脚本
wget https://raw.githubusercontent.com/digwis/digwis-panel/main/install-digwis-panel.sh

# 设置执行权限
chmod +x install-digwis-panel.sh

# 运行安装
./install-digwis-panel.sh
```

## 🛠️ 开发者部署

如果你是开发者，想要部署自己的版本：

### 1. 准备部署

```bash
# 克隆项目
git clone https://github.com/digwis/digwis-panel.git
cd digwis-panel

# 检查状态
make git-status
```

### 2. 部署到远程仓库

```bash
# 完整部署（推送代码 + 生成安装脚本）
make deploy

# 或者分步执行
make push           # 仅推送代码
make install-script # 仅生成安装脚本
```

### 3. 使用部署脚本

```bash
# 完整部署
./scripts/deploy.sh

# 仅生成安装脚本
./scripts/deploy.sh --install-script-only

# 强制部署（忽略未提交更改）
./scripts/deploy.sh --force

# 查看帮助
./scripts/deploy.sh --help
```

## 🔧 系统要求

### 最低要求

- **操作系统**: Linux (Ubuntu 18.04+, CentOS 7+, Debian 9+)
- **内存**: 512MB RAM
- **存储**: 1GB 可用空间
- **网络**: 互联网连接

### 推荐配置

- **操作系统**: Ubuntu 22.04 LTS
- **内存**: 2GB RAM
- **存储**: 10GB 可用空间
- **CPU**: 2 核心

## 📋 安装过程

安装脚本会自动执行以下步骤：

1. **检查系统要求**
   - 验证操作系统
   - 检查权限

2. **安装依赖**
   - 更新包管理器
   - 安装 Git, Curl, Wget
   - 安装 Go 语言环境

3. **下载项目**
   - 克隆 Git 仓库
   - 设置文件权限

4. **构建项目**
   - 下载 Go 依赖
   - 编译二进制文件

5. **配置服务**
   - 创建 systemd 服务
   - 配置自动启动

6. **配置防火墙**
   - 开放必要端口
   - 配置安全规则

7. **启动服务**
   - 启动 DigWis Panel
   - 验证运行状态

## 🔐 安全配置

### 防火墙设置

```bash
# UFW (Ubuntu)
sudo ufw allow 9091/tcp

# Firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-port=9091/tcp
sudo firewall-cmd --reload
```

### SSL/TLS 配置

建议使用反向代理（如 Nginx）来提供 HTTPS：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://127.0.0.1:9091;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 📊 服务管理

### 基本命令

```bash
# 查看服务状态
sudo systemctl status digwis-panel

# 启动服务
sudo systemctl start digwis-panel

# 停止服务
sudo systemctl stop digwis-panel

# 重启服务
sudo systemctl restart digwis-panel

# 查看日志
sudo journalctl -u digwis-panel -f

# 开机自启
sudo systemctl enable digwis-panel
```

### 配置文件位置

- **安装目录**: `/opt/digwis-panel/`
- **服务文件**: `/etc/systemd/system/digwis-panel.service`
- **日志文件**: `journalctl -u digwis-panel`

## 🗑️ 卸载

### 使用卸载脚本

```bash
# 下载卸载脚本
wget https://raw.githubusercontent.com/digwis/digwis-panel/main/uninstall-digwis-panel.sh

# 运行卸载
chmod +x uninstall-digwis-panel.sh
./uninstall-digwis-panel.sh
```

### 手动卸载

```bash
# 停止服务
sudo systemctl stop digwis-panel
sudo systemctl disable digwis-panel

# 删除服务文件
sudo rm /etc/systemd/system/digwis-panel.service
sudo systemctl daemon-reload

# 删除安装目录
sudo rm -rf /opt/digwis-panel
```

## 🔧 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   sudo lsof -i :9091
   sudo netstat -tlnp | grep 9091
   ```

2. **权限问题**
   ```bash
   sudo chown -R root:root /opt/digwis-panel
   sudo chmod +x /opt/digwis-panel/digwis-panel
   ```

3. **服务启动失败**
   ```bash
   sudo journalctl -u digwis-panel --no-pager
   sudo systemctl status digwis-panel
   ```

### 获取帮助

- **GitHub Issues**: https://github.com/digwis/digwis-panel/issues
- **文档**: https://github.com/digwis/digwis-panel/wiki
- **邮件**: support@digwis.com

## 📈 更新

### 自动更新

```bash
# 重新运行安装脚本即可更新
curl -fsSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-digwis-panel.sh | bash
```

### 手动更新

```bash
cd /opt/digwis-panel
sudo git pull origin main
sudo go build -o digwis-panel .
sudo systemctl restart digwis-panel
```

# DigWis 面板部署脚本说明

本项目提供了多个部署脚本，用于在VPS上快速安装和部署DigWis服务器管理面板。

## 📁 脚本文件说明

### 1. `install-quick.sh` - 快速安装脚本（推荐）

**特点：**
- 🚀 快速简洁，适合生产环境
- 📦 自动安装Go环境
- 🔄 从GitHub拉取最新源码
- ⚡ 自动编译和配置
- 🎯 一键完成所有安装步骤

**使用方法：**
```bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-quick.sh | sudo bash
```

### 2. `install-remote.sh` - 完整安装脚本

**特点：**
- 📋 详细的安装日志和进度显示
- 🔍 完整的系统检查和验证
- 🛡️ 更多的安全检查和错误处理
- 🎨 美观的界面和Logo显示
- 📊 详细的安装结果展示

**使用方法：**
```bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-remote.sh | sudo bash
```

### 3. `verify-install.sh` - 安装验证脚本

**功能：**
- ✅ 验证二进制文件是否正确安装
- ✅ 检查配置文件和目录结构
- ✅ 验证系统服务状态
- ✅ 测试HTTP服务响应
- ✅ 显示访问地址和管理命令

**使用方法：**
```bash
./verify-install.sh
```

### 4. `test-install.sh` - 脚本测试工具

**功能：**
- 🧪 检查安装脚本语法
- 🔍 验证关键函数和变量
- 📋 测试架构检测逻辑
- 🌐 验证GitHub仓库地址

**使用方法：**
```bash
./test-install.sh
```

## 🚀 推荐部署流程

### 生产环境部署

1. **快速安装（推荐）**
   ```bash
   curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-quick.sh | sudo bash
   ```

2. **验证安装**
   ```bash
   curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/verify-install.sh | bash
   ```

3. **访问面板**
   ```
   http://YOUR_SERVER_IP:8080
   ```

### 开发环境部署

1. **克隆仓库**
   ```bash
   git clone https://github.com/digwis/digwis-panel.git
   cd digwis-panel
   ```

2. **测试脚本**
   ```bash
   ./test-install.sh
   ```

3. **本地安装**
   ```bash
   sudo ./install-quick.sh
   ```

4. **验证安装**
   ```bash
   ./verify-install.sh
   ```

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
- **内存**: 512MB+
- **磁盘**: 1GB+ 可用空间
- **权限**: root权限
- **网络**: 能访问GitHub和Go官方源

## 🔧 安装后管理

### 系统服务管理
```bash
# 查看状态
systemctl status digwis

# 启动/停止/重启
systemctl start digwis
systemctl stop digwis
systemctl restart digwis

# 开机自启
systemctl enable digwis
systemctl disable digwis
```

### 日志查看
```bash
# 实时日志
journalctl -u digwis -f

# 最近日志
journalctl -u digwis -n 50

# 错误日志
journalctl -u digwis -p err
```

### 配置文件
```bash
# 主配置文件
/etc/digwis/config.yaml

# 编辑配置
sudo nano /etc/digwis/config.yaml

# 重启生效
sudo systemctl restart digwis
```

## 🔒 安全配置

### 防火墙设置
```bash
# UFW (Ubuntu)
sudo ufw allow 8080/tcp

# firewalld (CentOS)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# iptables
sudo iptables -I INPUT -p tcp --dport 8080 -j ACCEPT
```

### SSL证书（可选）
面板支持自动申请Let's Encrypt证书，登录后在设置页面配置。

## 🔄 更新和维护

### 更新面板
重新运行安装脚本即可更新到最新版本：
```bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-quick.sh | bash
```

### 备份配置
```bash
# 备份配置文件
sudo cp /etc/digwis/config.yaml /etc/digwis/config.yaml.bak

# 备份数据目录
sudo tar -czf digwis-backup-$(date +%Y%m%d).tar.gz /var/lib/digwis
```

### 卸载面板
```bash
# 停止服务
sudo systemctl stop digwis
sudo systemctl disable digwis

# 删除文件
sudo rm -rf /opt/digwis
sudo rm -rf /etc/digwis
sudo rm -rf /var/lib/digwis
sudo rm -rf /var/log/digwis
sudo rm -f /etc/systemd/system/digwis.service

# 重新加载systemd
sudo systemctl daemon-reload
```

## ❓ 故障排除

### 常见问题

1. **安装失败**
   - 检查网络连接
   - 确保有root权限
   - 查看错误日志

2. **服务启动失败**
   ```bash
   journalctl -u digwis -n 20
   ```

3. **无法访问面板**
   - 检查服务状态
   - 检查防火墙设置
   - 检查端口占用

4. **编译失败**
   - 检查Go版本
   - 检查网络连接
   - 清理缓存重试

### 获取帮助

- GitHub Issues: https://github.com/digwis/digwis-panel/issues
- 详细文档: [INSTALL.md](INSTALL.md)

---

**注意**: 所有脚本都需要root权限运行，请确保在可信环境中执行。

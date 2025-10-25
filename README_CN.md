# DigWis Panel - 一键安装

[English](README.md) | 简体中文

DigWis Panel 是一个现代化的服务器管理面板，支持网站管理、数据库管理、SSL证书、备份等功能。

## 快速安装

### 一键安装命令

```bash
curl -fsSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

或使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

### 系统要求

- **操作系统**: 
  - **优先推荐**: Ubuntu 18.04+ 或 Debian 10+（兼容性和支持最佳）
  - 同样支持: CentOS 7+、RHEL 7+、Fedora 30+
- **架构**: x86_64 (AMD64)
- **内存**: 至少 512MB RAM
- **磁盘**: 至少 100MB 可用空间
- **权限**: root 或 sudo 访问权限

> 💡 **提示**: 我们推荐使用 Ubuntu 20.04 LTS 或 Debian 11 以获得最佳体验和稳定性。

### 安装后

安装完成后，脚本会显示：
- 面板访问地址
- 登录用户名
- 登录密码
- 访问端口

**重要**: 请妥善保管登录信息！

## 管理命令

```bash
# 查看服务状态
sudo systemctl status digwis-panel

# 启动服务
sudo systemctl start digwis-panel

# 停止服务
sudo systemctl stop digwis-panel

# 重启服务
sudo systemctl restart digwis-panel

# 查看实时日志
sudo journalctl -u digwis-panel -f

# 查看登录信息
sudo cat /etc/digwis-panel/install.conf
```

## 功能特性

- 🌐 **网站管理**: 创建和管理多个网站
- 🗄️ **数据库管理**: MySQL/MariaDB 数据库管理
- 🔒 **SSL证书**: Let's Encrypt 自动证书申请
- 📦 **备份恢复**: 自动备份和一键恢复
- 🔧 **系统监控**: 实时监控服务器状态
- 🤖 **AI助手**: 集成AI辅助管理功能
- 🔐 **安全加固**: 防火墙、SSH安全配置

## 卸载

```bash
sudo systemctl stop digwis-panel
sudo systemctl disable digwis-panel
sudo rm -rf /opt/digwis-panel
sudo rm -f /etc/systemd/system/digwis-panel.service
sudo rm -rf /etc/digwis-panel
sudo rm -rf /var/log/digwis-panel
sudo systemctl daemon-reload
```

## 常见问题

### 1. 忘记登录密码怎么办？

```bash
sudo cat /etc/digwis-panel/install.conf
```

### 2. 如何修改访问端口？

编辑服务配置文件：
```bash
sudo nano /etc/systemd/system/digwis-panel.service
```

修改 `ExecStart` 行的 `-port` 参数，然后重启服务：
```bash
sudo systemctl daemon-reload
sudo systemctl restart digwis-panel
```

### 3. 防火墙配置

确保开放面板端口（查看 `/etc/digwis-panel/install.conf` 获取端口号）：

**firewalld:**
```bash
sudo firewall-cmd --permanent --add-port=端口号/tcp
sudo firewall-cmd --reload
```

**ufw:**
```bash
sudo ufw allow 端口号/tcp
```

### 4. 更新面板

重新运行安装命令即可更新到最新版本，不会影响现有配置和数据：
```bash
curl -fsSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

## 支持

- GitHub: https://github.com/digwis/digwis-panel
- 问题反馈: https://github.com/digwis/digwis-panel/issues

---

© 2025 DigWis Panel. 保留所有权利。

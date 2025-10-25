# DigWis Panel - One-Click Installation

English | [简体中文](README_CN.md)

DigWis Panel is a modern server management panel with support for website management, database administration, SSL certificates, automated backups, and more.

## Quick Installation

### One-Click Installation

```bash
curl -fsSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

Or using wget:

```bash
wget -qO- https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

### System Requirements

- **Operating System**: 
  - **Recommended**: Ubuntu 18.04+ or Debian 10+ (best compatibility and support)
  - Also supported: CentOS 7+, RHEL 7+, Fedora 30+
- **Architecture**: x86_64 (AMD64)
- **Memory**: At least 512MB RAM
- **Disk Space**: At least 100MB available
- **Permissions**: root or sudo access

> 💡 **Tip**: We recommend using Ubuntu 20.04 LTS or Debian 11 for the best experience and stability.

### After Installation

The installation script will display:
- Panel access URL
- Login username
- Login password
- Access port

**Important**: Keep your login credentials safe!

## Management Commands

```bash
# Check service status
sudo systemctl status digwis-panel

# Start service
sudo systemctl start digwis-panel

# Stop service
sudo systemctl stop digwis-panel

# Restart service
sudo systemctl restart digwis-panel

# View real-time logs
sudo journalctl -u digwis-panel -f

# View login credentials
sudo cat /etc/digwis-panel/install.conf
```

## Features

- 🌐 **Website Management**: Create and manage multiple websites
- 🗄️ **Database Management**: MySQL/MariaDB database administration
- 🔒 **SSL Certificates**: Automatic Let's Encrypt certificate issuance
- 📦 **Backup & Restore**: Automated backups and one-click recovery
- 🔧 **System Monitoring**: Real-time server status monitoring
- 🤖 **AI Assistant**: Integrated AI-powered management assistance
- 🔐 **Security Hardening**: Firewall and SSH security configuration

## Uninstallation

```bash
sudo systemctl stop digwis-panel
sudo systemctl disable digwis-panel
sudo rm -rf /opt/digwis-panel
sudo rm -f /etc/systemd/system/digwis-panel.service
sudo rm -rf /etc/digwis-panel
sudo rm -rf /var/log/digwis-panel
sudo systemctl daemon-reload
```

## FAQ

### 1. Forgot your login password?

```bash
sudo cat /etc/digwis-panel/install.conf
```

### 2. How to change the access port?

Edit the service configuration file:
```bash
sudo nano /etc/systemd/system/digwis-panel.service
```

Modify the `-port` parameter in the `ExecStart` line, then restart:
```bash
sudo systemctl daemon-reload
sudo systemctl restart digwis-panel
```

### 3. Firewall Configuration

Make sure to open the panel port (check `/etc/digwis-panel/install.conf` for the port number):

**firewalld:**
```bash
sudo firewall-cmd --permanent --add-port=PORT_NUMBER/tcp
sudo firewall-cmd --reload
```

**ufw:**
```bash
sudo ufw allow PORT_NUMBER/tcp
```

### 4. Update Panel

Simply run the installation command again to update to the latest version without affecting existing configurations:
```bash
curl -fsSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
```

## Support

- GitHub: https://github.com/digwis/digwis-panel
- Issues: https://github.com/digwis/digwis-panel/issues
- Website: https://digwis.github.io/digwis-panel

---

© 2025 DigWis Panel. All rights reserved.

# DigWis Panel Scripts

这个目录包含了 DigWis Panel 项目的各种脚本文件。

## 📁 目录结构

```
scripts/
├── build/              # 构建相关脚本
│   └── build-release.sh    # 多平台发布构建脚本
├── install/            # 安装相关脚本
│   └── install.sh          # 一键安装脚本
└── management/         # 管理工具脚本
    └── digwis              # 面板管理工具
```

## 🚀 使用方法

### 构建脚本
```bash
# 构建发布包
./scripts/build/build-release.sh

# 或者使用 Makefile
make release
```

### 安装脚本
```bash
# 一键安装
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/scripts/install/install.sh | sudo bash

# 或者本地安装
sudo ./scripts/install/install.sh
```

### 管理工具
```bash
# 面板管理（需要先安装到系统）
./scripts/management/digwis

# 或者直接使用系统命令（安装后）
digwis
```

## 📋 脚本说明

### build-release.sh
- **用途**: 构建多平台发布包
- **支持平台**: linux/amd64, linux/arm64, linux/arm
- **输出**: releases/ 目录下的压缩包

### install.sh
- **用途**: 一键安装 DigWis Panel
- **支持系统**: Ubuntu/Debian/CentOS/RHEL/Fedora
- **功能**: 自动检测系统、下载、安装、配置服务

### digwis
- **用途**: 面板管理工具
- **功能**: 安装、卸载、启动、停止、查看状态、查看日志
- **模式**: 支持命令行参数和交互式菜单

## 🔧 开发说明

- 所有脚本都应该放在对应的子目录中
- 脚本应该有可执行权限：`chmod +x script_name.sh`
- 新增脚本时请更新此 README 文件

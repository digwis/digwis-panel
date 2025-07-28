# DigWis Panel 数据存储策略

## 📍 存储位置

### 🔧 开发环境
```bash
# 项目根目录下的 data 文件夹
/path/to/digwis-panel/data/digwis-panel.db

# 检测条件：
- 存在 .air.toml, go.mod, package.json 等开发文件
- 或设置 DIGWIS_MODE=development
```

### 🏭 生产环境（一键安装后）
```bash
# 首选位置（推荐）
/opt/digwis-panel/data/digwis-panel.db

# 备选位置（按优先级）
/var/lib/digwis-panel/digwis-panel.db
/usr/local/var/digwis-panel/digwis-panel.db
/etc/digwis-panel/data/digwis-panel.db
~/.digwis-panel/digwis-panel.db

# 检测条件：
- 没有开发文件标识
- 或设置 DIGWIS_MODE=production
- 按优先级尝试创建目录并测试写权限
```

## 🗄️ 数据库结构

### SQLite 统一数据库
```sql
-- 用户会话表
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    expires DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 系统配置表
CREATE TABLE system_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 操作日志表
CREATE TABLE operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    action TEXT NOT NULL,
    resource TEXT,
    details TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 🚀 一键安装脚本配置

### 安装目录结构
```bash
/opt/digwis-panel/
├── digwis-panel           # 主程序
├── data/                  # 数据目录
│   ├── digwis-panel.db   # 主数据库
│   ├── digwis-panel.db-wal
│   ├── digwis-panel.db-shm
│   └── backups/          # 备份目录
├── logs/                 # 日志文件
└── static/               # 静态资源

/etc/digwis-panel/
└── config.yaml           # 配置文件

/var/log/digwis-panel/     # 系统日志
```

### systemd 服务配置
```ini
[Unit]
Description=DigWis Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/digwis-panel
Environment=DIGWIS_MODE=production
Environment=DIGWIS_DATA_DIR=/opt/digwis-panel/data
ExecStart=/opt/digwis-panel/digwis-panel -config /etc/digwis-panel/config.yaml -port 8080
Restart=always
RestartSec=5
```

## 🔧 环境变量控制

### 开发环境
```bash
export DIGWIS_MODE=development
export DIGWIS_DATA_DIR="/custom/path"  # 可选：自定义数据目录
```

### 生产环境
```bash
export DIGWIS_MODE=production
export DIGWIS_DATA_DIR="/opt/digwis-panel/data"  # 一键安装脚本会设置
```

## 💡 与主流面板对比

| 面板 | 数据库 | 存储位置 | 特点 |
|------|--------|----------|------|
| **宝塔面板** | SQLite | `/www/server/panel/data/` | 应用目录存储 |
| **1Panel** | SQLite | `/opt/1panel/db/` | 标准系统位置 |
| **Webmin** | 文件 | `/etc/webmin/` | 配置目录存储 |
| **DigWis Panel** | SQLite | `/opt/digwis-panel/data/` | 标准位置 + 智能备选 |

## ✅ 优势特点

### 1. **零配置**
- 无需安装额外数据库服务（MySQL、PostgreSQL、Redis）
- 开箱即用，自动初始化

### 2. **智能选择**
- 自动检测开发/生产环境
- 智能选择最佳存储位置
- 权限检测和备选方案

### 3. **数据持久化**
- 重启服务器不丢失会话
- 支持 WAL 模式，提高并发性能
- 自动清理过期数据

### 4. **备份简单**
- 单个文件包含所有数据
- 支持热备份（VACUUM INTO）
- 易于迁移和恢复

### 5. **开发友好**
- 开发和生产环境分离
- 支持热重载时数据保持
- 便于调试和测试

## 🔄 数据迁移

### 从开发到生产
```bash
# 1. 停止开发服务器
# 2. 复制数据库文件
cp /path/to/dev/data/digwis-panel.db /opt/digwis-panel/data/
# 3. 设置权限
chown root:root /opt/digwis-panel/data/digwis-panel.db
chmod 644 /opt/digwis-panel/data/digwis-panel.db
```

### 备份和恢复
```bash
# 备份
cp /opt/digwis-panel/data/digwis-panel.db /backup/location/

# 恢复
systemctl stop digwis-panel
cp /backup/location/digwis-panel.db /opt/digwis-panel/data/
systemctl start digwis-panel
```

## 🛡️ 安全考虑

### 文件权限
```bash
# 数据目录权限
chmod 750 /opt/digwis-panel/data/
chown root:root /opt/digwis-panel/data/

# 数据库文件权限
chmod 644 /opt/digwis-panel/data/digwis-panel.db
chown root:root /opt/digwis-panel/data/digwis-panel.db
```

### 访问控制
- 只有 root 用户可以访问数据目录
- 数据库文件不可被普通用户读取
- 会话数据加密存储（JSON 格式）

## 📊 性能优化

### SQLite 配置
- 启用 WAL 模式提高并发性能
- 设置合理的连接池参数
- 定期清理过期数据
- 支持热备份不影响服务

### 监控指标
- 数据库文件大小
- 活跃会话数量
- 查询响应时间
- 磁盘空间使用

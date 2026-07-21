# DineTogether

一个基于 Go + Gin + SQLite 的多人聚餐点餐系统。

## 功能

- 用户注册/登录，基于 Session 的认证
- 管理员管理菜品（CRUD）、Party（CRUD）、用户（CRUD）
- 用户加入/离开 Party，提交/删除订单
- 基于"精力值"的 Party 点餐机制
- 菜品图片上传/预览
- CSRF 防护、登录频率限制

## 技术栈

- **后端**: Go 1.24 + Gin 框架
- **数据库**: SQLite (go-sqlite3)
- **前端**: HTML + Tailwind CSS（CDN）
- **Session**: Cookie-based session store
- **密码**: bcrypt 加密

## 快速启动

```bash
# 确保已安装 Go 1.24+
# 确保已安装 C 编译器（go-sqlite3 需要 CGO）
# 如 Windows 无 CGO，可安装 TDM-GCC 或使用 WSL

# 克隆项目
git clone <repo-url> && cd DineTogether

# 确保 CGO 已启用
set CGO_ENABLED=1

# 运行
go run main.go

# 或使用 Docker（端口 8081）
docker compose up -d --build
```

访问 http://localhost:8081

## 项目结构

```
DineTogether/
├── main.go                 # 入口，路由注册，数据库迁移
├── config.yaml             # 数据库路径、Session 密钥
├── schema.sql              # 数据库结构定义
├── handlers/               # 业务逻辑处理
│   ├── auth.go             # 登录/注册/中间件
│   ├── user.go             # 用户 CRUD
│   ├── menu.go             # 菜品 CRUD
│   ├── party.go            # Party CRUD + 加入/离开
│   ├── order.go            # 点餐/删除订单
│   ├── party_orders.go     # 订单列表
│   ├── image.go            # 图片上传/删除
│   └── response.go         # 统一响应格式
├── middleware/
│   ├── csrf.go             # CSRF 防护
│   ├── ratelimit.go        # 速率限制
│   └── error_handler.go    # 全局错误处理
├── models/
│   └── models.go           # 数据模型
├── templates/              # HTML 模板
├── static/
│   ├── utils.js            # 前端工具函数
│   └── uploads/            # 菜品图片
└── db/                     # SQLite 数据库文件（自动创建）
```

## API 接口

### 公开接口
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /register | 用户注册 |
| POST | /login | 用户登录 |
| POST | /logout | 退出登录 |
| GET  | /api/csrf-token | 获取 CSRF Token |
| GET  | /menus | 菜品列表 |
| GET  | /menu/:id | 菜品详情 |
| GET  | /api/party | 当前用户 Party 信息 |
| GET  | /api/party-orders | Party 订单列表 |
| POST | /order | 提交订单 |
| DELETE | /order/:id | 删除订单 |
| POST | /join-party | 加入 Party |
| POST | /leave-party | 离开 Party |
| POST | /change-password | 修改密码 |

### 管理员接口（需 Session）
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /upload-image | 上传图片 |
| POST | /delete-image | 删除图片 |
| GET/POST | /menus | 菜品管理 |
| PUT/DELETE | /menu/:id | 菜品管理 |
| GET/POST | /parties | Party 管理 |
| PUT/DELETE | /party/:id | Party 管理 |
| GET/POST | /users | 用户管理 |
| PUT/DELETE | /user/:id | 用户管理 |

## 安全性

- 密码使用 bcrypt 加密存储
- Session 使用随机密钥签名
- CSRF Token 防护（除登录/注册外所有 POST/PUT/DELETE）
- 登录接口速率限制（每分钟 10 次）
- Session Cookie 设置 HttpOnly + SameSite=Lax
- CORS 限制为本地开发域名

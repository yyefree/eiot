<div align="center">

# EIOT 物联网综合管理平台

轻量化、高性能、响应式的物联网设备综合管理平台

提供从 **设备接入 → 数据采集 → 状态监控 → 控制下发 → 设备共享 → 移动端管理** 的完整闭环

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3.4-42b883?logo=vue.js&logoColor=white)
![Flutter](https://img.shields.io/badge/Flutter-3.11-02569B?logo=flutter&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?logo=mysql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?logo=redis&logoColor=white)
![EMQX](https://img.shields.io/badge/EMQX-5.7-6F02B5?logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)

</div>

---

## 功能特性

- **产品即物模型** — 一个产品 = 一套物模型 + 一套移动端 UI 模板，结构极简
- **飞燕标准兼容** — 完整对接阿里云飞燕 Topic 体系与 Alink JSON 格式
- **严格多租户** — 用户数据完全隔离，仅可通过设备共享机制跨用户授权
- **全响应式** — Web 管理后台适配桌面 / 平板 / 手机三端
- **可视化拖拽** — 产品级移动端 UI 通过拖拽方式生成，无需编写前端代码
- **云原生部署** — Docker Compose 一键拉起，Alpine 基础镜像，秒级启动
- **跨平台移动端** — Flutter 构建，一套代码支持 Android/iOS

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| **后端** | Go + Gin + GORM | 高性能 HTTP 路由 + ORM |
| **数据库** | MySQL 8.0 | 用户/产品/设备等结构化数据 |
| **缓存** | Redis 7.2 | 设备在线状态、验证码、限流 |
| **MQTT** | EMQX 5.7 | 设备消息接入/下发 |
| **鉴权** | JWT (HS256) + bcrypt | 无状态会话 + 密码哈希 |
| **Web 前端** | Vue 3 + Vite + Element Plus | 响应式管理后台 |
| **图表** | ECharts 5 | 温湿度/能耗趋势曲线 |
| **移动端** | Flutter + Provider | 跨平台 App |
| **部署** | Docker Compose + Nginx | 一键容器化部署 |

## 架构图

```
┌──────────────────────────────────────────────────────────┐
│              移动端 App (Flutter)                          │
│      Android/iOS，/mobile 目录                            │
└────────────────────────────┬─────────────────────────────┘
                             │ HTTP REST
┌────────────────────────────▼─────────────────────────────┐
│              前端管理台 (Web)                              │
│   Vue3 + Vite + Element Plus + ECharts，/frontend 目录   │
└────────────────────────────┬─────────────────────────────┘
                             │ RESTful API (:8080)
┌────────────────────────────▼─────────────────────────────┐
│              后端服务 (Go + Gin)                           │
│   Gin 路由 / GORM / JWT，/backend 目录                    │
└──────────────┬─────────────┬─────────────┬───────────────┘
               │             │             │
        ┌──────▼─────┐ ┌─────▼──────┐ ┌────▼─────┐
        │   MySQL     │ │   Redis    │ │  EMQX    │
        │  业务数据   │ │ 缓存/验证码 │ │ MQTT消息 │
        └────────────┘ └────────────┘ └──────────┘
```

## 快速开始

### 环境要求

- Docker + Docker Compose
- Go 1.22+ （仅构建后端时需要）
- Node.js 18+ （仅构建前端时需要）
- Flutter 3.11+ （仅构建移动端时需要）

### 一键部署（Docker）

```bash
# 1. 克隆仓库
git clone https://github.com/yyefree/eiot.git
cd eiot

# 2. 构建后端二进制（Linux/amd64）
cd backend
$env:GOOS="linux"; $env:GOARCH="amd64"   # Windows PowerShell
go build -o ../deploy/eiot-server ./cmd/api-server
cd ..

# 3. 构建前端
cd frontend
npm install && npm run build
cd ..

# 4. 启动所有服务
cd deploy
docker compose up -d --build
```

### 访问入口

| 服务 | 地址 | 凭证 |
|------|------|------|
| Web 管理台 | http://localhost:8088 | `13800000000` / `admin123` |
| 后端 API | http://localhost:8080 | 同上 |
| EMQX 控制台 | http://localhost:18083 | `admin` / `public` |

### 演示账号

| 手机号 | 密码 | 角色 |
|--------|------|------|
| `13800000000` | `admin123` | 管理员 |
| `13900000001` | `123456` | 普通用户 |

首次启动时后端会自动执行种子数据（5个产品 + 50台设备 + 5个用户 + 20条共享记录）。

## 项目结构

```
eiot/
├── backend/                    # Go 后端
│   ├── api/config.yaml         # 配置文件
│   ├── cmd/api-server/         # 程序入口
│   ├── internal/
│   │   ├── dao/                # 数据访问层
│   │   ├── handler/            # HTTP 请求处理
│   │   ├── logic/              # 业务逻辑
│   │   ├── model/              # 数据模型
│   │   └── svc/                # 服务上下文
│   └── pkg/
│       ├── cache/              # Redis 封装
│       ├── config/             # 配置解析
│       ├── middleware/         # JWT 鉴权
│       ├── mqtt/               # EMQX 客户端
│       ├── tsdb/               # 时序数据库
│       └── util/               # 工具函数
│
├── frontend/                   # Vue3 前端
│   ├── src/
│   │   ├── router/             # 路由配置
│   │   ├── utils/              # Axios 封装
│   │   └── views/              # 页面组件
│   │       ├── Login.vue       # 登录
│   │       ├── Layout.vue      # 响应式布局
│   │       ├── Dashboard.vue   # 看板
│   │       ├── Project.vue     # 项目管理
│   │       ├── Product.vue     # 产品+物模型
│   │       ├── Device.vue      # 设备列表
│   │       ├── DeviceDetail.vue# 设备详情+控制
│   │       ├── MobileUI.vue    # 拖拽式UI配置
│   │       ├── Share.vue       # 设备共享
│   │       └── Users.vue       # 用户管理
│   └── vite.config.js
│
├── mobile/                     # Flutter 移动端
│   ├── lib/
│   │   ├── api/                # HTTP 客户端
│   │   ├── models/             # 数据模型
│   │   ├── pages/              # 页面
│   │   │   ├── login_page.dart
│   │   │   ├── device_list_page.dart
│   │   │   ├── device_detail_page.dart
│   │   │   ├── share_page.dart
│   │   │   └── profile_page.dart
│   │   ├── providers/          # 状态管理
│   │   └── main.dart
│   └── android/                # Android 原生工程
│
├── deploy/                     # Docker 部署
│   ├── docker-compose.yml
│   ├── Dockerfile.backend
│   ├── Dockerfile.frontend
│   ├── nginx.conf
│   └── eiot-server             # 后端二进制
│
├── 功能说明文档.md
└── 物联网综合管理平台需求文档.md
```

## 核心功能

### 设备模型（飞燕三级层级）

**项目 (Project) → 产品 (Product) → 设备 (Device)**

产品本身就是物模型，包含属性、事件、服务三要素：

```json
{
  "identifier": "temp_01",
  "name": "温度",
  "accessMode": "r",
  "dataType": {
    "type": "float",
    "specs": { "min": -40, "max": 85, "step": 0.1, "unit": "℃" }
  }
}
```

### MQTT Topic（飞燕标准）

| 方向 | Topic | 说明 |
|------|-------|------|
| 设备上报属性 | `/sys/{ProductKey}/{DeviceName}/prop/post` | 设备主动推送数据 |
| 云端下发控制 | `/sys/{ProductKey}/{DeviceName}/prop/set` | 云端控制设备 |
| 固件版本上报 | `/ota/device/inform/{ProductKey}/{DeviceName}` | 设备上报版本 |

### API 概览

**认证接口**

| 方法 | 路由 | 说明 |
|------|------|------|
| POST | `/api/auth/login` | 密码登录 |
| POST | `/api/auth/send-code` | 发送验证码 |
| POST | `/api/auth/login-code` | 验证码登录 |
| GET | `/api/user/info` | 用户信息 |

**管理员接口** (`/api/admin/*`)

| 方法 | 路由 | 说明 |
|------|------|------|
| GET/POST | `/api/admin/project` | 项目管理 |
| GET/POST | `/api/admin/product` | 产品管理 |
| PUT | `/api/admin/product/:id/thing-model` | 物模型配置 |
| GET/PUT | `/api/admin/product/:id/mobile-ui` | 移动端UI配置 |
| POST | `/api/admin/device/batch` | 批量生成设备 |
| GET | `/api/admin/stats` | 看板统计 |

**用户接口** (`/api/device/*`)

| 方法 | 路由 | 说明 |
|------|------|------|
| POST | `/api/device/bind` | 绑定设备 |
| GET | `/api/device` | 设备列表 |
| GET | `/api/device/:id` | 设备详情 |
| POST | `/api/device/:id/control` | 控制下发 |
| POST | `/api/device/share` | 设备共享 |

完整 API 文档见 [功能说明文档.md](功能说明文档.md)。

## 移动端 App

Flutter 跨平台移动端，支持：

- 密码登录 / 验证码登录
- 设备列表（分页加载 + 下拉刷新）
- 设备详情（实时数据 + 控制面板）
- 设备共享管理
- 个人中心

### 构建 APK

```bash
cd mobile
flutter pub get
flutter build apk --release
# 产物：build/app/outputs/flutter-apk/app-release.apk
```

## 响应式设计

| 断点 | 场景 | 布局策略 |
|------|------|---------|
| < 768px | 手机竖屏 | 汉堡菜单 + 卡片列表 |
| 768–1200px | 平板 | 侧边栏固定 + 两列栅格 |
| ≥ 1200px | 桌面 | 多列布局 + 完整表格 |

## 关键设计决策

| 决策 | 理由 |
|------|------|
| 产品即物模型 | 简化结构，避免 N+1 关联 |
| JWT 无状态会话 | 便于横向扩展 |
| 多租户数据隔离 | 用户仅可见自有 + 共享设备 |
| 拖拽式 UI 生成 | 新增设备无需改 App 代码 |
| 种子数据幂等 | 重复启动不报错 |
| Alpine + 静态编译 | 镜像 < 20MB，秒级启动 |

## 开发

### 后端开发

```bash
cd backend
go run ./cmd/api-server
```

### 前端开发

```bash
cd frontend
npm install
npm run dev    # 启动开发服务器 (热更新)
```

### 移动端开发

```bash
cd mobile
flutter pub get
flutter run
```

## License

MIT License - 详见 [LICENSE](LICENSE)

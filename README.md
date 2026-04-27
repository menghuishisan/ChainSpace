# 链境 ChainSpace

区块链技术教学与实验平台

## 项目简介

链境是一个面向高校的区块链技术教学与实验平台，支持多租户架构，提供容器化实验环境和CTF竞赛系统。

## 技术栈

### 后端
- Go 1.21+ / Gin / GORM
- PostgreSQL / Redis
- Kubernetes

### 前端
- React 18 / TypeScript / Vite
- Ant Design 5.x / TailwindCSS
- Monaco Editor / xterm.js

## 项目结构

```
ChainPrac/
├── backend/          # 后端服务
├── frontend/         # 前端应用
├── docs/             # 设计文档
├── deploy/           # 部署配置
```

## 设计文档

| 文档 | 说明 |
|------|------|
| [01-项目概述](docs/01-项目概述.md) | 项目背景、目标 |
| [02-需求规格说明书](docs/02-需求规格说明书.md) | 功能需求 |
| [03-系统架构设计](docs/03-系统架构设计.md) | 技术架构 |
| [04-数据模型设计](docs/04-数据模型设计.md) | 数据库设计 |
| [05-实验环境设计](docs/05-实验环境设计.md) | 容器方案 |
| [06-接口设计](docs/06-接口设计.md) | API定义 |
| [07-功能流程图](docs/07-功能流程图.md) | 业务流程 |
| [08-UI界面设计](docs/08-UI界面设计.md) | 页面布局 |

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Docker & Kubernetes

### 本地开发

```bash
# 启动依赖服务
docker-compose -f deploy/docker-compose.yml up -d

# 启动后端
cd backend
go run cmd/server/main.go

# 启动前端
cd frontend
npm install
npm run dev
```

### 访问地址

- 前端: http://localhost:5173
- 后端: http://localhost:8080
- API文档: http://localhost:8080/swagger


## License

MIT

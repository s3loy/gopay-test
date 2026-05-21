# gopay — CLAUDE.md

## 项目概述

gopay 是一个生产级 Go 全栈支付网关，支持微信支付 V3 和支付宝 V3。采用 Clean Architecture 分层架构，提供 RESTful API 和 Next.js 收银台前端。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.24 + Gin + GORM + PostgreSQL |
| 支付 SDK | gopay v1.5.107 |
| 配置 | Viper |
| 日志 | Zap + Lumberjack |
| 校验 | validator/v10 |
| DI | Wire（声明式，当前手动注入） |
| 前端 | Next.js 14 App Router + TypeScript + Tailwind CSS |
| 部署 | Docker + Docker Compose |

## 目录结构

```
gopay/
├── cmd/api/              # 服务入口
├── internal/
│   ├── domain/           # 实体 + 仓储接口 + 领域服务接口
│   ├── usecase/          # 业务用例编排
│   ├── infrastructure/   # GORM 实现 + 支付 Provider 封装
│   ├── interfaces/http/  # Handler + Middleware + Router + DTO
│   └── pkg/              # 配置、日志、错误码、响应封装等
├── web/                  # Next.js 前端
├── configs/              # YAML 配置
└── deployments/docker/   # Dockerfile + docker-compose
```

## 开发命令

```bash
# 编译
go build ./cmd/api

# 测试（全部）
go test ./...

# 测试（业务逻辑）
go test ./internal/domain/entity/... ./internal/usecase/... ./internal/pkg/apperror/...

# 运行（需本地 PostgreSQL）
go run ./cmd/api -config configs/config.dev.yaml

# 前端
cd web && npm install && npm run dev
```

## 编码约定

- **Commit message**: `type(scope): description`，Conventional Commits，英文小写，≤72 字符
- **分支命名**: `fix/security-and-transaction`（无编号，描述改动方向）
- **Clean Architecture**: domain → usecase → infrastructure，内层不依赖外层
- **错误码**: 6 位数字 `XXYYYY`，定义在 `internal/pkg/apperror/codes.go`
- **测试**: 表驱动（table-driven），entity + usecase 目标 100% 语句覆盖

## Scope 定义

| Scope | 说明 |
|-------|------|
| `entity` | 领域实体 |
| `usecase` | 业务用例 |
| `repo` | 仓储实现 |
| `handler` | HTTP handler |
| `provider` | 支付 provider |
| `config` | 配置 |
| `middleware` | HTTP 中间件 |

## 本地开发环境

1. 启动 PostgreSQL（Docker）：`docker run -d -p 5432:5432 -e POSTGRES_USER=gopay -e POSTGRES_PASSWORD=gopay123 -e POSTGRES_DB=gopay postgres:16`
2. 配置 `configs/config.dev.yaml`
3. 运行 `go run ./cmd/api -config configs/config.dev.yaml`
4. 前端 `cd web && npm run dev`

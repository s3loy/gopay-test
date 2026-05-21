# gopay

生产级 Go 全栈支付网关，支持微信支付 V3 和支付宝 V3。

## 特性

- 双渠道支付：微信（Native/JSAPI）+ 支付宝（PC/扫码/WAP/APP）
- Clean Architecture 分层：domain → usecase → infrastructure → interfaces
- 统一错误码体系（6 位数字编码）+ 统一响应格式
- 数据库事务保护核心业务操作
- Webhook 安全：签名验证 + 请求体大小限制 + 基础防重放
- Next.js 14 收银台前端

## 快速开始

### 前提条件

- Go 1.24+
- PostgreSQL 16+
- Node.js 20+（前端）

### 后端

```bash
# 1. 启动 PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_USER=gopay \
  -e POSTGRES_PASSWORD=gopay123 \
  -e POSTGRES_DB=gopay \
  postgres:16

# 2. 配置
cp configs/config.dev.yaml configs/config.local.yaml
# 编辑 configs/config.local.yaml，填入支付密钥

# 3. 运行
go run ./cmd/api -config configs/config.local.yaml
```

### 前端

```bash
cd web
npm install
npm run dev
```

## API 概览

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/orders` | 创建订单 |
| GET | `/api/v1/orders/:order_no` | 订单详情 |
| POST | `/api/v1/orders/:order_no/close` | 关闭订单 |
| POST | `/api/v1/payments` | 创建支付 |
| GET | `/api/v1/payments/:payment_no` | 支付详情 |
| POST | `/api/v1/refunds` | 申请退款 |
| GET | `/api/v1/refunds/:refund_no` | 退款详情 |
| POST | `/webhook/wechat/notify` | 微信支付回调 |
| POST | `/webhook/alipay/notify` | 支付宝回调 |
| GET | `/health` | 健康检查 |

统一响应格式：`{"code": 0, "message": "success", "data": {}, "request_id": "...", "timestamp": 1234567890}`

## 项目结构

```
gopay/
├── cmd/api/                    # 服务入口
├── internal/
│   ├── domain/entity/          # Order, Payment, Refund 实体
│   ├── domain/repository/      # 仓储接口
│   ├── domain/service/         # PaymentProvider 接口
│   ├── usecase/                # 业务用例
│   ├── infrastructure/
│   │   ├── persistence/postgresql/  # GORM 实现
│   │   └── payment/            # 微信/支付宝 Provider
│   ├── interfaces/http/        # Handler, Middleware, Router, DTO
│   └── pkg/                    # 配置、日志、错误码、响应等
├── web/                        # Next.js 前端
├── configs/                    # YAML 配置
└── deployments/docker/         # Dockerfile + docker-compose
```

## 环境变量

配置通过 YAML 文件 + 环境变量覆盖。环境变量前缀 `GOPAY_`，使用 `_` 代替 `.`：

| 变量 | 说明 |
|------|------|
| `GOPAY_APP_ENV` | 环境：`development` / `production` |
| `GOPAY_SERVER_PORT` | 服务端口 |
| `GOPAY_DATABASE_HOST` | 数据库主机 |
| `GOPAY_DATABASE_PASSWORD` | 数据库密码 |
| `GOPAY_PAYMENT_WECHAT_APPID` | 微信 AppID |
| `GOPAY_PAYMENT_WECHAT_API_V3_KEY` | 微信 APIv3 密钥 |
| `GOPAY_PAYMENT_ALIPAY_APPID` | 支付宝 AppID |

## 测试

```bash
go test ./...
```

重点覆盖：entity 状态机、usecase 业务逻辑、apperror 错误处理。

## 部署

```bash
cd deployments/docker
docker-compose up --build
```

## License

MIT

# gopay- test

Go 支付网关

封装微信 V3 和支付宝 V3 接口

学习和练习项目

- 支付接口封装（微信 Native/JSAPI、支付宝 PC/扫码/WAP/APP）

> 项目实现了完整的支付流程代码，但需要配置真实的商户密钥和证书才能调用微信支付/支付宝的正式环境。

## 快速开始

### 前提条件

- Go 1.26
- PostgreSQL 16+
- Node.js 20+（前端）

### 支付宝沙箱

项目已内置支付宝沙箱测试配置，certs/ 目录下包含示例密钥对。要本地跑通支付流程：

**1. 获取沙箱账号**

访问 [支付宝开放平台沙箱](https://open.alipay.com/develop/sandbox/app)，登录后获取：
- **沙箱应用 AppID**（替换 `configs/config.local.yaml` 中的 `appid`）
- **支付宝公钥**（替换 `certs/alipay_public_key.pem`）
- **应用公钥** (`certs/app_public_key.pem`)
- **应用私钥** (`certs/app_private_key.pem`)


**3. 配置文件**

`configs/config.local.yaml` 中支付宝部分已预填沙箱配置，只需确认：

```yaml
payment:
  alipay:
    enabled: true
    appid: "你的沙箱APPID"
    private_key_path: "certs/app_private_key.pem"
    public_key_path: "certs/alipay_public_key.pem"
    notify_url: "http://localhost:8080/webhook/alipay/notify"
    is_prod: false   # 沙箱环境必须 false
```

**4. 验证支付流程**

启动服务后，用以下 curl 命令走完整流程：

```bash
# 1. 创建订单
ORDER_RESP=$(curl -s -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"subject":"测试商品","amount":100,"currency":"CNY"}')
ORDER_NO=$(echo $ORDER_RESP | grep -o '"order_no":"[^"]*"' | cut -d'"' -f4)
echo "订单号: $ORDER_NO"

# 2. 创建支付宝支付（PC 端）
curl -s -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d "{\"order_no\":\"$ORDER_NO\",\"channel\":\"alipay\",\"method\":\"pc\",\"client_ip\":\"127.0.0.1\",\"return_url\":\"http://localhost:3000/pay/result\"}" | jq .
```

返回的 `pay_params` 中会有 `pay_url`，复制到浏览器打开，用**沙箱支付宝 App**（安卓）扫码或登录沙箱买家账号完成支付。

支付成功后，服务端会收到支付宝回调，订单状态变为 `paid`。可通过以下接口查询确认：

```bash
# 3. 查询支付状态
curl -s http://localhost:8080/api/v1/payments/{payment_no}

# 4. 查询订单状态
curl -s http://localhost:8080/api/v1/orders/$ORDER_NO
```

如果订单状态从 `pending` → `paid`，说明支付链路完全通畅。

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
# 编辑 configs/config.local.yaml，填入支付密钥（或保持空值只跑本地接口）

# 3. 运行
go run ./cmd/api -config configs/config.local.yaml
```

### 前端

未测试，仅占位

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

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.26 + Gin + GORM + PostgreSQL |
| 支付 SDK | gopay |
| 配置 | Viper |
| 日志 | Zap + Lumberjack |
| 校验 | validator/v10 |
| DI | Wire |
| Lint | golangci-lint |
| 前端 | Next.js 16 + React 18 + TypeScript + Tailwind CSS |
| 部署 | Docker + Docker Compose |

## 项目结构

```
gopay/
├── cmd/api/                         # 服务入口
├── internal/
│   ├── domain/
│   │   ├── entity/                  # Order, Payment, Refund 实体及状态机
│   │   ├── repository/              # 仓储接口
│   │   └── service/                 # PaymentProvider 接口定义
│   ├── usecase/                     # 业务用例编排
│   ├── infrastructure/
│   │   ├── persistence/postgresql/  # GORM 仓储实现
│   │   └── payment/                 # 微信/支付宝 Provider 封装
│   │       ├── alipay/
│   │       └── wechat/
│   ├── interfaces/http/
│   │   ├── handler/                 # HTTP Handler
│   │   ├── middleware/              # CORS、日志、恢复等
│   │   ├── router/                  # 路由定义
│   │   └── dto/                     # 请求/响应 DTO
│   └── pkg/
│       ├── apperror/                # 错误码定义
│       ├── config/                  # 配置加载
│       ├── logger/                  # Zap 日志
│       ├── response/                # 统一响应封装
│       └── validator/               # 校验器封装
├── web/                             # Next.js 前端
├── configs/                         # YAML 配置文件
└── deployments/docker/              # Dockerfile + docker-compose
```

## 部署

```bash
cd deployments/docker
docker compose up --build
```

镜像分层：
- **Builder**: `golang:1.26-alpine`
- **Runtime**: `alpine:3.22`（非 root 用户运行）

## 开发注意事项

### 支付配置

项目使用 [gopay](https://github.com/go-pay/gopay) SDK 封装微信/支付宝接口。要调用真实支付环境，需在配置文件中提供：

- 微信：MchID、APIv3Key、商户证书路径、平台公钥路径
- 支付宝：AppID、应用私钥路径、公钥/根证书路径

未配置密钥时服务可正常启动，但支付接口会返回配置不可用的错误。

## License

[MIT](LICENSE)

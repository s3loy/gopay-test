# 测试质量审计报告 —— gopay 项目

**审计分支**: main  
**审计日期**: 2026-05-16  
**审计 Agent**: 测试质量审计 Agent  
**项目**: github.com/s3loy/gopay (Go 1.24)

---

## 1. 测试覆盖率统计表（按包）

| 包路径 | Go 文件数 | 测试文件数 | 测试函数数 | 覆盖率 | 状态 |
|--------|----------|-----------|-----------|--------|------|
| `internal/pkg/config` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/pkg/logger` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/pkg/apperror` | 2 | 0 | 0 | 0% | 缺失 |
| `internal/pkg/validator` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/pkg/pagination` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/pkg/timeutil` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/pkg/response` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/domain/entity` | 4 | 0 | 0 | 0% | 缺失 |
| `internal/domain/repository` | 3 | 0 | 0 | 0% | 缺失 |
| `internal/domain/service` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/usecase` | 3 | 0 | 0 | 0% | 缺失 |
| `internal/infrastructure/persistence/postgresql` | 5 | 0 | 0 | 0% | 缺失 |
| `internal/infrastructure/payment` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/infrastructure/payment/wechat` | 2 | 0 | 0 | 0% | 缺失 |
| `internal/infrastructure/payment/alipay` | 2 | 0 | 0 | 0% | 缺失 |
| `internal/interfaces/http/handler` | 5 | 0 | 0 | 0% | 缺失 |
| `internal/interfaces/http/dto` | 4 | 0 | 0 | 0% | 缺失 |
| `internal/interfaces/http/router` | 1 | 0 | 0 | 0% | 缺失 |
| `internal/interfaces/http/middleware` | 5 | 0 | 0 | 0% | 缺失 |
| `cmd/api` | 2 | 0 | 0 | 0% | 缺失 |
| `pkg/constants` | 1 | 0 | 0 | 0% | 缺失 |
| **总计** | **45** | **0** | **0** | **0%** | **全部缺失** |

> 注：纯常量/类型定义文件（如 `codes.go`、`constants.go`、`common.go` 等）在覆盖率统计中可豁免，但当前项目**无任何测试文件**，不存在豁免空间。

---

## 2. 未测试的关键代码路径清单

### 2.1 核心业务逻辑（Critical）

| 文件 | 关键路径 | 风险等级 |
|------|---------|---------|
| `internal/usecase/order_usecase.go` | `Create` — 金额校验、默认货币/过期时间、订单号生成 | P0 |
| `internal/usecase/order_usecase.go` | `Close` — 订单状态流转（MarkClosed 错误路径） | P0 |
| `internal/usecase/payment_usecase.go` | `Create` — 订单可支付校验、provider 调用、失败回滚 | P0 |
| `internal/usecase/payment_usecase.go` | `HandleWechatNotify` — 签名验证、幂等处理、状态更新 | P0 |
| `internal/usecase/payment_usecase.go` | `HandleAlipayNotify` — 同上，支付宝通知处理 | P0 |
| `internal/usecase/refund_usecase.go` | `Create` — 退款金额校验、累计退款校验、provider 退款调用 | P0 |
| `internal/domain/entity/order.go` | `CanPay/CanClose/CanRefund/MarkPaid/MarkClosed` 全部状态机方法 | P0 |
| `internal/domain/entity/payment.go` | `CanRefund/IsSuccess/IsExpired` | P0 |

### 2.2 数据持久层（Critical）

| 文件 | 关键路径 | 风险等级 |
|------|---------|---------|
| `internal/infrastructure/persistence/postgresql/order_repo.go` | Create/GetByID/GetByOrderNo/Update/UpdateStatus/List + RecordNotFound 处理 | P0 |
| `internal/infrastructure/persistence/postgresql/payment_repo.go` | 全部 CRUD + List + GetByThirdPartyNo/GetByOrderID | P0 |
| `internal/infrastructure/persistence/postgresql/refund_repo.go` | 全部 CRUD + GetTotalRefundAmount + List | P0 |
| `internal/infrastructure/persistence/postgresql/db.go` | NewDB 连接池配置、autoMigrate | P1 |
| `internal/infrastructure/persistence/postgresql/model.go` | JSONMap Value/Scan 方法 | P1 |

### 2.3 支付 Provider（Critical）

| 文件 | 关键路径 | 风险等级 |
|------|---------|---------|
| `internal/infrastructure/payment/factory.go` | Get — channel 路由、不支持 channel 错误 | P0 |
| `internal/infrastructure/payment/wechat/provider.go` | CreatePayment(Native/JSAPI)、QueryPayment、Refund、VerifyNotify | P0 |
| `internal/infrastructure/payment/wechat/client.go` | NewClient、IsAvailable | P1 |
| `internal/infrastructure/payment/alipay/provider.go` | CreatePayment(PC/WAP/APP)、QueryPayment、Refund、VerifyNotify | P0 |
| `internal/infrastructure/payment/alipay/client.go` | NewClient、IsAvailable、ParseAlipayNotify | P1 |

### 2.4 HTTP Handler（High）

| 文件 | 关键路径 | 风险等级 |
|------|---------|---------|
| `internal/interfaces/http/handler/order_handler.go` | Create/Get/Close + 参数绑定 + 响应转换 | P1 |
| `internal/interfaces/http/handler/payment_handler.go` | Create/Get + 参数绑定 | P1 |
| `internal/interfaces/http/handler/refund_handler.go` | Create/Get + 参数绑定 | P1 |
| `internal/interfaces/http/handler/webhook_handler.go` | WechatNotify/AlipayNotify + body 读取 + header 提取 | P1 |
| `internal/interfaces/http/handler/health_handler.go` | Check | P2 |

### 2.5 中间件（Medium）

| 文件 | 关键路径 | 风险等级 |
|------|---------|---------|
| `internal/interfaces/http/middleware/recovery.go` | Panic 恢复 | P1 |
| `internal/interfaces/http/middleware/timeout.go` | 超时取消 + done channel | P1 |
| `internal/interfaces/http/middleware/request_id.go` | Header 读取/生成 | P2 |
| `internal/interfaces/http/middleware/logger.go` | 请求日志字段组装 | P2 |
| `internal/interfaces/http/middleware/cors.go` | CORS header 设置 | P2 |

### 2.6 工具包（Medium）

| 文件 | 关键路径 | 风险等级 |
|------|---------|---------|
| `internal/pkg/apperror/error.go` | New/Wrap/Is/WithDetail/WithCause/WithHTTPStatus/Error/Unwrap | P1 |
| `internal/pkg/apperror/codes.go` | GetMessage/GetHTTPStatus — 所有 code 映射 | P2 |
| `internal/pkg/validator/validator.go` | ValidateStruct — 验证错误转换 | P1 |
| `internal/pkg/response/response.go` | OK/Error/ErrorWithCode/Page/JSON — 含非 AppError 分支 | P1 |
| `internal/pkg/config/config.go` | Load — 默认值设置、文件读取、环境变量、Unmarshal | P1 |
| `internal/pkg/logger/logger.go` | Init — 文件/stdout 分支、parseLevel、buildEncoder | P1 |
| `internal/pkg/pagination/pagination.go` | FromContext — 边界值（0、负数、超过 MaxPageSize） | P2 |
| `internal/pkg/timeutil/timeutil.go` | Now/NowPtr/Parse/Format/FormatPtr | P2 |

### 2.7 基础设施（Medium）

| 文件 | 关键路径 | 风险等级 |
|------|---------|---------|
| `cmd/api/main.go` | 启动流程、依赖注入顺序、graceful shutdown | P1 |
| `cmd/api/wire.go` | wire 构建图 | P2 |
| `internal/interfaces/http/router/router.go` | 路由注册、中间件顺序 | P2 |

---

## 3. 测试质量评价

### 3.1 表驱动测试（Table-Driven Tests）

**状态**: 完全缺失

项目无任何测试文件，因此不存在表驱动测试。Go 社区最佳实践强烈推荐使用表驱动测试来覆盖同一函数的多组输入/预期输出，当前项目完全未采用。

### 3.2 边界 Case 覆盖

**状态**: 完全缺失

以下边界 case 均未测试：

- **OrderUsecase.Create**: amount = 0（应拒绝）、currency = ""（应默认 CNY）、expireMinutes = 0（应默认 30）
- **RefundUsecase.Create**: amount = 0（应拒绝）、amount = payment.Amount（边界全额退款）、amount > remaining（应拒绝）
- **PaymentUsecase.Create**: 已过期订单（CanPay = false）、不支持的 channel/method
- **Pagination.FromContext**: page = 0（应默认 1）、pageSize = 0（应默认 20）、pageSize = 999（应截断为 100）
- **apperror.Wrap**: err = nil（应返回 nil）、err 已是 *AppError（应直接返回）
- **entity.Order.MarkPaid**: 已支付订单再次调用、已关闭订单调用
- **entity.Order.MarkClosed**: 已关闭订单再次调用、已支付订单调用

### 3.3 错误路径测试

**状态**: 完全缺失

以下错误路径均未测试：

- Repository 层 `gorm.ErrRecordNotFound` 转换为业务错误码
- Provider 调用失败后的 payment/refund 状态回滚（`UpdateStatus` 到 Failed）
- Webhook 通知处理中的幂等逻辑（已成功的 payment 再次收到通知）
- Config.Load 文件不存在、解析失败路径
- Logger.Init 目录创建失败路径
- Response.Error 非 *AppError 分支（返回 500）
- 微信/支付宝 provider 的 `!IsAvailable()` 分支

### 3.4 Mock 使用合理性

**状态**: 无法评价（无测试即无 mock）

**建议策略**（当开始编写测试时）：

| 层级 | 建议 | 理由 |
|------|------|------|
| Usecase | mock `repository` 接口 + mock `PaymentProviderFactory` | 纯业务逻辑，应隔离外部依赖 |
| Repository | **不 mock GORM**，使用 testcontainers/嵌入式 PostgreSQL 做集成测试 | 数据库操作需要验证 SQL 正确性 |
| Provider | mock `wechatv3.ClientV3` / `alipayv3.ClientV3` | 第三方 API 调用，mock 合理 |
| Handler | 使用 `gin` 的 `httptest` + mock usecase | HTTP 层测试标准做法 |
| Webhook | mock usecase + 构造真实 HTTP 请求 | 验证 body/header 解析正确性 |

---

## 4. 集成测试

### 4.1 数据库操作集成测试

**状态**: 完全缺失

建议：
- 使用 `testcontainers-go` 启动 PostgreSQL 容器
- 或使用 `github.com/ory/dockertest` 做集成测试
- 每个测试用例在事务中执行，结束后回滚，保证测试隔离性

### 4.2 支付 Provider 集成测试

**状态**: 完全缺失

建议：
- 微信/支付宝 provider 应提供 **mock provider** 实现（基于 `service.PaymentProvider` 接口）
- 沙箱环境测试可作为可选的 CI 步骤（非阻塞）
- 至少应测试 provider factory 的路由逻辑和不支持 channel 的错误处理

---

## 5. 测试结构检查

### 5.1 _test.go 文件位置

**状态**: 完全缺失

所有 45 个 Go 源文件均无成对的 `*_test.go` 文件。Go 惯例要求测试文件与源文件同包（或 `package xxx_test` 用于黑盒测试）。

### 5.2 测试命名

**状态**: 完全缺失

无测试函数，无法检查命名。Go 惯例：
- 单元测试：`TestXxx`（对应被测函数/方法）
- 子测试：`TestXxx/yyy`（使用 `t.Run`）
- Benchmark：`BenchmarkXxx`
- 示例：`ExampleXxx`

### 5.3 Benchmark 测试

**状态**: 完全缺失

以下函数建议添加 benchmark：
- `generateOrderNo` / `generatePaymentNo` / `generateRefundNo` — 高并发场景下的性能基线
- `apperror.Wrap` — 错误包装在热路径上的开销
- `JSONMap.Value/Scan` — JSON 序列化/反序列化性能
- `OrderUsecase.Create` — 完整业务流程的吞吐基线

---

## 6. 高优先级测试补充建议

按优先级排序的测试补充计划：

### Phase 1 — 实体层（最简单，建立测试基线）
- `internal/domain/entity/order_test.go` — 状态机全部方法
- `internal/domain/entity/payment_test.go` — CanRefund/IsExpired
- `internal/pkg/apperror/error_test.go` — New/Wrap/Is/链式方法

### Phase 2 — Usecase 层（核心业务）
- `internal/usecase/order_usecase_test.go` — Create/Get/Close，mock repo
- `internal/usecase/payment_usecase_test.go` — Create/HandleWechatNotify/HandleAlipayNotify，mock repo + mock provider
- `internal/usecase/refund_usecase_test.go` — Create/Get，mock repo + mock provider

### Phase 3 — 基础设施层
- `internal/infrastructure/persistence/postgresql/*_test.go` — 集成测试（testcontainers）
- `internal/infrastructure/payment/factory_test.go` — provider 路由

### Phase 4 — HTTP 层
- `internal/interfaces/http/handler/*_test.go` — gin httptest + mock usecase
- `internal/interfaces/http/middleware/*_test.go` — 中间件行为验证

### Phase 5 — 工具包
- `internal/pkg/config/config_test.go` — 默认值、环境变量覆盖
- `internal/pkg/validator/validator_test.go` — 验证错误转换
- `internal/pkg/response/response_test.go` — 响应格式
- `internal/pkg/pagination/pagination_test.go` — 边界值

---

## 7. 总体评分与结论

### 评分维度

| 维度 | 权重 | 得分 | 加权得分 |
|------|------|------|---------|
| 测试覆盖率 | 30% | 0/10 | 0.0 |
| 测试质量（表驱动/边界/mock） | 25% | 0/10 | 0.0 |
| 错误路径覆盖 | 20% | 0/10 | 0.0 |
| 集成测试 | 15% | 0/10 | 0.0 |
| 测试结构规范 | 10% | 0/10 | 0.0 |
| **总分** | **100%** | — | **0.0/10** |

### 结论

**不通过。**

项目当前处于**零测试状态**——45 个 Go 源文件、0 个测试文件、0% 语句覆盖率。核心业务逻辑（订单状态机、支付创建、退款校验、Webhook 处理、数据库 CRUD）均无任何测试保护。

这是一个支付系统，涉及资金流转，测试缺失的风险极高。任何代码变更都可能引入回归缺陷，且无法在 CI 中捕获。

### 建议行动

1. **立即补充** Phase 1（实体层）+ Phase 2（usecase 层）测试，建立最低测试基线
2. **引入 mock 框架**（如 `github.com/golang/mock` 或 `github.com/vektra/mockery`）为 repository 接口生成 mock
3. **设置 CI 覆盖率门禁**，目标：usecase + entity 包 100% 语句覆盖
4. **添加集成测试框架**（testcontainers-go）用于 repository 层
5. **补充 benchmark**，为高并发支付场景建立性能基线

---

*报告生成时间: 2026-05-16*  
*审计 Agent: 测试质量审计 Agent*

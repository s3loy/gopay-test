# 测试质量审计报告 — fix/security-and-transaction 分支

**审计日期**: 2026/05/19  
**审计范围**: 所有 `*_test.go` 文件 + 无测试包  
**审计目标**: entity + usecase 100% 语句覆盖

---

## 1. 覆盖率统计表

| 包 | 语句覆盖率 | 目标 | 状态 |
|---|---|---|---|
| `internal/domain/entity` | 45.2% | 100% | ❌ 未达标 |
| `internal/pkg/apperror` | 90.0% | 100% | ❌ 未达标 |
| `internal/usecase` | 34.0% | 100% | ❌ 未达标 |
| `internal/infrastructure/persistence/postgresql` | 0.0% | — | ⚠️ 无测试 |
| `internal/infrastructure/payment` | 0.0% | — | ⚠️ 无测试 |
| `internal/interfaces/http/handler` | 0.0% | — | ⚠️ 无测试 |
| `internal/interfaces/http/middleware` | 0.0% | — | ⚠️ 无测试 |

### 函数级覆盖率明细

| 文件 | 函数 | 覆盖率 | 备注 |
|---|---|---|---|
| `entity/order.go` | `IsExpired` | 100% | ✅ |
| `entity/order.go` | `CanPay` | 100% | ✅ |
| `entity/order.go` | `CanClose` | 100% | ✅ |
| `entity/order.go` | `CanRefund` | 100% | ✅ |
| `entity/order.go` | `MarkPaid` | 100% | ✅ |
| `entity/order.go` | `MarkClosed` | 100% | ✅ |
| `entity/payment.go` | `IsSuccess` | 100% | ✅ |
| `entity/payment.go` | `CanRefund` | 100% | ✅ |
| `entity/payment.go` | `IsExpired` | 100% | ✅ |
| `entity/common.go` | `OrderStatus.String` | 0.0% | ❌ 未测试 |
| `entity/common.go` | `PaymentStatus.String` | 0.0% | ❌ 未测试 |
| `entity/common.go` | `RefundStatus.String` | 0.0% | ❌ 未测试 |
| `apperror/error.go` | `New` | 100% | ✅ |
| `apperror/error.go` | `NewWithStatus` | 0.0% | ❌ 未测试 |
| `apperror/error.go` | `Wrap` | 100% | ✅ |
| `apperror/error.go` | `Is` | 100% | ✅ |
| `apperror/error.go` | `Error` | 100% | ✅ |
| `apperror/error.go` | `Unwrap` | 100% | ✅ |
| `apperror/error.go` | `WithDetail` | 100% | ✅ |
| `apperror/error.go` | `WithCause` | 100% | ✅ |
| `apperror/error.go` | `WithHTTPStatus` | 100% | ✅ |
| `apperror/codes.go` | `GetMessage` | 66.7% | ⚠️ 缺未知 code 分支 |
| `apperror/codes.go` | `GetHTTPStatus` | 66.7% | ⚠️ 缺未知 code 分支 |
| `usecase/order_usecase.go` | `NewOrderUsecase` | 100% | ✅ |
| `usecase/order_usecase.go` | `Create` | 100% | ✅ |
| `usecase/order_usecase.go` | `Get` | 100% | ✅ |
| `usecase/order_usecase.go` | `Close` | 83.3% | ⚠️ 缺 MarkClosed 返回 error 后的路径 |
| `usecase/order_usecase.go` | `generateOrderNo` | 100% | ✅ |
| `usecase/payment_usecase.go` | `NewPaymentUsecase` | 0.0% | ❌ 未测试 |
| `usecase/payment_usecase.go` | `Create` | 0.0% | ❌ 未测试 |
| `usecase/payment_usecase.go` | `Get` | 0.0% | ❌ 未测试 |
| `usecase/payment_usecase.go` | `HandleWechatNotify` | 0.0% | ❌ 未测试 |
| `usecase/payment_usecase.go` | `HandleAlipayNotify` | 0.0% | ❌ 未测试 |
| `usecase/payment_usecase.go` | `generatePaymentNo` | 0.0% | ❌ 未测试 |
| `usecase/refund_usecase.go` | `NewRefundUsecase` | 100% | ✅ |
| `usecase/refund_usecase.go` | `Create` | 82.1% | ⚠️ 缺部分错误路径 |
| `usecase/refund_usecase.go` | `Get` | 0.0% | ❌ 未测试 |
| `usecase/refund_usecase.go` | `generateRefundNo` | 100% | ✅ |

---

## 2. 发现项

### [critical] payment_usecase.go 零测试覆盖

- **位置**: `internal/usecase/payment_usecase.go`
- **问题**: 支付核心用例（Create、HandleWechatNotify、HandleAlipayNotify）无任何测试，是整个支付链路最关键的业务逻辑。
- **影响**: 支付创建、回调通知处理等核心流程无自动化验证，风险极高。
- **建议**: 立即补充 `payment_usecase_test.go`，覆盖：
  - 正常创建支付流程
  - 订单不可支付时的错误
  - Provider 创建失败后的状态回滚
  - 微信支付回调（成功、重复通知、验证失败）
  - 支付宝回调（成功、重复通知、验证失败）

### [critical] entity 包整体覆盖率仅 45.2%

- **位置**: `internal/domain/entity/`
- **问题**: `refund.go`（纯 struct，无方法，可豁免）、`common.go` 中所有 `String()` 方法未测试。
- **影响**: 状态字符串化逻辑未验证，未知状态值的分支未覆盖。
- **建议**: 补充 `common.go` 中三个 `String()` 方法的表驱动测试，覆盖所有枚举值 + 未知值（如 `OrderStatus(99)`）。

### [warning] apperror 包缺 `NewWithStatus` 测试

- **位置**: `internal/pkg/apperror/error_test.go`
- **问题**: `NewWithStatus` 函数（`error.go:35`）无测试覆盖。
- **建议**: 补充单测验证 `NewWithStatus` 正确设置 `HTTPStatus`。

### [warning] apperror 包 `GetMessage` / `GetHTTPStatus` 缺未知 code 分支

- **位置**: `internal/pkg/apperror/codes.go:183-195`
- **问题**: 两个函数在 code 不存在于 map 时返回默认值的逻辑未测试。
- **建议**: 补充未知 code（如 `999999`）的测试用例，验证返回 `CodeUnknown` 和 `500`。

### [warning] order_usecase.go `Close` 缺 MarkClosed error 路径

- **位置**: `internal/usecase/order_usecase.go:72-81`
- **问题**: `Close` 方法中 `order.MarkClosed()` 返回 error 后的路径未测试（覆盖率 83.3%）。
- **建议**: 补充一个 case：订单状态为 `paid` 时调用 `Close`，验证返回 `CodeOrderCannotClose` 错误。

### [warning] refund_usecase.go `Create` 缺多个错误路径

- **位置**: `internal/usecase/refund_usecase.go:48-135`
- **问题**: 以下路径未覆盖：
  - `providerFact.Get` 返回 error
  - `refundRepo.Create` 返回 error
  - `refundRepo.Update` 返回 error
  - `refundRepo.GetTotalRefundAmount` 第二次调用返回 error（日志路径）
  - `paymentRepo.UpdateStatus` 返回 error（日志路径）
- **建议**: 补充上述错误路径的 mock 测试。

### [warning] refund_usecase.go `Get` 未测试

- **位置**: `internal/usecase/refund_usecase.go:137-139`
- **问题**: `Get` 方法无任何测试。
- **建议**: 补充成功和 not found 两个 case。

### [info] 无 Benchmark 测试

- **问题**: 所有测试文件中无任何 `Benchmark*` 函数。
- **建议**: 对 `generateOrderNo`、`generatePaymentNo`、`generateRefundNo` 等可能产生并发冲突的函数补充 benchmark，验证并发安全性。

### [info] 无集成测试

- **问题**: 项目仅有单元测试，无数据库/HTTP/Provider 集成测试。
- **建议**: 在 `tests/integration/` 下补充：
  - 数据库集成测试（使用 testcontainers 或内存 PostgreSQL）
  - HTTP handler 端到端测试（使用 `httptest`）
  - 支付 Provider 的 mock server 测试

### [info] Mock 实现位于测试文件中

- **问题**: `mockOrderRepo`、`mockPaymentRepo`、`mockRefundRepo`、`mockProvider`、`mockProviderFactory`、`mockTxMgr` 均内联定义在 `_test.go` 中。
- **评估**: 当前规模下可接受，但随着项目增长建议：
  - 使用 `mockery` 或 `gomock` 生成标准 mock
  - 或提取到 `internal/mocks/` 包中复用

### [info] 测试命名规范良好

- **评估**: 测试命名遵循 `Test<Struct>_<Method>` 或 `Test<Struct>_<Method>/case_name` 规范，符合 Go 惯例。

---

## 3. 高优先级补充计划

按优先级排序：

| 优先级 | 任务 | 预计新增测试文件 | 预计覆盖提升 |
|---|---|---|---|
| P0 | 创建 `payment_usecase_test.go` | 1 | usecase: 34% → ~70% |
| P0 | 补充 `entity/common_test.go`（String 方法） | 1 | entity: 45.2% → ~85% |
| P1 | 补充 `apperror` 缺失测试（NewWithStatus、未知 code） | 0（追加到现有文件） | apperror: 90% → 100% |
| P1 | 补充 `order_usecase_test.go` Close error 路径 | 0（追加到现有文件） | — |
| P1 | 补充 `refund_usecase_test.go` 缺失错误路径 + Get | 0（追加到现有文件） | usecase: ~70% → ~85% |
| P2 | 补充 benchmark（订单/支付/退款号生成） | 0（追加到现有文件） | — |
| P2 | 创建集成测试目录结构 | 1+ | — |

---

## 4. 审计结论

当前测试质量**不达标**。entity + usecase 的目标 100% 语句覆盖远未达成，核心支付流程（`payment_usecase.go`）完全无测试是最大的风险点。建议在合入 main 前至少完成 P0 和 P1 级别的补充。

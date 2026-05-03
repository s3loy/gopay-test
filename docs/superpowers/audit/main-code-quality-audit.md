# Go 代码质量审计报告 —— gopay

**审计日期**: 2026-05-16
**审计范围**: 完整代码库（internal/、cmd/、pkg/）
**审计维度**: Clean Architecture、设计精简度、接口隔离、错误处理、命名规范、代码重复、空值/边界处理、GORM 使用

---

## 1. Clean Architecture 合规性

### 1.1 依赖方向检查

| 层级 | 依赖方向 | 状态 |
|------|----------|------|
| domain (entity/repository/service) | 无外部依赖 | 合规 |
| usecase | 仅依赖 domain | 合规 |
| infrastructure (postgresql/wechat/alipay) | 依赖 domain + 外部库 | 合规 |
| interfaces (http handler/router/middleware) | 依赖 usecase + domain + gin | 合规 |
| cmd/api | 依赖所有层 | 合规 |

**结论**: 依赖方向正确，无反向依赖。domain 层纯净，符合 Clean Architecture 核心原则。

### 1.2 发现项

#### [info] wire.go 中 build tag 重复声明
- **文件**: `cmd/api/wire.go:1-2`
- **问题**: `//go:build wireinject` 和 `// +build wireinject` 同时存在，前者为 Go 1.17+ 新语法，后者为兼容旧版本。项目使用 Go 1.24，可移除旧语法。
- **建议**: 移除 `// +build wireinject` 行。

---

## 2. 设计精简度

### 2.1 发现项

#### [warning] `internal/pkg/timeutil` 包过度封装
- **文件**: `internal/pkg/timeutil/timeutil.go`
- **问题**: `Now()` 直接返回 `time.Now()`，`Parse()` 和 `Format()` 只是标准库函数的薄包装，未提供额外价值。`NowPtr()` 分配新指针，但每次调用都新建，不如调用处直接处理。
- **建议**: 评估是否真正需要此包。标准库 `time` 包已足够表达力，过度封装增加认知负担。若保留，考虑合并到更通用的工具包中。

#### [warning] `pkg/constants/constants.go` 与 `internal/domain/entity/common.go` 重复定义
- **文件**: `pkg/constants/constants.go` vs `internal/domain/entity/common.go:91-104`
- **问题**: 常量在两个位置重复定义（ChannelWechat/Alipay, MethodNative/JSAPI/PC/WAP/APP, CurrencyCNY）。`pkg/constants` 被谁引用需确认，若无人引用则为死代码。
- **建议**: 删除 `pkg/constants/constants.go`，统一使用 entity 包中的定义。单一事实来源原则。

#### [warning] `internal/pkg/pagination` 包未被使用
- **文件**: `internal/pkg/pagination/pagination.go`
- **问题**: 该包提供从 gin context 解析分页参数的功能，但 repository 的 `List` 方法直接接收 `Filter` 结构体（内含 Page/Size），Handler 层未调用 `pagination.FromContext()`。
- **建议**: 若当前设计（Filter 直接传参）是 intentional 的，则删除此包；若计划让 Handler 自动解析分页参数，则应在 Handler 中接入。

#### [info] `cmd/api/main.go` 手动装配与 wire.go 并存
- **文件**: `cmd/api/main.go:45-82`
- **问题**: main.go 中手动执行所有依赖注入，wire.go 也定义了注入图，但 main() 并未调用 `initializeApp()`。wire 工具形同虚设。
- **建议**: 统一使用 wire 生成代码，或移除 wire 依赖。当前两者并存造成维护负担。

---

## 3. 接口隔离

### 3.1 Repository 接口评估

| 接口 | 方法数 | 评价 |
|------|--------|------|
| `OrderRepository` | 6 | 合理，CRUD + List |
| `PaymentRepository` | 8 | 合理，含 ThirdPartyNo 查询 |
| `RefundRepository` | 8 | 合理，含 GetTotalRefundAmount |

**结论**: 接口粒度适中，无臃肿接口（God Interface）。

### 3.2 PaymentProvider 接口评估

- **文件**: `internal/domain/service/payment_provider.go:45-52`
- **评价**: 接口定义 6 个方法，覆盖支付、退款、查询、通知验证，是合理的抽象。但 `VerifyNotify` 的签名 `([]byte, map[string]string)` 对支付宝不友好——支付宝通知是 form 数据，wechat 是 JSON body，导致支付宝 provider 的 `VerifyNotify` 实现直接返回 headers（第 170 行），逻辑被拆分到 `ParseAlipayNotify` 函数。

#### [warning] PaymentProvider.VerifyNotify 签名对支付宝不自然
- **文件**: `internal/domain/service/payment_provider.go:50`
- **问题**: `VerifyNotify(ctx, body, headers)` 的签名源于微信支付（需要 body + headers 验签），但支付宝通知是 URL-encoded form 数据，验签方式完全不同。导致 `alipay/provider.go:166-171` 中 `VerifyNotify` 几乎为空实现，真正逻辑在 `ParseAlipayNotify`。
- **建议**: 考虑将通知处理拆分为渠道特定的 Handler，或在 usecase 层统一处理不同渠道的解析差异，而非强制套入同一接口。

### 3.3 发现项

#### [warning] wechat/alipay Client 结构重复
- **文件**: `internal/infrastructure/payment/wechat/client.go` vs `alipay/client.go`
- **问题**: 两个 Client 结构高度相似（都含 `client` 指针、`cfg`、 `IsAvailable()`、`V3()`、`NotifyURL()`），仅初始化逻辑和配置类型不同。
- **建议**: 提取公共的 `BaseClient[T any]` 或接口，减少重复。例如：
  ```go
  type PaymentClient interface {
      IsAvailable() bool
      NotifyURL() string
  }
  ```

---

## 4. 错误处理

### 4.1 错误码体系评估

- **文件**: `internal/pkg/apperror/codes.go`
- **评价**: 错误码按领域分层（System/Order/Payment/Refund/Webhook/WeChat/Alipay），结构清晰。HTTP 状态码映射合理。

### 4.2 发现项

#### [critical] `apperror.Wrap` 对 nil error 返回 nil，但调用处未处理可能 panic
- **文件**: `internal/pkg/apperror/error.go:43-56`
- **问题**: `Wrap(nil, code)` 返回 `nil` 是正确的，但需确保所有调用处不直接解引用返回值。经检查，所有调用都立即返回或判断，无此问题。

#### [warning] `Order.MarkPaid()` 错误码复用不当
- **文件**: `internal/domain/entity/order.go:41-52`
- **问题**: 当订单已支付时返回 `CodeOrderAlreadyPaid`（正确），但当订单不能支付（如已关闭/过期）时返回 `CodeOrderCannotClose`（`101006`），语义不匹配。"CannotClose" 用于关闭操作，不应复用于支付校验。
- **建议**: 新增 `CodeOrderCannotPay` 错误码，或在 `MarkPaid` 中返回更通用的 `CodeOrderExpired` / `CodeOrderAlreadyClosed`。

#### [warning] 多处错误被静默忽略
- **文件及位置**:
  - `internal/usecase/payment_usecase.go:156-159` — `order.MarkPaid()` 和 `orderRepo.UpdateStatus()` 错误被 `_` 忽略
  - `internal/usecase/payment_usecase.go:199-202` — 同上
  - `internal/usecase/refund_usecase.go:111-114` — `GetTotalRefundAmount` 和 `UpdateStatus` 错误被忽略
- **问题**: 支付通知处理中，订单状态更新失败被静默吞掉。这会导致支付已成功但订单状态仍为 pending 的数据不一致。
- **建议**: 至少记录日志；理想情况下应返回错误让调用方重试（webhook 应返回非 200 触发渠道重试）。

#### [warning] `RefundUsecase.Create` 中退款成功后更新订单状态逻辑错误
- **文件**: `internal/usecase/refund_usecase.go:111-114`
- **问题**: 当全额退款后，将 payment 状态更新为 `PaymentStatusFailed`（`102002`），语义错误。退款成功后的支付状态应为 "已退款" 或保持 "成功"，标记为 Failed 会混淆。
- **建议**: 新增 `PaymentStatusRefunded` 状态，或在 Order 层面标记 `OrderStatusFullRefund` / `OrderStatusPartialRefund`。

#### [warning] `alipay/client.go:68-70` 读取公钥错误被忽略
- **文件**: `internal/infrastructure/payment/alipay/client.go:68-70`
- **问题**: `PublicKey()` 方法中 `os.ReadFile` 错误被 `_` 忽略，返回空字符串。调用方 `ParseAlipayNotify` 若传入空 publicKey，验签会失败。
- **建议**: 返回 `(string, error)`，或在初始化时缓存公钥内容到内存。

---

## 5. 命名规范

### 5.1 总体评价

命名整体符合 Go 惯例，包名简洁，接口名以 `-er` 结尾（`OrderRepository`、`PaymentProvider`）。

### 5.2 发现项

#### [info] 工厂函数命名不一致
- **文件**: `internal/infrastructure/payment/factory.go:16`
- **问题**: `NewProviderFactory` 返回 `service.PaymentProviderFactory`，但函数名未体现 "Payment"。
- **建议**: 改为 `NewPaymentProviderFactory` 以与接口名保持一致，或保持现状（包名 `payment` 已提供上下文）。

#### [info] `generateOrderNo()` / `generatePaymentNo()` / `generateRefundNo()` 未导出但分散在多个文件
- **文件**: `internal/usecase/order_usecase.go:71-73`, `payment_usecase.go:209-211`, `refund_usecase.go:123-125`
- **问题**: 三个函数逻辑几乎相同（前缀 + 时间戳 + 纳秒尾数），分散在不同 usecase 文件中。
- **建议**: 提取到 `internal/pkg/idgen` 或类似包中，统一编号生成策略。

#### [info] `JSONMap` 类型命名
- **文件**: `internal/infrastructure/persistence/postgresql/model.go:9`
- **问题**: `JSONMap` 作为 GORM 自定义类型名合理，但 `map[string]any` 在多处直接使用，可考虑在 entity 中也使用此类型别名以保持一致。

---

## 6. 代码重复

### 6.1 wechat/alipay provider 重复模式

| 重复模式 | wechat 位置 | alipay 位置 | 重复度 |
|----------|-------------|-------------|--------|
| `IsAvailable()` 检查 | 每个方法开头 | 每个方法开头 | 100% |
| API 错误处理（调用→判错→判状态码） | 每个方法 | 每个方法 | ~80% |
| BodyMap 构建 | CreatePayment/Refund | CreatePayment/Refund | ~60% |

#### [warning] provider 方法中重复的可用性检查 + 错误处理
- **文件**: `wechat/provider.go` 和 `alipay/provider.go`
- **问题**: 每个 provider 方法都以相同的 3 行代码开头：
  ```go
  if !p.client.IsAvailable() {
      return nil, apperror.New(..., "xxx client not available")
  }
  ```
  以及几乎相同的 API 响应错误处理模式。
- **建议**: 提取装饰器或辅助函数：
  ```go
  func (p *provider) checkAvailable() error { ... }
  func wrapAPIError(err error, code int) error { ... }
  ```

### 6.2 Repository 重复模式

三个 repository 文件（order_repo.go / payment_repo.go / refund_repo.go）结构几乎完全相同：
- Create: model 转换 + Create + 回填 ID
- GetByXxx: Where + First + ErrRecordNotFound 处理
- Update: Model + Where + Updates
- UpdateStatus: Model + Where + Update("status")
- List: 条件构建 + Count + Offset/Limit/Find

#### [info] Repository 实现高度重复
- **评价**: 这是 GORM repository 的典型模式，一定程度的重复可接受。若需精简，可考虑使用泛型基类或代码生成。
- **建议**: 当前重复度在可接受范围内，但若继续增加更多 entity，建议引入泛型 repository 基类。

### 6.3 Handler 重复模式

三个 Handler（order/payment/refund）的 Create/Get 方法结构几乎相同：bind → validate → call usecase → response。

#### [info] Handler 结构重复
- **评价**: HTTP Handler 的重复是常见现象，当前每 Handler 仅 2 个方法，提取公共抽象的收益有限。
- **建议**: 保持现状，待 Handler 数量增长后再考虑提取通用 CRUD Handler。

---

## 7. 空值/边界处理

### 7.1 发现项

#### [critical] `pagination.FromContext` 中 `pageSize > MaxPageSize` 但未处理 `Size = 0` 导致的除零/负数
- **文件**: `internal/pkg/pagination/pagination.go:22-42`
- **问题**: 虽然 `pageSize <= 0` 时会被重置为 `DefaultPageSize`，但 `Filter.Size` 若被直接构造为 0（不经过 FromContext），`List` 方法中 `offset := (filter.Page - 1) * filter.Size` 不会除零，但 `Limit(0)` 在 GORM 中可能返回空结果而非错误。
- **建议**: 在 Repository `List` 方法入口处对 Page/Size 做防御性校验。

#### [warning] `Order.ExpiredAt` 边界未校验
- **文件**: `internal/usecase/order_usecase.go:27-54`
- **问题**: `Create` 方法中 `expireMinutes` 被限制为 `>0`，但未设上限。若传入极大值（如 `1<<31`），`time.Now().Add()` 可能溢出（Go 中 time.Duration 为 int64 纳秒，约 290 年上限，实际不易溢出，但应做上限约束）。
- **建议**: 增加 `expireMinutes <= 1440`（24小时）或业务允许的最大值校验。

#### [warning] `WebhookHandler.WechatNotify` 未限制 body 大小
- **文件**: `internal/interfaces/http/handler/webhook_handler.go:21-27`
- **问题**: `io.ReadAll(c.Request.Body)` 无大小限制，恶意请求可发送超大 body 导致 OOM。
- **建议**: 使用 `io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))` 限制读取大小（如 1MB）。

#### [warning] `Timeout` middleware 中 goroutine 泄漏风险
- **文件**: `internal/interfaces/http/middleware/timeout.go:11-33`
- **问题**: 当请求超时时，`c.Abort()` 被调用，但 `c.Next()` 所在的 goroutine 仍在运行，可能继续执行后续 handler 逻辑并尝试写入已关闭的 response writer。
- **建议**: Gin 的 `c.Abort()` 不会停止已启动的 goroutine。考虑使用 Gin 的 `context.Timeout()` 替代自定义实现，或确保 handler 中检查 `c.IsAborted()`。

#### [info] `RefundUsecase.Create` 中 `totalRefunded >= payment.Amount` 使用 `>=`
- **文件**: `internal/usecase/refund_usecase.go:112`
- **问题**: 前面已校验 `req.Amount <= remaining`，所以 `totalRefunded` 理论上不会超过 `payment.Amount`。使用 `>=` 是防御性的，但逻辑上应为 `==`。
- **建议**: 改为 `==` 更精确，或保留 `>=` 作为防御编程。

---

## 8. GORM 使用

### 8.1 事务处理

#### [critical] 核心业务操作无事务保护
- **文件**:
  - `internal/usecase/payment_usecase.go:53-117` — CreatePayment：创建 payment 记录 → 调用第三方 → 更新 payment（非原子操作）
  - `internal/usecase/refund_usecase.go:43-117` — CreateRefund：查询 payment → 创建 refund → 调用第三方 → 更新 refund → 更新 payment（多步无事务）
  - `internal/usecase/payment_usecase.go:123-164` — HandleWechatNotify：更新 payment → 更新 order（两步更新无事务）
- **问题**: 所有涉及多表/多步更新的业务操作均无事务包裹。任何中间步骤失败都会导致数据不一致。
- **建议**:
  1. Repository 层增加 `WithTransaction(fn func(ctx context.Context) error) error` 方法
  2. Usecase 层将多步操作包裹在事务中：
     ```go
     err := u.orderRepo.WithTransaction(ctx, func(ctx context.Context) error {
         // 所有操作使用同一 ctx 中的 tx
     })
     ```

### 8.2 N+1 查询风险

#### [info] `PaymentRepository.GetByOrderID` 无 N+1 风险
- **文件**: `internal/infrastructure/persistence/postgresql/payment_repo.go:62-72`
- **评价**: 单次查询返回所有 payments，无 N+1。

#### [info] `OrderRepository.List` / `PaymentRepository.List` / `RefundRepository.List` 无 N+1 风险
- **评价**: 分页查询使用 `Find(&ms)` 一次性获取，无关联预加载，无 N+1。

### 8.3 其他发现

#### [warning] `Updates` 可能更新零值
- **文件**: `internal/infrastructure/persistence/postgresql/order_repo.go:51-57`
- **问题**: `Updates(m)` 会更新 model 中的所有字段，包括零值（如空字符串、0）。若 entity 中某些字段为默认值，可能意外覆盖数据库中的已有值。
- **建议**: 使用 `Select` 指定更新字段，或在 model 转换时明确只更新允许修改的字段。更安全的做法是在 entity 层明确变更字段：
  ```go
  db.Model(&OrderModel{}).Where("id = ?", order.ID).Updates(map[string]any{
      "status": order.Status,
      "updated_at": time.Now(),
  })
  ```

#### [warning] `UpdateStatus` 未更新 `updated_at`
- **文件**: `internal/infrastructure/persistence/postgresql/order_repo.go:59-64`
- **问题**: `Update("status", ...)` 只更新 status 字段，GORM 的 `UpdatedAt` 不会自动更新（因为不是 `Save` 或 `Updates` 调用）。
- **建议**: 显式更新 `updated_at`：
  ```go
  .Update("updated_at", gorm.Expr("NOW()"))
  ```
  或启用 GORM 的 `UpdateColumn` 配合自动时间戳。

---

## 9. 安全审计（快速检查）

### 9.1 发现项

#### [warning] `alipay/client.go:68-70` 公钥读取无错误处理
- 已在 4.2 节提及。

#### [warning] `wechat/provider.go:35` 硬编码 appid
- **文件**: `internal/infrastructure/payment/wechat/provider.go:35`
- **问题**: `"wx_appid_placeholder"` 为硬编码占位符，生产环境会导致支付失败。
- **建议**: 从配置读取 appid。

#### [warning] Webhook 通知未做幂等性保护
- **文件**: `internal/usecase/payment_usecase.go:143-145`
- **问题**: 虽然检查了 `payment.Status == PaymentStatusSuccess` 后返回 nil（防止重复处理），但未记录通知日志/去重表。若同一通知在极短时间内并发到达，可能产生竞态条件。
- **建议**: 使用数据库唯一索引（payment_no + notify_id）或分布式锁保证幂等。

---

## 10. 其他发现

### 10.1 发现项

#### [info] `config.go` 中 `RedisConfig` 定义但未使用
- **文件**: `internal/pkg/config/config.go:15,53-58`
- **问题**: Redis 配置结构存在，但代码中无 Redis 客户端初始化或使用。
- **建议**: 若暂不需要，删除 RedisConfig 以减少配置复杂度。

#### [info] `wire.go` 中 `newApp` 未定义
- **文件**: `cmd/api/wire.go:39`
- **问题**: `newApp` 函数在 wire.go 中被引用，但代码中不存在。wire generate 会失败。
- **建议**: 补充 `newApp` 函数或修正 wire 图。

#### [info] `OrderHandler.toOrderResponse` 中 `PaidAt` 转换重复
- **文件**: `internal/interfaces/http/handler/order_handler.go:72-90`
- **问题**: 指针转 int64 指针的逻辑可提取为通用 helper。
- **建议**: 在 dto 包中添加 `TimeToInt64Ptr(t *time.Time) *int64` 辅助函数。

---

## 11. 评分与结论

### 评分维度

| 维度 | 权重 | 得分 | 说明 |
|------|------|------|------|
| Clean Architecture | 15% | 9/10 | 依赖方向正确，wire 工具未实际使用扣 1 分 |
| 设计精简度 | 15% | 7/10 | 存在 timeutil、constants 等冗余包，wire/main 双轨 |
| 接口隔离 | 15% | 8/10 | 接口粒度合理，VerifyNotify 签名对支付宝不自然 |
| 错误处理 | 15% | 6/10 | 多处错误被静默忽略，错误码复用不当 |
| 命名规范 | 10% | 8/10 | 整体良好，工厂函数命名和编号生成可优化 |
| 代码重复 | 10% | 7/10 | provider 重复模式可提取，repository 重复可接受 |
| 空值/边界处理 | 10% | 6/10 | webhook body 无限制，timeout middleware 有泄漏风险 |
| GORM 使用 | 10% | 5/10 | **无事务保护**是最大问题，Updates 零值风险 |

### 总体评分

**加权总分: 7.0 / 10**

### 是否通过

**有条件通过** —— 代码结构良好，Clean Architecture 合规，但存在以下必须修复的问题后方可视为高质量代码：

1. **【critical】为关键业务操作添加数据库事务保护**（payment 创建、refund 创建、通知处理）
2. **【warning】修复静默忽略错误的代码路径**（notification handler 中的 order 更新、refund 后的 payment 状态更新）
3. **【warning】修复 refund 成功后错误地将 payment 标记为 Failed 的问题**
4. **【warning】webhook body 添加大小限制**
5. **【warning】移除或统一 wire.go 与 main.go 的依赖注入方式**
6. **【info】删除未使用的包**（`pkg/constants`、`internal/pkg/pagination`、`internal/pkg/timeutil` 若未引用）

---

*报告生成完毕。*

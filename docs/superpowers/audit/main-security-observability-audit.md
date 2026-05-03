# Security & Observability Audit Report

**Project:** gopay  
**Branch:** main  
**Auditor:** Security & Observability Agent  
**Date:** 2026-05-16  
**Scope:** Full codebase — OWASP Top 10, Payment Security, Input Validation, Concurrency Safety, Resource Safety, Observability

---

## Executive Summary

| Category | Findings |
|----------|----------|
| Critical | 3 |
| Warning | 9 |
| Info | 5 |
| **Total** | **17** |

**Overall Score:** 5.5 / 10  
**Verdict:** **NOT PASS** — Multiple critical payment security gaps and missing observability instrumentation must be resolved before production deployment.

---

## 1. OWASP Top 10

### CRITICAL-1: Webhook Body Size Unbounded — Potential DoS
- **Severity:** critical
- **File:** `internal/interfaces/http/handler/webhook_handler.go:22`
- **Issue:** `io.ReadAll(c.Request.Body)` reads the entire request body without any size limit. An attacker can send an arbitrarily large payload causing OOM.
- **Fix:** Use `io.LimitReader(c.Request.Body, maxBodySize)` before `io.ReadAll`, e.g. `maxBodySize = 1 << 20` (1MB).

### CRITICAL-2: Alipay Notify Signature Verification Bypassed in Provider
- **Severity:** critical
- **File:** `internal/infrastructure/payment/alipay/provider.go:166-171`
- **Issue:** `provider.VerifyNotify()` simply returns `headers` without performing any signature verification. The actual verification (`ParseAlipayNotify` + `VerifySign`) is implemented as a standalone function but never called by the usecase layer. This means any forged Alipay notify will be accepted.
- **Fix:** Move `ParseAlipayNotify` logic into `provider.VerifyNotify`, or have the usecase call `alipay.ParseAlipayNotify` before processing. Ensure the public key is loaded securely.

### WARNING-1: DSN Construction via String Concatenation (Low Risk)
- **Severity:** warning
- **File:** `internal/pkg/config/config.go:48-51`
- **Issue:** `DSN()` uses `fmt.Sprintf` with raw config values. While GORM/PostgreSQL driver does not support stacked queries, special characters in password could cause connection failures or unexpected behavior.
- **Fix:** Use `url.URL` with `UserPassword` escaped, or rely on GORM's connection string parser which handles escaping.

### WARNING-2: CORS Allows All Origins
- **Severity:** warning
- **File:** `internal/interfaces/http/middleware/cors.go:11`
- **Issue:** `Access-Control-Allow-Origin: *` allows any origin to make cross-origin requests to payment endpoints. Combined with missing authentication middleware, this increases attack surface.
- **Fix:** Restrict to known frontend domains in production; add authentication middleware.

---

## 2. Payment Security

### CRITICAL-3: Missing Anti-Replay Protection for Webhooks
- **Severity:** critical
- **File:** `internal/usecase/payment_usecase.go:123-207`
- **Issue:** Neither WeChat nor Alipay webhook handlers implement replay-attack protection. There is no nonce/timestamp validation, no notification ID deduplication, and no `CodeWebhookDuplicate` logic is actually used. An attacker can replay the same notify request multiple times to trigger duplicate state transitions.
- **Fix:**
  1. Add a unique constraint on `(payment_no, third_party_no, notify_time)` or maintain a `processed_notifications` table.
  2. Validate WeChat timestamp (`Wechatpay-Timestamp`) against current time (±5 min).
  3. Validate Alipay notify ID uniqueness.

### WARNING-3: Payment Amount Not Verified Against Order Amount
- **Severity:** warning
- **File:** `internal/usecase/payment_usecase.go:53-117`
- **Issue:** When creating a payment, `payment.Amount` is set directly from `order.Amount` without any additional verification. While this is internally consistent, the webhook handlers (`HandleWechatNotify`, `HandleAlipayNotify`) do not verify the actual paid amount returned by the payment provider against the expected order amount. A provider bug or MITM could result in partial payment being treated as full payment.
- **Fix:** In webhook handlers, compare the notified amount with `payment.Amount` before marking success.

### WARNING-4: Order Status Machine Race Condition
- **Severity:** warning
- **File:** `internal/usecase/payment_usecase.go:143-164`, `186-207`
- **Issue:** The check `if payment.Status == entity.PaymentStatusSuccess { return nil }` and subsequent update are not atomic. Two concurrent webhook requests for the same payment can both pass the check and both attempt to update the order, potentially causing double inventory/release issues downstream.
- **Fix:** Use database-level optimistic locking (`version` column) or `UPDATE ... WHERE status = pending` pattern. Wrap in a transaction with `SELECT FOR UPDATE`.

### WARNING-5: Refund Status Immediately Set to Success
- **Severity:** warning
- **File:** `internal/usecase/refund_usecase.go:103-108`
- **Issue:** After calling `provider.Refund()`, the refund status is immediately set to `RefundStatusSuccess` without waiting for async notification or querying the refund status. If the provider refund fails asynchronously, the local state will be incorrect.
- **Fix:** Set status to `RefundStatusProcessing` after provider call; update to `Success` only upon webhook notification or explicit status query.

### WARNING-6: Refund Updates Payment Status to Failed on Full Refund
- **Severity:** warning
- **File:** `internal/usecase/refund_usecase.go:111-114`
- **Issue:** When total refunded >= payment amount, the payment status is updated to `PaymentStatusFailed`. This is semantically incorrect — a fully refunded payment should transition to `PaymentStatusRefunded` (or similar), not "Failed".
- **Fix:** Introduce `PaymentStatusRefunded` and use it here.

### INFO-1: WeChat Notify Signature Headers Not Fully Verified
- **Severity:** info
- **File:** `internal/infrastructure/payment/wechat/provider.go:180-200`
- **Issue:** `VerifyNotify` decrypts the ciphertext but does not explicitly verify the `Wechatpay-Signature` header against the response body. The `go-pay` library may handle this internally, but the code does not document or assert this assumption.
- **Fix:** Add explicit signature verification using `wechatv3.VerifySignByPK` or document the library's behavior in a code comment.

### INFO-2: Hardcoded WeChat AppID Placeholder
- **Severity:** info
- **File:** `internal/infrastructure/payment/wechat/provider.go:35`, `76`
- **Issue:** `"wx_appid_placeholder"` is hardcoded. This will cause production payment failures.
- **Fix:** Load `AppID` from config.

---

## 3. Input Validation

### WARNING-7: DTO Validation Gaps
- **Severity:** warning
- **File:** `internal/interfaces/http/dto/payment_dto.go:10-11`
- **Issue:** `OpenID` and `BuyerID` lack validation tags. `OpenID` can be up to 128 chars; `BuyerID` similarly. No `max` or `regexp` constraints.
- **Fix:** Add `binding:"omitempty,max=128"` or appropriate validation.

### WARNING-8: OrderNo / PaymentNo / RefundNo Path Params Lack Format Validation
- **Severity:** warning
- **File:** `internal/interfaces/http/handler/order_handler.go:42`, `58`; `payment_handler.go:54`; `refund_handler.go:54`
- **Issue:** Path parameters are only checked for non-emptiness. No regex validation (e.g., `ORD\d{14}\d{4}`) to prevent injection or enumeration.
- **Fix:** Add format validation in handler or usecase layer.

### INFO-3: Webhook Headers Not Validated
- **Severity:** info
- **File:** `internal/interfaces/http/handler/webhook_handler.go:29-34`
- **Issue:** WeChat webhook headers are collected without checking presence. Empty `Wechatpay-Signature` will be passed downstream and may cause silent verification failures.
- **Fix:** Check required headers exist and return 400 early if missing.

---

## 4. Concurrency Safety

### WARNING-9: Timeout Middleware Goroutine Leak on Panic
- **Severity:** warning
- **File:** `internal/interfaces/http/middleware/timeout.go:11-33`
- **Issue:** The timeout middleware spawns a goroutine that calls `c.Next()`. If the handler panics and is recovered by `Recovery()` middleware, the `done` channel may never be closed, causing a goroutine leak.
- **Fix:** Use `defer close(done)` or better, use Gin's built-in `c.Request.Context()` timeout without a separate goroutine.

### INFO-4: Global Validator Singleton is Race-Safe but Not Customized
- **Severity:** info
- **File:** `internal/pkg/validator/validator.go:8-12`
- **Issue:** The global `validate` is initialized in `init()` and never modified. This is safe, but lacks custom validators (e.g., for order number format).
- **Fix:** Register custom validators for domain-specific formats.

---

## 5. Resource Safety

### INFO-5: Database Connection Pool Configured but No Health Check
- **Severity:** info
- **File:** `internal/infrastructure/persistence/postgresql/db.go:15-44`
- **Issue:** Connection pool settings are applied, but there is no periodic health check or ping to detect stale connections.
- **Fix:** Add `sqlDB.Ping()` at startup and optionally a background ping loop.

### INFO-6: HTTP Client Timeout Not Explicitly Set for Payment SDKs
- **Severity:** info
- **File:** `internal/infrastructure/payment/wechat/client.go:27`; `alipay/client.go:26`
- **Issue:** The `go-pay` clients are initialized without explicit HTTP client timeout configuration. Default timeouts may be too long or too short.
- **Fix:** Configure custom `http.Client` with explicit timeouts and pass to SDK if supported.

---

## 6. Observability

### WARNING-10: Missing Metrics Instrumentation
- **Severity:** warning
- **File:** Multiple — `usecase/payment_usecase.go`, `usecase/refund_usecase.go`, `usecase/order_usecase.go`
- **Issue:** No metrics (counters, histograms) are emitted for:
  - Payment created / succeeded / failed counts
  - Refund created / succeeded / failed counts
  - Order created / closed counts
  - Webhook received / processed / failed counts
  - Third-party API latency
- **Fix:** Integrate Prometheus client. Add counters like `payments_total{status="success|failed", channel="wechat|alipay"}` and histograms for provider API latency.

### WARNING-11: Missing Distributed Tracing
- **Severity:** warning
- **File:** Entire codebase
- **Issue:** No OpenTelemetry or similar tracing is integrated. Payment flows (order -> payment -> provider -> webhook) cannot be correlated across services.
- **Fix:** Add OpenTelemetry middleware for Gin and propagate trace context through usecase -> repository -> provider layers.

### WARNING-12: Insufficient Structured Logging in Success Paths
- **Severity:** warning
- **File:** `internal/usecase/payment_usecase.go`, `refund_usecase.go`, `order_usecase.go`
- **Issue:** Only error paths have logs. Success paths (e.g., payment created, order paid) are not logged, making debugging and audit difficult.
- **Fix:** Add `Info` level logs for all state transitions with relevant IDs (order_no, payment_no, refund_no).

### WARNING-13: Webhook Handler Logs Do Not Include Key Identifiers
- **Severity:** warning
- **File:** `internal/interfaces/http/handler/webhook_handler.go:24`, `37`, `47`, `60`
- **Issue:** Error logs only include the raw error. They do not include `out_trade_no`, `trade_status`, or request headers, making it impossible to correlate failures with specific transactions.
- **Fix:** Add structured fields (`zap.String("out_trade_no", ...)`) to all webhook logs.

### INFO-7: Logger May Leak Sensitive Data
- **Severity:** info
- **File:** `internal/pkg/config/config.go:42`; `internal/infrastructure/payment/wechat/client.go:24`
- **Issue:** `config.go:42` logs DSN which includes password (line 42 in db.go actually logs host:port/dbname, not password — but DSN construction includes it). More critically, payment request/response bodies and notify payloads may be logged if debug mode is enabled.
- **Fix:** Ensure `DebugSwitch` is never enabled in production. Add log redaction for sensitive fields (API keys, certificates, user PII).

### INFO-8: Request Logger Does Not Log Request Body
- **Severity:** info
- **File:** `internal/interfaces/http/middleware/logger.go:11-45`
- **Issue:** The access log middleware captures method, path, status, latency but not request/response bodies. For payment debugging, this is insufficient.
- **Fix:** Optionally log request body for non-webhook endpoints (with PII redaction). Use a separate audit log for payment-critical operations.

---

## Appendix: File Checklist

| File | Status |
|------|--------|
| `internal/infrastructure/payment/wechat/client.go` | Reviewed |
| `internal/infrastructure/payment/wechat/provider.go` | Reviewed |
| `internal/infrastructure/payment/alipay/client.go` | Reviewed |
| `internal/infrastructure/payment/alipay/provider.go` | Reviewed |
| `internal/infrastructure/persistence/postgresql/db.go` | Reviewed |
| `internal/infrastructure/persistence/postgresql/order_repo.go` | Reviewed |
| `internal/infrastructure/persistence/postgresql/payment_repo.go` | Reviewed |
| `internal/infrastructure/persistence/postgresql/refund_repo.go` | Reviewed |
| `internal/interfaces/http/handler/order_handler.go` | Reviewed |
| `internal/interfaces/http/handler/payment_handler.go` | Reviewed |
| `internal/interfaces/http/handler/refund_handler.go` | Reviewed |
| `internal/interfaces/http/handler/webhook_handler.go` | Reviewed |
| `internal/interfaces/http/middleware/request_id.go` | Reviewed |
| `internal/interfaces/http/middleware/cors.go` | Reviewed |
| `internal/interfaces/http/middleware/logger.go` | Reviewed |
| `internal/interfaces/http/middleware/recovery.go` | Reviewed |
| `internal/interfaces/http/middleware/timeout.go` | Reviewed |
| `internal/usecase/order_usecase.go` | Reviewed |
| `internal/usecase/payment_usecase.go` | Reviewed |
| `internal/usecase/refund_usecase.go` | Reviewed |
| `internal/pkg/config/config.go` | Reviewed |
| `internal/pkg/logger/logger.go` | Reviewed |
| `internal/pkg/apperror/error.go` | Reviewed |
| `internal/pkg/apperror/codes.go` | Reviewed |
| `cmd/api/main.go` | Reviewed |

---

*Report generated by Security & Observability Audit Agent.*

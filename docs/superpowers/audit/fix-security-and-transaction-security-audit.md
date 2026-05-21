# Security Audit Report — gopay Payment Gateway

**Branch:** `fix/security-and-transaction`  
**Date:** 2026-05-19  
**Auditor:** Security Audit Agent  
**Scope:** Full Go backend (Clean Architecture), with emphasis on payment infrastructure, webhooks, usecases, middleware, and configuration.

---

## Executive Summary

This audit reviewed the gopay payment gateway codebase against OWASP Top 10 and payment-industry security requirements. **3 critical, 4 high, 6 medium, and 4 low severity findings** were identified. The most severe issues involve missing webhook signature verification, weak CORS defaults, missing nonce/replay protection, and insufficient transaction isolation for concurrent payment notifications. All findings are actionable and should be addressed before production deployment.

---

## Findings Table

| # | Severity | Location | Category | Issue Description | Risk Analysis | Fix Recommendation |
|---|----------|----------|----------|-------------------|---------------|-------------------|
| 1 | **critical** | `internal/infrastructure/payment/wechat/provider.go:187-219` | Signature Verification | `VerifyNotify` decrypts the WeChat payload using `APIV3Key` but **does NOT verify the `Wechatpay-Signature` header** against the platform certificate. | An attacker can forge payment notifications by sending arbitrary encrypted payloads. The timestamp check alone does not cryptographically prove the message came from WeChat. | Use `wechatv3.V3ParseNotify` + `client.VerifySignByCert` (or `client.VerifySign`) to verify the signature before decryption. |
| 2 | **critical** | `internal/infrastructure/payment/alipay/provider.go:167-202` | Signature Verification | `VerifyNotify` parses URL values and calls `alipay.VerifySign`, but the **public key is read from disk on every invocation** (`PublicKey()` in `client.go:64-70`) with **silent error swallowing** (`data, _ := os.ReadFile(...)`). | If the public key file becomes unreadable (permissions, deletion), verification silently fails with an empty key, allowing signature bypass. | Cache the public key in memory at client initialization; return error if unreadable. Ensure `VerifySign` is called with a non-empty key. |
| 3 | **critical** | `internal/interfaces/http/middleware/cors.go:9-42` | CSRF / CORS | When `origins` is empty, the middleware **reflects the incoming `Origin` header back** (`allowedOrigin = origin`) without validation. | Any website can make cross-origin requests to the payment API. Combined with missing CSRF tokens, this enables malicious sites to trigger payments/refunds on behalf of authenticated users. | Default to same-origin only in production. Never reflect arbitrary origins. Add `Vary: Origin` header. |
| 4 | **high** | `internal/usecase/payment_usecase.go:138-199` | Idempotency / Replay | `HandleWechatNotify` checks `payment.Status == PaymentStatusSuccess` for duplicates, but **does not use a unique index or atomic compare-and-swap**. Concurrent notifications for the same payment can race. | Two simultaneous webhooks for the same out_trade_no could both pass the duplicate check, update the payment to success, and trigger duplicate order fulfillment. | Add DB-level unique constraint on `(payment_no, status=success)` or use `SELECT FOR UPDATE` within the transaction. Consider an idempotency key table. |
| 5 | **high** | `internal/usecase/payment_usecase.go:202-264` | Idempotency / Replay | `HandleAlipayNotify` has the same race condition as WeChat notify handler. | Same as #4 — concurrent Alipay notifications can cause duplicate order status updates. | Same fix as #4. |
| 6 | **high** | `internal/infrastructure/payment/wechat/provider.go:192-201` | Anti-Replay | Timestamp check allows **±5 minutes** but does **not validate the `Wechatpay-Nonce` header** against a nonce cache. | Replay attacks within the 5-minute window are possible. An attacker can replay a valid notification multiple times. | Maintain a short-lived cache (e.g., Redis, 10 min TTL) of seen nonces. Reject duplicates. |
| 7 | **high** | `internal/usecase/refund_usecase.go:48-135` | Transaction Safety | `Create` does not wrap the refund record creation + provider API call in a transaction. If the provider succeeds but the local `Update` fails, state becomes inconsistent. | Partial refund state: money returned but system shows "processing". Also, `GetTotalRefundAmount` is called twice (before and after) without locking, allowing concurrent refunds to exceed the payment amount. | Wrap the entire refund flow in `txMgr.WithTransaction`. Use `SELECT FOR UPDATE` on the payment row when computing remaining refundable amount. |
| 8 | **medium** | `internal/infrastructure/payment/wechat/client.go:22-24` | Key Management | Private key is read from filesystem path configured in YAML. No runtime check for file permissions (e.g., `0400`). | If the private key file has world-readable permissions (`644`), any local user can exfiltrate the merchant private key. | After reading the key, `os.Stat` the file and enforce `mode.Perm() & 0077 == 0`. Log a warning/error if permissions are too open. |
| 9 | **medium** | `internal/infrastructure/payment/alipay/client.go:21-23` | Key Management | Same as #8 — Alipay private key read without permission checks. | Same risk as #8. | Same fix as #8. |
| 10 | **medium** | `internal/pkg/config/config.go:100-136` | Configuration | `Load` does not validate that `APIV3Key` or private key paths are non-empty when `Enabled=true`. | Misconfiguration can lead to runtime panics or silent failures (e.g., empty APIV3Key causes decryption to fail opaquely). | Add validation in `Load`: if `Wechat.Enabled`, require `AppID`, `MchID`, `APIV3Key`, `PrivateKeyPath`. If `Alipay.Enabled`, require `AppID`, `PrivateKeyPath`. |
| 11 | **medium** | `internal/usecase/payment_usecase.go:56-132` | Amount Safety | `CreatePayment` copies `order.Amount` directly to the payment without re-validation. The order amount is validated at creation, but there is no check that the payment amount matches the current order state. | If order amount is modified between creation and payment (e.g., by admin or race condition), the payment could be for a different amount. | Re-fetch the order within the payment creation flow and assert `order.Amount == payment.Amount`. |
| 12 | **medium** | `internal/usecase/order_usecase.go:29-66` | Input Validation | `Create` accepts `subject`, `description` from user input without HTML/sanitization. While DTO has `max` length, no XSS filtering is applied. | If these fields are rendered in the frontend or returned in API responses without escaping, stored XSS is possible. | Sanitize or escape user-provided `subject` and `description` before storage. Ensure frontend uses proper escaping. |
| 13 | **medium** | `internal/interfaces/http/handler/webhook_handler.go:23-54` | Input Validation | `WechatNotify` reads body with `io.LimitReader(..., 1MB)` but does not check if the body was truncated. | A 1MB+ body would be silently truncated, potentially causing JSON unmarshal to fail or process partial data. | Check `len(body) == maxWebhookBodySize` and reject if exact limit hit (indicates truncation). |
| 14 | **low** | `internal/interfaces/http/middleware/cors.go:32-33` | CSRF / CORS | `Access-Control-Allow-Headers` includes `Authorization` but `Access-Control-Allow-Credentials` is never set. | Cookies / Authorization headers cannot be sent cross-origin anyway, but the header list implies credentials support. Either add `Allow-Credentials: true` (with strict origin validation) or remove `Authorization` from allowed headers. | Add `Access-Control-Allow-Credentials: true` only when origin is explicitly allowed and not `*`. |
| 15 | **low** | `internal/interfaces/http/middleware/timeout.go:11-33` | Timeout / Resource Leak | `Timeout` middleware spawns a goroutine for every request but does not handle the case where `c.Next()` panics — the `done` channel may never close. | A panicking handler could leak the goroutine and leave the timeout select hanging until the context deadline. | Add `defer recover()` inside the goroutine or use `gin`'s built-in timeout middleware which handles this. |
| 16 | **low** | `internal/infrastructure/payment/wechat/provider.go:56-57` | Sensitive Data Exposure | Error logs include raw `zap.Error(err)` from the WeChat SDK, which may contain sensitive request details. | Potential leakage of payment parameters, merchant IDs, or partial keys in log aggregation systems. | Review and redact sensitive fields from error logs. Use structured logging with explicit allowed fields. |
| 17 | **low** | `internal/infrastructure/persistence/postgresql/db.go:46` | Sensitive Data Exposure | Database DSN is logged including host and dbname (no password, but still infra info). | Low risk, but reveals internal network topology in logs. | Remove DSN logging or log only sanitized connection info (host:port masked). |

---

## Detailed Analysis by Dimension

### 1. Input Validation
- **SQL Injection:** GORM parameterized queries are used throughout (`Where("payment_no = ?", paymentNo)`). No raw SQL concatenation found. **Clean.**
- **Command Injection:** No shell execution or `os/exec` usage. **Clean.**
- **Path Traversal:** `config.go` reads file paths from config YAML. No path sanitization, but paths are controlled by deployment config, not user input. **Acceptable with deployment controls.**
- **XSS:** `subject` and `description` are stored as-is. Frontend responsibility to escape, but backend should sanitize. See finding #12.

### 2. Certificate / Key Management
- Private keys are file-based, not hardcoded. **Good.**
- No file permission checks. See findings #8, #9.
- `APIV3Key` is stored in YAML config. Ensure this file has `0600` permissions in production.

### 3. Signature Verification
- **WeChat:** Missing signature verification entirely. See finding #1 (critical).
- **Alipay:** Signature verification present but brittle due to on-demand key loading with swallowed errors. See finding #2 (critical).

### 4. Idempotency
- Duplicate payment check is status-based, not atomic. See findings #4, #5 (high).
- No idempotency key for `CreatePayment` or `CreateRefund` API calls. Retrying a failed request could create duplicate payments.

### 5. Amount Safety
- Amounts stored as `int64` (cents). No floating-point precision issues. **Good.**
- Refund amount checked against payment amount and total refunded. See finding #7 for concurrency gap.
- No verification that payment amount matches order amount at payment time. See finding #11.

### 6. Transaction Safety
- Payment notification handlers use `txMgr.WithTransaction`. **Good.**
- Refund creation does NOT use transaction manager. See finding #7.
- Order status transitions (`MarkPaid`, `MarkClosed`) have entity-level guards but no DB-level constraints.

### 7. Sensitive Information Leakage
- Error logs may contain SDK-internal errors. See finding #16.
- Config does not appear to log keys, but DSN is partially logged. See finding #17.
- `ThirdPartyResp` is stored in DB as JSONB. Ensure this does not contain PII or keys.

### 8. CSRF / CORS
- CORS middleware reflects arbitrary origins when config is empty. See finding #3 (critical).
- No CSRF token mechanism for state-changing API calls (`POST /payments`, `POST /refunds`). Relies on CORS + Authorization header, but CORS is broken. See finding #3.

### 9. Timeout and Retry Strategy
- HTTP server has `ReadTimeout`, `WriteTimeout`, `IdleTimeout`. **Good.**
- Gin middleware timeout of 30s. See finding #15 for goroutine leak risk.
- No retry policy for payment provider API calls. Network blips could fail payments without recovery.

### 10. Anti-Replay
- WeChat has timestamp check (±5 min) but no nonce tracking. See finding #6 (high).
- Alipay has no timestamp or nonce check at all.

---

## Recommendations Priority Matrix

| Priority | Findings | Action |
|----------|----------|--------|
| **P0 — Block Production** | #1, #2, #3 | Fix immediately. These are exploitable vulnerabilities. |
| **P1 — High Risk** | #4, #5, #6, #7 | Fix before public launch. Race conditions and replay attacks. |
| **P2 — Medium Risk** | #8, #9, #10, #11, #12, #13 | Address in next sprint. Hardening and defense in depth. |
| **P3 — Low Risk** | #14, #15, #16, #17 | Nice-to-have improvements. |

---

## Appendix: Files Reviewed

- `internal/infrastructure/payment/wechat/provider.go`
- `internal/infrastructure/payment/alipay/provider.go`
- `internal/infrastructure/payment/wechat/client.go`
- `internal/infrastructure/payment/alipay/client.go`
- `internal/interfaces/http/handler/webhook_handler.go`
- `internal/usecase/payment_usecase.go`
- `internal/usecase/refund_usecase.go`
- `internal/usecase/order_usecase.go`
- `internal/interfaces/http/middleware/*`
- `internal/pkg/config/config.go`
- `internal/interfaces/http/router/router.go`
- `internal/domain/entity/*`
- `internal/infrastructure/persistence/postgresql/*`
- `cmd/api/main.go`
- `configs/config.dev.yaml`

---

*End of Report*

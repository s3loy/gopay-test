# Payment Link Audit Report

**Branch**: `fix/security-and-transaction`  
**Scope**: WeChat Pay V3 + Alipay V3 payment, refund, webhook, and transaction flow  
**Date**: 2026-05-19  
**Auditor**: Payment Audit Agent

---

## Summary

| Priority | Count | Description |
|----------|-------|-------------|
| P0-blocker | 3 | Critical bugs that will cause payment/refund failures or security vulnerabilities |
| P1-major | 5 | Significant issues affecting correctness, reliability, or compliance |
| P2-minor | 4 | Code quality, observability, or edge-case improvements |

---

## P0 - Blocker Issues

### P0-1: WeChat Refund `amount.total` incorrectly set to refund amount instead of original order amount

- **File**: `internal/infrastructure/payment/wechat/provider.go:141-145`
- **Code**:
  ```go
  SetBodyMap("amount", func(bm gopay.BodyMap) {
      bm.Set("refund", req.Amount).
          Set("total", req.Amount).   // BUG: should be original order amount
          Set("currency", "CNY")
  })
  ```
- **Problem**: `amount.total` must be the **original order total amount**, not the refund amount. WeChat V3 API requires `total` to match the original order's total amount for validation. Passing the refund amount here will cause the refund API to reject the request with an amount mismatch error.
- **Impact**: All WeChat refunds will fail with API error.
- **Fix**: Pass the original payment amount as `total`. The `ProviderRefundRequest` struct needs to include the original payment amount, or the usecase must provide it.
  ```go
  bm.Set("refund", req.Amount).
      Set("total", req.OriginalAmount). // original order amount
  ```

### P0-2: WeChat notify signature verification is completely missing

- **File**: `internal/infrastructure/payment/wechat/provider.go:187-219`
- **Code**: `VerifyNotify` only parses the notify body, checks timestamp anti-replay, and decrypts the ciphertext. It **never verifies the WeChatPay-Signature** against the platform certificate.
- **Problem**: Per gopay SDK docs, WeChat V3 async notifications **must** be verified using `notifyReq.VerifySignByPKMap(certMap)` before decryption. Without signature verification, an attacker can forge payment notifications and mark orders as paid.
- **Impact**: **Critical security vulnerability** — anyone can forge WeChat payment success notifications and trigger order fulfillment without actual payment.
- **Fix**: Add signature verification before decryption:
  ```go
  certMap := p.client.V3().WxPublicKeyMap()
  if err := notifyReq.VerifySignByPKMap(certMap); err != nil {
      return nil, apperror.Wrap(err, apperror.CodeWebhookInvalidSignature)
  }
  result, err := notifyReq.DecryptPayCipherText(p.client.cfg.APIV3Key)
  ```

### P0-3: Alipay notify verification uses wrong public key in cert mode

- **File**: `internal/infrastructure/payment/alipay/provider.go:167-202`
- **Code**: `VerifyNotify` always calls `alipay.VerifySign(publicKey, notifyReq)` (public key mode), regardless of whether the client was initialized with certificates.
- **Problem**: When `SetCert()` is called during client initialization (cert mode), Alipay notifications **must** be verified with `alipay.VerifySignWithCert(aliPublicCert, notifyReq)`. Using `VerifySign` with the app public key in cert mode will fail verification.
- **Impact**: Alipay notifications will fail verification in production cert mode, or the system will incorrectly reject legitimate notifications. Conversely, if running in public-key mode but using cert-mode verification logic, it would also fail.
- **Fix**: Detect cert mode (e.g., store a flag in `Client` when `SetCert` succeeds) and branch:
  ```go
  if p.client.IsCertMode() {
      ok, err = alipay.VerifySignWithCert(p.client.PublicCert(), notifyReq)
  } else {
      ok, err = alipay.VerifySign(p.client.PublicKey(), notifyReq)
  }
  ```

---

## P1 - Major Issues

### P1-1: WeChat `V3TransactionQueryOrder` called with `thirdPartyNo` but uses `OutTradeNo` enum

- **File**: `internal/infrastructure/payment/wechat/provider.go:113`
- **Code**:
  ```go
  wxRsp, err := p.client.V3().V3TransactionQueryOrder(ctx, wechatv3.OutTradeNo, thirdPartyNo)
  ```
- **Problem**: The `QueryPayment` interface receives `thirdPartyNo` (WeChat's `transaction_id`), but passes `wechatv3.OutTradeNo` as the query type. This means it will query by merchant order number (`out_trade_no`) using the `transaction_id` value, which will fail if the two values differ (which they always do).
- **Impact**: WeChat payment queries will return "order not found" errors.
- **Fix**: Use `wechatv3.TransactionId` when the parameter is a WeChat transaction ID, or change the interface to accept `outTradeNo` and document it clearly.
  ```go
  // Option A: if thirdPartyNo is transaction_id
  wxRsp, err := p.client.V3().V3TransactionQueryOrder(ctx, wechatv3.TransactionId, thirdPartyNo)
  // Option B: change interface to accept outTradeNo for queries
  ```

### P1-2: Payment creation is not atomic — DB record exists before API call

- **File**: `internal/usecase/payment_usecase.go:92-114`
- **Code**: `paymentRepo.Create()` is called **before** `provider.CreatePayment()`. If the provider API call fails, the payment record remains in "pending" status in the database.
- **Problem**: This creates orphaned pending payments. While the code updates status to "failed" on error, there's a race condition where concurrent requests could create multiple payments for the same order. Also, the order's payment state is not checked before creating a new payment.
- **Impact**: Database pollution with failed payment records; potential duplicate payment creation for the same order.
- **Fix**: Wrap the entire flow in a transaction, or check for existing pending payments for the same order before creating a new one. Alternatively, create the payment record only after a successful API response (but then you lose the audit trail). The better approach:
  1. Check if an active pending payment already exists for this order.
  2. If yes, return the existing payment (idempotency).
  3. If no, create payment record and call provider API within a transaction, with rollback on failure.

### P1-3: Alipay `TradeCreate` for WAP/APP uses wrong `product_code`

- **File**: `internal/infrastructure/payment/alipay/provider.go:64`
- **Code**:
  ```go
  bm.Set("product_code", "QUICK_MSECURITY_PAY")
  ```
- **Problem**: `QUICK_MSECURITY_PAY` is the product code for **APP** payments (mobile app). For **WAP** payments, the correct product code is `QUICK_WAP_WAY`. Using the wrong product code may cause Alipay to reject WAP payment requests or return incorrect payment parameters.
- **Impact**: WAP payments may fail or behave incorrectly.
- **Fix**: Branch by method:
  ```go
  switch req.Method {
  case entity.MethodWAP:
      bm.Set("product_code", "QUICK_WAP_WAY")
  case entity.MethodAPP:
      bm.Set("product_code", "QUICK_MSECURITY_PAY")
  }
  ```

### P1-4: Alipay `QueryRefund` uses `out_request_no` but should also support `trade_no`

- **File**: `internal/infrastructure/payment/alipay/provider.go:142-165`
- **Code**: `QueryRefund` only sets `out_request_no` (refund request number). However, Alipay's `TradeFastPayRefundQuery` requires either `out_request_no` + `out_trade_no` or `trade_no`.
- **Problem**: Without `out_trade_no`, Alipay may not be able to locate the refund record, especially in edge cases.
- **Impact**: Refund queries may fail intermittently.
- **Fix**: Also pass `out_trade_no` (the original payment's `PaymentNo`) in the query:
  ```go
  bm.Set("out_trade_no", req.PaymentNo).
      Set("out_request_no", thirdPartyNo)
  ```
  This requires the `QueryRefund` interface to accept the original payment number, or the usecase to provide it.

### P1-5: `ProviderRefundRequest` lacks `OriginalAmount` field needed for WeChat refund

- **File**: `internal/domain/service/payment_provider.go:30-37`
- **Code**:
  ```go
  type ProviderRefundRequest struct {
      PaymentNo    string
      ThirdPartyNo string
      RefundNo     string
      Amount       int64
      Reason       string
      NotifyURL    string
  }
  ```
- **Problem**: As noted in P0-1, WeChat V3 refund requires the original order total amount in `amount.total`. The current struct only carries the refund amount.
- **Impact**: Blocks the fix for P0-1.
- **Fix**: Add `OriginalAmount int64` field to `ProviderRefundRequest` and populate it in `refund_usecase.go`.

---

## P2 - Minor Issues

### P2-1: WeChat `QueryPayment` returns raw `TradeState` string without mapping to internal status

- **File**: `internal/infrastructure/payment/wechat/provider.go:121-129`
- **Code**: Returns `wxRsp.Response.TradeState` directly as `Status`.
- **Problem**: WeChat states like `NOTPAY`, `CLOSED`, `REVOKED`, `USERPAYING`, `PAYERROR` are passed through as-is. The usecase and domain layer may not handle all these states consistently.
- **Impact**: Potential state machine inconsistencies.
- **Fix**: Map WeChat states to internal `PaymentStatus` values in the provider layer.

### P2-2: Alipay refund status hardcoded to "SUCCESS"

- **File**: `internal/infrastructure/payment/alipay/provider.go:135`
- **Code**:
  ```go
  Status: "SUCCESS",
  ```
- **Problem**: The refund result status is hardcoded instead of using the actual response status. If Alipay returns a failure or pending status, the system incorrectly records it as success.
- **Impact**: Incorrect refund status tracking.
- **Fix**: Use the actual response status if available, or query the refund status after creation.

### P2-3: Webhook handler returns "fail" for all errors, including duplicate notifications

- **File**: `internal/interfaces/http/handler/webhook_handler.go`
- **Problem**: Both WeChat and Alipay handlers return HTTP 400 + "fail" for any error from the usecase. For duplicate notifications (which are harmless), this causes the payment provider to retry unnecessarily.
- **Impact**: Unnecessary webhook retries, increased load.
- **Fix**: Distinguish between verification failures (return 400/fail), duplicate notifications (return 200/success), and processing errors (return 500/fail or 200/success depending on retry strategy).

### P2-4: Missing `NotifyURL` for refunds

- **File**: `internal/usecase/refund_usecase.go:95-102`
- **Code**:
  ```go
  NotifyURL: "",
  ```
- **Problem**: Refund notifications are sent with an empty `notify_url`. If the payment provider supports refund notifications (WeChat does), the system will never receive them.
- **Impact**: Refund status updates rely solely on polling instead of async notifications.
- **Fix**: Configure refund notify URLs in config and pass them through.

---

## Configuration Audit

### WeChat Config (`config.go:77-87`)

| Field | Used In | Status |
|-------|---------|--------|
| `AppID` | `provider.go` CreatePayment, PaySignOfJSAPI | OK |
| `MchID` | `client.go` NewClientV3 | OK |
| `SerialNo` | `client.go` NewClientV3 | OK |
| `APIV3Key` | `client.go` NewClientV3, `provider.go` DecryptPayCipherText | OK |
| `PrivateKeyPath` | `client.go` | OK |
| `PublicKeyPath` | `client.go` AutoVerifySignByCert, `alipay/client.go` PublicKey | **Conflict** — same field name used for different purposes |
| `PublicKeyID` | `client.go` AutoVerifySignByCert | OK |
| `NotifyURL` | `client.go` | OK |

**Note**: `PublicKeyPath` in `AlipayConfig` is overloaded — it stores the Alipay public key for public-key mode verification, but is also used as the Alipay public certificate path in cert mode. This is confusing and error-prone. Consider renaming to `AliPublicCertPath` for cert mode clarity.

### Alipay Config (`config.go:89-98`)

| Field | Used In | Status |
|-------|---------|--------|
| `AppID` | `client.go` NewClientV3 | OK |
| `PrivateKeyPath` | `client.go` | OK |
| `PublicKeyPath` | `provider.go` VerifySign, `client.go` PublicKey | OK for public-key mode |
| `AppCertPath` | `client.go` SetCert | OK |
| `RootCertPath` | `client.go` SetCert | OK |
| `NotifyURL` | `client.go` | OK |
| `IsProd` | `client.go` NewClientV3 | OK |

---

## Transaction Safety Assessment

| Flow | Transaction Boundary | Assessment |
|------|---------------------|------------|
| Payment Create | No transaction | **Risky** — DB record created before API call; orphaned records possible |
| WeChat Notify | `txMgr.WithTransaction` | **OK** — payment + order update atomic |
| Alipay Notify | `txMgr.WithTransaction` | **OK** — payment + order update atomic |
| Refund Create | No transaction | **Risky** — refund record created before API call; similar to payment create |

**Recommendation**: Add transaction boundaries around payment creation and refund creation, or implement idempotency checks to prevent duplicate records.

---

## Amount Safety Assessment

| Check | Location | Status |
|-------|----------|--------|
| Refund amount > 0 | `refund_usecase.go:57-59` | OK |
| Refund amount <= payment amount | `refund_usecase.go:60-62` | OK |
| Cumulative refund <= payment amount | `refund_usecase.go:64-72` | OK |
| WeChat amount unit (fen) | `provider.go` | OK |
| Alipay amount conversion (fen -> yuan) | `alipay/provider.go:231-233` | OK |
| WeChat refund `amount.total` | `wechat/provider.go:143` | **BUG** — see P0-1 |

---

## Recommendations

1. **Fix P0 issues immediately** before any production deployment. The missing WeChat signature verification is a critical security flaw.
2. **Add WeChat signature verification** using `VerifySignByPKMap` with platform certificates.
3. **Fix Alipay cert-mode verification** by detecting the initialization mode and using the appropriate verify function.
4. **Add `OriginalAmount` to `ProviderRefundRequest`** and fix WeChat refund `amount.total`.
5. **Fix WeChat query** to use the correct query type (`TransactionId` vs `OutTradeNo`).
6. **Add transaction boundaries** or idempotency checks for payment and refund creation.
7. **Map external payment states** to internal states consistently in the provider layer.
8. **Distinguish webhook error types** to avoid unnecessary retries on duplicates.
9. **Add refund notify URL configuration** to support async refund status updates.

---

*End of Audit Report*

package apperror

import "net/http"

// System (00xxxx)
const (
	CodeSuccess               = 0
	CodeUnknown               = 10001
	CodeInternal              = 10002
	CodeServiceUnavailable    = 10003
	CodeRequestTimeout        = 10004
	CodeRateLimit             = 10005
	CodeInvalidParams         = 100100
	CodeMissingParam          = 100101
	CodeInvalidParamFormat    = 100102
	CodeNotFound              = 100200
	CodeAlreadyExists         = 100201
	CodeUnauthorized          = 100300
	CodeForbidden             = 100301
	CodeDatabaseError         = 100400
	CodeThirdPartyError       = 100500
)

// Order (01xxxx)
const (
	CodeOrderNotFound              = 101001
	CodeOrderAlreadyExists         = 101002
	CodeOrderExpired               = 101003
	CodeOrderAlreadyPaid           = 101004
	CodeOrderAlreadyClosed         = 101005
	CodeOrderCannotClose           = 101006
	CodeOrderAmountInvalid         = 101007
)

// Payment (02xxxx)
const (
	CodePaymentNotFound            = 102001
	CodePaymentAlreadyExists       = 102002
	CodePaymentExpired             = 102003
	CodePaymentAlreadyPaid         = 102004
	CodePaymentAlreadyCancelled    = 102005
	CodePaymentAlreadyClosed       = 102006
	CodeInvalidPaymentChannel      = 102007
	CodeInvalidPaymentMethod       = 102008
	CodePaymentAmountMismatch      = 102009
	CodePaymentCreationFailed      = 102010
	CodePaymentNotificationVerifyFailed = 102012
)

// Refund (03xxxx)
const (
	CodeRefundNotFound             = 103001
	CodeRefundAlreadyExists        = 103002
	CodeRefundAmountExceedsPayment = 103003
	CodeRefundAmountExceedsRemaining = 103004
	CodePaymentNotRefundable       = 103005
	CodeRefundCreationFailed       = 103006
)

// Webhook (04xxxx)
const (
	CodeWebhookInvalidSignature    = 104001
	CodeWebhookProcessingFailed    = 104002
	CodeWebhookDuplicate           = 104003
)

// WeChat (10xxxx)
const (
	CodeWeChatAPICallFailed        = 110001
	CodeWeChatSignVerifyFailed     = 110002
	CodeWeChatCertError            = 110003
)

// Alipay (11xxxx)
const (
	CodeAlipayAPICallFailed        = 111001
	CodeAlipaySignVerifyFailed     = 111002
	CodeAlipayCertError            = 111003
)

var messageMap = map[int]string{
	CodeSuccess:               "success",
	CodeUnknown:               "unknown error",
	CodeInternal:              "internal server error",
	CodeServiceUnavailable:    "service unavailable",
	CodeRequestTimeout:        "request timeout",
	CodeRateLimit:             "rate limit exceeded",
	CodeInvalidParams:         "invalid request parameters",
	CodeMissingParam:          "missing required parameter",
	CodeInvalidParamFormat:    "invalid parameter format",
	CodeNotFound:              "resource not found",
	CodeAlreadyExists:         "resource already exists",
	CodeUnauthorized:          "unauthorized",
	CodeForbidden:             "forbidden",
	CodeDatabaseError:         "database error",
	CodeThirdPartyError:       "third-party service error",
	CodeOrderNotFound:              "order not found",
	CodeOrderAlreadyExists:         "order already exists",
	CodeOrderExpired:               "order expired",
	CodeOrderAlreadyPaid:           "order already paid",
	CodeOrderAlreadyClosed:         "order already closed",
	CodeOrderCannotClose:           "order cannot be closed",
	CodeOrderAmountInvalid:         "order amount invalid",
	CodePaymentNotFound:            "payment not found",
	CodePaymentAlreadyExists:       "payment already exists",
	CodePaymentExpired:             "payment expired",
	CodePaymentAlreadyPaid:         "payment already paid",
	CodePaymentAlreadyCancelled:    "payment already cancelled",
	CodePaymentAlreadyClosed:       "payment already closed",
	CodeInvalidPaymentChannel:      "invalid payment channel",
	CodeInvalidPaymentMethod:       "invalid payment method",
	CodePaymentAmountMismatch:      "payment amount mismatch",
	CodePaymentCreationFailed:      "payment creation failed",
	CodePaymentNotificationVerifyFailed: "payment notification verification failed",
	CodeRefundNotFound:             "refund not found",
	CodeRefundAlreadyExists:        "refund already exists",
	CodeRefundAmountExceedsPayment: "refund amount exceeds payment amount",
	CodeRefundAmountExceedsRemaining: "refund amount exceeds remaining refundable amount",
	CodePaymentNotRefundable:       "payment not refundable",
	CodeRefundCreationFailed:       "refund creation failed",
	CodeWebhookInvalidSignature:    "invalid notification signature",
	CodeWebhookProcessingFailed:    "notification processing failed",
	CodeWebhookDuplicate:           "duplicate notification",
	CodeWeChatAPICallFailed:        "WeChat API call failed",
	CodeWeChatSignVerifyFailed:     "WeChat signature verification failed",
	CodeWeChatCertError:            "WeChat certificate error",
	CodeAlipayAPICallFailed:        "Alipay API call failed",
	CodeAlipaySignVerifyFailed:     "Alipay signature verification failed",
	CodeAlipayCertError:            "Alipay certificate error",
}

var httpStatusMap = map[int]int{
	CodeSuccess:               http.StatusOK,
	CodeUnknown:               http.StatusInternalServerError,
	CodeInternal:              http.StatusInternalServerError,
	CodeServiceUnavailable:    http.StatusServiceUnavailable,
	CodeRequestTimeout:        http.StatusRequestTimeout,
	CodeRateLimit:             http.StatusTooManyRequests,
	CodeInvalidParams:         http.StatusBadRequest,
	CodeMissingParam:          http.StatusBadRequest,
	CodeInvalidParamFormat:    http.StatusBadRequest,
	CodeNotFound:              http.StatusNotFound,
	CodeAlreadyExists:         http.StatusConflict,
	CodeUnauthorized:          http.StatusUnauthorized,
	CodeForbidden:             http.StatusForbidden,
	CodeDatabaseError:         http.StatusInternalServerError,
	CodeThirdPartyError:       http.StatusBadGateway,
	CodeOrderNotFound:              http.StatusNotFound,
	CodeOrderAlreadyExists:         http.StatusConflict,
	CodeOrderExpired:               http.StatusBadRequest,
	CodeOrderAlreadyPaid:           http.StatusConflict,
	CodeOrderAlreadyClosed:         http.StatusConflict,
	CodeOrderCannotClose:           http.StatusBadRequest,
	CodeOrderAmountInvalid:         http.StatusBadRequest,
	CodePaymentNotFound:            http.StatusNotFound,
	CodePaymentAlreadyExists:       http.StatusConflict,
	CodePaymentExpired:             http.StatusBadRequest,
	CodePaymentAlreadyPaid:         http.StatusConflict,
	CodePaymentAlreadyCancelled:    http.StatusConflict,
	CodePaymentAlreadyClosed:       http.StatusConflict,
	CodeInvalidPaymentChannel:      http.StatusBadRequest,
	CodeInvalidPaymentMethod:       http.StatusBadRequest,
	CodePaymentAmountMismatch:      http.StatusBadRequest,
	CodePaymentCreationFailed:      http.StatusInternalServerError,
	CodePaymentNotificationVerifyFailed: http.StatusBadRequest,
	CodeRefundNotFound:             http.StatusNotFound,
	CodeRefundAlreadyExists:        http.StatusConflict,
	CodeRefundAmountExceedsPayment: http.StatusBadRequest,
	CodeRefundAmountExceedsRemaining: http.StatusBadRequest,
	CodePaymentNotRefundable:       http.StatusBadRequest,
	CodeRefundCreationFailed:       http.StatusInternalServerError,
	CodeWebhookInvalidSignature:    http.StatusBadRequest,
	CodeWebhookProcessingFailed:    http.StatusInternalServerError,
	CodeWebhookDuplicate:           http.StatusOK,
	CodeWeChatAPICallFailed:        http.StatusBadGateway,
	CodeWeChatSignVerifyFailed:     http.StatusBadRequest,
	CodeWeChatCertError:            http.StatusInternalServerError,
	CodeAlipayAPICallFailed:        http.StatusBadGateway,
	CodeAlipaySignVerifyFailed:     http.StatusBadRequest,
	CodeAlipayCertError:            http.StatusInternalServerError,
}

func GetMessage(code int) string {
	if msg, ok := messageMap[code]; ok {
		return msg
	}
	return messageMap[CodeUnknown]
}

func GetHTTPStatus(code int) int {
	if status, ok := httpStatusMap[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

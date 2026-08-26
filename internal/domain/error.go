package domain

const (
	CodeInvalidInput         = "INVALID_INPUT"
	CodeNotFound             = "NOT_FOUND"
	CodeAccessRestricted     = "ACCESS_RESTRICTED"
	CodeRateLimited          = "RATE_LIMITED"
	CodeNetworkFailure       = "NETWORK_FAILURE"
	CodeTimeout              = "TIMEOUT"
	CodeUnsupportedOperation = "UNSUPPORTED_OPERATION"
	CodeUpstreamChanged      = "UPSTREAM_CHANGED"
	CodeInternalError        = "INTERNAL_ERROR"
)

type ProviderError struct {
	Code              string
	Message           string
	Retryable         bool
	RetryAfterSeconds *int
	Cause             error
}

func (e *ProviderError) Error() string {
	return e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Cause
}

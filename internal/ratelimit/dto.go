package ratelimit

// CheckRequest represents the rate limit check request
type CheckRequest struct {
	Key     string `json:"key" binding:"required"`
	Profile string `json:"profile" binding:"required"`
}

// CheckResponse represents the rate limit check response
type CheckResponse struct {
	Allowed    bool  `json:"allowed"`
	Limit      int   `json:"limit"`
	Remaining  int   `json:"remaining"`
	Reset      int64 `json:"reset"`
	RetryAfter int64 `json:"retry_after,omitempty"`
}

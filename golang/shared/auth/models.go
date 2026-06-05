package models

const (
	RequestTypeRegister = "register"
	RequestTypeLogin    = "login"
)

// Запрос от Gateway к Auth Service
type AuthRequest struct {
	CorrelationID string `json:"correlation_id"`
	Type          string `json:"type"` // "register" или "login"
	Username      string `json:"username,omitempty"`
	Email         string `json:"email"`
	Password      string `json:"password"`
}

// Ответ от Auth Service к Gateway
type AuthResponse struct {
	CorrelationID string `json:"correlation_id"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Token         string `json:"token,omitempty"`
	ExpiresAt     int64  `json:"expires_at,omitempty"` // Unix timestamp
	UserID        string `json:"user_id,omitempty"`
}

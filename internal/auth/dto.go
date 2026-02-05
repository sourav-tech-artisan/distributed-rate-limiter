package auth

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	APIKey  string `json:"api_key"`
	Message string `json:"message"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	APIKey string `json:"api_key"`
}

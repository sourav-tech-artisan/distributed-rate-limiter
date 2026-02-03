package model

import (
	"time"
)

// Tenant represents a user/organization using the rate limiter service
type Tenant struct {
	ID                string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email             string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash      string    `gorm:"type:varchar(255);not null"`
	APIKey            string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	MaxProfiles       int       `gorm:"default:10"`
	MaxRequestsPerDay int       `gorm:"default:10000"`
	IsActive          bool      `gorm:"default:true"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (Tenant) TableName() string {
	return "tenants"
}

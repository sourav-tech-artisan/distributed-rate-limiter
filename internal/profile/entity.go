package profile

import (
	"time"
)

// Profile represents a rate limiting configuration for a tenant
type Profile struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID  string    `gorm:"type:uuid;index;not null"`
	Name      string    `gorm:"type:varchar(64);not null"`
	Algorithm string    `gorm:"type:varchar(32);not null"`
	Limit     int       `gorm:"column:limit_value;not null"`
	Window    string    `gorm:"column:window;type:varchar(32);not null"`
	Capacity  *int      `gorm:"type:int"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (Profile) TableName() string {
	return "profiles"
}

// GetCapacity returns capacity or defaults to limit if not set
func (p *Profile) GetCapacity() int {
	if p.Capacity != nil {
		return *p.Capacity
	}
	return p.Limit
}

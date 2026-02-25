package db

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time `gorm:"type:timestamp" json:"createdAt"`
}

type APIKey struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Key        string     `gorm:"type:varchar(255);not null" json:"key"`
	UserID     uint       `gorm:"not null" json:"user_id"`
	User       User       `gorm:"foreignKey:UserID;references:ID" json:"user"`
	IsRevoked  bool       `gorm:"default:false" json:"is_revoked"`
	ExpiresAt  *time.Time `gorm:"type:timestamp" json:"expires_at"`
	Name       string     `gorm:"type:varchar(255)" json:"name"`
	LastUsedAt *time.Time `gorm:"type:timestamp" json:"last_used_at"`
	CreatedAt  *time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"type:timestamp" json:"updated_at"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

type AccessLogs struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	Method     string    `gorm:"type:varchar(10);not null" json:"method"`
	Path       string    `gorm:"type:varchar(500);not null" json:"path"`
	StatusCode int       `gorm:"not null" json:"status_code"`
	IPAddress  string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent  string    `gorm:"type:varchar(500)" json:"user_agent"`
	Duration   int64     `gorm:"not null" json:"duration"`
	Timestamp  time.Time `gorm:"type:timestamp;not null" json:"timestamp"`
}

func (AccessLogs) TableName() string {
	return "access_logs"
}

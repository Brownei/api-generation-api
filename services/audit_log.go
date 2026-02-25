package services

import (
	"time"

	"github.com/Brownei/api-generation-api/db"
	"gorm.io/gorm"
)

type AuditLogService struct {
	db *gorm.DB
}

func NewAuditLogService(db *gorm.DB) *AuditLogService {
	return &AuditLogService{db: db}
}

func (s *AuditLogService) CreateLog(log *db.AccessLogs) error {
	return s.db.Create(log).Error
}

type AuditLogEntry struct {
	UserID     uint
	Method     string
	Path       string
	StatusCode int
	IPAddress  string
	UserAgent  string
	Duration   int64
}

func (s *AuditLogService) LogRequest(entry AuditLogEntry) {
	log := &db.AccessLogs{
		UserID:     entry.UserID,
		Method:     entry.Method,
		Path:       entry.Path,
		StatusCode: entry.StatusCode,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		Duration:   entry.Duration,
		Timestamp:  time.Now(),
	}

	go func() {
		if err := s.CreateLog(log); err != nil {
			_ = err
		}
	}()
}

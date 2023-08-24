package storage

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Storage struct {
	AIAgentStorage *AIAgentStorage
}

func NewStorage(postgresDriver *gorm.DB, logger *zap.SugaredLogger) *Storage {
	aiAgentStorage := NewAIAgentStorage(postgresDriver, logger)
	return &Storage{
		AIAgentStorage: aiAgentStorage,
	}
}

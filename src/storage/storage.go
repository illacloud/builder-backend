package storage

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Storage struct {
	AppStorage         *AppStorage
	ActionStorage      *ActionStorage
	AppSnapshotStorage *AppSnapshotStorage
	KVStateStorage     *KVStateStorage
	ResourceStorage    *ResourceStorage
	SetStateStorage    *SetStateStorage
	TreeStateStorage   *TreeStateStorage
}

func NewStorage(postgresDriver *gorm.DB, logger *zap.SugaredLogger) *Storage {
	aiAgentStorage := NewAIAgentStorage(postgresDriver, logger)
	return &Storage{
		AIAgentStorage: aiAgentStorage,
	}
}

package storage

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Storage struct {
	AppStorage         *AppStorage
	ActionStorage      *ActionStorage
	FlowActionStorage  *FlowActionStorage
	AppSnapshotStorage *AppSnapshotStorage
	KVStateStorage     *KVStateStorage
	ResourceStorage    *ResourceStorage
	SetStateStorage    *SetStateStorage
	TreeStateStorage   *TreeStateStorage
}

func NewStorage(postgresDriver *gorm.DB, logger *zap.SugaredLogger) *Storage {
	return &Storage{
		AppStorage:         NewAppStorage(logger, postgresDriver),
		ActionStorage:      NewActionStorage(logger, postgresDriver),
		FlowActionStorage:  NewFlowActionStorage(logger, postgresDriver),
		AppSnapshotStorage: NewAppSnapshotStorage(logger, postgresDriver),
		KVStateStorage:     NewKVStateStorage(logger, postgresDriver),
		ResourceStorage:    NewResourceStorage(logger, postgresDriver),
		SetStateStorage:    NewSetStateStorage(logger, postgresDriver),
		TreeStateStorage:   NewTreeStateStorage(logger, postgresDriver),
	}
}

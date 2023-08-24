package storage

import (
	"time"

	"github.com/google/uuid"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/illacloud/illa-resource-manager-backend/src/model"
)

type AIAgentStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewAIAgentStorage(db *gorm.DB, logger *zap.SugaredLogger) *AIAgentStorage {
	return &AIAgentStorage{
		logger: logger,
		db:     db,
	}
}

func (d *AIAgentStorage) Create(u *model.AIAgent) (int, error) {
	if err := d.db.Create(u).Error; err != nil {
		return 0, err
	}
	return u.ID, nil
}

func (d *AIAgentStorage) RetrieveByID(id int) (*model.AIAgent, error) {
	u := &model.AIAgent{}
	if err := d.db.First(u, id).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (d *AIAgentStorage) RetrieveByIDs(ids []int) ([]*model.AIAgent, error) {
	cts := []*model.AIAgent{}
	if err := d.db.Where("(id) IN ?", ids).Find(&cts).Error; err != nil {
		return nil, err
	}
	return cts, nil
}

func (d *AIAgentStorage) RetrieveByUID(uid uuid.UUID) (*model.AIAgent, error) {
	u := &model.AIAgent{}
	if err := d.db.Where("uid = ?", uid).First(&u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (d *AIAgentStorage) RetrieveByTeamIDAndID(teamID int, id int) (*model.AIAgent, error) {
	var aiAgent *model.AIAgent
	if err := d.db.Where("id = ? AND team_id = ?", id, teamID).First(&aiAgent).Error; err != nil {
		return nil, err
	}
	return aiAgent, nil
}

func (d *AIAgentStorage) RetrievePublishedByID(id int) (*model.AIAgent, error) {
	var aiAgent *model.AIAgent
	if err := d.db.Where("id = ? AND published_to_marketplace = 1", id).First(&aiAgent).Error; err != nil {
		return nil, err
	}
	return aiAgent, nil
}

func (d *AIAgentStorage) RetrieveByUIDAndStatus(uid uuid.UUID, status int) (*model.AIAgent, error) {
	u := &model.AIAgent{}
	if err := d.db.Where("uid = ? AND transaction_status = ?", uid, status).First(&u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (d *AIAgentStorage) RetrieveByTeamIDAndSortByCreatedAtDesc(teamID int) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	if err := d.db.Where("team_id = ?", teamID).Order("created_at desc").Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) RetrieveByTeamIDAndSortByUpdatedAtDesc(teamID int) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	if err := d.db.Where("team_id = ?", teamID).Order("updated_at desc").Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) RetrieveByKeywordsAndSortByCreatedAtDesc(teamID int, keywords string) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	fuzzyKeywords := "%" + keywords + "%"
	if err := d.db.Where("team_id = ? AND (name ilike ? OR model_payload->>'prompt' ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords, fuzzyKeywords).Order("created_at desc").Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) RetrieveByKeywordsAndSortByUpdatedAtDesc(teamID int, keywords string) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	fuzzyKeywords := "%" + keywords + "%"
	if err := d.db.Where("team_id = ? AND (name ilike ? OR model_payload->>'prompt' ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords, fuzzyKeywords).Order("updated_at desc").Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) RetrieveByTeamIDSortByCreatedAtDescByPage(teamID int, pagination *Pagination) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	if err := d.db.Scopes(paginate(d.db, pagination)).Where("team_id = ?", teamID).Order("created_at desc").Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) RetrieveByTeamIDSortByUpdatedAtDescByPage(teamID int, pagination *Pagination) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	if err := d.db.Scopes(paginate(d.db, pagination)).Where("team_id = ?", teamID).Order("updated_at desc").Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) RetrieveByKeywordsAndSortByCreatedAtDescByPage(teamID int, keywords string, pagination *Pagination) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	fuzzyKeywords := "%" + keywords + "%"
	if err := d.db.Scopes(paginate(d.db, pagination)).Where("team_id = ? AND (name ilike ? OR model_payload->>'prompt' ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords, fuzzyKeywords).Order("created_at desc").Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) RetrieveByKeywordsAndSortByUpdatedAtDescByPage(teamID int, keywords string, pagination *Pagination) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	fuzzyKeywords := "%" + keywords + "%"
	if err := d.db.Scopes(paginate(d.db, pagination)).Where("team_id = ? AND (name ilike ? OR model_payload->>'prompt' ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords, fuzzyKeywords).Order("updated_at desc").Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) RetrieveByTeamIDAppIDAndPage(teamID int, appID int, pagination *Pagination) ([]*model.AIAgent, error) {
	var aiAgents []*model.AIAgent
	if err := d.db.Scopes(paginate(d.db, pagination)).Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&aiAgents).Error; err != nil {
		return nil, err
	}
	return aiAgents, nil
}

func (d *AIAgentStorage) UpdateByID(u *model.AIAgent) error {
	if err := d.db.Model(&model.AIAgent{}).Where("id = ?", u.ID).UpdateColumns(u).Error; err != nil {
		return err
	}
	return nil
}

func (d *AIAgentStorage) UpdateByUID(u *model.AIAgent) error {
	if err := d.db.Model(&model.AIAgent{}).Where("uid = ?", u.UID).UpdateColumns(u).Error; err != nil {
		return err
	}
	return nil
}

func (d *AIAgentStorage) DeleteByID(id int) error {
	if err := d.db.Delete(&model.AIAgent{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (d *AIAgentStorage) DeleteByUID(uid string) error {
	if err := d.db.Where("uid = ?", uid).Delete(&model.AIAgent{}).Error; err != nil {
		return err
	}
	return nil
}

func (d *AIAgentStorage) DeleteByTeamID(teamID int) error {
	if err := d.db.Where("team_id = ?", teamID).Delete(&model.AIAgent{}).Error; err != nil {
		return err
	}
	return nil
}

func (d *AIAgentStorage) CountByTeamID(teamID int) (int64, error) {
	var count int64
	if err := d.db.Model(&model.AIAgent{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (d *AIAgentStorage) CountByTeamIDAndKeywords(teamID int, keywords string) (int64, error) {
	var count int64
	fuzzyKeywords := "%" + keywords + "%"
	if err := d.db.Model(&model.AIAgent{}).Where("team_id = ? AND (name ilike ? OR model_payload->>'prompt' ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords, fuzzyKeywords).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (d *AIAgentStorage) RetrieveLastModifiedTime(teamID int) (time.Time, error) {
	var aiAgent *model.AIAgent
	if err := d.db.Where("team_id = ?", teamID).Order("updated_at desc").First(&aiAgent).Error; err != nil {
		return time.Time{}, err
	}
	return aiAgent.ExportUpdatedAt(), nil
}

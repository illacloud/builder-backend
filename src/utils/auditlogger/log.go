// Copyright 2023 Illa Soft, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auditlogger

import (
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
)

const (
	AUDIT_LOG_CREATE_APP = 1
	AUDIT_LOG_EDIT_APP   = 2
	AUDIT_LOG_DELETE_APP = 3
	AUDIT_LOG_VIEW_APP   = 4
	AUDIT_LOG_DEPLOY_APP = 5

	AUDIT_LOG_CREATE_RESOURCE = 6
	AUDIT_LOG_UPDATE_RESOURCE = 7
	AUDIT_LOG_DELETE_RESOURCE = 8

	AUDIT_LOG_RUN_ACTION = 9

	AUDIT_LOG_TRIGGER_TASK = 10
)

const (
	TASK_GENERATE_SQL = "sqlGenerate"
)

type LogInfo struct {
	// event
	EventType int
	// team
	TeamID   int
	TeamName string
	// user
	UserID int
	// ip address
	IP string
	// app
	AppID   int
	AppName string
	// resource
	ResourceID   int
	ResourceName string
	ResourceType string
	// action
	ActionID        int
	ActionName      string
	ActionParameter interface{}
	// task
	TaskName  string
	TaskInput map[string]interface{}
}

type AuditLog struct {
	ID           int       `gorm:"column:id;type:bigserial;primary_key"`
	UID          uuid.UUID `gorm:"column:uid;type:uuid;not null"`
	Type         int       `gorm:"column:type;type:smallint;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp;not null"`
	BasicData    string    `gorm:"column:basic_data;type:jsonb;not null"`
	Context      string    `gorm:"column:context;type:jsonb;not null"`
	AppID        int       `gorm:"column:app_id;type:bigint"`
	AppName      string    `gorm:"column:app_name;type:varchar;size:255"`
	ResourceID   int       `gorm:"column:resource_id;type:bigint"`
	ResourceName string    `gorm:"column:resource_name;type:varchar;size: 255"`
	ActionID     int       `gorm:"column:action_id;type:bigint"`
	ActionName   string    `gorm:"column:action_name;type:varchar;size: 255"`
	UserID       int       `gorm:"column:user_id;type:bigint;not null"`
	Username     string    `gorm:"column:username;type:varchar;size:32;not null"`
	Email        string    `gorm:"column:email;type:varchar;size:255;not null"`
	Who          string    `gorm:"column:who;type:jsonb;not null"`
	TeamID       int       `gorm:"column:team_id;type:bigint;not null"`
	TeamName     string    `gorm:"column:team_name;type:varchar;size:255;not null"`
}

func (*AuditLog) TableName() string {
	return "audit_logs"
}

type User struct {
	ID             int       `gorm:"column:id;type:bigserial;primary_key;index:users_ukey"`
	UID            uuid.UUID `gorm:"column:uid;type:uuid;not null;index:users_ukey"`
	Nickname       string    `gorm:"column:nickname;type:varchar;size:15"`
	PasswordDigest string    `gorm:"column:password_digest;type:varchar;size:60;not null"`
	Email          string    `gorm:"column:email;type:varchar;size:255;not null"`
	Avatar         string    `gorm:"column:avatar;type:varchar;size:255;not null"`
	SSOConfig      string    `gorm:"column:sso_config;type:jsonb"`    // for single sign-on data
	Customization  string    `gorm:"column:customization;type:jsonb"` // for user itself customization config, including: Language, IsSubscribed
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp"`
}

func (*User) TableName() string {
	return "users"
}

type UserSSOConfig struct {
	Github map[string]string `json:"github"`
}

type UserRole struct {
	UserRole  int
	CreatedAt time.Time
}

type UserInfo struct {
	UserId            int       `json:"userId"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	ConnectedToGithub bool      `json:"connectedToGithub"`
	HasPassword       bool      `json:"hasPassword"`
	PermissionType    int       `json:"permissionType"`
	JoinTeamAt        time.Time `json:"joinTeamAt,omitempty"`
}

func (a *AuditLogger) getTeamName(teamId int) string {
	var teamName string
	a.db.Raw("SELECT name FROM teams WHERE id = ?", teamId).Scan(&teamName)
	return teamName
}

func (a *AuditLogger) getUserInfo(userId, teamId int) *UserInfo {
	u := &User{}
	if err := a.db.First(u, userId).Error; err != nil {
		return anonymousUser()
	}

	var userRole UserRole
	a.db.Raw("SELECT user_role, created_at FROM team_members WHERE team_id = ? AND user_id = ?", teamId, userId).Scan(&userRole)

	connectedToGithub := false
	var userSSOConfig UserSSOConfig
	if err := json.Unmarshal([]byte(u.SSOConfig), &userSSOConfig); err == nil {
		if userSSOConfig.Github["name"] == "github" {
			connectedToGithub = true
		}
	}

	return &UserInfo{
		UserId:            userId,
		Username:          u.Nickname,
		Email:             u.Email,
		ConnectedToGithub: connectedToGithub,
		HasPassword:       u.PasswordDigest != "",
		PermissionType:    userRole.UserRole,
		JoinTeamAt:        userRole.CreatedAt,
	}
}

func anonymousUser() *UserInfo {
	return &UserInfo{
		UserId:            -1,
		Username:          "Anonymous User",
		Email:             "anonymous",
		ConnectedToGithub: false,
		HasPassword:       false,
		PermissionType:    -1,
	}
}

func (a *AuditLogger) Log(logInfo *LogInfo) {
	if os.Getenv("ILLA_DEPLOY_MODE") != ILLA_DEPLOY_MODE_CLOUD {
		return
	}
	// Get teamName via teamId
	teamName := a.getTeamName(logInfo.TeamID)

	// Get UserInfo via userId
	userInfo := a.getUserInfo(logInfo.UserID, logInfo.TeamID)

	// Basic data
	basicData := map[string]interface{}{
		"ipAddress": logInfo.IP,
	}

	// Context data
	contextData := make(map[string]interface{})
	switch logInfo.EventType {
	case AUDIT_LOG_CREATE_APP, AUDIT_LOG_EDIT_APP, AUDIT_LOG_DELETE_APP, AUDIT_LOG_VIEW_APP, AUDIT_LOG_DEPLOY_APP:
		if logInfo.AppName == "" {
			logInfo.AppID = -1
			logInfo.AppName = "Tutorial App"
		}
		contextData["appInfo"] = map[string]interface{}{
			"appId":   logInfo.AppID,
			"appName": logInfo.AppName,
		}
	case AUDIT_LOG_CREATE_RESOURCE, AUDIT_LOG_UPDATE_RESOURCE, AUDIT_LOG_DELETE_RESOURCE:
		contextData["resourceInfo"] = map[string]interface{}{
			"resourceID":   logInfo.ResourceID,
			"resourceName": logInfo.ResourceName,
			"resourceType": logInfo.ResourceType,
		}
	case AUDIT_LOG_RUN_ACTION:
		if logInfo.AppName == "" {
			logInfo.AppID = -1
			logInfo.AppName = "Tutorial App"
		}
		contextData["appInfo"] = map[string]interface{}{
			"appId":   logInfo.AppID,
			"appName": logInfo.AppName,
		}
		contextData["resourceInfo"] = map[string]interface{}{
			"resourceID":   logInfo.ResourceID,
			"resourceName": logInfo.ResourceName,
			"resourceType": logInfo.ResourceType,
		}
		contextData["actionInfo"] = map[string]interface{}{
			"actionID":        logInfo.ActionID,
			"actionName":      logInfo.ActionName,
			"actionParameter": logInfo.ActionParameter,
		}
	case AUDIT_LOG_TRIGGER_TASK:
		contextData["taskInfo"] = map[string]interface{}{
			"taskName":  logInfo.TaskName,
			"taskInput": logInfo.TaskInput,
		}
	}

	// Convert data
	basicDataBytes, _ := json.Marshal(basicData)
	basicDataString := string(basicDataBytes)

	userInfoBytes, _ := json.Marshal(userInfo)
	userInfoString := string(userInfoBytes)

	contextDataBytes, _ := json.Marshal(contextData)
	contextDataString := string(contextDataBytes)

	// Log to PostgreSQL
	auditLog := AuditLog{
		UID:       uuid.New(),
		Type:      logInfo.EventType,
		CreatedAt: time.Now().UTC(),
		BasicData: basicDataString,
		Context:   contextDataString,
		UserID:    logInfo.UserID,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		Who:       userInfoString,
		TeamID:    logInfo.TeamID,
		TeamName:  teamName,
	}

	if logInfo.AppID != 0 {
		auditLog.AppID = logInfo.AppID
		auditLog.AppName = logInfo.AppName
	}
	if logInfo.ResourceID != 0 {
		auditLog.ResourceID = logInfo.ResourceID
		auditLog.ResourceName = logInfo.ResourceName
	}
	if logInfo.ActionID != 0 {
		auditLog.ActionID = logInfo.ActionID
		auditLog.ActionName = logInfo.ActionName
	}

	a.db.Create(&auditLog)
}

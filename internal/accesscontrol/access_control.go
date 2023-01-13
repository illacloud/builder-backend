package accesscontrol

import (
	supervisior "github.com/illacloud/builder-backend/internal/util/supervisior"
)

// default
const DEFAULT_TEAM_ID = 0
const DEFAULT_UNIT_ID = 0

// user status in team
const STATUS_OK = 1
const STATUS_PENDING = 2
const STATUS_SUSPEND = 3

// Attirbute Unit List
const (
	UNIT_TYPE_TEAM    = iota + 1  // cloud team
	UNIT_TYPE_TEAM_MEMBER         // cloud team member
	UNIT_TYPE_USER                // cloud user
	UNIT_TYPE_INVITE              // cloud invite
	UNIT_TYPE_DOMAIN              // cloud domain
	UNIT_TYPE_BILLING             // cloud billing
	UNIT_TYPE_BUILDER_DASHBOARD   // builder dabshboard
	UNIT_TYPE_APP                 // builder app
	UNIT_TYPE_COMPONENTS          // builder components
	UNIT_TYPE_RESOURCE            // resource resource
	UNIT_TYPE_ACTION              // resource action
	UNIT_TYPE_TRANSFORMER         // resource transformer
	UNIT_TYPE_JOB                 // hub job
)

// User Role ID in Team
// @note: this will extend as role system later.
const (
	USER_ROLE_OWNER  = 1
	USER_ROLE_ADMIN  = 2
	USER_ROLE_EDITOR = 3
	USER_ROLE_VIEWER = 4
)

// global invite permission config
// owner & admin -> can invite admin, editor, viewer
// editor 	     -> can invite editor, viewer
// viewer 	     -> can invite viewer
// map[nowUserRole]map[atrgetUserRole]attribute

// this config map target role to target invite role attribute
// e.g. you want invite USER_ROLE_ADMIN, so it's mapped attribute is ACTION_ACCESS_INVITE_ADMIN
var InviteRoleAttributeMap = map[int]int{
	USER_ROLE_OWNER: ACTION_ACCESS_INVITE_OWNER, USER_ROLE_ADMIN: ACTION_ACCESS_INVITE_ADMIN, USER_ROLE_EDITOR: ACTION_ACCESS_INVITE_EDITOR, USER_ROLE_VIEWER: ACTION_ACCESS_INVITE_VIEWER,
}

// this config map target role to target manage user role attribute
// e.g. you want modify a user to role USER_ROLE_EDITOR, so it's mapped attribute is ACTION_MANAGE_ROLE_TO_EDITOR
var ModifyRoleFromAttributeMap = map[int]int{
	USER_ROLE_OWNER: ACTION_MANAGE_ROLE_FROM_OWNER, USER_ROLE_ADMIN: ACTION_MANAGE_ROLE_FROM_ADMIN, USER_ROLE_EDITOR: ACTION_MANAGE_ROLE_FROM_EDITOR, USER_ROLE_VIEWER: ACTION_MANAGE_ROLE_FROM_VIEWER,
}
var MadifyRoleToAttributeMap = map[int]int{
	USER_ROLE_OWNER: ACTION_MANAGE_ROLE_TO_OWNER, USER_ROLE_ADMIN: ACTION_MANAGE_ROLE_TO_ADMIN, USER_ROLE_EDITOR: ACTION_MANAGE_ROLE_TO_EDITOR, USER_ROLE_VIEWER: ACTION_MANAGE_ROLE_TO_VIEWER,
}

// Attribute List
// action access
const (
	// Basic Attribute
	ACTION_ACCESS_VIEW = iota + 1 // 访问 Attribute
	// Invite Attribute
	ACTION_ACCESS_INVITE_BY_LINK  // invite team member by link
	ACTION_ACCESS_INVITE_BY_EMAIL // invite team member by email
	ACTION_ACCESS_INVITE_OWNER    // can invite team member as an owner
	ACTION_ACCESS_INVITE_ADMIN    // can invite team member as an admin
	ACTION_ACCESS_INVITE_EDITOR   // can invite team member as an editor
	ACTION_ACCESS_INVITE_VIEWER   // can invite team member as a viewer
)

// action manage
const (
	// Team Attribute
	ACTION_MANAGE_TEAM_NAME          = iota + 1 // rename Team Attribute
	ACTION_MANAGE_TEAM_ICON                     // update icon
	ACTION_MANAGE_TEAM_CONFIG                   // update team config
	ACTION_MANAGE_UPDATE_TEAM_DOMAIN            // update team domain

	// Team Member Attribute
	ACTION_MANAGE_REMOVE_MEMBER    // remove member from a team
	ACTION_MANAGE_ROLE             // manage role of team member
	ACTION_MANAGE_ROLE_FROM_OWNER  // modify team member role from owner ..
	ACTION_MANAGE_ROLE_FROM_ADMIN  // modify team member role from admin ..
	ACTION_MANAGE_ROLE_FROM_EDITOR // modify team member role from editor ..
	ACTION_MANAGE_ROLE_FROM_VIEWER // modify team member role from viewer ..
	ACTION_MANAGE_ROLE_TO_OWNER    // modify team member role to owner
	ACTION_MANAGE_ROLE_TO_ADMIN    // modify team member role to admin
	ACTION_MANAGE_ROLE_TO_EDITOR   // modify team member role to editor
	ACTION_MANAGE_ROLE_TO_VIEWER   // modify team member role to viewer

	// User Attribute
	ACTION_MANAGE_RENAME_USER        // rename
	ACTION_MANAGE_UPDATE_USER_AVATAR // update avatar

	// Invite Attribute
	ACTION_MANAGE_CONFIG_INVITE // config invite
	ACTION_MANAGE_INVITE_LINK   // config invite link, open, close and renew

	// Domain Attribute
	ACTION_MANAGE_TEAM_DOMAIN // update team domain
	ACTION_MANAGE_APP_DOMAIN  // update app domain

	// Billing Attribute
	ACTION_MANAGE_PAYMENT_INFO // manage team payment info

	// Dashboard Attribute
	ACTION_MANAGE_DASHBOARD_BROADCAST

	// App Attribute
	ACTION_MANAGE_CREATE_APP // create APP
	ACTION_MANAGE_EDIT_APP   // edit APP

	// Resource Attribute
	ACTION_MANAGE_CREATE_RESOURCE // create resource
	ACTION_MANAGE_EDIT_RESOURCE   // edit resource

	// Action Attribute
	ACTION_MANAGE_CREATE_ACTION  // create action
	ACTION_MANAGE_EDIT_ACTION    // edit action
	ACTION_MANAGE_PREVIEW_ACTION // preview action
	ACTION_MANAGE_RUN_ACTION     // run action
)

// action delete
const (
	// Basic Attribute
	ACTION_DELETE = iota + 1 // 删除 Attribute

	// Domain Attribute
	ACTION_DELETE_TEAM_DOMAIN // 删除 Team Domain
	ACTION_DELETE_APP_DOMAIN  // 删除 App Domain

)

// action manage special (only owner and admin can access by default)
const (
	// Team Attribute
	ACTION_SPECIAL_EDITOR_AND_VIEWER_CAN_INVITE_BY_LINK_SW = iota + 1 // editor 和 viewer 可以使用链接邀请的 Attribute
	// Team Member Attribute
	ACTION_SPECIAL_TRANSFER_OWNER // 转移 owner 的 Attribute
	// Invite Attribute
	ACTION_SPECIAL_INVITE_LINK_RENEW // 更新邀请链接
	// APP Attribute
	ACTION_SPECIAL_RELEASE_APP // release APP

)

type AttributeGroup struct {
	TeamID        int
	UserAuthToken string
	UserRole      int
	UnitType      int
	UnitID        int
	Remote        *supervisior.Supervisior
}

func (attrg *AttributeGroup) Init() {
	attrg.TeamID = 0
	attrg.UserAuthToken = ""
	attrg.UserRole = 0
	attrg.UnitType = 0
	attrg.UnitID = 0
}

func (attrg *AttributeGroup) SetTeamID(teamID int) {
	attrg.TeamID = teamID
}

func (attrg *AttributeGroup) SetUserAuthToken(token string) {
	attrg.UserAuthToken = token
}

func (attrg *AttributeGroup) SetUserRole(userRole int) {
	attrg.UserRole = userRole
}

func (attrg *AttributeGroup) SetUnitType(unitType int) {
	attrg.UnitType = unitType
}

func (attrg *AttributeGroup) SetUnitID(unitID int) {
	attrg.UnitID = unitID
}

func (attrg *AttributeGroup) CanAccess(attribute int) (bool, error) {
	// remote method
	req, errInRemote := attrg.Remote.CanAccess(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanDelete(attribute int) (bool, error) {
	req, errInRemote := attrg.Remote.CanDelete(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanManage(attribute int) (bool, error) {
	req, errInRemote := attrg.Remote.CanManage(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanManageSpecial(attribute int) (bool, error) {
	req, errInRemote := attrg.Remote.CanManageSpecial(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanModify(attribute, fromID, toID int) (bool, error) {
	req, errInRemote := attrg.Remote.CanModify(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute, fromID, toID)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanInvite(userRole int) (bool, error) {
	// convert to attribute
	attribute, hit := InviteRoleAttributeMap[userRole]
	if !hit {
		return false, nil
	}
	// check attirbute
	return attrg.CanAccess(attribute)
}

func (attrg *AttributeGroup) DoesNowUserAreEditorOrViewer() bool {
	if attrg.UserRole == USER_ROLE_EDITOR || attrg.UserRole == USER_ROLE_VIEWER {
		return true
	}
	return false
}

func NewAttributeGroup(teamID int, userAuthToken string, userRole int, unitType int, unitID int) (*AttributeGroup, error) {
	// init sdk
	instance, err := supervisior.NewSupervisior()
	if err != nil {
		return nil, err
	}
	// init
	attrg := &AttributeGroup{
		TeamID:        teamID, // 0 for self-host mode by default
		UserRole:      userRole,
		UserAuthToken: userAuthToken,
		UnitType:      unitType,
		UnitID:        unitID,
		Remote:        instance,
	}
	return attrg, nil
}

func NewAttributeGroupForController(teamID int, userAuthToken string, unitType int) (*AttributeGroup, error) {
	// init sdk
	instance, err := supervisior.NewSupervisior()
	if err != nil {
		return nil, err
	}
	// init
	attrg := &AttributeGroup{
		TeamID:        teamID, // 0 for self-host mode by default
		UserAuthToken: userAuthToken,
		UnitType:      unitType,
		Remote:        instance,
	}
	return attrg, nil
}

func NewRawAttributeGroup() (*AttributeGroup, error) {
	// init sdk
	instance, err := supervisior.NewSupervisior()
	if err != nil {
		return nil, err
	}
	// init
	attrg := &AttributeGroup{
		Remote: instance,
	}
	return attrg, nil
}

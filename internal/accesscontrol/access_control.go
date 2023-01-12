package accesscontrol

import (
	"github.com/illacloud/builder-backend/internal/const"
	cloudsdk "github.com/illacloud/builder-backend/internal/util/illacloudbackendsdk"
)

// user status in team
const STATUS_OK = 1
const STATUS_PENDING = 2
const STATUS_SUSPEND = 3

// Attirbute Unit List
const (
	UNIT_TYPE_TEAM        = 1  // cloud team
	UNIT_TYPE_TEAM_MEMBER = 2  // cloud team member
	UNIT_TYPE_USER        = 3  // cloud user
	UNIT_TYPE_INVITE      = 4  // cloud invite
	UNIT_TYPE_DOMAIN      = 5  // cloud domain
	UNIT_TYPE_BILLING     = 6  // cloud billing
	UNIT_TYPE_APP         = 7  // builder app
	UNIT_TYPE_COMPONENTS  = 8  // builder components
	UNIT_TYPE_RESOURCE    = 9  // resource resource
	UNIT_TYPE_ACTION      = 10 // resource action
	UNIT_TYPE_TRANSFORMER = 11 // resource transformer
	UNIT_TYPE_JOB         = 12 // hub job
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

const (
	ATTRIBUTE_CATEGORY_ACCESS  = 1
	ATTRIBUTE_CATEGORY_DELETE  = 2
	ATTRIBUTE_CATEGORY_MANAGE  = 3
	ATTRIBUTE_CATEGORY_SPECIAL = 4
)

// Attribute List
// action access
const (
	// Basic Attribute
	ACTION_ACCESS_VIEW = iota + 1 // 访问 Attribute
	// Invite Attribute
	ACTION_ACCESS_INVITE_BY_LINK  // 使用链接邀请用户
	ACTION_ACCESS_INVITE_BY_EMAIL // 使用邮件邀请用户
	ACTION_ACCESS_INVITE_OWNER
	ACTION_ACCESS_INVITE_ADMIN
	ACTION_ACCESS_INVITE_EDITOR
	ACTION_ACCESS_INVITE_VIEWER
)

// action manage
const (
	// Team Attribute
	ACTION_MANAGE_TEAM_NAME          = iota + 1 // 重命名 Team Attribute
	ACTION_MANAGE_TEAM_ICON                     // 更新 icon
	ACTION_MANAGE_TEAM_CONFIG                   // 更新 team 设置
	ACTION_MANAGE_UPDATE_TEAM_DOMAIN            // 更新 team domain
	// Team Member Attribute
	ACTION_MANAGE_REMOVE_MEMBER    // 移除团队成员的 Attribute
	ACTION_MANAGE_ROLE             // 修改团队成员角色的 Attribute
	ACTION_MANAGE_ROLE_FROM_OWNER  // 将用户角色修改为 owner
	ACTION_MANAGE_ROLE_FROM_ADMIN  // 将用户角色修改为 admin
	ACTION_MANAGE_ROLE_FROM_EDITOR // 将用户角色修改为 editor
	ACTION_MANAGE_ROLE_FROM_VIEWER // 将用户角色修改为 viewer
	ACTION_MANAGE_ROLE_TO_OWNER    // 将用户角色修改为 owner
	ACTION_MANAGE_ROLE_TO_ADMIN    // 将用户角色修改为 admin
	ACTION_MANAGE_ROLE_TO_EDITOR   // 将用户角色修改为 editor
	ACTION_MANAGE_ROLE_TO_VIEWER   // 将用户角色修改为 viewer
	// User Attribute
	ACTION_MANAGE_RENAME_USER        // 重命名用户
	ACTION_MANAGE_UPDATE_USER_AVATAR // 更新 avatar
	// Invite Attribute
	ACTION_MANAGE_CONFIG_INVITE // 配置邀请选项和参数
	ACTION_MANAGE_INVITE_LINK   // 配置 invite link
	// Domain Attribute
	ACTION_MANAGE_TEAM_DOMAIN // 更新 Team Domain
	ACTION_MANAGE_APP_DOMAIN  // 更新 App domain
	// Billing Attribute
	ACTION_MANAGE_PAYMENT_INFO // 编辑付款信息
	// App Attribute
	ACTION_MANAGE_CREATE_APP // 创建 APP
	ACTION_MANAGE_EDIT_APP   // 编辑 APP
	// Resource Attribute
	ACTION_MANAGE_CREATE_RESOURCE // 创建 Resource
	ACTION_MANAGE_EDIT_RESOURCE   // 编辑 Resource
)

// action delete
const (
	// Basic Attribute
	ACTION_DELETE = iota + 1 // 删除 Attribute
	// Domain Attribute
	ACTION_DELETE_TEAM_DOMAIN // 删除 Team Domain
	ACTION_DELETE_APP_DOMAIN  // 删除 App Domain

)

// action special
const (
	// Team Attribute
	ACTION_SPECIAL_EDITOR_AND_VIEWER_CAN_INVITE_BY_LINK_SW = iota + 1 // editor 和 viewer 可以使用链接邀请的 Attribute
	// Team Member Attribute
	ACTION_SPECIAL_TRANSFER_OWNER // 转移 owner 的 Attribute
	// Invite Attribute
	ACTION_SPECIAL_INVITE_LINK_RENEW // 更新邀请链接
)

// Attribute Config List
// Only define avaliable attribute here
// map[AttributeCategory][role][unitType][Attribute]status
var AttributeConfigList = map[int]map[int]map[int]map[int]bool{
	ATTRIBUTE_CATEGORY_ACCESS: {
		USER_ROLE_OWNER: {
			UNIT_TYPE_TEAM:        {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_TEAM_MEMBER: {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_USER:        {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_INVITE:      {ACTION_ACCESS_VIEW: true, ACTION_ACCESS_INVITE_BY_LINK: true, ACTION_ACCESS_INVITE_BY_EMAIL: true, ACTION_ACCESS_INVITE_ADMIN: true, ACTION_ACCESS_INVITE_EDITOR: true, ACTION_ACCESS_INVITE_VIEWER: true},
			UNIT_TYPE_DOMAIN:      {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_BILLING:     {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_APP:         {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_RESOURCE:    {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_JOB:         {ACTION_ACCESS_VIEW: true},
		},
		USER_ROLE_ADMIN: {
			UNIT_TYPE_TEAM:        {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_TEAM_MEMBER: {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_USER:        {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_INVITE:      {ACTION_ACCESS_VIEW: true, ACTION_ACCESS_INVITE_BY_LINK: true, ACTION_ACCESS_INVITE_BY_EMAIL: true, ACTION_ACCESS_INVITE_ADMIN: true, ACTION_ACCESS_INVITE_EDITOR: true, ACTION_ACCESS_INVITE_VIEWER: true},
			UNIT_TYPE_DOMAIN:      {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_APP:         {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_RESOURCE:    {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_JOB:         {ACTION_ACCESS_VIEW: true},
		},
		USER_ROLE_EDITOR: {
			UNIT_TYPE_TEAM_MEMBER: {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_USER:        {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_INVITE:      {ACTION_ACCESS_VIEW: true, ACTION_ACCESS_INVITE_BY_LINK: true, ACTION_ACCESS_INVITE_BY_EMAIL: true, ACTION_ACCESS_INVITE_EDITOR: true, ACTION_ACCESS_INVITE_VIEWER: true},
			UNIT_TYPE_APP:         {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_RESOURCE:    {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_JOB:         {ACTION_ACCESS_VIEW: true},
		},
		USER_ROLE_VIEWER: {
			UNIT_TYPE_TEAM_MEMBER: {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_USER:        {ACTION_ACCESS_VIEW: true},
			UNIT_TYPE_INVITE:      {ACTION_ACCESS_VIEW: true, ACTION_ACCESS_INVITE_BY_LINK: true, ACTION_ACCESS_INVITE_BY_EMAIL: true, ACTION_ACCESS_INVITE_VIEWER: true},
			UNIT_TYPE_APP:         {ACTION_ACCESS_VIEW: true},
		},
	},
	ATTRIBUTE_CATEGORY_DELETE: {
		USER_ROLE_OWNER: {
			UNIT_TYPE_TEAM:        {ACTION_DELETE: true},
			UNIT_TYPE_TEAM_MEMBER: {ACTION_DELETE: true},
			UNIT_TYPE_USER:        {ACTION_DELETE: true},
			UNIT_TYPE_INVITE:      {ACTION_DELETE: true},
			UNIT_TYPE_DOMAIN:      {ACTION_DELETE_TEAM_DOMAIN: true, ACTION_DELETE_APP_DOMAIN: true},
			UNIT_TYPE_BILLING:     {ACTION_DELETE: true},
			UNIT_TYPE_APP:         {ACTION_DELETE: true},
			UNIT_TYPE_RESOURCE:    {ACTION_DELETE: true},
			UNIT_TYPE_JOB:         {ACTION_DELETE: true},
		},
		USER_ROLE_ADMIN: {
			UNIT_TYPE_TEAM:        {ACTION_DELETE: true},
			UNIT_TYPE_TEAM_MEMBER: {ACTION_DELETE: true},
			UNIT_TYPE_USER:        {ACTION_DELETE: true},
			UNIT_TYPE_INVITE:      {ACTION_DELETE: true},
			UNIT_TYPE_DOMAIN:      {ACTION_DELETE_TEAM_DOMAIN: true, ACTION_DELETE_APP_DOMAIN: true},
			UNIT_TYPE_APP:         {ACTION_DELETE: true},
			UNIT_TYPE_RESOURCE:    {ACTION_DELETE: true},
			UNIT_TYPE_JOB:         {ACTION_DELETE: true},
		},
		USER_ROLE_EDITOR: {
			UNIT_TYPE_USER:     {ACTION_DELETE: true},
			UNIT_TYPE_APP:      {ACTION_DELETE: true},
			UNIT_TYPE_RESOURCE: {ACTION_DELETE: true},
			UNIT_TYPE_JOB:      {ACTION_DELETE: true},
		},
		USER_ROLE_VIEWER: {
			UNIT_TYPE_USER: {ACTION_DELETE: true},
		},
	},
	ATTRIBUTE_CATEGORY_MANAGE: {
		USER_ROLE_OWNER: {
			UNIT_TYPE_TEAM:        {ACTION_MANAGE_TEAM_NAME: true, ACTION_MANAGE_TEAM_ICON: true, ACTION_MANAGE_UPDATE_TEAM_DOMAIN: true, ACTION_MANAGE_TEAM_CONFIG: true},
			UNIT_TYPE_TEAM_MEMBER: {ACTION_MANAGE_REMOVE_MEMBER: true, ACTION_MANAGE_ROLE: true, ACTION_MANAGE_ROLE_FROM_OWNER: true, ACTION_MANAGE_ROLE_FROM_ADMIN: true, ACTION_MANAGE_ROLE_FROM_EDITOR: true, ACTION_MANAGE_ROLE_FROM_VIEWER: true, ACTION_MANAGE_ROLE_TO_OWNER: true, ACTION_MANAGE_ROLE_TO_ADMIN: true, ACTION_MANAGE_ROLE_TO_EDITOR: true, ACTION_MANAGE_ROLE_TO_VIEWER: true},
			UNIT_TYPE_USER:        {ACTION_MANAGE_RENAME_USER: true, ACTION_MANAGE_UPDATE_USER_AVATAR: true},
			UNIT_TYPE_INVITE:      {ACTION_MANAGE_CONFIG_INVITE: true, ACTION_MANAGE_INVITE_LINK: true},
			UNIT_TYPE_DOMAIN:      {ACTION_MANAGE_TEAM_DOMAIN: true, ACTION_MANAGE_APP_DOMAIN: true},
			UNIT_TYPE_BILLING:     {ACTION_MANAGE_PAYMENT_INFO: true},
			UNIT_TYPE_APP:         {ACTION_MANAGE_CREATE_APP: true, ACTION_MANAGE_EDIT_APP: true},
			UNIT_TYPE_RESOURCE:    {ACTION_MANAGE_CREATE_RESOURCE: true, ACTION_MANAGE_EDIT_RESOURCE: true},
			UNIT_TYPE_JOB:         {},
		},
		USER_ROLE_ADMIN: {
			UNIT_TYPE_TEAM:        {ACTION_MANAGE_TEAM_NAME: true, ACTION_MANAGE_TEAM_ICON: true, ACTION_MANAGE_UPDATE_TEAM_DOMAIN: true, ACTION_MANAGE_TEAM_CONFIG: true},
			UNIT_TYPE_TEAM_MEMBER: {ACTION_MANAGE_REMOVE_MEMBER: true, ACTION_MANAGE_ROLE: true, ACTION_MANAGE_ROLE_FROM_ADMIN: true, ACTION_MANAGE_ROLE_FROM_EDITOR: true, ACTION_MANAGE_ROLE_FROM_VIEWER: true, ACTION_MANAGE_ROLE_TO_ADMIN: true, ACTION_MANAGE_ROLE_TO_EDITOR: true, ACTION_MANAGE_ROLE_TO_VIEWER: true},
			UNIT_TYPE_USER:        {ACTION_MANAGE_RENAME_USER: true, ACTION_MANAGE_UPDATE_USER_AVATAR: true},
			UNIT_TYPE_INVITE:      {ACTION_MANAGE_CONFIG_INVITE: true, ACTION_MANAGE_INVITE_LINK: true},
			UNIT_TYPE_DOMAIN:      {ACTION_MANAGE_TEAM_DOMAIN: true, ACTION_MANAGE_APP_DOMAIN: true},
			UNIT_TYPE_APP:         {ACTION_MANAGE_CREATE_APP: true, ACTION_MANAGE_EDIT_APP: true},
			UNIT_TYPE_RESOURCE:    {ACTION_MANAGE_CREATE_RESOURCE: true, ACTION_MANAGE_EDIT_RESOURCE: true},
			UNIT_TYPE_JOB:         {},
		},
		USER_ROLE_EDITOR: {
			UNIT_TYPE_TEAM_MEMBER: {ACTION_MANAGE_REMOVE_MEMBER: true, ACTION_MANAGE_ROLE: true, ACTION_MANAGE_ROLE_FROM_EDITOR: true, ACTION_MANAGE_ROLE_FROM_VIEWER: true, ACTION_MANAGE_ROLE_TO_EDITOR: true, ACTION_MANAGE_ROLE_TO_VIEWER: true},
			UNIT_TYPE_USER:        {ACTION_MANAGE_RENAME_USER: true, ACTION_MANAGE_UPDATE_USER_AVATAR: true},
			UNIT_TYPE_APP:         {ACTION_MANAGE_CREATE_APP: true, ACTION_MANAGE_EDIT_APP: true},
			UNIT_TYPE_RESOURCE:    {ACTION_MANAGE_CREATE_RESOURCE: true, ACTION_MANAGE_EDIT_RESOURCE: true},
			UNIT_TYPE_JOB:         {},
		},
		USER_ROLE_VIEWER: {
			UNIT_TYPE_TEAM_MEMBER: {ACTION_MANAGE_REMOVE_MEMBER: true, ACTION_MANAGE_ROLE: true, ACTION_MANAGE_ROLE_FROM_VIEWER: true, ACTION_MANAGE_ROLE_TO_VIEWER: true},
			UNIT_TYPE_USER:        {ACTION_MANAGE_RENAME_USER: true, ACTION_MANAGE_UPDATE_USER_AVATAR: true},
			UNIT_TYPE_JOB:         {},
		},
	},
	ATTRIBUTE_CATEGORY_SPECIAL: {
		USER_ROLE_OWNER: {
			UNIT_TYPE_TEAM:        {ACTION_SPECIAL_EDITOR_AND_VIEWER_CAN_INVITE_BY_LINK_SW: true},
			UNIT_TYPE_TEAM_MEMBER: {ACTION_SPECIAL_TRANSFER_OWNER: true},
			UNIT_TYPE_INVITE:      {ACTION_SPECIAL_INVITE_LINK_RENEW: true},
		},
		USER_ROLE_ADMIN: {
			UNIT_TYPE_TEAM:   {ACTION_SPECIAL_EDITOR_AND_VIEWER_CAN_INVITE_BY_LINK_SW: true},
			UNIT_TYPE_INVITE: {ACTION_SPECIAL_INVITE_LINK_RENEW: true},
		},
		USER_ROLE_EDITOR: {},
		USER_ROLE_VIEWER: {},
	},
}

type Attribute struct {
	Access  map[int]bool
	Delete  map[int]bool
	Manage  map[int]bool
	Special map[int]bool
}

func NewAttribute(userRole int, unitType int) *Attribute {
	attr := &Attribute{
		Access:  AttributeConfigList[ATTRIBUTE_CATEGORY_ACCESS][userRole][unitType],
		Delete:  AttributeConfigList[ATTRIBUTE_CATEGORY_DELETE][userRole][unitType],
		Manage:  AttributeConfigList[ATTRIBUTE_CATEGORY_MANAGE][userRole][unitType],
		Special: AttributeConfigList[ATTRIBUTE_CATEGORY_SPECIAL][userRole][unitType],
	}
	return attr
}

type AttributeGroup struct {
	TeamID 	      int
	UserAuthToken string
	UserRole      int
	UnitType      int
	UnitID        int
	Attribute     *Attribute
	Remote        *IllaCloudSDK
	DeployMode    string
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
	if attrg.DeployMode == const.DEPLOY_MODE_CLOUD {
		req, errInRemote := attrg.Remote.CanAccess(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute)
		if errInRemote != nil {
			return false, errInRemote
		}
		return req
	}
	// local method
	r, match := attrg.Attribute.Access[attribute]
	if !match {
		return false
	}
	return r
}

func (attrg *AttributeGroup) CanDelete(attribute int) (bool, error) {
	// remote method
	if attrg.DeployMode == const.DEPLOY_MODE_CLOUD {
		req, errInRemote := attrg.Remote.CanDelete(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute)
		if errInRemote != nil {
			return false, errInRemote
		}
		return req
	}
	// local method
	r, match := attrg.Attribute.Delete[attribute]
	if !match {
		return false
	}
	return r
}

func (attrg *AttributeGroup) CanManage(attribute int) (bool, error) {
	// remote method
	if attrg.DeployMode == const.DEPLOY_MODE_CLOUD {
		req, errInRemote := attrg.Remote.CanManage(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute)
		if errInRemote != nil {
			return false, errInRemote
		}
		return req
	}
	// local method
	r, match := attrg.Attribute.Manage[attribute]
	if !match {
		return false
	}
	return r
}

func (attrg *AttributeGroup) CanManageSpecial(attribute int) (bool, error) {
	// remote method
	if attrg.DeployMode == const.DEPLOY_MODE_CLOUD {
		req, errInRemote := attrg.Remote.CanManageSpecial(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute)
		if errInRemote != nil {
			return false, errInRemote
		}
		return req
	}
	// local method
	r, match := attrg.Attribute.Special[attribute]
	if !match {
		return false
	}
	return r
}

func (attrg *AttributeGroup) CanModify(attribute, fromID, toID int) (bool, error) {
	// remote method
	if attrg.DeployMode == const.DEPLOY_MODE_CLOUD {
		req, errInRemote := attrg.Remote.CanModify(attrg.UserAuthToken, attrg.TeamID, attrg.UnitType, attrg.UnitID, attribute, fromID, toID)
		if errInRemote != nil {
			return false, errInRemote
		}
		return req
	}
	// local method
	// @todo: extend this method, now only support modify user role check.
	if attribute == ACTION_MANAGE_ROLE {
		return attrg.canModifyRoleFromTo(fromID, toID)
	}
	return false
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

// @note: this is private method, does not include remote call.
func (attrg *AttributeGroup) canModifyRoleFromTo(fromRole, toRole int) bool {
	// convert to attribute
	fromRoleAttribute, fromHit := ModifyRoleFromAttributeMap[fromRole]
	toRoleAttribute, toHit := MadifyRoleToAttributeMap[toRole]
	if !fromHit || !toHit {
		return false
	}
	// check attirbute
	fromResult, fromMatch := attrg.Attribute.Manage[fromRoleAttribute]
	toResult, toMatch := attrg.Attribute.Manage[toRoleAttribute]
	if !fromMatch || !toMatch {
		return false
	}
	return fromResult && toResult
}

func (attrg *AttributeGroup) DoesNowUserAreEditorOrViewer() bool {
	if attrg.UserRole == USER_ROLE_EDITOR || attrg.UserRole == USER_ROLE_VIEWER {
		return true
	}
	return false
}

func NewAttributeGroup(userRole int, unitType int) (*AttributeGroup, error) {
	// init sdk
	sdk, err := cloudsdk.NewIllaCloudSDK()
	if err != nil {
		return nil, err
	}
	// init
	attr := NewAttribute(userRole, unitType)
	attrg := &AttributeGroup{
		TeamID:    0, // 0 for self-host mode by default
		UserRole:  userRole,
		UnitType:  unitType,
		UnitID:    0, // 0 for placeholder, this feature has not implemented.
		Attribute: attr,
		Remote:    sdk,
	}
	// init deploy mode
	deployMode := os.Getenv("ILLA_DEPLOY_MODE")
	if deployMode != const.DEPLOY_MODE_CLOUD {
		attrg.DeployMode =const.DEPLOY_MODE_SELF_HOST
	} else {
		attrg.DeployMode =const.DEPLOY_MODE_CLOUD
	}
	return attrg, nil
}

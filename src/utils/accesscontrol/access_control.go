package accesscontrol

import (
	"github.com/illacloud/builder-backend/src/utils/supervisor"
)

// default
const ANONYMOUS_AUTH_TOKEN = ""
const DEFAULT_TEAM_ID = 0
const DEFAULT_UNIT_ID = 0

// user status in team
const STATUS_OK = 1
const STATUS_PENDING = 2
const STATUS_SUSPEND = 3

// Attirbute Unit List
const (
	UNIT_TYPE_TEAM                      = 1  // cloud team
	UNIT_TYPE_TEAM_MEMBER               = 2  // cloud team member
	UNIT_TYPE_USER                      = 3  // cloud user
	UNIT_TYPE_INVITE                    = 4  // cloud invite
	UNIT_TYPE_DOMAIN                    = 5  // cloud domain
	UNIT_TYPE_BILLING                   = 6  // cloud billing
	UNIT_TYPE_BUILDER_DASHBOARD         = 7  // builder dabshboard
	UNIT_TYPE_APP                       = 8  // builder app
	UNIT_TYPE_COMPONENTS                = 9  // builder components
	UNIT_TYPE_RESOURCE                  = 10 // resource resource
	UNIT_TYPE_ACTION                    = 11 // resource action
	UNIT_TYPE_TRANSFORMER               = 12 // resource transformer
	UNIT_TYPE_JOB                       = 13 // hub job
	UNIT_TYPE_TREE_STATES               = 14 // components tree states
	UNIT_TYPE_KV_STATES                 = 15 // components k-v states
	UNIT_TYPE_SET_STATES                = 16 // components set states
	UNIT_TYPE_PROMOTE_CODES             = 17 // promote codes
	UNIT_TYPE_PROMOTE_CODE_USAGES       = 18 // promote codes usage table
	UNIT_TYPE_ROLES                     = 19 // team roles table
	UNIT_TYPE_USER_ROLE_RELATIONS       = 20 // user role relation table
	UNIT_TYPE_UNIT_ROLE_RELATIONS       = 21 // unit role relation table
	UNIT_TYPE_COMPENSATING_TRANSACTIONS = 22 // compensating transactions
	UNIT_TYPE_TRANSACTION_SERIALS       = 23 // transaction serials
	UNIT_TYPE_CAPACITIES                = 24 // capacity
	UNIT_TYPE_DRIVE                     = 25 // drive
	UNIT_TYPE_PERIPHERAL_SERVICE        = 26 // Peripheral service, including sql generate, STMP etc.
	UNIT_TYPE_AUDIT_LOG                 = 27 // cloud audit log
	UNIT_TYPE_MARKETPLACE               = 28 // marketplace
	UNIT_TYPE_AI_AGENT                  = 29 // ai-agent
	UNIT_TYPE_WORKFLOW                  = 30 // workflow
	UNIT_TYPE_FLOW_NODE                 = 31 // workflow node
	UNIT_TYPE_FLOW_ACTION               = 32 // workflow action
)

// User Role ID in Team
// @note: this will extend as role system later.
const (
	USER_ROLE_ANONYMOUS = -1
	USER_ROLE_OWNER     = 1
	USER_ROLE_ADMIN     = 2
	USER_ROLE_EDITOR    = 3
	USER_ROLE_VIEWER    = 4
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
	ACTION_ACCESS_VIEW = iota + 1 // access Attribute
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
	ACTION_MANAGE_PAYMENT      // manage payment, including create, update, cancel subscribe and purchase
	ACTION_MANAGE_PAYMENT_INFO // manage team payment info, including get portal info band billing info

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

	// Drive Attribute
	ACTION_MANAGE_CREATE_FILE      // create file
	ACTION_MANAGE_EDIT_FILE        // edit file
	ACTION_MANAGE_CREATE_SHARELINK // create sharelink

	// Marketplace Attribute
	ACTION_MANAGE_CONTRIBUTE_MARKETPLACE // contribute marketplace
	ACTION_MANAGE_UNLIST_MARKETPLACE     // contribute marketplace

	// AI-Agent Attribute
	ACTION_MANAGE_CREATE_AI_AGENT // create AI-Agent
	ACTION_MANAGE_EDIT_AI_AGENT   // edit AI-Agent
	ACTION_MANAGE_FORK_AI_AGENT   // fork AI-Agent
	ACTION_MANAGE_RUN_AI_AGENT    // run ai-agent

	ACTION_MANAGE_FORK_APP // for app

	// workflow
	ACTION_MANAGE_CREATE_WORKFLOW
	ACTION_MANAGE_EDIT_WORKFLOW

	// Flow Action Attribute
	ACTION_MANAGE_CREATE_FLOW_ACTION  // create flow action
	ACTION_MANAGE_EDIT_FLOW_ACTION    // edit flow action
	ACTION_MANAGE_PREVIEW_FLOW_ACTION // preview flow action
	ACTION_MANAGE_RUN_FLOW_ACTION     // run flow action
)

// action delete
const (
	// Basic Attribute
	ACTION_DELETE = iota + 1 // delete Attribute

	// Domain Attribute
	ACTION_DELETE_TEAM_DOMAIN // delete Team Domain
	ACTION_DELETE_APP_DOMAIN  // delete App Domain

)

// action manage special (only owner and admin can access by default)
const (
	ACTION_SPECIAL_EDITOR_AND_VIEWER_CAN_INVITE_BY_LINK_SW = iota + 1 // the "editor and viewer can invite" switch
	ACTION_SPECIAL_TRANSFER_OWNER                                     // transfer team owner to others
	ACTION_SPECIAL_INVITE_LINK_RENEW                                  // renew the invite link
	ACTION_SPECIAL_RELEASE_APP                                        // release APP
	ACTION_SPECIAL_GENERATE_SQL                                       //  paid functions, generate sql
	ACTOIN_SPECIAL_TAKE_SNAPSHOT                                      //  paid functions
	ACTOIN_SPECIAL_RECOVER_SNAPSHOT                                   //  paid functions
	ACTOIN_SPECIAL_RUN_SPECIAL_AI_AGENT_MODEL                         //  paid functions, AI-Agent Run special AI-Agent model like GPT-4
	ACTION_SPECIAL_RELEASE_PUBLIC_APP                                 //  paid functions, release public APP
)

type AttributeGroup struct {
	Remote *supervisor.Supervisor
}

func (attrg *AttributeGroup) CanAccess(teamID int, userAuthToken string, unitType int, unitID int, attribute int) (bool, error) {
	// remote method
	req, errInRemote := attrg.Remote.CanAccess(userAuthToken, teamID, unitType, unitID, attribute)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanDelete(teamID int, userAuthToken string, unitType int, unitID int, attribute int) (bool, error) {
	req, errInRemote := attrg.Remote.CanDelete(userAuthToken, teamID, unitType, unitID, attribute)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanManage(teamID int, userAuthToken string, unitType int, unitID int, attribute int) (bool, error) {
	req, errInRemote := attrg.Remote.CanManage(userAuthToken, teamID, unitType, unitID, attribute)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanManageSpecial(teamID int, userAuthToken string, unitType int, unitID int, attribute int) (bool, error) {
	req, errInRemote := attrg.Remote.CanManageSpecial(userAuthToken, teamID, unitType, unitID, attribute)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func (attrg *AttributeGroup) CanModify(teamID int, userAuthToken string, unitType int, unitID int, attribute int, fromID int, toID int) (bool, error) {
	req, errInRemote := attrg.Remote.CanModify(userAuthToken, teamID, unitType, unitID, attribute, fromID, toID)
	if errInRemote != nil {
		return false, errInRemote
	}
	return req, nil
}

func NewAttributeGroup() (*AttributeGroup, error) {
	// init sdk
	instance := supervisor.GetInstance()

	// init
	attrg := &AttributeGroup{
		Remote: instance,
	}
	return attrg, nil
}

func NewAttributeGroupForController() (*AttributeGroup, error) {
	// init sdk
	instance := supervisor.GetInstance()

	// init
	attrg := &AttributeGroup{
		Remote: instance,
	}
	return attrg, nil
}

func NewRawAttributeGroup() (*AttributeGroup, error) {
	// init sdk
	instance := supervisor.GetInstance()

	// init
	attrg := &AttributeGroup{
		Remote: instance,
	}
	return attrg, nil
}

package websocket

import (
	"github.com/illacloud/builder-backend/src/model"
)

const DEFAULT_ROOM_SLOT = 0

type InRoomUsers struct {
	RoomID           int
	All              []*UserForCooperateFeedback            // []*UserForCooperateFeedback
	AllUsers         map[string]*UserForCooperateFeedback   // map[user.ID]*UserForCooperateFeedback
	AttachedUserList map[string][]*UserForCooperateFeedback // map[component.DisplayName][]*UserForCooperateFeedback
}

func NewInRoomUsers(roomID int) *InRoomUsers {
	iru := &InRoomUsers{}
	iru.All = make([]*UserForCooperateFeedback, DEFAULT_ROOM_SLOT)
	iru.AllUsers = make(map[string]*UserForCooperateFeedback)
	iru.AttachedUserList = make(map[string][]*UserForCooperateFeedback)
	return iru
}

func (iru *InRoomUsers) EnterRoom(user *model.User) {
	fuser := NewUserForCooperateFeedbackByUser(user)
	// check if user already in room
	if _, hit := iru.AllUsers[fuser.ID]; hit {
		return
	}
	iru.All = append(iru.All, fuser)
	iru.AllUsers[fuser.ID] = fuser

}

func (iru *InRoomUsers) LeaveRoom(userID string) {
	targetFuser, hit := iru.AllUsers[userID]
	if !hit { // invalied user input, just ignore
		return
	}
	// remove user in room
	for i, fuser := range iru.All {
		if fuser.ID == userID {
			iru.All = append(iru.All[:i], iru.All[i+1:]...)
			break
		}
	}
	// remove user in components
	iru.DisattachComponent(userID, targetFuser.ExportAttachedComponentDisplayName())
	// remove from AllUsers
	delete(iru.AllUsers, userID)
}

func (iru *InRoomUsers) Count() int {
	return len(iru.AllUsers)
}

func (iru *InRoomUsers) AttachComponent(userID string, componentDisplayNames []string) {
	fuser, hit := iru.AllUsers[userID]
	if !hit { // invalied user input, just ignore
		return
	}
	for _, displayName := range componentDisplayNames {
		// check if components not recorded, insert it
		if _, alreadyAttached := fuser.AttachedComponents[displayName]; !alreadyAttached {
			iru.AttachedUserList[displayName] = append(iru.AttachedUserList[displayName], fuser)
		}
		// record user attached components
		fuser.AttachedComponents[displayName] = displayName
	}
}

func (iru *InRoomUsers) DisattachComponent(userID string, componentDisplayNames []string) {
	fuser, hit := iru.AllUsers[userID]
	if !hit { // invalied user input, just ignore
		return
	}
	for _, displayName := range componentDisplayNames {
		if _, hit := iru.AttachedUserList[displayName]; !hit { // skip user unavaliable input
			continue
		}
		// remove from AttachedUserList
		for i, fuser := range iru.AttachedUserList[displayName] {
			if fuser.ID == userID {
				iru.AttachedUserList[displayName] = append(iru.AttachedUserList[displayName][:i], iru.AttachedUserList[displayName][i+1:]...)
				break
			}
		}
		// remove fuser attached components
		delete(fuser.AttachedComponents, displayName)
	}
}

type InRoomUsersFeedback struct {
	InRoomUsers []*UserForCooperateFeedback `json:"inRoomUsers"`
}

func (iru *InRoomUsers) FetchAllInRoomUsers() *InRoomUsersFeedback {
	return &InRoomUsersFeedback{
		InRoomUsers: iru.All,
	}
}

type ComponentAttachedUsers struct {
	ComponentAttachedUsers map[string][]*UserForCooperateFeedback `json:"componentAttachedUsers"`
}

func (iru *InRoomUsers) FetchAllAttachedUsers() *ComponentAttachedUsers {
	return &ComponentAttachedUsers{
		ComponentAttachedUsers: iru.AttachedUserList,
	}

}

type UserForCooperateFeedback struct {
	ID                 string            `json:"id"`
	Nickname           string            `json:"nickname"`
	Avatar             string            `json:"avatar"`
	AttachedComponents map[string]string `json:"-"`
}

func NewUserForCooperateFeedbackByUser(user *model.User) *UserForCooperateFeedback {
	return &UserForCooperateFeedback{
		ID:                 user.ExportIDToString(),
		Nickname:           user.Nickname,
		Avatar:             user.Avatar,
		AttachedComponents: make(map[string]string),
	}
}

func (fuser *UserForCooperateFeedback) ExportAttachedComponentDisplayName() []string {
	l := len(fuser.AttachedComponents)
	ret := make([]string, l)
	for _, displayName := range fuser.AttachedComponents {
		ret = append(ret, displayName)
	}
	return ret
}

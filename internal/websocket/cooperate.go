package ws

import (
	"github.com/illacloud/builder-backend/internal/repository"
)

const DEFAULT_ROOM_SLOT = 4

type InRoomUsers struct {
	RoomID           int
	All              []*UserForCooperateFeedback // []*UserForCooperateFeedback
	AllUsers         map[int]*UserForCooperateFeedback
	AttachedUserList map[string][]*UserForCooperateFeedback // map[component.DisplayName][]*UserForCooperateFeedback
}

func NewInRoomUsers(roomID int) *InRoomUsers {
	iru := &InRoomUsers{}
	iru.All = make([]*UserForCooperateFeedback, DEFAULT_ROOM_SLOT)
	iru.AllUsers = make(map[int]*UserForCooperateFeedback)
	return iru
}

func (iru *InRoomUsers) EnterRoom(user *repository.User) {
	fuser := NewUserForCooperateFeedbackByUser(user)
	iru.All = append(iru.All, fuser)
	iru.AllUsers[fuser.ID] = fuser
}

func (iru *InRoomUsers) LeaveRoom(userID int) {
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

func (iru *InRoomUsers) AttachComponent(userID int, componentDisplayNames []string) {
	fuser, hit := iru.AllUsers[userID]
	if !hit { // invalied user input, just ignore
		return
	}
	for _, displayName := range componentDisplayNames {
		// check if components not recorded, insert it
		if _, hit := iru.AttachedUserList[displayName]; !hit {
			iru.AttachedUserList[displayName] = make([]*UserForCooperateFeedback, DEFAULT_ROOM_SLOT)
		}
		iru.AttachedUserList[displayName] = append(iru.AttachedUserList[displayName], fuser)
		// record user attached components
		fuser.AttachedComponents[displayName] = displayName
	}
}

func (iru *InRoomUsers) DisattachComponent(userID int, componentDisplayNames []string) {
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
	}
}

func (iru *InRoomUsers) FetchAllInRoomUsers() []*UserForCooperateFeedback {
	return iru.All
}

func (iru *InRoomUsers) FetchAllAttachedUsers() map[string][]*UserForCooperateFeedback {
	return iru.AttachedUserList
}

type UserForCooperateFeedback struct {
	ID                 int               `json:"id"`
	Nickname           string            `json:"nickname"`
	Avatar             string            `json:"avatar"`
	AttachedComponents map[string]string `json:"-"`
}

func NewUserForCooperateFeedbackByUser(user *repository.User) *UserForCooperateFeedback {
	return &UserForCooperateFeedback{
		ID:                 user.ID,
		Nickname:           user.Nickname,
		Avatar:             repository.USER_DEFAULT_AVATAR,
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

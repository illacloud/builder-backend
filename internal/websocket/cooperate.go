package ws

import (
	"github.com/illacloud/builder-backend/internal/repository"
)

const DEFAULT_ROOM_SLOT = 4

type InRoomUsers struct {
	RoomID           int
	All              []*UserForCooperateFeedback            // []*UserForCooperateFeedback
	AttachedUserList map[string][]*UserForCooperateFeedback // map[component.DisplayName][]*UserForCooperateFeedback
}

type UserForCooperateFeedback struct {
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func NewUserForCooperateFeedbackByUser(user *repository.User) *UserForCooperateFeedback {
	return &UserForCooperateFeedback{
		ID:       user.ID,
		Nickname: user.Nickname,
		Avatar:   repository.USER_DEFAULT_AVATAR,
	}
}

func NewInRoomUsers(roomID int) *InRoomUsers {
	iru := &InRoomUsers{}
	iru.All = make([]*UserForCooperateFeedback, DEFAULT_ROOM_SLOT)
	return iru
}

func (iru *InRoomUsers) EnterRoom(user *repository.User) {
	fuser := NewUserForCooperateFeedbackByUser(user)
	iru.All = append(iru.All, fuser)
}

func (iru *InRoomUsers) LeaveRoom(userID int) {
	for i, fuser := range iru.All {
		if fuser.ID == userID {
			iru.All = append(iru.All[:i], iru.All[i+1:]...)
			break
		}
	}
}

func (iru *InRoomUsers) AttachComponent(userID int, componentDisplayNames []string) {
	fuser := NewUserForCooperateFeedbackByUser(user)
	for _, displayName := range componentDisplayNames {
		// check if components not recorded, insert it
		if _, hit := iru.AttachedUserList[displayName]; !hit {
			iru.AttachedUserList[displayName] = make([]*UserForCooperateFeedback, DEFAULT_ROOM_SLOT)
		}
		iru.AttachedUserList[displayName] = append(iru.AttachedUserList[displayName], fuser)
	}
}

func (iru *InRoomUsers) DisattachComponent(userID int, componentDisplayNames []string) {
	for _, displayName := range componentDisplayNames {
		if _, hit := iru.AttachedUserList[displayName]; !hit { // skip user unavaliable input
			continue
		}
		// remove from AttachedUserList
		for i, fuser := range iru.AttachedUserList[displayName] {
			if fuser.ID == user.ID {
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

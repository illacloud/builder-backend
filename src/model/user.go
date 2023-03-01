package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/illa-builder-backend/src/utils/idconvertor"
)

const CUSTOMIZATION_LANGUAGE_EN_US = "en-US"
const CUSTOMIZATION_LANGUAGE_ZH_CN = "zh-CN"
const CUSTOMIZATION_LANGUAGE_KO_KR = "ko-KR"
const CUSTOMIZATION_LANGUAGE_JA_JP = "ja-JP"

const PENDING_USER_NICKNAME = "pending"
const PENDING_USER_PASSWORDDIGEST = "pending"
const PENDING_USER_AVATAR = ""

type RawUser struct {
	ID             string    `json:"id"`
	UID            uuid.UUID `json:"uid"`
	Nickname       string    `json:"nickname"`
	PasswordDigest string    `json:"passworddigest"`
	Email          string    `json:"email"`
	Avatar         string    `json:"avatar"`
	SSOConfig      string    `json:"SSOConfig"`     // for single sign-on data
	Customization  string    `json:"customization"` // for user itself customization config, including: Language, IsSubscribed
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

type User struct {
	ID             int       `json:"id"`
	UID            uuid.UUID `json:"uid"`
	Nickname       string    `json:"nickname"`
	PasswordDigest string    `json:"passworddigest"`
	Email          string    `json:"email"`
	Avatar         string    `json:"avatar"`
	SSOConfig      string    `json:"SSOConfig"`     // for single sign-on data
	Customization  string    `json:"customization"` // for user itself customization config, including: Language, IsSubscribed
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func NewUser(u *RawUser) *User {
	return &User{
		ID:             idconvertor.ConvertStringToInt(u.ID),
		UID:            u.UID,
		Nickname:       u.Nickname,
		PasswordDigest: u.PasswordDigest,
		Email:          u.Email,
		Avatar:         u.Avatar,
		SSOConfig:      u.SSOConfig,
		Customization:  u.Customization,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

func NewUserByDataControlRawData(rawUserString string) (*User, error) {
	rawUser := RawUser{}
	errInUnmarshal := json.Unmarshal([]byte(rawUserString), &rawUser)
	if errInUnmarshal != nil {
		return nil, errInUnmarshal
	}
	return NewUser(&rawUser), nil
}

func (u *User) ExportIDToString() string {
	return idconvertor.ConvertIntToString(u.ID)
}

package repository

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const CUSTOMIZATION_LANGUAGE_EN_US = "en-US"
const CUSTOMIZATION_LANGUAGE_ZH_CN = "zh-CN"

const PENDING_USER_NICKNAME = "pending"
const PENDING_USER_PASSWORDDIGEST = "pending"
const PENDING_USER_AVATAR = ""

type User struct {
	ID             int       `json:"id" gorm:"column:id;type:bigserial;primary_key;index:users_ukey"`
	UID            uuid.UUID `json:"uid" gorm:"column:uid;type:uuid;not null;index:users_ukey"`
	Nickname       string    `json:"nickname" gorm:"column:nickname;type:varchar;size:15"`
	PasswordDigest string    `json:"passworddigest" gorm:"column:password_digest;type:varchar;size:60;not null"`
	Email          string    `json:"email" gorm:"column:email;type:varchar;size:255;not null"`
	Avatar         string    `json:"avatar" gorm:"column:avatar;type:varchar;size:255;not null"`
	SSOConfig      string    `json:"SSOConfig" gorm:"column:sso_config;type:jsonb"`        // for single sign-on data
	Customization  string    `json:"customization" gorm:"column:customization;type:jsonb"` // for user itself customization config, including: Language, IsSubscribed
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp"`
}


func NewUserByDataControlRawData(rawUser string) (*User, error) {
	user := User{}
    errInUnmarshal := json.Unmarshal([]byte(rawUser), &user)
	if errInUnmarshal != nil {
		return nil, errInUnmarshal
	}
	return &user, nil
}

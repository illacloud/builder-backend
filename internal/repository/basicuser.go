// Copyright 2022 The ILLA Authors.
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

package repository

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const USER_DEFAULT_AVATAR = ""

type User struct {
	ID             int       `gorm:"column:id;type:bigserial;primary_key;index:users_ukey"`
	UID            uuid.UUID `gorm:"column:uid;type:uuid;not null;index:users_ukey"`
	Nickname       string    `gorm:"column:nickname;type:varchar;size:15;not null"`
	PasswordDigest string    `gorm:"column:password_digest;type:varchar;size:60;not null"`
	Email          string    `gorm:"column:email;type:varchar;size:255;not null"`
	Language       int       `gorm:"column:language;type:smallint;not null"`
	IsSubscribed   bool      `gorm:"column:is_subscribed;type:boolean;default:false;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp"`
}

type UserRepository interface {
	CreateUser(user *User) (int, error)
	UpdateUser(user *User) error
	FetchUserByEmail(email string) (*User, error)
	RetrieveByID(id int) (*User, error)
	FetchUserByUKey(id int, uid uuid.UUID) (*User, error)
}

type UserRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewUserRepositoryImpl(db *gorm.DB, logger *zap.SugaredLogger) *UserRepositoryImpl {
	return &UserRepositoryImpl{logger: logger, db: db}
}

func (impl *UserRepositoryImpl) CreateUser(user *User) (int, error) {
	if err := impl.db.Create(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (impl *UserRepositoryImpl) UpdateUser(user *User) error {
	if err := impl.db.Model(user).UpdateColumns(User{
		Nickname:       user.Nickname,
		PasswordDigest: user.PasswordDigest,
		Language:       user.Language,
		UpdatedAt:      user.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *UserRepositoryImpl) FetchUserByEmail(email string) (*User, error) {
	user := &User{}
	if err := impl.db.Where("email = ?", email).First(user).Error; err != nil {
		return &User{}, err
	}
	return user, nil
}

func (impl *UserRepositoryImpl) RetrieveByID(id int) (*User, error) {
	user := &User{}
	if err := impl.db.First(user, id).Error; err != nil {
		return &User{}, err
	}
	return user, nil
}

func (impl *UserRepositoryImpl) FetchUserByUKey(id int, uid uuid.UUID) (*User, error) {
	var user User
	if err := impl.db.Where("uid = ? AND id = ?", uid, id).First(&user).Error; err != nil {
		return &User{}, err
	}

	return &user, nil
}

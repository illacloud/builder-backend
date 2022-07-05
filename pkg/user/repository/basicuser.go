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

type User struct {
	ID             uuid.UUID `gorm:"column:id;type:uuid;default:uuid_generate_v4();primary_key;unique"`
	Username       string    `gorm:"column:user_name;type:varchar"`
	PasswordDigest string    `gorm:"column:password_digest;type:varchar"`
	Email          string    `gorm:"column:email;type:varchar"`
	Language       string    `gorm:"column:language;type:varchar"`
	IsSubscribed   bool      `gorm:"column:is_subscribed;type:boolean"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp"`
}

type UserRepository interface {
	CreateUser(user *User) error
	UpdateUser(user *User) error
	FetchUserByEmail(email string) (*User, error)
	RetrieveById(userId uuid.UUID) (*User, error)
}

type UserRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewUserRepositoryImpl(db *gorm.DB, logger *zap.SugaredLogger) *UserRepositoryImpl {
	return &UserRepositoryImpl{logger: logger, db: db}
}

func (impl *UserRepositoryImpl) CreateUser(user *User) error {
	if err := impl.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (impl *UserRepositoryImpl) UpdateUser(user *User) error {
	if err := impl.db.Model(user).Updates(User{
		Username:       user.Username,
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

func (impl *UserRepositoryImpl) RetrieveById(userId uuid.UUID) (*User, error) {
	user := &User{}
	if err := impl.db.First(user, userId).Error; err != nil {
		return &User{}, err
	}
	return user, nil
}

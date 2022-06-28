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

package db

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type JSONB map[string]interface{}

// Value implements the database/sql/driver Valuer interface.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	res, err := json.Marshal(j)
	return string(res), err
}

// Scan implements the database/sql Scanner interface.
func (j *JSONB) Scan(src interface{}) error {

	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	srcCopy := make([]byte, len(source))
	copy(srcCopy, source)
	return j.DecodeText(nil, srcCopy)
}

// MarshalJSON to output non base64 encoded []byte
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	t := (map[string]interface{})(j)
	return json.Marshal(t)
}

// UnmarshalJSON to deserialize []byte
func (j *JSONB) UnmarshalJSON(b []byte) error {
	t := map[string]interface{}{}
	err := json.Unmarshal(b, &t)
	*j = t
	return err
}

// GormDataType gorm common data type
func (j JSONB) GormDataType() string {
	return "jsonb"
}

// GormDBDataType gorm db data type
func (JSONB) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

func (j JSONB) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	src, err := j.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return append(buf, src...), nil
}

func (j *JSONB) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return nil
	}
	t := map[string]interface{}{}
	err := json.Unmarshal(src, &t)
	*j = t
	return err
}

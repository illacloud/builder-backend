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

package smtp

type Resource struct {
	Host     string `validate:"required"`
	Port     int    `validate:"gt=0"`
	Username string
	Password string
}

type Action struct {
	From        string   `validate:"required"`
	To          []string `validate:"required,gt=0,dive,required"`
	Bcc         []string
	Cc          []string
	SetReplyTo  bool
	ReplyTo     string `validate:"required_unless=SetReplyTo false"`
	Subject     string `validate:"required"`
	ContentType string `validate:"required,oneof=text/plain text/html"`
	Body        string `validate:"required"`
	Attachment  []Attachment
}

type Attachment struct {
	Data        string
	Name        string
	ContentType string
}

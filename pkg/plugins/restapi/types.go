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

package restapi

import (
	"github.com/illa-family/builder-backend/internal/util"
)

type RESTOptions struct {
	BaseURL        string `validate:"required"`
	URLParams      []map[string]string
	Headers        []map[string]string
	Cookies        []map[string]string
	Authentication string            `validate:"oneof=none basic bearer"`
	AuthContent    map[string]string `validate:"required_unless=Authentication none"`
}

type RESTTemplate struct {
	URL       string
	Method    string `validate:"oneof=GET POST PUT PATCH DELETE"`
	BodyType  string `validate:"oneof=none form-data x-www-form-urlencoded raw json binary"`
	UrlParams []map[string]string
	Headers   []map[string]string
	Body      interface{} `validate:"required_unless=BodyType none"`
	Cookies   []map[string]string
}

type RawBody struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type BinaryBody string

type RecordBody struct {
	Key   string
	Value string
}

func (t *RESTTemplate) ReflectBodyToRaw() *RawBody {
	rbd := &RawBody{}
	rb, _ := t.Body.(map[string]interface{})
	for k, v := range rb {
		switch k {
		case "type":
			rbd.Type, _ = v.(string)
		case "content":
			rbd.Content, _ = v.(string)
		}
	}
	return rbd
}

func (t *RESTTemplate) ReflectBodyToBinary() []byte {
	bs, _ := t.Body.(string)
	bi := util.StringToBinary(bs, 8)
	return bi
}

func (t *RESTTemplate) ReflectBodyToRecord() []*RecordBody {
	rs := make([]*RecordBody, 0)
	objs, _ := t.Body.([]interface{})
	for _, v := range objs {
		obj, _ := v.(map[string]interface{})
		record := &RecordBody{}
		for k, v2 := range obj {
			switch k {
			case "key":
				record.Key, _ = v2.(string)
			case "value":
				record.Value, _ = v2.(string)
			}
		}
		rs = append(rs, record)
	}
	return rs
}

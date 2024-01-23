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

package condition

import "errors"

type ConditionTemplate struct {
	URL       string
	Method    string `validate:"oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
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

func (t *ConditionTemplate) ReflectBodyToRaw() *RawBody {
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

func resolveIntFieldsFromActionOptions(actionOptions map[string]interface{}, fieldName string) (int, error) {
	raw, hit := actionOptions[fieldName]
	if !hit {
		return 0, errors.New("missing " + fieldName + " field")

	}
	numberInFloat, numberAssertPass := raw.(float64)
	number := int(numberInFloat)
	if !numberAssertPass {
		return 0, errors.New(fieldName + " field which in action options assert failed")

	}
	return number, nil
}

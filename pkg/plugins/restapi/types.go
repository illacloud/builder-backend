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

type RESTOptions struct {
	BaseURL        string `validate:"required"`
	URLParams      []map[string]string
	Headers        []map[string]string
	Cookies        []map[string]string
	Authentication string            `validate:"oneof=none basic bearer"`
	AuthContent    map[string]string `validate:"required_unless=Authentication none"`
}

type RESTTemplate struct {
	URL       string `validate:"required"`
	Method    string `validate:"oneof=GET POST PUT PATCH DELETE"`
	BodyType  string `validate:"oneof=none form-data x-www-form-urlencoded json"`
	UrlParams []map[string]string
	Headers   []map[string]string
	Body      map[string]string `validate:"required_unless=BodyType none"`
	Cookies   []map[string]string
}

// Copyright 2023 Illa Soft, Inc.
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

package huggingface

const (
	METHOD_GET    = "GET"
	METHOD_POST   = "POST"
	METHOD_PUT    = "PUT"
	METHOD_DELETE = "DELETE"
	METHOD_PATCH  = "PATCH"

	BODY_NONE   = "none"
	BODY_RAW    = "raw"
	BODY_BINARY = "binary"
	BODY_FORM   = "form-data"
	BODY_XWFU   = "x-www-form-urlencoded"

	AUTH_NONE   = "none"
	AUTH_BASIC  = "basic"
	AUTH_BEARER = "bearer"

	HF_API_ADDRESS = "https://api-inference.huggingface.co/models/"

	INPUT_PAIRS_MODE  = "pairs"
	INPUT_TEXT_MODE   = "text"
	INPUT_JSON_MODE   = "json"
	INPUT_BINARY_MODE = "binary"
)

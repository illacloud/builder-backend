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

package tencentcos

const (
	BUCKET_URL_TEMPLATE = "https://%s.cos.%s.myqcloud.com" // the template is {bucket name}, {region}
)

const (
	LIST_COMMAND             = "list"
	GET_DOWNLOAD_URL_COMMAND = "getDownloadURL"
)

type Resource struct {
	BucketName      string `validate:"required"`
	Region          string `validate:"required"`
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}

var ACLs = map[string]bool{
	"private":                   true,
	"public-read":               true,
	"public-read-write":         true,
	"authenticated-read":        true,
	"aws-exec-read":             true,
	"bucket-owner-read":         true,
	"bucket-owner-full-control": true,
}

type Action struct {
	Commands    string                 `validate:"required,oneof=list getDownloadURL"`
	CommandArgs map[string]interface{} `validate:"required"`
}

type ListCommandArgs struct {
	MaxKeys int `json:"maxKeys"`
}

type GetPresignedURLCommandArgs struct {
	FileName string `json:"fileName"`
}

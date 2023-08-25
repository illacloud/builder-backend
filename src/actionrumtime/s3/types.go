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

package s3

const (
	LIST_COMMAND         = "list"
	READ_COMMAND         = "read"
	DOWNLOAD_COMMAND     = "download"
	DELETE_COMMAND       = "delete"
	BATCH_DELETE_COMMAND = "batchDelete"
	UPLOAD_COMMAND       = "upload"
	BATCH_UPLOAD_COMMAND = "batchUpload"
)

type Resource struct {
	BucketName      string
	Region          string `validate:"required"`
	ACL             string
	Endpoint        bool
	BaseURL         string `validate:"required_unless=Endpoint false"`
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
	Commands    string                 `validate:"required,oneof=list read download delete batchDelete upload batchUpload"`
	CommandArgs map[string]interface{} `validate:"required"`
}

type ListCommandArgs struct {
	BucketName string `json:"bucketName"`
	Prefix     string `json:"prefix"`
	Delimiter  string `json:"delimiter"`
	SignedURL  bool   `json:"signedURL"`
	Expiry     int64  `json:"expiry" validate:"required_unless=SignedURL false"`
	MaxKeys    int32  `json:"maxKeys"`
}

type BaseCommandArgs struct {
	BucketName string `json:"bucketName"`
	SignedURL  bool   `json:"signedURL"`
	Expiry     int64  `json:"expiry" validate:"required_unless=SignedURL false"`
	ObjectKey  string `json:"objectKey" validate:"required"`
}

type BatchDeleteCommandArgs struct {
	BucketName    string   `json:"bucketName"`
	ObjectKeyList []string `json:"objectKeyList" validate:"required,gt=0,dive,required"`
}

type UploadCommandArgs struct {
	BucketName  string `json:"bucketName"`
	ContentType string `json:"contentType"`
	Expiry      int64  `json:"expiry"`
	ObjectKey   string `json:"objectKey" validate:"required"`
	ObjectData  string `json:"objectData"`
}

type BatchUploadCommandArgs struct {
	BucketName     string   `json:"bucketName"`
	ContentType    string   `json:"contentType"`
	Expiry         int64    `json:"expiry"`
	ObjectKeyList  []string `json:"objectKeyList" validate:"required,gt=0,dive,required"`
	ObjectDataList []string `json:"objectDataList"`
}

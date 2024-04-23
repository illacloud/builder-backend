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

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/mitchellh/mapstructure"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
)

func (s *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*cos.Client, error) {
	if err := mapstructure.Decode(resourceOptions, &s.ResourceOpts); err != nil {
		return nil, err
	}

	// init connect string
	bucketURL := fmt.Sprintf(BUCKET_URL_TEMPLATE, s.ResourceOpts.BucketName, s.ResourceOpts.Region)
	bucketURLParsed, errInParse := url.Parse(bucketURL)
	if errInParse != nil {
		return nil, errInParse
	}
	baseURL := &cos.BaseURL{
		BucketURL: bucketURLParsed,
	}

	// init client
	cosClient := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  s.ResourceOpts.AccessKeyID,
			SecretKey: s.ResourceOpts.SecretAccessKey,
			Transport: &debug.DebugRequestTransport{
				RequestHeader:  true,
				RequestBody:    true,
				ResponseHeader: true,
				ResponseBody:   true,
			},
		},
	})

	return cosClient, nil
}

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

import (
	"os"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/gomail.v2"
)

func (s *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*gomail.Dialer, error) {
	if err := mapstructure.Decode(resourceOptions, &s.ResourceOpts); err != nil {
		return nil, err
	}

	smtpDialer := gomail.NewDialer(s.ResourceOpts.Host, s.ResourceOpts.Port, s.ResourceOpts.Username, s.ResourceOpts.Password)

	return smtpDialer, nil
}

func attachSizeLimiter(contentLength int64) bool {
	limitStr := os.Getenv("ILLA_S3_LIMIT")
	limit64, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		limit64 = 5
	}
	objectSize := contentLength / 1024
	objectSize /= 1024

	if objectSize > limit64 {
		return true
	}
	return false
}

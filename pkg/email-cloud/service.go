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

package email_cloud

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	BASEURL          = "http://k8s-producti-cloud-48ebc94d00-1633650720.ap-northeast-1.elb.amazonaws.com/v1/"
	SUBSCRIBE        = "subscribe"
	VERIFICATIONCODE = "code"
)

func SendSubscriptionEmail(email string) error {
	client := resty.New()
	resp, err := client.R().
		SetBody(map[string]string{"email": email}).
		Post(BASEURL + SUBSCRIBE)
	if resp.StatusCode() != http.StatusOK || err != nil {
		return errors.New("failed to send subscription email")
	}
	fmt.Printf("response: %+v, err: %+v", resp, err)
	return nil
}

func SendVerificationEmail(email, code, usage string) error {
	client := resty.New()
	resp, err := client.R().
		SetBody(map[string]string{"email": email, "code": code, "usage": usage}).
		Post(BASEURL + VERIFICATIONCODE)
	if resp.StatusCode() != http.StatusOK || err != nil {
		return errors.New("failed to send verification code email")
	}
	fmt.Printf("response: %+v, err: %+v", resp, err)
	return nil
}

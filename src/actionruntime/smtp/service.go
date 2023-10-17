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
	"encoding/base64"
	"errors"
	"io"
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/gomail.v2"
)

type Connector struct {
	ResourceOpts Resource
	ActionOpts   Action
}

func (s *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &s.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate smtp options
	validate := validator.New()
	if err := validate.Struct(s.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get smtp dialer
	smtpDialer, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	// Dial dials and authenticates to an SMTP server
	sendCloser, err := smtpDialer.Dial()
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer sendCloser.Close()

	return common.ConnectionResult{Success: true}, nil
}

func (s *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{
		Success: true,
		Schema:  nil,
	}, nil
}

func (s *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get smtp dialer
	smtpDialer, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get smtp dialer")
	}

	// format smtp action
	if err := mapstructure.Decode(actionOptions, &s.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate smtp options
	validate := validator.New()
	if err := validate.Struct(s.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build message
	emailMessage := gomail.NewMessage()

	// set header
	emailMessage.SetHeader("From", s.ActionOpts.From)
	emailMessage.SetHeader("To", s.ActionOpts.To...)
	emailMessage.SetHeader("Subject", s.ActionOpts.Subject)
	if len(s.ActionOpts.Bcc) != 0 {
		emailMessage.SetHeader("Bcc", s.ActionOpts.Bcc...)
	}
	if len(s.ActionOpts.Cc) != 0 {
		emailMessage.SetHeader("Cc", s.ActionOpts.Cc...)
	}
	if s.ActionOpts.SetReplyTo {
		emailMessage.SetHeader("Reply-To", s.ActionOpts.ReplyTo)
	}

	// set body
	emailMessage.SetBody(s.ActionOpts.ContentType, s.ActionOpts.Body)

	// attach
	if len(s.ActionOpts.Attachment) != 0 {
		for _, attach := range s.ActionOpts.Attachment {
			attachDataBytes, err := base64.StdEncoding.DecodeString(attach.Data)
			if err != nil {
				continue
			}
			decodedAttachDataString, err := url.QueryUnescape(string(attachDataBytes))
			if err != nil {
				continue
			}
			contentLength := len(decodedAttachDataString)
			if attachSizeLimiter(int64(contentLength)) {
				continue
			}
			emailMessage.Attach(attach.Name, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write([]byte(decodedAttachDataString))
				return err
			}))
		}
	}

	if err := smtpDialer.DialAndSend(emailMessage); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{{"message": "email sent successfully"}},
	}, nil
}

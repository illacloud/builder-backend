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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

type RESTOptions struct {
	BaseURL        string `validate:"required"`
	URLParams      []map[string]string
	Headers        []map[string]string
	Cookies        []map[string]string
	SelfSignedCert bool
	Certs          map[string]string `validate:"required_unless=SelfSignedCert false"`
	Authentication string            `validate:"oneof=none basic bearer digest oauth1.0 hawk aws"`
	AuthContent    map[string]string `validate:"required_unless=Authentication none"`
}

type RESTTemplate struct {
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
	sdec, _ := base64.StdEncoding.DecodeString(bs)
	return sdec
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

func (t *RESTTemplate) ReflectBodyToMap() map[string]string {
	rs := make(map[string]string)
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
		rs[record.Key] = record.Value
	}
	return rs
}

type FormDataBody struct {
	Key   string
	Type  string
	Value string
}

func (t *RESTTemplate) ReflectBodyToMultipart() (texts, files map[string]string) {
	rs := make(map[string]string)
	fs := make(map[string]string)
	objs, _ := t.Body.([]interface{})
	for _, v := range objs {
		obj, _ := v.(map[string]interface{})
		record := &FormDataBody{}
		for k, v2 := range obj {
			switch k {
			case "key":
				record.Key, _ = v2.(string)
			case "value":
				record.Value, _ = v2.(string)
			case "type":
				record.Type, _ = v2.(string)
			}
		}
		if record.Type == "text" {
			rs[record.Key] = record.Value
		} else if record.Type == "file" {
			v3, _ := base64.StdEncoding.DecodeString(record.Value)
			fs[record.Key] = string(v3)
		}
	}
	return rs, fs
}

func loadSelfSignedCerts(server string, certs map[string]string) (*tls.Config, error) {
	cfg := &tls.Config{}
	switch certs["mode"] {
	case VERIFY_MODE_CA:
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(certs["caCert"])) {
			return nil, fmt.Errorf("failed to append caCert")
		}
		cfg = &tls.Config{
			RootCAs:    caCertPool,
			ServerName: server,
		}
	case VERIFY_MODE_FULL:
		cert, err := tls.X509KeyPair([]byte(certs["clientCert"]), []byte(certs["clientKey"]))
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(certs["caCert"])) {
			return nil, fmt.Errorf("failed to append caCert")
		}
		cfg = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ServerName:   server,
		}
	case VERIFY_MODE_SKIP:
		cfg = &tls.Config{
			InsecureSkipVerify: true,
		}
	default:
		break
	}

	return cfg, nil
}

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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

const (
	FIELD_CONTEXT    = "context"
	FIELD_TEMPLATE   = "template"
	FIELD_URL        = "url"
	FIELD_BODY       = "body"
	FIELD_METHOD     = "method"
	FIELD_COOKIES    = "cookies"
	FIELD_HEADERS    = "headers"
	FIELD_BODY_TYPE  = "bodyType"
	FIELD_URL_PARAMS = "urlParams"
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
	Context   map[string]interface{}
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
	fmt.Printf("[DUMP] ReflectBodyToRaw().rb: %+v\n", rb)
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

func (r *RawBody) UnmarshalRawBody() (interface{}, string) {
	switch r.Type {
	case "json":
		rawStr, _ := strconv.Unquote(r.Content)
		fmt.Printf("[UnmarshalRawBody()] r.Content: %+v\n", r.Content)
		fmt.Printf("[UnmarshalRawBody()] rawStr: %+v\n", rawStr)
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(rawStr), &data); err == nil {
			return data, "application/json"
		}
		var listData []map[string]interface{}
		if err := json.Unmarshal([]byte(rawStr), &listData); err == nil {
			return listData, "application/json"
		}
		// try non-unquote
		if err := json.Unmarshal([]byte(r.Content), &data); err == nil {
			return data, "application/json"
		}
		if err := json.Unmarshal([]byte(r.Content), &listData); err == nil {
			return listData, "application/json"
		}
		return rawStr, "application/json"
	case "xml":
		rawStr, err := strconv.Unquote(r.Content)
		if err != nil {
			return r.Content, "application/xml"
		}
		return rawStr, "application/xml"
	case "html":
		return []byte(r.Content), "text/html"
	case "text":
		return []byte(r.Content), "text/plain"
	case "javascript":
		return []byte(r.Content), "application/javascript"
	}
	return nil, ""
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
	Value interface{}
}

func (t *RESTTemplate) ReflectBodyToMultipart() (texts map[string]string, files map[string]map[string]string) {
	rs := make(map[string]string)
	fs := make(map[string]map[string]string)
	objs, _ := t.Body.([]interface{})
	for _, v := range objs {
		obj, _ := v.(map[string]interface{})
		record := &FormDataBody{}
		for k, v2 := range obj {
			switch k {
			case "key":
				record.Key, _ = v2.(string)
			case "value":
				record.Value = v2
			case "type":
				record.Type, _ = v2.(string)
			}
		}
		if record.Type == "text" {
			rs[record.Key], _ = record.Value.(string)
		} else if record.Type == "file" {
			fileData, ok := record.Value.(map[string]interface{})
			if !ok {
				fs[record.Key] = map[string]string{"filename": "", "data": ""}
			}
			strData, ok := fileData["data"].(string)
			if !ok {
				strData = ""
			}
			filename, ok := fileData["filename"].(string)
			if !ok {
				filename = ""
			}
			v3, _ := base64.StdEncoding.DecodeString(strData)
			fs[record.Key] = map[string]string{"filename": filename, "data": string(v3)}
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

func (q *RESTTemplate) DoesContextValied(rawTemplate map[string]interface{}) bool {
	contextRaw, hit := rawTemplate[FIELD_CONTEXT]
	if !hit {
		return false
	}
	contextAsserted, assertPass := contextRaw.(map[string]interface{})
	if !assertPass {
		return false
	}
	return len(contextAsserted) > 0
}

func (q *RESTTemplate) SetRawQueryAndContext(rawTemplate map[string]interface{}) error {
	// set context
	contextRaw, hit := rawTemplate[FIELD_CONTEXT]
	if !hit {
		return errors.New("missing context field SetRawQueryAndContext() in query")
	}
	contextAsserted, assertPass := contextRaw.(map[string]interface{})
	if !assertPass {
		return errors.New("context field assert failed in SetRawQueryAndContext() method")
	}
	q.Context = contextAsserted
	return nil
}

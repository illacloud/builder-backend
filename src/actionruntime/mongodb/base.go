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

package mongodb

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*mongo.Client, error) {
	if err := mapstructure.Decode(resourceOptions, &m.Resource); err != nil {
		return nil, err
	}

	// format connection string
	uri := ""
	if m.Resource.ConfigType == GUI_OPTIONS {
		mOptions := GUIOptions{}
		if err := mapstructure.Decode(m.Resource.ConfigContent, &mOptions); err != nil {
			return nil, err
		}
		if mOptions.DatabaseUsername != "" && mOptions.DatabasePassword != "" {
			escapedPassword := url.QueryEscape(mOptions.DatabasePassword)
			uri = fmt.Sprintf("%s://%s:%s@%s", CONNECTION_FORMAT[mOptions.ConnectionFormat],
				mOptions.DatabaseUsername, escapedPassword, mOptions.Host)
		} else {
			uri = fmt.Sprintf("%s://%s", CONNECTION_FORMAT[mOptions.ConnectionFormat], mOptions.Host)
		}
		if mOptions.ConnectionFormat == STANDARD_FORMAT {
			uri = uri + ":" + mOptions.Port
		}
		if mOptions.DatabaseName != "" {
			uri = uri + "/" + mOptions.DatabaseName
			if mOptions.DatabaseName != "admin" {
				uri += "?authSource=admin"
			}
		}
	} else if m.Resource.ConfigType == URI_OPTIONS {
		mOptions := URIOptions{}
		if err := mapstructure.Decode(m.Resource.ConfigContent, &mOptions); err != nil {
			return nil, err
		}
		uri = mOptions.URI
	} else {
		return nil, errors.New("unsupported mongodb options")
	}

	var client *mongo.Client
	var err error

	// TLS: self-signed certificate
	var credential options.Credential
	var tlsConfig tls.Config
	if m.Resource.SSL.Open == true && m.Resource.SSL.CA != "" {
		credential = options.Credential{AuthMechanism: "MONGODB-X509"}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM([]byte(m.Resource.SSL.CA)); !ok {
			return nil, errors.New("format MongoDB TLS CA Cert failed")
		}
		tlsConfig = tls.Config{RootCAs: pool}
		if m.Resource.SSL.Client != "" {
			splitIndex := bytes.Index([]byte(m.Resource.SSL.Client), []byte("-----\n-----"))
			if splitIndex <= 0 {
				return nil, errors.New("format MongoDB TLS Client Key Pair failed")
			}
			clientKeyPairSlice := []string{m.Resource.SSL.Client[:splitIndex+6], m.Resource.SSL.Client[splitIndex+6:]}
			clientCert := ""
			clientKey := ""
			if strings.Contains(clientKeyPairSlice[0], "CERTIFICATE") {
				clientCert = clientKeyPairSlice[0]
				clientKey = clientKeyPairSlice[1]
			} else {
				clientCert = clientKeyPairSlice[1]
				clientKey = clientKeyPairSlice[0]
			}
			ccBlock, _ := pem.Decode([]byte(clientCert))
			ckBlock, _ := pem.Decode([]byte(clientKey))
			if (ccBlock != nil && ccBlock.Type == "CERTIFICATE") && (ckBlock != nil || strings.Contains(ckBlock.Type, "PRIVATE KEY")) {
				cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
				if err != nil {
					return nil, err
				}
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}
	}

	// connect to mongodb
	clientOptions := options.Client().ApplyURI(uri)
	if m.Resource.SSL.Open == true && m.Resource.SSL.CA != "" {
		clientOptions = clientOptions.SetTLSConfig(&tlsConfig).SetAuth(credential)
	}
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}

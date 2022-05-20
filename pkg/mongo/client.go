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

package mongo

import (
	"context"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

type Config struct {
	MongodbURI string `env:"ILLA_MONGODB_URI"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

// NewMongoClient return a new mongo connection to work with
func NewMongoClient(cfg *Config, logger *zap.SugaredLogger) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cfg.MongodbURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logger.Errorw("error in connecting mongo ", "mongoDB", cfg, "err", err)
		return nil, err
	}
	// check mongoDB connection
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		logger.Errorw("error in testing connection with mongo", "mongoDB", cfg, "err", err)
		return nil, err
	} else {
		logger.Infow("connected with mongo", "mongoDB", cfg)
	}
	return client, err
}

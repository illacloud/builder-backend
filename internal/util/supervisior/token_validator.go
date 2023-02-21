package supervisior

import (
	"encoding/base64"
	"errors"
	"sort"

	"crypto/md5"

	"github.com/caarlos0/env"
)

type Config struct {
	SupervisiorInternalAPI string `env:"ILLA_SUPERVISIOR_INTERNAL_API" envDefault:"http://127.0.0.1:9001/api/v1"`
	Secret                 string `env:"ILLA_SECRET_KEY" envDefault:""`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

type RequestTokenValidator struct {
	Config *Config
}

func NewRequestTokenValidator() (*RequestTokenValidator, error) {
	cfg, err := GetConfig()
	if err != nil {
		return nil, errors.New("can not get config.")
	}
	return &RequestTokenValidator{
		Config: cfg,
	}, nil
}

func (r *RequestTokenValidator) GenerateValidateToken(input ...string) string {
	return r.GenerateValidateTokenBySliceParam(input)
}

func (r *RequestTokenValidator) GenerateValidateTokenBySliceParam(input []string) string {
	var concatr string
	sort.Strings(input)
	for _, str := range input {
		concatr += str
	}
	concatr += r.Config.Secret
	hash := md5.Sum([]byte(concatr))
	var hashConverted []byte = hash[:]

	return base64.StdEncoding.EncodeToString(hashConverted)
}

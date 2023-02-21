package tokenvalidator

import (
	"encoding/base64"
	"sort"

	"crypto/md5"

	"github.com/caarlos0/env"
)

type Config struct {
	Secret string `env:"ILLA_SECRET_KEY" envDefault:""`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

type RequestTokenValidator struct {
	Config *Config
}

func NewRequestTokenValidator() *RequestTokenValidator {
	cfg, err := GetConfig()
	if err != nil {
		panic("this environment param ILLA_SECRET_KEY must be setted.")
	}
	return &RequestTokenValidator{
		Config: cfg,
	}
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

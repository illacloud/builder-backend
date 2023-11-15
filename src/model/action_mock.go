package model

type MockConfig struct {
	Enabled              bool   `json:"enabled"`
	MockData             string `json:"mockData"`
	EnableForReleasedApp bool   `json:"enableForReleasedApp"`
}

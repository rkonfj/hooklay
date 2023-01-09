package internal

import (
	"os"

	"gopkg.in/yaml.v3"
)

func NewConfig() *Config {
	file, err := os.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}
	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}
	return &config
}

type Config struct {
	ApiVersion int `yaml:"apiVersion"`
	ServerPort int `yaml:"serverPort"`
	Hook       string
	Security   RelaySecurity
	Targets    []*RelayTarget
	Templates  []*Template
}

type RelaySecurity struct {
	Token SecurityToken
}

type SecurityToken struct {
	Header string
	Value  string
}

type RelayTarget struct {
	Name               string
	Enabled            bool
	Url                string
	BodyTemplate       string
	IdempotentTemplate string
	Conditions         []Condition
}

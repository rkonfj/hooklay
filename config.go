package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

var config Config

func init() {
	file, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	ApiVersion int `yaml:"apiVersion"`
	ServerPort int `yaml:"serverPort"`
	Relays     []Relay
}

type Relay struct {
	Hook     string
	Enabled  bool
	Security RelaySecurity
	Targets  []RelayTarget
}

type RelaySecurity struct {
	Token     SecurityToken
	Signature SecuritySignature
}

type SecurityToken struct {
	Header string
	Value  string
}

type SecuritySignature struct {
	Method   string
	Password string
	Header   string
}

type RelayTarget struct {
	Name       string
	Enabled    bool
	Url        string
	Body       string
	Conditions []TargetCondition
}

type TargetCondition struct {
	Key      string
	Operator string
	Value    string
}

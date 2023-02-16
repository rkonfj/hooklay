package config

import (
	"os"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
)

var Conf *Config = new(Config)

func init() {
	b, err := os.ReadFile("config.yml")
	if err != nil {
		b, err = os.ReadFile("/etc/kdt.yml")
		if err != nil {
			log.Fatal(err)
		}
	}
	yaml.Unmarshal(b, Conf)
}

type Config struct {
	Kubeconfig     string
	Environments   map[string]*DeployContext
	IgnoreWarnning bool `yaml:"ignoreWarnning"`
	Jifa           string
	Gitlab         Gitlab
	Notification   Notification
}

type DeployContext struct {
	From        Env
	To          Env
	GitlabGroup string `yaml:"gitlabGroup"`
}

type Env struct {
	Context   string
	Namespace string
}

type Gitlab struct {
	BaseURL string `yaml:"baseURL"`
	Token   string
	Groups  map[string]map[string]int
}

type Notification struct {
	Dingtalk string
}

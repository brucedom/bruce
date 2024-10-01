package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Execution struct {
	Name    string `yaml:"name"`
	Action  string `yaml:"action"`
	Type    string `yaml:"type"`
	Cadence int    `yaml:"cadence"`
	Target  string `yaml:"target"`
}

type ServerConfig struct {
	RunnerID      string      `yaml:"runner-id"`
	Authorization string      `yaml:"authorization"`
	Endpoint      string      `yaml:"endpoint"`
	Execution     []Execution `yaml:"execution"`
}

func ReadServerConfig(l string, sc *ServerConfig) error {
	data, err := os.ReadFile(l)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &sc)
}

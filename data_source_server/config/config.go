package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"quote/common/kafka"
	"quote/common/log"
)

type Etcd struct {
	Addr string `yaml:"addr"`
}

type GlobalConfig struct {
	IP    string                `yaml:"ip"`
	Port  int                   `yaml:"port"`
	Log   map[string]log.Config `yaml:"log"`
	Etcd  Etcd                  `yaml:"etcd"`
	Kafka kafka.Config          `yaml:"kafka"`
}

func Parse(f string) (*GlobalConfig, error) {
	d, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var r GlobalConfig
	if err = yaml.Unmarshal(d, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

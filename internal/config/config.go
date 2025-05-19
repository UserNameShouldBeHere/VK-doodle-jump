package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type ClientCfg struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DBCfg struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Admins   []int  `yaml:"admins"`
}

type ServerCfg struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	DB   DBCfg  `yaml:"db"`
}

type Config struct {
	Client ClientCfg `yaml:"client"`
	Server ServerCfg `yaml:"server"`
}

func Parse(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func DefaultConfig() *Config {
	return &Config{
		Client: ClientCfg{
			Host: "109.120.183.59",
			Port: 443,
		},
		Server: ServerCfg{
			Host: "127.0.0.1",
			Port: 3001,
			DB: DBCfg{
				Host:     "localhost",
				Port:     3301,
				User:     "admin",
				Password: "pass",
				Admins:   []int{},
			},
		},
	}
}

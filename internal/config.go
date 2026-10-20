package config

import (
	"encoding/json"
	"io"
	"os"
)

const configFilePath = "/.gatorconfig.json"

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func Read() (*Config, error) {
	config := Config{}
	HOME, _ := os.UserHomeDir()

	file, err := os.Open(HOME + configFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) Setuser(name string) error {
	c.Current_user_name = name
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	HOME, _ := os.UserHomeDir()
	err = os.WriteFile(HOME+configFilePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

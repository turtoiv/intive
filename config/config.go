package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"fmt"
)

type DBConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	User string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func NewDBConfig(configPath string) (string, error) {
	config := &DBConfig{}

	file, err := os.Open(configPath)
    if err != nil {
        return "", err
    }
	defer file.Close()
	
	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		fmt.Println("unable to decode file")
		return "", err
	}

	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.User, config.Password, config.Host, config.Port, config.Database)
	
	return connString, nil
}

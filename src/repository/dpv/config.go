package dpv

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Language struct {
	Key  string `yaml:"key"`
	Name string `yaml:"name"`
	Icon string `yaml:"icon"`
}
type Config struct {
	DB struct {
		Host string `yaml:"host"`
		Root string `yaml:"root"`
		Port int    `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"db"`
	Auth struct {
		DpvSecretKey    string `yaml:"dpv_secret_key"`
		DpvTokenSeconds int    `yaml:"dpv_token_seconds"`
	} `yaml:"auth"`
	Server struct {
		Words1 string `yaml:"words1"`
		Words2 string `yaml:"words2"`
		Words3 string `yaml:"words3"`
		Words4 string `yaml:"words4"`
	}
	Settings struct {
		Version   string
		UserTypes []string `yaml:"user_types"`
	} `yaml:"settings"`
	Path string
}

var ConfigInstance *Config

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		wd, _ := os.Getwd()
		return nil, fmt.Errorf("could not load config file, looking for %v in %v: %w", configPath, wd, err)
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, fmt.Errorf("could not decode config file: %w", err)
	}
	config.Path = configPath[:len(configPath)-len("config.yml")]
	return config, nil
}

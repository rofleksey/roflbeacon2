package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	BaseApiURL string `yaml:"baseApiURL" validate:"required"`

	Log struct {
		TelegramToken  string `yaml:"telegramToken"`
		TelegramChatID string `yaml:"telegramChatID"`
	} `yaml:"log"`

	Telegram struct {
		Token       string `yaml:"token" validate:"required"`
		AdminChatID string `yaml:"adminChatID" validate:"required"`
	} `yaml:"telegram"`

	DB struct {
		User     string `yaml:"user" validate:"required"`
		Pass     string `yaml:"pass" validate:"required"`
		Host     string `yaml:"host" validate:"required"`
		Database string `yaml:"database" validate:"required"`
	} `yaml:"db"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var result Config
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	if result.BaseApiURL == "" {
		result.BaseApiURL = "https://beacon.rofleksey.ru"
	}
	if result.DB.User == "" {
		result.DB.User = "postgres"
	}
	if result.DB.Pass == "" {
		result.DB.Pass = "postgres"
	}
	if result.DB.Host == "" {
		result.DB.Host = "localhost:5432"
	}
	if result.DB.Database == "" {
		result.DB.Database = "roflbeacon2"
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(result); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &result, nil
}

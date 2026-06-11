package config

import (
	"encoding/json" // converts Go structs to JSON and back
	"os"            // read/write files
	"path/filepath" // creates persistent file paths for any os
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {

	configPath, err := GetConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(fileData, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil // success!
}
func Write(cfg Config) error {

	configPath, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	fileData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, fileData, 0644)
	if err != nil {
		return err
	}

	return nil // success!
}

func (cfg *Config) SetUser(userName string) error {

	cfg.CurrentUserName = userName

	return Write(*cfg)
}

func GetConfigFilePath() (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, configFileName), nil
}

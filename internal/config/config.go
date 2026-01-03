package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type DBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type JWTConfig struct {
	SecretKey  string `json:"secretKey"`
	Issuer     string `json:"issuer"`
	Audience   string `json:"audience"`
	CookieName string `json:"cookieName"`
}

type Config struct {
	JWT JWTConfig `json:"jwt"`
	DB  DBConfig  `json:"db"`
}

func Load(cfgType string) (*Config, error) {
	var cfg *Config
	var err error
	switch cfgType {
	case "dev":
		cfg, err = loadConfigJSON()
	case "prod":
		cfg, err = loadConfigEnv()
	default:
		return nil, fmt.Errorf("config loading error: invalid config type: %v, must be 'dev' or 'prod'", cfgType)
	}

	return cfg, err
}

func loadConfigEnv() (*Config, error) {
	var err error
	getEnv := func(key string) string {
		if err != nil {
			return ""
		}
		val, ok := os.LookupEnv(key)
		if !ok {
			err = fmt.Errorf("missing environment variable: %v", key)
			return ""
		}
		return val
	}

	cfg := &Config{
		JWT: JWTConfig{
			SecretKey:  getEnv("JWT_SECRET"),
			Issuer:     getEnv("JWT_ISSUER"),
			Audience:   getEnv("JWT_AUDIENCE"),
			CookieName: getEnv("JWT_COOKIE"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			DBName:   getEnv("DB_NAME"),
		},
	}

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadConfigJSON() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("json config file oppening error: %v", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("json config file reading error: %v", err)
	}

	return &cfg, nil
}

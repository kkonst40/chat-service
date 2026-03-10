package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type DBConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
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

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type RateLimiterConfig struct {
	Limit                  int
	MaxBurst               int
	CleanupIntervalSeconds int
	IPIdleLifetimeSeconds  int
}

type Config struct {
	Env                   string            `json:"env"`
	Port                  string            `json:"port"`
	SSOAddress            string            `json:"ssoAddress"`
	RequestTimeoutSeconds int               `json:"requestTimeout"`
	LoginCacheTTLHours    int               `json:"loginCacheTTL"`
	WSConnsPerIP          int               `json:"wsConnsPerIP"`
	JWT                   JWTConfig         `json:"jwt"`
	DB                    DBConfig          `json:"db"`
	Redis                 RedisConfig       `json:"redis"`
	RateLimiter           RateLimiterConfig `json:"rateLimiter"`
}

func Load() (*Config, error) {
	var cfg *Config
	var err error
	switch runtime.GOOS {
	case "windows":
		cfg, err = loadConfigJSON()
	case "linux":
		cfg, err = loadConfigEnv()
	default:
		return nil, fmt.Errorf("config loading error")
	}

	return cfg, err
}

func loadConfigEnv() (*Config, error) {
	var (
		errMissing error
		errNotInt  error
		errResult  error
	)

	getEnvString := func(key string) string {
		val, ok := os.LookupEnv(key)
		if !ok {
			if errMissing == nil {
				errMissing = fmt.Errorf("missing environment variables: %s", key)
			} else {
				errMissing = fmt.Errorf("%w, %s", errMissing, key)
			}
			return ""
		}
		return val
	}

	getEnvInt := func(key string) int {
		val, ok := os.LookupEnv(key)
		if !ok {
			if errMissing == nil {
				errMissing = fmt.Errorf("missing environment variables: %s", key)
			} else {
				errMissing = fmt.Errorf("%w, %s", errMissing, key)
			}
			return 0
		}

		valInt, err := strconv.Atoi(val)
		if err != nil {
			if errNotInt == nil {
				errNotInt = fmt.Errorf("environment variables must be integer: %s", val)
			} else {
				errNotInt = fmt.Errorf("%w, %s", errNotInt, val)
			}
			return 0
		}

		return valInt
	}

	cfg := &Config{
		Env:                   getEnvString("ENV"),
		Port:                  getEnvString("PORT"),
		SSOAddress:            getEnvString("SSO_URL"),
		RequestTimeoutSeconds: getEnvInt("REQUEST_TIMEOUT"),
		LoginCacheTTLHours:    getEnvInt("LOGIN_CACHE_TTL"),
		WSConnsPerIP:          getEnvInt("WSCONNS_PER_IP"),
		JWT: JWTConfig{
			SecretKey:  getEnvString("JWT_SECRET"),
			Issuer:     getEnvString("JWT_ISSUER"),
			Audience:   getEnvString("JWT_AUDIENCE"),
			CookieName: getEnvString("JWT_COOKIE"),
		},
		DB: DBConfig{
			Host:     getEnvString("DB_HOST"),
			Port:     getEnvString("DB_PORT"),
			User:     getEnvString("DB_USER"),
			Password: getEnvString("DB_PASSWORD"),
			DBName:   getEnvString("DB_NAME"),
		},
		Redis: RedisConfig{
			Host:     getEnvString("REDIS_HOST"),
			Port:     getEnvString("REDIS_PORT"),
			Password: getEnvString("REDIS_PASSWORD"),
			DB:       getEnvInt("REDIS_DB"),
		},
		RateLimiter: RateLimiterConfig{
			Limit:                  getEnvInt("IP_RATE_LIMIT"),
			MaxBurst:               getEnvInt("IP_RATE_MAX_BURST"),
			CleanupIntervalSeconds: getEnvInt("IP_RATE_CLEANUP_INTERVAL"),
			IPIdleLifetimeSeconds:  getEnvInt("IP_RATE_IDLE_LIFETIME"),
		},
	}
	if errMissing != nil {
		errResult = errMissing
	}

	if errNotInt != nil {
		if errResult == nil {
			errResult = errNotInt
		} else {
			errResult = fmt.Errorf("%w; %w", errResult, errNotInt)
		}
	}

	if errResult != nil {
		return nil, errResult
	}

	return cfg, nil
}

func loadConfigJSON() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	currDir := filepath.Dir(exePath)
	file, err := os.Open(filepath.Join(currDir, "config.json"))
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

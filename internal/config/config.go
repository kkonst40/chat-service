package config

import (
	"fmt"
	"os"
	"strconv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	SecretKey  string
	Issuer     string
	Audience   string
	CookieName string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Host string
	Port string
}

type RateLimiterConfig struct {
	Limit                  int
	MaxBurst               int
	CleanupIntervalSeconds int
	IPIdleLifetimeSeconds  int
}

type Config struct {
	Env                   string
	Port                  string
	SSOAddress            string
	RequestTimeoutSeconds int
	LoginCacheTTLHours    int
	WSConnsPerIP          int
	JWT                   JWTConfig
	DB                    DBConfig
	Redis                 RedisConfig
	Kafka                 KafkaConfig
	RateLimiter           RateLimiterConfig
}

func Load() (*Config, error) {
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
			User:     getEnvString("POSTGRES_USER"),
			Password: getEnvString("POSTGRES_PASSWORD"),
			DBName:   getEnvString("POSTGRES_DB"),
		},
		Redis: RedisConfig{
			Host:     getEnvString("REDIS_HOST"),
			Port:     getEnvString("REDIS_PORT"),
			Password: getEnvString("REDIS_PASSWORD"),
			DB:       getEnvInt("REDIS_DB"),
		},
		Kafka: KafkaConfig{
			Host: getEnvString("KAFKA_HOST"),
			Port: getEnvString("KAFKA_PORT"),
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

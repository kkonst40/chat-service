package config

import (
	"path/filepath"

	"github.com/kkonst40/ichat/internal/util"
)

type JWTConfig struct {
	SecretKey  string `json:"secretKey"`
	Issuer     string `json:"issuer"`
	Audience   string `json:"audience"`
	CookieName string `json:"cookieName"`
}

func LoadJwtConfig() (*JWTConfig, error) {
	dir, err := util.GetCurrentDir()
	if err != nil {
		return nil, err
	}

	var jwtConfig JWTConfig
	jwtPath := filepath.Join(dir, "config", "jwt_config.json")
	err = util.ReadJson(jwtPath, &jwtConfig)
	if err != nil {
		return nil, err
	}

	return &jwtConfig, nil
}

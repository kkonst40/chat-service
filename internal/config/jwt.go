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

func MustLoad() *JWTConfig {
	dir, err := util.GetCurrentDir()
	if err != nil {
		panic(err)
	}

	var jwtConfig JWTConfig
	jwtPath := filepath.Join(dir, "config", "jwt_config.json")
	err = util.ReadJson(jwtPath, &jwtConfig)
	if err != nil {
		panic(err)
	}

	return &jwtConfig
}

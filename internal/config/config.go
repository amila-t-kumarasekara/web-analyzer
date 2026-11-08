package config

import (
	"os"

	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	AdminAddr string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = consts.DefaultPort
	}

	adminAddr := os.Getenv("ADMIN_ADDR")
	if adminAddr == "" {
		adminAddr = consts.DefaultAdminAddr
	}

	return &Config{
		Port:      port,
		AdminAddr: adminAddr,
	}
}

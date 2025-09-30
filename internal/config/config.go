package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceHost string
	ServicePort int
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()
	host := os.Getenv("SERVICE_HOST")
	if host == "" {
		host = "localhost"
	}
	portStr := os.Getenv("SERVICE_PORT")
	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 8080
	}
	return &Config{ServiceHost: host, ServicePort: port}, nil
}

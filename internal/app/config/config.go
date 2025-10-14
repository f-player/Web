package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	User     string
}

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
	Redis       RedisConfig
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

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpiresIn, err := time.ParseDuration(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil {
		jwtExpiresIn = time.Hour * 1
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisUser := os.Getenv("REDIS_USER")

	return &Config{
		ServiceHost: host,
		ServicePort: port,
		JWT: JWTConfig{
			Secret:    jwtSecret,
			ExpiresIn: jwtExpiresIn,
		},
		Redis: RedisConfig{
			Host:     redisHost,
			Port:     redisPort,
			Password: redisPassword,
			User:     redisUser,
		},
	}, nil
}

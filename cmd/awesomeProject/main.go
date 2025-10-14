package main

import (
	"context"
	"fmt"
	"lab1-design-backend/internal/app/config"
	"lab1-design-backend/internal/app/dsn"
	"lab1-design-backend/internal/app/handler"
	"lab1-design-backend/internal/app/redis"
	"lab1-design-backend/internal/app/repository"
	"lab1-design-backend/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//	@title						API для системы DietComposition
//	@version					1.0
//	@description				API-сервер для управления заявками и продуктами в системе DietComposition.
//	@contact.name				API Support
//	@contact.email				support@example.com
//	@host						localhost:8090
//	@BasePath					/api
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	redisClient, errRedis := redis.New(context.Background(), conf.Redis)
	if errRedis != nil {
		logrus.Fatalf("error initializing redis: %v", errRedis)
	}

	hand := handler.NewHandler(rep, redisClient, &conf.JWT)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}

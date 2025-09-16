package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"lab1-design-backend/internal/app/handler"
	"lab1-design-backend/internal/app/repository"
	"log"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("../../templates/*")
	r.Static("/static", "../../resources")

	// маршруты
	r.GET("/hello", handler.GetOrders)  // список услуг
	r.GET("/order/:id", handler.GetOrder) // детально
	r.GET("/cart/:id", handler.CartHandler) // корзина


	r.Run(":8080")
	log.Println("Server down")
}

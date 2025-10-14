package handler

import (
	"lab1-design-backend/internal/app/repository"
	"lab1-design-backend/internal/app/config"
	"lab1-design-backend/internal/app/redis"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWTConfig  *config.JWTConfig
}

func NewHandler(r *repository.Repository, redis *redis.Client, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redis,
		JWTConfig:  jwtConfig,
	}
}

func (h *Handler) RegisterAPI(r *gin.RouterGroup) {

	// Доступны всем
	r.POST("/users", h.Register)
	r.POST("/auth/login", h.Login)
	r.GET("/products", h.GetProducts)
	r.GET("/products/:id", h.GetProduct)

	// Эндпоинты, доступные только авторизованным пользователям
	auth := r.Group("/")
	auth.Use(h.AuthMiddleware)
	{
		// Пользователи
		auth.POST("/auth/logout", h.Logout)
		auth.GET("/users/:id", h.GetUserData)
		auth.PUT("/users/:id", h.UpdateUserData)

		// Заявки
		auth.POST("/diet_composition/draft/products/:product_id", h.AddProductToDraft)
		auth.GET("/diet_composition/cart", h.GetCartBadge)
		auth.GET("/diet_composition", h.ListDietComposition)
		auth.GET("/diet_composition/:id", h.GetDietComposition)
		auth.PUT("/diet_composition/:id", h.UpdateDietComposition)
		auth.PUT("/diet_composition/:id/form", h.FormDietComposition)
		auth.DELETE("/diet_composition/:id", h.DeleteDietComposition)
		auth.DELETE("/diet_composition/:id/products/:product_id", h.RemoveProductFromDietComposition)
		auth.PUT("/diet_composition/:id/products/:product_id", h.UpdateMM)
	}

	// Эндпоинты, доступные только модераторам
	moderator := r.Group("/")
	moderator.Use(h.AuthMiddleware, h.ModeratorMiddleware)
	{
		// Управление продуктами (создание, изменение, удаление)
		moderator.POST("/products", h.CreateProduct)
		moderator.PUT("/products/:id", h.UpdateProduct)
		moderator.DELETE("/products/:id", h.DeleteProduct)
		moderator.POST("/products/:id/image", h.UploadProductImage)

		// Управление заявками (завершение/отклонение)
		moderator.PUT("/diet_composition/:id/resolve", h.ResolveDietComposition)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
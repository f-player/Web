package handler

import (
	"lab1-design-backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const hardcodedUserID = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// Регистрация только API роутов
func (h *Handler) RegisterAPI(r *gin.RouterGroup) {
	// Домен услуг (факторов)
	r.GET("/products", h.GetProducts)
	r.GET("/products/:id", h.GetProduct)
	r.POST("/products", h.CreateProduct)
	r.PUT("/products/:id", h.UpdateProduct)
	r.DELETE("/products/:id", h.DeleteProduct)
	r.POST("/calculation/draft/products/:product_id", h.AddProductToDraft)
	r.POST("/products/:id/image", h.UploadProductImage)

	// Домен заявок (FRAX)
	r.GET("/calculation/cart", h.GetCartBadge)
	r.GET("/calculation", h.ListCalculation)
	r.GET("/calculation/:id", h.GetCalculation)
	r.PUT("/calculation/:id", h.UpdateCalculation)
	r.PUT("/calculation/:id/form", h.FormCalculation)
	r.PUT("/calculation/:id/resolve", h.ResolveCalculation)
	r.DELETE("/calculation/:id", h.DeleteCalculation)

	// Домен м-м
	r.DELETE("/calculation/:id/products/:product_id", h.RemoveProductFromCalculation)
	r.PUT("/calculation/:id/products/:product_id", h.UpdateMM)

	// Домен пользователь
	r.POST("/users", h.Register)
	r.GET("/users/:id", h.GetUserData)
	r.PUT("/users/:id", h.UpdateUserData)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
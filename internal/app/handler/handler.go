package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"lab1-design-backend/internal/app/repository"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// 1. Главная страница
func (h *Handler) GetOrders(ctx *gin.Context) {
	searchQuery := ctx.Query("query")
	var orders []repository.Order
	var err error

	if searchQuery == "" {
		orders, err = h.Repository.GetOrders()
	} else {
		orders, err = h.Repository.GetOrdersByTitle(searchQuery)
	}

	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":         time.Now().Format("15:04:05"),
		"orders":       orders,
		"query":        searchQuery,
		"cartId":    h.Repository.GetCartId(),
    	"cartCount": h.Repository.GetCartCount(1),
	})
}

// 2. Детальная страница
func (h *Handler) GetOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	order, err := h.Repository.GetOrder(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"order":        order,
		
	})
}


// // Добавить в корзину
// func (h *Handler) AddToCart(ctx *gin.Context) {
// 	idStr := ctx.Param("id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		ctx.String(http.StatusBadRequest, "Invalid ID")
// 		return
// 	}

// 	order, err := h.Repository.GetOrder(id)
// 	if err != nil {
// 		ctx.String(http.StatusNotFound, "Order not found")
// 		return
// 	}

// 	h.Repository.AddToCart(order)
// 	ctx.Redirect(http.StatusFound, "/hello")
// }

// // 3. Страница корзины
// func (h *Handler) GetCartPage(ctx *gin.Context) {
// 	cart := h.Repository.GetCart()

// 	if len(cart) == 0 {
// 		ctx.HTML(http.StatusOK, "cart.html", gin.H{
// 			"cart":           cart,
// 			"plantRatio":     0,
// 			"animalRatio":    0,
// 			"hasCalculation": false,
// 		})
// 		return
// 	}

// 	// расчёт
// 	var totalC, totalN int
// 	for _, item := range cart {
// 		totalC += item.C
// 		totalN += item.N
// 	}
// 	avgC := float64(totalC) / float64(len(cart))
// 	avgN := float64(totalN) / float64(len(cart))

// 	plantRatio := 100 / (1 + (avgN / 10))
// 	if plantRatio < 0 {
// 		plantRatio = 0
// 	}
// 	if plantRatio > 100 {
// 		plantRatio = 100
// 	}
// 	animalRatio := 100 - plantRatio

// 	ctx.HTML(http.StatusOK, "cart.html", gin.H{
// 		"cart":           cart,
// 		"plantRatio":     int(plantRatio),
// 		"animalRatio":    int(animalRatio),
// 		"avgC":           avgC,
// 		"avgN":           avgN,
// 		"hasCalculation": true,
// 	})
// }


func (h *Handler) CartHandler(ctx *gin.Context) {
  idStr := ctx.Param("id")
  id, err := strconv.Atoi(idStr)
  if err != nil {
    logrus.Error(err)
  }
  cartorder := h.Repository.GetCartOrder(id)
  ctx.HTML(http.StatusOK, "cart.html", gin.H{
    "cartorder": cartorder,
    "count":           len(cartorder),
  })
}

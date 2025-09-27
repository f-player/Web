package handler

import (

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"lab1-design-backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// router.GET("/", h.GetProducts)
	// router.GET("/order/:id", h.GetProduct)

	router.GET("/group_product", h.GetProducts)
	router.GET("/product/:id", h.GetProduct)
	router.GET("/calculation/:calculation_id", h.GetCalculation)
	router.POST("/calculation/add/product/:product_id", h.AddProductToCalculation)
	router.POST("/calculation/:calculation_id/delete", h.DeleteCalculation)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "C:/Users/Ladmin/Desktop/РИП/lab1-design-backend/resources/styles")
}

// errorHandler для более удобного вывода ошибок 
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
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





// func (h *Handler) CalculationHandler(ctx *gin.Context) {
//   idStr := ctx.Param("id")
//   id, err := strconv.Atoi(idStr)
//   if err != nil {
//     logrus.Error(err)
//   }
//   calculationproduct := h.Repository.GetCalculationProduct(id)
//   ctx.HTML(http.StatusOK, "calculation.html", gin.H{
//     "calculationproduct": calculationproduct,
//     "count":           len(calculationproduct),
//   })
// }
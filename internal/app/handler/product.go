package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"lab1-design-backend/internal/app/ds"
)

func (h *Handler) GetProducts(ctx *gin.Context) {
	var products []ds.Products
	var err error

	searchQuery := ctx.Query("query") // получаем значение из нашего поля
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		products, err = h.Repository.GetProducts()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		products, err = h.Repository.GetProductsByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	draftCalculation, err := h.Repository.GetDraftCalculation(hardcodedUserID)
		var calculationID uint = 0
		var productsCount int = 0

	if err == nil && draftCalculation != nil {
		fullCalculation, err := h.Repository.GetCalculationWithProducts(draftCalculation.ID)
		if err == nil {
			calculationID = fullCalculation.ID
			productsCount = len(fullCalculation.ProductsLink)
		}
	}

	ctx.HTML(http.StatusOK, "group_product.html", gin.H{
		"time":   time.Now().Format("15:04:05"),
		"products": products,
		"query":  searchQuery, // передаем введенный запрос обратно на страницу
		// в ином случае оно будет очищаться при нажатии на кнопку
		"calculationId":    calculationID,
    	"productsCount": productsCount,
	})
}

func (h *Handler) GetProduct(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /order/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	product, err := h.Repository.GetProduct(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "product.html", gin.H{
		"product": product,
	})
}

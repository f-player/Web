package handler

import (
	"lab1-design-backend/internal/app/ds"
	"net/http"
	"strconv"
	"fmt"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)


func CalculateProductRatio(product ds.Products) map[string]int {
    animalRatio := 50 // базовое значение 50%
    
    // Корректировка на основе δ¹⁵N (азот)
    if product.N_pol > 5 {
        animalRatio += 30
    } else if product.N_pol > 0 {
        animalRatio += 15
    } else if product.N_pol < -5 {
        animalRatio -= 20
    }
    
    // Корректировка на основе δ¹³C (углерод)
    if product.C_pol < -25 {
        animalRatio -= 15
    } else if product.C_pol > -15 {
        animalRatio += 10
    }
    
    // Ограничение значений 0-100%
    if animalRatio > 100 {
        animalRatio = 100
    }
    if animalRatio < 0 {
        animalRatio = 0
    }
    
    plantRatio := 100 - animalRatio
    
    return map[string]int{
        "Animal": animalRatio,
        "Plant":  plantRatio,
    }
}


// GET /api/diet_composition/cart - иконка корзины

// GetCartBadge godoc
// @Summary      Получить информацию для иконки корзины (авторизованный пользователь)
// @Description  Возвращает ID черновика текущего пользователя и количество продуктов в нем.
// @Tags         diet_composition
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} ds.CartBadgeDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /diet_composition/cart [get]
func (h *Handler) GetCartBadge(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	draft, err := h.Repository.GetDraftDietComposition(userID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			DietCompositionID: nil,
			Count:  0,
		})
		return
	}

	fullDietComposition, err := h.Repository.GetDietCompositionWithProducts(draft.ID)
	if err != nil {
		logrus.Error("Error getting DietComposition with products:", err)
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			DietCompositionID: nil,
			Count:  0,
		})
		return
	}

	c.JSON(http.StatusOK, ds.CartBadgeDTO{
		DietCompositionID: &fullDietComposition.ID,
		Count:  len(fullDietComposition.ProductsLink),
	})
}


// GET /api/diet_composition - список заявок с фильтрацией
// ListDietComposition godoc
// @Summary      Получить список заявок (авторизованный пользователь)
// @Description  Возвращает отфильтрованный список всех сформированных заявок (кроме черновиков и удаленных).
// @Tags         diet_composition
// @Produce      json
// @Security     ApiKeyAuth
// @Param        status query int false "Фильтр по статусу заявки"
// @Param        from query string false "Фильтр по дате 'от' (формат YYYY-MM-DD)"
// @Param        to query string false "Фильтр по дате 'до' (формат YYYY-MM-DD)"
// @Success      200 {array} ds.DietCompositionDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /diet_composition [get]
func (h *Handler) ListDietComposition(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}
	isModerator := isUserModerator(c)

	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	dietCompositionList, err := h.Repository.DietCompositionListFiltered(userID, isModerator, status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, dietCompositionList)
}



// GET /api/diet_composition/:id - одна заявка с услугами
// GetDietComposition godoc
// @Summary      Получить одну заявку по ID (авторизованный пользователь)
// @Description  Возвращает полную информацию о заявке, включая привязанные продукты.
// @Tags         diet_composition
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Success      200 {object} ds.DietCompositionDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      404 {object} map[string]string "Заявка не найдена"
// @Router       /diet_composition/{id} [get]
func (h *Handler) GetDietComposition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	dietComposition, err := h.Repository.GetDietCompositionWithProducts(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}
	
	var products []ds.ProductInDietCompositionDTO
	for _, link := range dietComposition.ProductsLink {
		ratio := CalculateProductRatio(link.Product)
		products = append(products, ds.ProductInDietCompositionDTO{
			ID:          link.ID,
			ProductID:   link.ProductID,
			Title:       link.Product.Title,
			Image:       link.Product.Image,
			C_pol:       link.Product.C_pol,
			N_pol:       link.Product.N_pol,
			Description: link.Description,
			Ratio:       ratio,
		})
		fmt.Printf("Calculated ratio for product %d: Animal=%d%%, Plant=%d%%\n", 
			link.Product.ID, ratio["Animal"], ratio["Plant"])
	}

	dietCompositionDTO := ds.DietCompositionDTO{
		ID:             dietComposition.ID,
		Status:         dietComposition.Status,
		CreationDate:   dietComposition.CreationDate,
		CreatorID:      dietComposition.Creator.ID,
		ModeratorID:    nil,
		FormingDate:    dietComposition.FormingDate,
		ComplitionDate: dietComposition.ComplitionDate,
		C_pol: dietComposition.C_pol,
		N_pol: dietComposition.N_pol,
		PRP: dietComposition.PRP,
		PGP: dietComposition.PGP,
		Products:       products,
	}

	if dietComposition.ModeratorID != nil {
		dietCompositionDTO.ModeratorID = &dietComposition.Moderator.ID
	}

	c.JSON(http.StatusOK, dietCompositionDTO)
}


// PUT /api/diet_composition/:id - изменение полей заявки
// UpdateDietComposition godoc
// @Summary      Обновить данные заявки (авторизованный пользователь)
// @Description  Позволяет пользователю обновить поля своей заявки .
// @Tags         diet_composition
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        updateData body ds.DietCompositionUpdateRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /diet_composition/{id} [put]
func (h *Handler) UpdateDietComposition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.DietCompositionUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateDietCompositionUserFields(uint(id), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Данные заявки обновлены",
	})
}

// PUT /api/diet_composition/:id/form - сформировать заявку
// FormDietComposition godoc
// @Summary      Сформировать заявку (авторизованный пользователь)
// @Description  Переводит заявку из статуса "черновик" в "сформирована".
// @Tags         diet_composition
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки (черновика)"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Не все поля заполнены"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /diet_composition/{id}/form [put]
func (h *Handler) FormDietComposition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.FormDietComposition(uint(id), userID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка сформирована",
	})
}


// PUT /api/diet_composition/:id/resolve - завершить/отклонить заявку
// ResolveDietComposition godoc
// @Summary      Завершить или отклонить заявку (только модератор)
// @Description  Модератор завершает (с расчетом) или отклоняет заявку.
// @Tags         diet_composition
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        action body ds.DietCompositionResolveRequest true "Действие: 'complete' или 'reject'"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /diet_composition/{id}/resolve [put]
func (h *Handler) ResolveDietComposition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.DietCompositionResolveRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	moderatorID := uint(userID)
	if err := h.Repository.ResolveDietComposition(uint(id), moderatorID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка обработана модератором",
	})
}


// DELETE /api/diet_composition/:id - удаление заявки
// DeleteDietComposition godoc
// @Summary      Удалить заявку (авторизованный пользователь)
// @Description  Логически удаляет заявку, переводя ее в статус "удалена".
// @Tags         diet_composition
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /diet_composition/{id} [delete]
func (h *Handler) DeleteDietComposition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.LogicallyDeleteDietComposition(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка удалена",
	})
}


// DELETE /api/diet_composition/:id/products/:product_id - удаление продукта из заявки
// RemoveProductFromDietComposition godoc
// @Summary      Удалить продукт из заявки (авторизованный пользователь)
// @Description  Удаляет связь между заявкой и продуктом.
// @Tags         m-m
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        product_id path int true "ID продукта"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /diet_composition/{id}/products/{product_id} [delete]
func (h *Handler) RemoveProductFromDietComposition(c *gin.Context) {
	dietCompositionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.RemoveProductFromDietComposition(uint(dietCompositionID), uint(productID)); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Продукт удален из заявки",
	})
}



// PUT /api/diet_composition/:id/products/:product_id - изменение м-м связи
// UpdateMM godoc
// @Summary      Обновить описание продукта в заявке (авторизованный пользователь)
// @Description  Изменяет дополнительное описание для конкретного продукта в рамках одной заявки.
// @Tags         m-m
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        product_id path int true "ID продукта"
// @Param        updateData body ds.ProductToDietCompositionUpdateRequest true "Новое описание"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /diet_composition/{id}/products/{product_id} [put]
func (h *Handler) UpdateMM(c *gin.Context) {
	dietCompositionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.ProductToDietCompositionUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	updateData := ds.ProductToDietComposition{
		Description: req.Description,
	}

	if err := h.Repository.UpdateMM(uint(dietCompositionID), uint(productID), updateData); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Дополнительная информация к продукту обновлена",
	})
}


// GET /diet_composition/:diet_composition_id - HTML страница с визуализацией
func (h *Handler) GetDietCompositionPage(c *gin.Context) {
	dietCompositionID, err := strconv.Atoi(c.Param("diet_composition_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	dietComposition, err := h.Repository.GetDietCompositionWithProducts(uint(dietCompositionID))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	if len(dietComposition.ProductsLink) == 0 {
		h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty diet composition page, add products first"))
		return
	}

	// Рассчитываем соотношение для каждого продукта
	for i := range dietComposition.ProductsLink {
		ratio := CalculateProductRatio(dietComposition.ProductsLink[i].Product)
		dietComposition.ProductsLink[i].Ratio = ratio
		fmt.Printf("Calculated ratio for product %d: Animal=%d%%, Plant=%d%%\n", 
			dietComposition.ProductsLink[i].Product.ID, ratio["Animal"], ratio["Plant"])
	}

	c.HTML(http.StatusOK, "Diet_Composition.html", gin.H{
		"Diet_Composition": dietComposition,
		"productsCount": len(dietComposition.ProductsLink),
		"Diet_Compositionproduct": dietComposition.ProductsLink,
	})
}
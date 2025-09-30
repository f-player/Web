package handler

import (
	"lab1-design-backend/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GET /api/calculation/cart - иконка корзины
func (h *Handler) GetCartBadge(c *gin.Context) {
	draft, err := h.Repository.GetDraftCalculation(hardcodedUserID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			CalculationID: nil,
			Count:  0,
		})
		return
	}

	fullCalculation, err := h.Repository.GetCalculationWithProducts(draft.ID)
	if err != nil {
		logrus.Error("Error getting calculation with products:", err)
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			CalculationID: nil,
			Count:  0,
		})
		return
	}

	c.JSON(http.StatusOK, ds.CartBadgeDTO{
		CalculationID: &fullCalculation.ID,
		Count:  len(fullCalculation.ProductsLink),
	})
}

// GET /api/calculation - список заявок с фильтрацией
func (h *Handler) ListCalculation(c *gin.Context) {
	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	calculationList, err := h.Repository.CalculationListFiltered(status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, calculationList)
}

// GET /api/calculation/:id - одна заявка с услугами
func (h *Handler) GetCalculation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	calculation, err := h.Repository.GetCalculationWithProducts(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	var products []ds.ProductInCalculationDTO
	for _, link := range calculation.ProductsLink {
		products = append(products, ds.ProductInCalculationDTO{
			ID:          link.ID,
			ProductID:    link.ProductID,
			Title:       link.Product.Title,
			Image:       link.Product.Image,
			C_pol:       link.Product.C_pol,
			N_pol:       link.Product.N_pol,
			Description: link.Description,
		})
	}

	calculationDTO := ds.CalculationDTO{
		ID:             calculation.ID,
		Status:         calculation.Status,
		CreationDate:   calculation.CreationDate,
		CreatorID:      calculation.Creator.ID,
		ModeratorID:    nil,
		FormingDate:    calculation.FormingDate,
		ComplitionDate: calculation.ComplitionDate,
		C_pol: calculation.C_pol,
		N_pol: calculation.N_pol,
		PRP: calculation.PRP,
		PGP: calculation.PGP,
		
		Products:        products,
	}

	if calculation.ModeratorID != nil {
		calculationDTO.ModeratorID = &calculation.Moderator.ID
	}

	c.JSON(http.StatusOK, calculationDTO)
}

// PUT /api/calculation/:id - изменение полей заявки
func (h *Handler) UpdateCalculation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.CalculationUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateCalculationUserFields(uint(id), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PUT /api/calculation/:id/form - сформировать заявку
func (h *Handler) FormCalculation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.FormCalculation(uint(id), hardcodedUserID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PUT /api/calculation/:id/resolve - завершить/отклонить заявку
func (h *Handler) ResolveCalculation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.CalculationResolveRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	moderatorID := uint(hardcodedUserID)
	if err := h.Repository.ResolveCalculation(uint(id), moderatorID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// DELETE /api/calculation/:id - удаление заявки
func (h *Handler) DeleteCalculation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.LogicallyDeleteCalculation(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// DELETE /api/calculation/:id/products/:product_id - удаление фактора из заявки
func (h *Handler) RemoveProductFromCalculation(c *gin.Context) {
	calculationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.RemoveProductFromCalculation(uint(calculationID), uint(productID)); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PUT /api/calculation/:id/products/:product_id - изменение м-м связи
func (h *Handler) UpdateMM(c *gin.Context) {
	calculationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.ProductToCalculationUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	updateData := ds.ProductToCalculation{
		Description: req.Description,
	}

	if err := h.Repository.UpdateMM(uint(calculationID), uint(productID), updateData); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusNoContent)
}
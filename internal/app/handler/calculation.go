package handler

import (
	"lab1-design-backend/internal/app/ds"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const hardcodedUserID = 1

func (h *Handler) AddProductToCalculation(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	calculation, err := h.Repository.GetDraftCalculation(hardcodedUserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newCalculation := ds.CalculationSearching{
			CreatorID: hardcodedUserID,
			Status:    ds.StatusDraft,
		}
		if createErr := h.Repository.CreateCalculation(&newCalculation); createErr != nil {
			h.errorHandler(c, http.StatusInternalServerError, createErr)
			return
		}
		calculation = &newCalculation
	} else if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	if err = h.Repository.AddProductToCalculation(calculation.ID, uint(productID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err) // Добавьте обработку ошибки
        return
	}

	c.Redirect(http.StatusFound, "/calculation/"+strconv.Itoa(int(calculation.ID)))
}

func (h *Handler) GetCalculation(c *gin.Context) {
	calculationID, err := strconv.Atoi(c.Param("calculation_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	calculation, err := h.Repository.GetCalculationWithProducts(uint(calculationID))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	if len(calculation.ProductsLink) == 0 {
		h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty calculation page, add products first"))
		return
	}

	c.HTML(http.StatusOK, "calculation.html", gin.H{
        "calculation": calculation,
        "productsCount": len(calculation.ProductsLink),
        "calculationproduct": calculation.ProductsLink,
    })
}

func (h *Handler) DeleteCalculation(c *gin.Context) {
	calculationID, _ := strconv.Atoi(c.Param("calculation_id"))

	if err := h.Repository.LogicallyDeleteCalculation(uint(calculationID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/group_product")
}

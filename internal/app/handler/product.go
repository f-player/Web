package handler

import (
	"lab1-design-backend/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/products - список продуктов с фильтрацией
// GetProducts godoc
// @Summary      Получить список продуктов (все)
// @Description  Возвращает постраничный список продуктов.
// @Tags         products
// @Produce      json
// @Param        title query string false "Фильтр по названию продукта"
// @Success      200 {object} ds.PaginatedResponse
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /products [get]
func (h *Handler) GetProducts(c *gin.Context) {
	title := c.Query("title")

	products, total, err := h.Repository.ProductsList(title)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	var productDTOs []ds.ProductDTO
	for _, f := range products {
		productDTOs = append(productDTOs, ds.ProductDTO{
			ID:    f.ID,
			Title: f.Title,
			Image: f.Image,
			C_pol: f.C_pol,
			N_pol: f.N_pol,
		})
	}

	c.JSON(http.StatusOK, ds.PaginatedResponse{
		Items: productDTOs,
		Total: total,
	})
}


// GET /api/products/:id - один продукт
// GetProduct godoc
// @Summary      Получить один продукт по ID (все)
// @Description  Возвращает детальную информацию о продукте.
// @Tags         products
// @Produce      json
// @Param        id path int true "ID продукта"
// @Success      200 {object} ds.ProductDTO
// @Failure      404 {object} map[string]string "Продукт не найден"
// @Router       /products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	product, err := h.Repository.GetProductByID(id)
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	productDTO := ds.ProductDTO{
		ID:    product.ID,
		Title: product.Title,
		Image: product.Image,
		C_pol: product.C_pol,
		N_pol: product.N_pol,
	}

	c.JSON(http.StatusOK, productDTO)
}


// POST /api/products - создание продукта
// CreateProduct godoc
// @Summary      Создать новый продукт (только модератор)
// @Description  Создает новую запись о продукте.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        productData body ds.ProductCreateRequest true "Данные нового продукта"
// @Success      201 {object} ds.ProductDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен (не модератор)"
// @Router       /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var req ds.ProductCreateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}



	product := ds.Products{
		Title: req.Title,
		C_pol: req.C_pol,
		N_pol: req.N_pol,
	}

	if err := h.Repository.CreateProduct(&product); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	productDTO := ds.ProductDTO{
		ID:    product.ID,
		Title: product.Title,
		Image: product.Image,
		C_pol: product.C_pol,
		N_pol: product.N_pol,
	}

	c.JSON(http.StatusCreated, productDTO)
}



// PUT /api/products/:id - обновление продукта
// UpdateProduct godoc
// @Summary      Обновить продукт (только модератор)
// @Description  Обновляет информацию о существующем продукте.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID продукта"
// @Param        updateData body ds.ProductUpdateRequest true "Данные для обновления"
// @Success      200 {object} ds.ProductDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.ProductUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	product, err := h.Repository.UpdateProduct(uint(id), req)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	productDTO := ds.ProductDTO{
		ID:    product.ID,
		Title: product.Title,
		Image: product.Image,
		C_pol: product.C_pol,
		N_pol: product.N_pol,
	}

	c.JSON(http.StatusOK, productDTO)
}


// DELETE /api/products/:id - удаление продукта
// DeleteProduct godoc
// @Summary      Удалить продукт (только модератор)
// @Description  Удаляет продукт из системы.
// @Tags         products
// @Security     ApiKeyAuth
// @Param        id path int true "ID продукта для удаления"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteProduct(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Продукт удален",
	})
}


// POST /api/diet_composition/draft/products/:product_id - добавление продукта в черновик
// AddProductToDraft godoc
// @Summary      Добавить продукт в черновик заявки (все)
// @Description  Находит или создает черновик заявки для текущего пользователя и добавляет в него продукт.
// @Tags         products
// @Security     ApiKeyAuth
// @Param        product_id path int true "ID продукта для добавления"
// @Success      201 {object} map[string]string "Сообщение об успехе"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /diet_composition/draft/products/{product_id} [post]
func (h *Handler) AddProductToDraft(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.AddProductToDraft(userID, uint(productID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Черновик создан. Продукт добавлен в черновик.",
	})
}


// POST /api/products/:id/image - загрузка изображения продукта
// UploadFactorImage godoc
// @Summary      Загрузить изображение для продукта (только модератор)
// @Description  Загружает и привязывает изображение к продукту риска.
// @Tags         products
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID продукта"
// @Param        file formData file true "Файл изображения"
// @Success      200 {object} map[string]string "URL загруженного изображения"
// @Failure      400 {object} map[string]string "Файл не предоставлен"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /products/{id}/image [post]
func (h *Handler) UploadProductImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	imageURL, err := h.Repository.UploadProductImage(uint(id), file)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": imageURL})
}

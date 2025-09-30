package handler

import (
	"lab1-design-backend/internal/app/ds"
	"net/http"
	"strconv"
	"log"
	"github.com/gin-gonic/gin"
)

// GET /api/products - список факторов с фильтрацией
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
			ID:       f.ID,
			Title:    f.Title,
			Image:    f.Image,
			C_pol: f.C_pol,
			N_pol: f.N_pol,
			
		})
	}

	c.JSON(http.StatusOK, ds.PaginatedResponse{
		Items: productDTOs,
		Total: total,
	})
}

// GET /api/products/:id - один фактор
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
		ID:       product.ID,
		Title:    product.Title,
		Image:    product.Image,
		C_pol: product.C_pol,
		N_pol: product.N_pol,
	}

	c.JSON(http.StatusOK, productDTO)
}

// POST /api/products - создание фактора
func (h *Handler) CreateProduct(c *gin.Context) {
	var req ds.ProductCreateRequest
	if err := c.BindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	log.Printf("Creating product: Title=%s, C_pol=%d, N_pol=%d", req.Title, req.C_pol, req.N_pol)

	// statusValue := false

	product := ds.Products{
		Title:    req.Title,
		C_pol: req.C_pol,
		N_pol: req.N_pol,
	}

	if err := h.Repository.CreateProduct(&product); err != nil {
		log.Printf("Error creating product in repository: %v", err)
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	productDTO := ds.ProductDTO{
		ID:       product.ID,
		Title:    product.Title,
		Image:    product.Image,
		C_pol: product.C_pol,
		N_pol: product.N_pol,
	}

	log.Printf("Product created successfully: ID=%d", product.ID)
	c.JSON(http.StatusCreated, productDTO)
}

// PUT /api/products/:id - обновление фактора
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
		ID:       product.ID,
		Title:    product.Title,
		Image:    product.Image,
		C_pol: product.C_pol,
		N_pol: product.N_pol,
	}

	c.JSON(http.StatusOK, productDTO)
}

// DELETE /api/products/:id - удаление фактора
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

	c.Status(http.StatusNoContent)
}

// POST /api/calculation/draft/products/:product_id - добавление фактора в черновик
func (h *Handler) AddProductToDraft(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.AddProductToDraft(hardcodedUserID, uint(productID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// POST /api/products/:id/image - загрузка изображения фактора
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
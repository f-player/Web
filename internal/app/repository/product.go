package repository

import (
	"context"
	"errors"
	"fmt"
	"lab1-design-backend/internal/app/ds"
	"log"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GET /api/products - список продуктов с фильтрацией
func (r *Repository) ProductsList(title string) ([]ds.Products, int64, error) {
	var products []ds.Products
	var total int64

	query := r.db.Model(&ds.Products{})
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	productsQuery := query.Order("id asc")
	if err := productsQuery.Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GET /api/products/:id - один продукт
func (r *Repository) GetProductByID(id int) (*ds.Products, error) {
	var product ds.Products
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// POST /api/products - создание продукта
func (r *Repository) CreateProduct(product *ds.Products) error {
	return r.db.Create(product).Error
}

// PUT /api/products/:id - обновление продукта
func (r *Repository) UpdateProduct(id uint, req ds.ProductUpdateRequest) (*ds.Products, error) {
	var product ds.Products
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}

	if req.Title != nil {
		product.Title = *req.Title
	}
	if req.C_pol != nil {
		product.C_pol = *req.C_pol
	}
	if req.N_pol != nil {
		product.N_pol = *req.N_pol
	}

	if err := r.db.Save(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// DELETE /api/products/:id - удаление продукта
func (r *Repository) DeleteProduct(id uint) error {
	var product ds.Products
	var imageURLToDelete string

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&product, id).Error; err != nil {
			return err
		}
		if product.Image != nil {
			imageURLToDelete = *product.Image
		}
		if err := tx.Delete(&ds.Products{}, id).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	if imageURLToDelete != "" {
		parsedURL, err := url.Parse(imageURLToDelete)
		if err != nil {
			log.Printf("ERROR: could not parse image URL for deletion: %v", err)
			return nil
		}

		objectName := strings.TrimPrefix(parsedURL.Path, fmt.Sprintf("/%s/", r.bucketName))

		err = r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("ERROR: failed to delete object '%s' from MinIO: %v", objectName, err)
		}
	}

	return nil
}

// POST /api/diet_composition/draft/products/:product_id - добавление продукта в черновик
func (r *Repository) AddProductToDraft(userID, productID uint) error {
	log.Printf("DEBUG: AddProductToDraft called with userID=%d, productID=%d", userID, productID)

	return r.db.Transaction(func(tx *gorm.DB) error {
		var dietComposition ds.DietCompositionSearching
		err := tx.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&dietComposition).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("DEBUG: No draft diet composition found for user %d, creating new one", userID)

				newDietComposition := ds.DietCompositionSearching{
					CreatorID:    userID,
					Status:       ds.StatusDraft,
					CreationDate: time.Now(),
					C_pol:        0, // Добавляем значения по умолчанию
					N_pol:        0,
				}
				if err := tx.Create(&newDietComposition).Error; err != nil {
					log.Printf("ERROR: Failed to create draft diet composition: %v", err)
					return fmt.Errorf("failed to create draft diet composition: %w", err)
				}
				dietComposition = newDietComposition
				log.Printf("DEBUG: Created new diet composition with ID=%d", dietComposition.ID)
			} else {
				log.Printf("ERROR: Database error when searching for draft diet composition: %v", err)
				return err
			}
		} else {
			log.Printf("DEBUG: Found existing diet composition with ID=%d", dietComposition.ID)
		}

		// Проверяем, существует ли уже такой продукт в диет-композиции
		var count int64
		if err := tx.Model(&ds.ProductToDietComposition{}).Where("diet_composition_id = ? AND product_id = ?", dietComposition.ID, productID).Count(&count).Error; err != nil {
			log.Printf("ERROR: Failed to check if product exists in diet composition: %v", err)
			return err
		}

		if count > 0 {
			log.Printf("DEBUG: Product %d already exists in diet composition %d", productID, dietComposition.ID)
			return errors.New("product already in diet composition")
		}

		// Проверяем, существует ли продукт
		var product ds.Products
		if err := tx.First(&product, productID).Error; err != nil {
			log.Printf("ERROR: Product with ID %d not found: %v", productID, err)
			return fmt.Errorf("product not found: %w", err)
		}

		link := ds.ProductToDietComposition{
			DietCompositionID: dietComposition.ID,
			ProductID:         productID,
		}

		if err := tx.Create(&link).Error; err != nil {
			log.Printf("ERROR: Failed to add product to diet composition: %v", err)
			return fmt.Errorf("failed to add product to diet composition: %w", err)
		}

		log.Printf("DEBUG: Successfully added product %d to diet composition %d", productID, dietComposition.ID)
		return nil
	})
}

// POST /api/products/:id/image - загрузка изображения продукта
func (r *Repository) UploadProductImage(productID uint, fileHeader *multipart.FileHeader) (string, error) {
	var finalImageURL string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var product ds.Products
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productID).Error; err != nil {
			return fmt.Errorf("product with id %d not found: %w", productID, err)
		}

		const imagePathPrefix = "Images/"

		if product.Image != nil && *product.Image != "" {
			oldImageURL, err := url.Parse(*product.Image)
			if err == nil {
				oldObjectName := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.bucketName))
				r.minioClient.RemoveObject(context.Background(), r.bucketName, oldObjectName, minio.RemoveObjectOptions{})
			}
		}

		fileName := filepath.Base(fileHeader.Filename)
		objectName := imagePathPrefix + fileName

		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = r.minioClient.PutObject(context.Background(), r.bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		})

		if err != nil {
			return fmt.Errorf("failed to upload to minio: %w", err)
		}

		imageURL := fmt.Sprintf("http://%s/%s/%s", r.minioEndpoint, r.bucketName, objectName)

		if err := tx.Model(&product).Update("image", imageURL).Error; err != nil {
			return fmt.Errorf("failed to update product image url in db: %w", err)
		}

		finalImageURL = imageURL
		return nil
	})
	if err != nil {
		return "", err
	}
	return finalImageURL, nil
}

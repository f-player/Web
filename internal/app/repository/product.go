package repository

import (
	"lab1-design-backend/internal/app/ds"
	"context"
	"errors"
	"fmt"
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

// GET /api/products - список факторов с фильтрацией
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

// GET /api/products/:id - один фактор
func (r *Repository) GetProductByID(id int) (*ds.Products, error) {
	var product ds.Products
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// POST /api/products - создание фактора
func (r *Repository) CreateProduct(product *ds.Products) error {
	return r.db.Create(product).Error
}

// PUT /api/products/:id - обновление фактора
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

// DELETE /api/products/:id - удаление фактора
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

// POST /api/calculation/draft/products/:product_id - добавление фактора в черновик
func (r *Repository) AddProductToDraft(userID, productID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var calculation ds.CalculationSearching
		err := tx.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&calculation).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newCalculation := ds.CalculationSearching{
					CreatorID:    userID,
					Status:       ds.StatusDraft,
					CreationDate: time.Now(),
				}
				if err := tx.Create(&newCalculation).Error; err != nil {
					return fmt.Errorf("failed to create draft calculation: %w", err)
				}
				calculation = newCalculation
			} else {
				return err
			}
		}

		var count int64
		tx.Model(&ds.ProductToCalculation{}).Where("calculation_id = ? AND product_id = ?", calculation.ID, productID).Count(&count)
		if count > 0 {
			return errors.New("product already in calculation")
		}

		link := ds.ProductToCalculation{
			CalculationID:   calculation.ID,
			ProductID: productID,
		}

		if err := tx.Create(&link).Error; err != nil {
			return fmt.Errorf("failed to add product to calculation: %w", err)
		}

		return nil
	})
}

// POST /api/products/:id/image - загрузка изображения фактора
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
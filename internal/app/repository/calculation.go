package repository

import (
	"errors"
	"lab1-design-backend/internal/app/ds"
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// GET /api/diet_composition/cart - иконка корзины
func (r *Repository) GetDraftDietComposition(userID uint) (*ds.DietCompositionSearching, error) {
	var dietComposition ds.DietCompositionSearching
	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&dietComposition).Error
	if err != nil {
		return nil, err
	}
	return &dietComposition, nil
}


// GET /api/diet_composition/cart - иконка корзины
// GET /api/diet_composition/:id - одна заявка с услугами
func (r *Repository) GetDietCompositionWithProducts(dietCompositionID uint) (*ds.DietCompositionSearching, error) {
	var dietComposition ds.DietCompositionSearching
	err := r.db.Preload("ProductsLink.Product").Preload("Creator").Preload("Moderator").First(&dietComposition, dietCompositionID).Error
	if err != nil {
		return nil, err
	}

	if dietComposition.Status == ds.StatusDeleted {
		return nil, errors.New("diet composition page not found or has been deleted")
	}

	return &dietComposition, nil
}


// GET /api/diet_composition - список заявок с фильтрацией
func (r *Repository) DietCompositionListFiltered(userID uint, isModerator bool, status, from, to string) ([]ds.DietCompositionDTO, error) {
	var dietCompositionList []ds.DietCompositionSearching
	query := r.db.Preload("Creator").Preload("Moderator")

	query = query.Where("status != ? AND status != ?", ds.StatusDeleted, ds.StatusDraft)

	if !isModerator {
		query = query.Where("creator_id = ?", userID)
	}

	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	if from != "" {
		if fromTime, err := time.Parse("2006-01-02", from); err == nil {
			query = query.Where("forming_date >= ?", fromTime)
		}
	}

	if to != "" {
		if toTime, err := time.Parse("2006-01-02", to); err == nil {
			query = query.Where("forming_date <= ?", toTime)
		}
	}

	if err := query.Find(&dietCompositionList).Error; err != nil {
		return nil, err
	}

	var result []ds.DietCompositionDTO
	for _, dietComposition := range dietCompositionList {
		dto := ds.DietCompositionDTO{
			ID:             dietComposition.ID,
			Status:         dietComposition.Status,
			CreationDate:   dietComposition.CreationDate,
			CreatorID:      dietComposition.Creator.ID,
			ModeratorID:    nil,
			FormingDate:    dietComposition.FormingDate,
			ComplitionDate: dietComposition.ComplitionDate,
			C_pol:          dietComposition.C_pol,
			N_pol:          dietComposition.N_pol,
			PRP:            dietComposition.PRP,
			PGP:            dietComposition.PGP,
		}

		if dietComposition.ModeratorID != nil {
			dto.ModeratorID = &dietComposition.Moderator.ID
		}
		result = append(result, dto)
	}
	return result, nil
}



// PUT /api/diet_composition/:id - изменение полей заявки
func (r *Repository) UpdateDietCompositionUserFields(id uint, req ds.DietCompositionUpdateRequest) error {
	updates := make(map[string]interface{})

	updates["c_pol"] = req.C_pol
	updates["n_pol"] = req.N_pol
	if req.PGP != nil {
		updates["pgp"] = *req.PGP
	}
	if req.PRP != nil {
		updates["prp"] = *req.PRP
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&ds.DietCompositionSearching{}).Where("id = ?", id).Updates(updates).Error
}


// PUT /api/diet_composition/:id/form - сформировать заявку
func (r *Repository) FormDietComposition(id uint, creatorID uint) error {
	var dietComposition ds.DietCompositionSearching
	if err := r.db.First(&dietComposition, id).Error; err != nil {
		return err
	}

	if dietComposition.CreatorID != creatorID {
		return errors.New("only creator can form diet composition")
	}

	if dietComposition.Status != ds.StatusDraft {
		return errors.New("only draft diet composition can be formed")
	}

	if dietComposition.PGP == nil || dietComposition.PRP == nil {
		return errors.New("c_pol, n_pol, pgp and prp are required")
	}

	now := time.Now()
	return r.db.Model(&dietComposition).Updates(map[string]interface{}{
		"status":       ds.StatusFormed,
		"forming_date": now,
	}).Error
}



// PUT /api/diet_composition/:id/resolve - завершить/отклонить заявку
func (r *Repository) ResolveDietComposition(id uint, moderatorID uint, action string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		var dietComposition ds.DietCompositionSearching
		if err := tx.Preload("ProductsLink.Product").First(&dietComposition, id).Error; err != nil {
			return err
		}

		if dietComposition.Status != ds.StatusFormed {
			return errors.New("only formed diet composition can be resolved")
		}

		now := time.Now()
		updates := map[string]interface{}{
			"moderator_id":    moderatorID,
			"complition_date": now,
		}

		switch action {
		case "complete":
			{
				updates["status"] = ds.StatusCompleted
				pof, php := r.calculateDIETCOMPOSITION(dietComposition)
				updates["PGP"] = pof
				updates["PRP"] = php
			}
		case "reject":
			{
				updates["status"] = ds.StatusRejected
			}
		default:
			{
				return errors.New("invalid action, must be 'complete' or 'reject'")
			}
		}

		if err := tx.Model(&dietComposition).Updates(updates).Error; err != nil {
			return err
		}


		return nil
	})
}


// Функция расчета процента растительной и животной пищи
// Функция расчета процента растительной и животной пищи
func (r *Repository) calculateDIETCOMPOSITION(dietComposition ds.DietCompositionSearching) (float64, float64) {
	if len(dietComposition.ProductsLink) == 0 {
		return 0, 0 // Нет продуктов - нет расчета
	}

	// Значения C и N индивида
	individualC := float64(dietComposition.C_pol)
	individualN := float64(dietComposition.N_pol)

	var totalPlantSimilarity float64
	var totalAnimalSimilarity float64

	for _, item := range dietComposition.ProductsLink {
		productC := float64(item.Product.C_pol)
		productN := float64(item.Product.N_pol)

		// Рассчитываем "расстояние" между индивидом и продуктом
		// Чем меньше расстояние - тем больше сходство
		distance := math.Sqrt(math.Pow(individualC-productC, 2) + math.Pow(individualN-productN, 2))
		similarity := 1 / (1 + distance) // Преобразуем расстояние в сходство (0-1)

		// Определяем тип пищи на основе характеристик продукта
		// Растительная пища: более отрицательный C_pol (например, -25)
		// Животная пища: менее отрицательный C_pol (например, -15)
		if productC < -20 { // Растительная пища (более отрицательный C)
			totalPlantSimilarity += similarity
		} else { // Животная пища (менее отрицательный C)
			totalAnimalSimilarity += similarity
		}
	}

	// Рассчитываем проценты
	totalSimilarity := totalPlantSimilarity + totalAnimalSimilarity
	if totalSimilarity == 0 {
		return 0, 0
	}

	plantRatio := (totalPlantSimilarity / totalSimilarity) * 100
	animalRatio := (totalAnimalSimilarity / totalSimilarity) * 100

	// Ограничиваем значения 0-100
	plantRatio = math.Max(0, math.Min(100, plantRatio))
	animalRatio = math.Max(0, math.Min(100, animalRatio))

	return plantRatio, animalRatio
}

// Вспомогательная функция для определения растительной пищи
func isPlantFood(product ds.Products) bool {
	// Логика определения растительной пищи
	// Например: высокий C и низкий N
	return product.C_pol > 5 && product.N_pol < 3
}

// Вспомогательная функция для определения животной пищи
func isAnimalFood(product ds.Products) bool {
	// Логика определения животной пищи
	// Например: низкий C и высокий N
	return product.C_pol < 3 && product.N_pol > 5
}

// DELETE /api/diet_composition/:id - удаление заявки
func (r *Repository) LogicallyDeleteDietComposition(dietCompositionID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var dietComposition ds.DietCompositionSearching

		if err := tx.Preload("ProductsLink").First(&dietComposition, dietCompositionID).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":       ds.StatusDeleted,
			"forming_date": time.Now(),
		}

		if err := tx.Model(&ds.DietCompositionSearching{}).Where("id = ?", dietCompositionID).Updates(updates).Error; err != nil {
			return err
		}

		
		return nil
	})
}


// DELETE /api/diet_composition/:id/products/:product_id - удаление продукта из заявки
func (r *Repository) RemoveProductFromDietComposition(dietCompositionID, productID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		result := tx.Where("diet_composition_id = ? AND product_id = ?", dietCompositionID, productID).Delete(&ds.ProductToDietComposition{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("product not found in this diet composition")
		}
		

		var remainingCount int64
		if err := tx.Model(&ds.ProductToDietComposition{}).Where("diet_composition_id = ?", dietCompositionID).Count(&remainingCount).Error; err != nil {
			return err
		}

		if remainingCount == 0 {
			updates := map[string]interface{}{
				"status":       ds.StatusDeleted,
				"forming_date": time.Now(),
			}
			if err := tx.Model(&ds.DietCompositionSearching{}).Where("id = ?", dietCompositionID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}


// PUT /api/diet_composition/:id/products/:product_id - изменение м-м связи
func (r *Repository) UpdateMM(dietCompositionID, productID uint, updateData ds.ProductToDietComposition) error {
	var link ds.ProductToDietComposition
	if err := r.db.Where("diet_composition_id = ? AND product_id = ?", dietCompositionID, productID).First(&link).Error; err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if updateData.Description != nil {
		updates["description"] = *updateData.Description
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&link).Updates(updates).Error
}


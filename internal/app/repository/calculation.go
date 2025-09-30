package repository

import (
	"lab1-design-backend/internal/app/ds"
	"errors"
	"strconv"
	"time"
	"math"
	"gorm.io/gorm"
)

// GET /api/calculation/cart - иконка корзины
func (r *Repository) GetDraftCalculation(userID uint) (*ds.CalculationSearching, error) {
	var calculation ds.CalculationSearching
	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&calculation).Error
	if err != nil {
		return nil, err
	}
	return &calculation, nil
}

// GET /api/calculation/cart - иконка корзины
// GET /api/calculation/:id - одна заявка с услугами
func (r *Repository) GetCalculationWithProducts(calculationID uint) (*ds.CalculationSearching, error) {
	var calculation ds.CalculationSearching
	err := r.db.Preload("ProductsLink.Product").Preload("Creator").Preload("Moderator").First(&calculation, calculationID).Error
	if err != nil {
		return nil, err
	}

	if calculation.Status == ds.StatusDeleted {
		return nil, errors.New("frax page not found or has been deleted")
	}

	return &calculation, nil
}

// GET /api/calculation - список заявок с фильтрацией
func (r *Repository) CalculationListFiltered(status, from, to string) ([]ds.CalculationDTO, error) {
	var calculationList []ds.CalculationSearching
	query := r.db.Preload("Creator").Preload("Moderator")

	query = query.Where("status != ? AND status != ?", ds.StatusDeleted, ds.StatusDraft)

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

	if err := query.Find(&calculationList).Error; err != nil {
		return nil, err
	}

	var result []ds.CalculationDTO
	for _, calculation := range calculationList {
		dto := ds.CalculationDTO{
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
			
		}

		if calculation.ModeratorID != nil {
			dto.ModeratorID = &calculation.Moderator.ID
		}
		result = append(result, dto)
	}
	return result, nil
}

// PUT /api/calculation/:id - изменение полей заявки
func (r *Repository) UpdateCalculationUserFields(id uint, req ds.CalculationUpdateRequest) error {
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

	return r.db.Model(&ds.CalculationSearching{}).Where("id = ?", id).Updates(updates).Error
}

// PUT /api/calculation/:id/form - сформировать заявку
func (r *Repository) FormCalculation(id uint, creatorID uint) error {
	var calculation ds.CalculationSearching
	if err := r.db.First(&calculation, id).Error; err != nil {
		return err
	}

	if calculation.CreatorID != creatorID {
		return errors.New("only creator can form frax")
	}

	if calculation.Status != ds.StatusDraft {
		return errors.New("only draft frax can be formed")
	}

	if calculation.PGP == nil || calculation.PRP == nil {
		return errors.New("c_pol, n_pol, pgp and prp are required")
	}

	now := time.Now()
	return r.db.Model(&calculation).Updates(map[string]interface{}{
		"status":       ds.StatusFormed,
		"forming_date": now,
	}).Error
}

// PUT /api/calculation/:id/resolve - завершить/отклонить заявку
func (r *Repository) ResolveCalculation(id uint, moderatorID uint, action string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		var calculation ds.CalculationSearching
		if err := tx.Preload("ProductsLink.Product").First(&calculation, id).Error; err != nil {
			return err
		}

		if calculation.Status != ds.StatusFormed {
			return errors.New("only formed calculation can be resolved")
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
				pof, php := r.calculateCALCULATION(calculation)
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

		if err := tx.Model(&calculation).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
}







// Функция расчета процента растительной и животной пищи
func (r *Repository) calculateCALCULATION(calculation ds.CalculationSearching) (float64, float64) {
    if len(calculation.ProductsLink) == 0 {
        return 0, 0 // Нет продуктов - нет расчета
    }

    // Значения C и N индивида
    individualC := float64(calculation.C_pol)
    individualN := float64(calculation.N_pol)

    var totalPlantSimilarity float64
    var totalAnimalSimilarity float64
    var plantCount, animalCount int

    for _, item := range calculation.ProductsLink {
        productC := float64(item.Product.C_pol)
        productN := float64(item.Product.N_pol)

        // Рассчитываем "расстояние" между индивидом и продуктом
        // Чем меньше расстояние - тем больше сходство
        distance := math.Sqrt(math.Pow(individualC-productC, 2) + math.Pow(individualN-productN, 2))

        // Определяем тип пищи на основе характеристик продукта
        // (предположим, что высокий C и низкий N - растительная пища,
        // а низкий C и высокий N - животная)
        similarity := 1 / (1 + distance) // Преобразуем расстояние в сходство (0-1)

        if isPlantFood(item.Product) {
            totalPlantSimilarity += similarity
            plantCount++
        } else if isAnimalFood(item.Product) {
            totalAnimalSimilarity += similarity
            animalCount++
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








// DELETE /api/calculation/:id - удаление заявки
func (r *Repository) LogicallyDeleteCalculation(calculationID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var calculation ds.CalculationSearching

		if err := tx.Preload("ProductsLink").First(&calculation, calculationID).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":       ds.StatusDeleted,
			"forming_date": time.Now(),
		}

		if err := tx.Model(&ds.CalculationSearching{}).Where("id = ?", calculationID).Updates(updates).Error; err != nil {
			return err
		}


		return nil
	})
}

// DELETE /api/calculation/:id/products/:product_id - удаление фактора из заявки
func (r *Repository) RemoveProductFromCalculation(calculationID, productID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		result := tx.Where("calculation_id = ? AND product_id = ?", calculationID, productID).Delete(&ds.ProductToCalculation{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("product not found in this calculation")
		}


		var remainingCount int64
		if err := tx.Model(&ds.ProductToCalculation{}).Where("calculation_id = ?", calculationID).Count(&remainingCount).Error; err != nil {
			return err
		}

		if remainingCount == 0 {
			updates := map[string]interface{}{
				"status":       ds.StatusDeleted,
				"forming_date": time.Now(),
			}
			if err := tx.Model(&ds.CalculationSearching{}).Where("id = ?", calculationID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PUT /api/calculation/:id/products/:product_id - изменение м-м связи
func (r *Repository) UpdateMM(calculationID, productID uint, updateData ds.ProductToCalculation) error {
	var link ds.ProductToCalculation
	if err := r.db.Where("calculation_id = ? AND product_id = ?", calculationID, productID).First(&link).Error; err != nil {
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
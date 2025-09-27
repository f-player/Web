package repository

import (
	"lab1-design-backend/internal/app/ds"
	"errors"
)

func (r *Repository) GetDraftCalculation(userID uint) (*ds.CalculationSearching, error) {
	var calculation ds.CalculationSearching

	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&calculation).Error
	if err != nil {
		return nil, err
	}
	return &calculation, nil
}

func (r *Repository) CreateCalculation(calculation *ds.CalculationSearching) error {
	return r.db.Create(calculation).Error
}

func (r *Repository) AddProductToCalculation(calculationID, productID uint) error {
	var count int64

	r.db.Model(&ds.ProductToCalculation{}).Where("calculation_id = ? AND product_id = ?", calculationID, productID).Count(&count)
	if count > 0 {
		return errors.New("factor already in frax")
	}

	link := ds.ProductToCalculation{
		CalculationID:   calculationID,
		ProductID: productID,
	}
	return r.db.Create(&link).Error
}

func (r *Repository) GetCalculationWithProducts(calculationID uint) (*ds.CalculationSearching, error) {
	var calculation ds.CalculationSearching

	err := r.db.Preload("ProductsLink.Product").First(&calculation, calculationID).Error
	if err != nil {
		return nil, err
	}

	if calculation.Status == ds.StatusDeleted {
		return nil, errors.New("calculation page not found or has been deleted")
	}

	return &calculation, nil
}

func (r *Repository) LogicallyDeleteCalculation(calculationID uint) error {
	result := r.db.Exec("UPDATE calculation_searchings SET status = ? WHERE id = ?", ds.StatusDeleted, calculationID)
	return result.Error
}

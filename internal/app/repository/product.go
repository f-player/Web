package repository

import (
	"fmt"

	"lab1-design-backend/internal/app/ds"
)

func (r *Repository) GetProducts() ([]ds.Products, error) {
	var products []ds.Products
	err := r.db.Find(&products).Error
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return products, nil
}

func (r *Repository) GetProduct(id int) (ds.Products, error) {
	product := ds.Products{}
	err := r.db.Where("id = ?", id).First(&product).Error
	if err != nil {
		return ds.Products{}, err
	}
	return product, nil
}

func (r *Repository) GetProductsByTitle(title string) ([]ds.Products, error) {
	var products []ds.Products
	err := r.db.Where("Title ILIKE ?", "%"+title+"%").Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

package ds

// ProductToCalculation соответствует таблице "ProductToCalculation"
type ProductToCalculation struct {
	ID          uint    `gorm:"primaryKey;column:id"`
	CalculationID      uint    `gorm:"column:calculation_id;not null"`   // Внешний ключ к CalculationSearching
	ProductID    uint    `gorm:"column:product_id;not null"` // Внешний ключ к Products
	Description *string `gorm:"column:description;type:text"`

	// --- СВЯЗИ ---
	// Отношение "принадлежит к" для каждой из связанных таблиц.
	Calculation   CalculationSearching `gorm:"foreignKey:CalculationID"`
	Product Products       `gorm:"foreignKey:ProductID"`
}
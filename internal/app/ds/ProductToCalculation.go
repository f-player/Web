package ds

// ProductToDietComposition соответствует таблице "ProductToDietComposition"
type ProductToDietComposition struct {
	ID                uint           `gorm:"primaryKey;column:id"`
	DietCompositionID uint           `gorm:"column:diet_composition_id;not null"` // Внешний ключ к DietCompositionSearching
	ProductID         uint           `gorm:"column:product_id;not null"`          // Внешний ключ к Products
	Description       *string        `gorm:"column:description;type:text"`
	Ratio             map[string]int `gorm:"-"`

	// --- СВЯЗИ ---
	// Отношение "принадлежит к" для каждой из связанных таблиц.
	DietComposition DietCompositionSearching `gorm:"foreignKey:DietCompositionID"`
	Product         Products                 `gorm:"foreignKey:ProductID"`
}

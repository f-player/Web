package ds

// Products соответствует таблице "Products"
// Это справочник всех возможных факторов риска.
type Products struct {
	ID       uint     `gorm:"primaryKey;column:id"`
	Title    string   `gorm:"column:title;size:255;not null"`
	Image    *string  `gorm:"column:image;size:255"`
	C_pol         int       `gorm:"column:c_pol;not null"`
	N_pol         int       `gorm:"column:n_pol;not null"`

	// --- СВЯЗИ ---
	// Отношение "один-ко-многим" к связующей таблице:
	// Один фактор может быть использован во многих сессиях.
	CalculationLinks []ProductToCalculation `gorm:"foreignKey:ProductID"`
}
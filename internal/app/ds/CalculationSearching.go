package ds

import "time"

// CalculatonSearching соответствует таблице "CalculatonSearching".
type CalculationSearching struct {
	ID             uint       `gorm:"primaryKey;column:id"`
	Status         int        `gorm:"column:status;not null"`
	CreationDate   time.Time  `gorm:"column:creation_date;not null"`
	CreatorID      uint       `gorm:"column:creator_id;not null"` // Внешний ключ
	Moderator      *bool      `gorm:"column:moderator"`
	FormingDate    *time.Time `gorm:"column:forming_date"`
	ComplitionDate *time.Time `gorm:"column:complition_date"`
	
	C_pol         int       `gorm:"column:c_pol;not null"`
	N_pol         int       `gorm:"column:n_pol;not null"`
	PRP            *float64   `gorm:"column:PRP"`
	PGP            *float64   `gorm:"column:PGP"`

	// --- СВЯЗИ ---
	// Отношение "принадлежит к": каждая сессия принадлежит одному пользователю.
	Creator Users `gorm:"foreignKey:CreatorID"`
	// Отношение "один-ко-многим" к связующей таблице:
	// У одной сессии может быть много записей-факторов.
	ProductsLink []ProductToCalculation `gorm:"foreignKey:CalculationID"`
}

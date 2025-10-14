package ds

import "time"

// DietCompositionSearching соответствует таблице "DietCompositionSearching".
type DietCompositionSearching struct {
	ID             uint       `gorm:"primaryKey;column:id"`
	Status         int        `gorm:"column:status;not null"`
	CreationDate   time.Time  `gorm:"column:creation_date;not null"`
	CreatorID      uint       `gorm:"column:creator_id;not null"` // Внешний ключ
	ModeratorID    *uint      `gorm:"column:moderator_id"`
	FormingDate    *time.Time `gorm:"column:forming_date"`
	ComplitionDate *time.Time `gorm:"column:complition_date"`

	C_pol int      `gorm:"column:c_pol;not null"`
	N_pol int      `gorm:"column:n_pol;not null"`
	PRP   *float64 `gorm:"column:prp"`
	PGP   *float64 `gorm:"column:pgp"`

	// --- СВЯЗИ ---
	// Отношение "принадлежит к": каждая сессия принадлежит одному пользователю.
	Creator   Users  `gorm:"foreignKey:CreatorID"`
	Moderator *Users `gorm:"foreignKey:ModeratorID"`
	// Отношение "один-ко-многим" к связующей таблице:
	// У одной сессии может быть много записей-продуктов.
	ProductsLink []ProductToDietComposition `gorm:"foreignKey:DietCompositionID"`
}

package ds

import "time"

type TGSearching struct {
	ID             uint       `gorm:"primaryKey;column:id"`
	Status         int        `gorm:"column:status;not null"`
	CreationDate   time.Time  `gorm:"column:creation_date;not null"`
	CreatorID      uint       `gorm:"column:creator_id;not null"` // Внешний ключ
	Moderator      *bool      `gorm:"column:moderator"`
	FormingDate    *time.Time `gorm:"column:forming_date"`
	ComplitionDate *time.Time `gorm:"column:complition_date"`
	Description    *string    `gorm:"column:description;type:text"`
	Coverage       *float64   `gorm:"column:coverage"`
	Coefficient    *float64   `gorm:"column:coefficient"`

	// --- СВЯЗИ ---
	// Отношение "принадлежит к": каждая сессия принадлежит одному пользователю.
	Creator Users `gorm:"foreignKey:CreatorID"`
	// Отношение "один-ко-многим" к связующей таблице:
	ChannelsLink []ChannelToTG `gorm:"foreignKey:TGID"`
}

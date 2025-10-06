package ds

type Channels struct {
	ID          uint    `gorm:"primaryKey;column:id"`
	Title       string  `gorm:"column:title;size:255;not null"`
	Text        string  `gorm:"column:text;not null"`
	Image       *string `gorm:"column:image;size:255"`
	Subscribers *uint   `gorm:"column:subscribers"`
	Status      *bool   `gorm:"column:status"`

	// --- СВЯЗИ ---
	// Отношение "один-ко-многим" к связующей таблице:
	MsghistoryLinks []ChannelToMsghistory `gorm:"foreignKey:ChannelID"`
}

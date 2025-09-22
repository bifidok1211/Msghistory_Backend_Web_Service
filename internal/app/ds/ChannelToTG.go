package ds

type ChannelToTG struct {
	ID          uint  `gorm:"primaryKey;column:id"`
	TGID        uint  `gorm:"column:tg_id;not null"`      // Внешний ключ к ChannelSearching
	ChannelID   uint  `gorm:"column:channel_id;not null"` // Внешний ключ к Channels
	Views       *uint `gorm:"column:views;type:bigint"`
	RepostLevel *uint `gorm:"column:repost_level;type:bigint"`

	// --- СВЯЗИ ---
	TG      TGSearching `gorm:"foreignKey:TGID"`
	Channel Channels    `gorm:"foreignKey:ChannelID"`
}

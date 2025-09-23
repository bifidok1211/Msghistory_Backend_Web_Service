package ds

type ChannelToTG struct {
	TGID        uint  `gorm:"primaryKey;column:tg_id;not null"`      // Внешний ключ к ChannelSearching
	ChannelID   uint  `gorm:"primaryKey;column:channel_id;not null"` // Внешний ключ к Channels
	Views       *uint `gorm:"column:views;type:bigint"`
	RepostLevel *uint `gorm:"column:repost_level;type:bigint"`

	// --- СВЯЗИ ---
	TG      TGSearching `gorm:"foreignKey:TGID"`
	Channel Channels    `gorm:"foreignKey:ChannelID"`
}

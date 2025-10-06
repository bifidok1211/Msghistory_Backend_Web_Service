package ds

type ChannelToMsghistory struct {
	ID           uint  `gorm:"primaryKey;column:id"`
	MsghistoryID uint  `gorm:"column:msghistory_id;not null"` // Внешний ключ к ChannelSearching
	ChannelID    uint  `gorm:"column:channel_id;not null"`    // Внешний ключ к Channels
	Views        *uint `gorm:"column:views;type:bigint"`
	RepostLevel  *uint `gorm:"column:repost_level;type:bigint"`

	// --- СВЯЗИ ---
	Msghistory MsghistorySearching `gorm:"foreignKey:MsghistoryID"`
	Channel    Channels            `gorm:"foreignKey:ChannelID"`
}

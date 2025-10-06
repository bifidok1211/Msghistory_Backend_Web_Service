package ds

import "time"

type ChannelDTO struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Text        string  `json:"text"`
	Image       *string `json:"image"`
	Subscribers *uint   `json:"subscribers"`
	Status      *bool   `json:"status"`
}

type ChannelCreateRequest struct {
	Title       string `json:"title" binding:"required"`
	Text        string `json:"text" binding:"required"`
	Subscribers *uint  `json:"subscribers"`
}

type ChannelUpdateRequest struct {
	Title       *string `json:"title"`
	Text        *string `json:"text"`
	Subscribers *uint   `json:"subscribers"`
}

type MsghistoryDTO struct {
	ID             uint       `json:"id"`
	Status         int        `json:"status"`
	CreationDate   time.Time  `json:"creation_date"`
	CreatorID      uint       `json:"creator_login"`
	ModeratorID    *uint      `json:"moderator_login"`
	FormingDate    *time.Time `json:"forming_date"`
	ComplitionDate *time.Time `json:"complition_date"`
	Description    *string    `json:"description"`
	Coverage       *float64   `json:"coverage"`
	Coefficient    *float64   `json:"coefficient"`

	Channels []ChannelInMsghistoryDTO `json:"channels,omitempty"`
}

type ChannelInMsghistoryDTO struct {
	ChannelID   uint    `json:"channel_id"`
	Title       string  `json:"title"`
	Text        string  `json:"text"`
	Image       *string `json:"image"`
	Subscribers *uint   `json:"subscribers"`
	Views       *uint   `json:"views"`
	RepostLevel *uint   `json:"repost_level"`
}

type MsghistoryUpdateRequest struct {
	Description *string `json:"description"`
}

type MsghistoryResolveRequest struct {
	Action string `json:"action" binding:"required"` // "complete" | "reject"
}

type ChannelToMsghistoryUpdateRequest struct {
	Views       *uint `json:"views"`
	RepostLevel *uint `json:"repost_level"`
}

type CartBadgeDTO struct {
	MsghistoryID *uint `json:"msghistory_id"`
	Count        int   `json:"count"`
}

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserDTO struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Moderator bool   `json:"moderator"`
}

type UserUpdateRequest struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type PaginatedResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

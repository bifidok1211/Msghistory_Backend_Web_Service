package repository

import (
	"RIP/internal/app/ds"
	"errors"
)

func (r *Repository) GetDraftMsghistory(userID uint) (*ds.MsghistorySearching, error) {
	var msghistory ds.MsghistorySearching

	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&msghistory).Error
	if err != nil {
		return nil, err
	}
	return &msghistory, nil
}

func (r *Repository) CreateMsghistory(msghistory *ds.MsghistorySearching) error {
	return r.db.Create(msghistory).Error
}

func (r *Repository) AddChannelToMsghistory(msghistoryID, channelID uint) error {
	var count int64

	r.db.Model(&ds.ChannelToMsghistory{}).Where("msghistory_id = ? AND channel_id = ?", msghistoryID, channelID).Count(&count)
	if count > 0 {
		return errors.New("channel already in msghistory")
	}

	link := ds.ChannelToMsghistory{
		MsghistoryID: msghistoryID,
		ChannelID:    channelID,
	}
	return r.db.Create(&link).Error
}

func (r *Repository) GetMsghistoryWithChannels(msghistoryID uint) (*ds.MsghistorySearching, error) {
	var msghistory ds.MsghistorySearching

	err := r.db.Preload("ChannelsLink.Channel").First(&msghistory, msghistoryID).Error
	if err != nil {
		return nil, err
	}

	// if msghistory.Status == ds.StatusDeleted {
	// 	return nil, errors.New("msghistory page not found or has been deleted")
	// }

	return &msghistory, nil
}

func (r *Repository) LogicallyDeleteMsghistory(msghistoryID uint) error {
	result := r.db.Exec("UPDATE msghistory_searchings SET status = ? WHERE id = ?", ds.StatusDeleted, msghistoryID)
	return result.Error
}

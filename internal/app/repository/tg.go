package repository

import (
	"RIP/internal/app/ds"
	"errors"
)

func (r *Repository) GetDraftTG(userID uint) (*ds.TGSearching, error) {
	var tg ds.TGSearching

	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&tg).Error
	if err != nil {
		return nil, err
	}
	return &tg, nil
}

func (r *Repository) CreateTG(tg *ds.TGSearching) error {
	return r.db.Create(tg).Error
}

func (r *Repository) AddChannelToTG(tgID, channelID uint) error {
	var count int64

	r.db.Model(&ds.ChannelToTG{}).Where("tg_id = ? AND channel_id = ?", tgID, channelID).Count(&count)
	if count > 0 {
		return errors.New("channel already in tg")
	}

	link := ds.ChannelToTG{
		TGID:      tgID,
		ChannelID: channelID,
	}
	return r.db.Create(&link).Error
}

func (r *Repository) GetTGWithChannels(tgID uint) (*ds.TGSearching, error) {
	var tg ds.TGSearching

	err := r.db.Preload("ChannelsLink.Channel").First(&tg, tgID).Error
	if err != nil {
		return nil, err
	}

	if tg.Status == ds.StatusDeleted {
		return nil, errors.New("tg page not found or has been deleted")
	}

	return &tg, nil
}

func (r *Repository) LogicallyDeleteTG(tgID uint) error {
	result := r.db.Exec("UPDATE tg_searchings SET status = ? WHERE id = ?", ds.StatusDeleted, tgID)
	return result.Error
}

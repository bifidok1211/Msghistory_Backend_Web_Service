package repository

import (
	"RIP/internal/app/ds"
	"fmt"
)

func (r *Repository) GetAllChannels() ([]ds.Channels, error) {
	var channels []ds.Channels

	err := r.db.Find(&channels).Error
	if err != nil {
		return nil, err
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("channels not found")
	}
	return channels, nil
}

func (r *Repository) SearchChannelsByName(title string) ([]ds.Channels, error) {
	var channels []ds.Channels
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&channels).Error // добавили условие
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func (r *Repository) GetChannelByID(id int) (*ds.Channels, error) {
	var channel ds.Channels
	err := r.db.First(&channel, id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

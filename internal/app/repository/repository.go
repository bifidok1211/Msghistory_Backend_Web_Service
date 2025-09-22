package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Channels struct {
	ID          int
	Title       string
	Text        string
	Image       string
	Subscribers int
}

type ChannelsToTG struct {
	ChannelId   int
	Views       int
	Repostlevel int
}

type TG struct {
	Channels []ChannelsToTG
}

func (r *Repository) GetChannels() ([]Channels, error) {
	channels := []Channels{
		{
			ID:          1,
			Title:       "ИУ5",
			Text:        "Кафедра ИУ5 МГТУ им Баумана",
			Image:       "http://localhost:9000/images/tg_channels/IU5.jpg",
			Subscribers: 375,
		},
		{
			ID:          2,
			Title:       "МГТУ им. Н.Э. Баумана",
			Text:        "Официальный канал Бауманки.Здесь вы всегда найдете самые важные новости университета, информацию про мероприятия, интересные факты и многое другое!",
			Image:       "http://localhost:9000/images/tg_channels/main_baum.jpg",
			Subscribers: 24776,
		},
		{
			ID:          3,
			Title:       "Приемная коммиссия",
			Text:        "Здесь вы найдете всю самую необходимую информацию, связанную с поступлением в Бауманку.",
			Image:       "http://localhost:9000/images/tg_channels/priem.jpg",
			Subscribers: 23482,
		},
		{
			ID:          4,
			Title:       "Студенческий совет ИУ",
			Text:        "Новости о движе на факультете ИУ💙",
			Image:       "http://localhost:9000/images/tg_channels/stud_iu.jpg",
			Subscribers: 2019,
		},
		{
			ID:          5,
			Title:       "Профсоюз ИУ",
			Text:        "Официальный телеграм-канал Профсоюза студентов факультета ИУ МГТУ им. Н.Э.Баумана",
			Image:       "http://localhost:9000/images/tg_channels/prof.jpg",
			Subscribers: 1105,
		},
		{
			ID:          6,
			Title:       "Студенческий совет",
			Text:        "Самое студенческое СМИ Бауманки",
			Image:       "http://localhost:9000/images/tg_channels/stud.jpg",
			Subscribers: 7959,
		},
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return channels, nil
}

func (r *Repository) GetChannel(id int) (Channels, error) {
	channels, err := r.GetChannels()
	if err != nil {
		return Channels{}, err
	}

	for _, channel := range channels {
		if channel.ID == id {
			return channel, nil
		}
	}

	return Channels{}, fmt.Errorf("канал не найден")
}

func (r *Repository) GetChannelsByTitle(title string) ([]Channels, error) {
	channels, err := r.GetChannels()
	if err != nil {
		return []Channels{}, err
	}

	var result []Channels
	for _, channel := range channels {
		if strings.Contains(strings.ToLower(channel.Title), strings.ToLower(title)) {
			result = append(result, channel)
		}
	}

	return result, nil
}

var ChannelsInTG = map[int]TG{
	1: {
		Channels: []ChannelsToTG{
			{ChannelId: 0, Views: 25, Repostlevel: 0},
			{ChannelId: 1, Views: 777, Repostlevel: 1},
		},
	},
}

func (r *Repository) GetChannelsInTG(id int) (TG, error) {
	return ChannelsInTG[id], nil
}

func (r *Repository) GetArrayOfChannels(id int) ([]Channels, error) {
	channels, err := r.GetChannels()
	if err != nil {
		return []Channels{}, err
	}

	var result []Channels
	tg, err := r.GetChannelsInTG(id)
	if err != nil {
		return nil, err
	}
	for _, channelRef := range tg.Channels {
		result = append(result, channels[channelRef.ChannelId])
	}
	return result, nil
}

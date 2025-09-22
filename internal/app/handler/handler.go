package handler

import (
	"RIP/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetChannels(ctx *gin.Context) {
	var channels []repository.Channels
	var err error

	searchTG := ctx.Query("tg")
	if searchTG == "" {
		channels, err = h.Repository.GetChannels()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		channels, err = h.Repository.GetChannelsByTitle(searchTG)
		if err != nil {
			logrus.Error(err)
		}
	}

	items, _ := h.Repository.GetChannelsInTG(1)
	count := len(items.Channels)

	ctx.HTML(http.StatusOK, "channels.html", gin.H{
		"channels": channels,
		"tg":       searchTG,
		"count":    count,
		"id":       1,
	})
}

func (h *Handler) GetChannel(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	channel, err := h.Repository.GetChannel(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "single_channel.html", gin.H{
		"channel": channel,
	})
}

func (h *Handler) GetTG(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	tg, err := h.Repository.GetChannelsInTG(id)
	if err != nil {
		logrus.Error(err)
	}

	ChannelsInArray, err := h.Repository.GetArrayOfChannels(id)
	if err != nil {
		logrus.Error(err)
	}
	ChannelsInTG := tg.Channels
	ctx.HTML(http.StatusOK, "tg.html", gin.H{
		"Channels":     ChannelsInArray,
		"ChannelsInTG": ChannelsInTG,
	})
}

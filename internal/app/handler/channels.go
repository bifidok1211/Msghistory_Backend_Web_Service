package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllChannels(ctx *gin.Context) {
	var channels []ds.Channels
	var err error

	searchingChannels := ctx.Query("searchingChannels")
	if searchingChannels == "" {
		channels, err = h.Repository.GetAllChannels()
	} else {
		channels, err = h.Repository.SearchChannelsByName(searchingChannels)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	draftTG, err := h.Repository.GetDraftTG(hardcodedUserID)
	var tgID uint = 0
	var channelsCount int = 0

	if err == nil && draftTG != nil {
		fullTG, err := h.Repository.GetTGWithChannels(draftTG.ID)
		if err == nil {
			tgID = fullTG.ID
			channelsCount = len(fullTG.ChannelsLink)
		}
	}

	ctx.HTML(http.StatusOK, "channels.html", gin.H{
		"channels":       channels,
		"channelsSearch": searchingChannels,
		"tgID":           tgID,
		"channelsCount":  channelsCount,
	})
}

func (h *Handler) GetChannelByID(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	channel, err := h.Repository.GetChannelByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "single_channel.html", channel)
}

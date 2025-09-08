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

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		channels, err = h.Repository.GetChannels()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		channels, err = h.Repository.GetChannelsByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	orders, err := h.Repository.GetOrders()
	if err != nil {
		logrus.Error(err)
	}
	hchannels, err := h.Repository.GetChannels()
	if err != nil {
		logrus.Error(err)
	}

	count := 0
	for _, order := range orders {
		for _, сhannel := range order.Channels {
			for _, с := range hchannels {
				if сhannel.ID == с.ID {
					count++
				}
			}
		}
	}

	ctx.HTML(http.StatusOK, "channels.html", gin.H{
		"channels": channels,
		"query":    searchQuery,
		"count":    count,
		"orderID":  orders[0].ID_order,
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

func (h *Handler) GetOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	order, err := h.Repository.GetOrder(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"order": order,
	})
}

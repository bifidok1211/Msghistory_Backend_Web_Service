package handler

import (
	"RIP/internal/app/repository"

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

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/Msghistory", h.GetAllChannels)
	router.GET("/channel/:id", h.GetChannelByID)
	router.GET("/msghistory/:msghistory_id", h.GetMsghistory)
	router.POST("/msghistory/add/channel/:channel_id", h.AddChannelToMsghistory)
	router.POST("/msghistory/:msghistory_id/delete", h.DeleteMsghistory)

}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/resources", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

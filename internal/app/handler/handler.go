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
	router.GET("/TG", h.GetAllChannels)
	router.GET("/channel/:id", h.GetChannelByID)
	router.GET("/tg/:tg_id", h.GetTG)
	router.POST("/tg/add/channel/:channel_id", h.AddChannelToTG)
	router.POST("/tg/:tg_id/delete", h.DeleteTG)

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

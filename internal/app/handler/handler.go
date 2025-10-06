package handler

import (
	"RIP/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const hardcodedUserID = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterAPI(r *gin.RouterGroup) {
	// Домен услуг (каналов)
	r.GET("/channels", h.GetChannels)
	r.GET("/channels/:id", h.GetChannel)
	r.POST("/channels", h.CreateChannel)
	r.PUT("/channels/:id", h.UpdateChannel)
	r.DELETE("/channels/:id", h.DeleteChannel)
	r.POST("/msghistory/draft/channels/:channel_id", h.AddChannelToDraft)
	r.POST("/channels/:id/image", h.UploadChannelImage)

	// Домен заявок (Msghistory)
	r.GET("/msghistory/cart", h.GetCartBadge)
	r.GET("/msghistory", h.ListMsghistory)
	r.GET("/msghistory/:id", h.GetMsghistory)
	r.PUT("/msghistory/:id", h.UpdateMsghistory)
	r.PUT("/msghistory/:id/form", h.FormMsghistory)
	r.PUT("/msghistory/:id/resolve", h.ResolveMsghistory)
	r.DELETE("/msghistory/:id", h.DeleteMsghistory)

	// Домен м-м
	r.DELETE("/msghistory/:id/channels/:channel_id", h.RemoveChannelFromMsghistory)
	r.PUT("/msghistory/:id/channels/:channel_id", h.UpdateMM)

	// Домен пользователь
	r.POST("/users", h.Register)
	r.GET("/users/:id", h.GetUserData)
	r.PUT("/users/:id", h.UpdateUserData)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

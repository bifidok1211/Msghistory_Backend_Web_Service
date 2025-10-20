package handler

import (
	"RIP/internal/app/config"
	"RIP/internal/app/redis"
	"RIP/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWTConfig  *config.JWTConfig
}

func NewHandler(r *repository.Repository, redis *redis.Client, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redis,
		JWTConfig:  jwtConfig,
	}
}

func (h *Handler) RegisterAPI(r *gin.RouterGroup) {

	// Доступны всем
	r.POST("/users", h.Register)
	r.POST("/auth/login", h.Login)
	r.GET("/channels", h.GetChannels)
	r.GET("/channels/:id", h.GetChannel)

	// Эндпоинты, доступные только авторизованным пользователям
	auth := r.Group("/")
	auth.Use(h.AuthMiddleware)
	{
		// Пользователи
		auth.POST("/auth/logout", h.Logout)
		auth.GET("/users/:id", h.GetUserData)
		auth.PUT("/users/:id", h.UpdateUserData)

		// Заявки
		auth.POST("/msghistory/draft/channels/:channel_id", h.AddChannelToDraft)
		auth.GET("/msghistory/channelscart", h.GetCartBadge)
		auth.GET("/msghistory", h.ListMsghistory)
		auth.GET("/msghistory/:id", h.GetMsghistory)
		auth.PUT("/msghistory/:id", h.UpdateMsghistory)
		auth.PUT("/msghistory/:id/form", h.FormMsghistory)
		auth.DELETE("/msghistory/:id", h.DeleteMsghistory)
		auth.DELETE("/msghistory/:id/channels/:channel_id", h.RemoveChannelFromMsghistory)
		auth.PUT("/msghistory/:id/channels/:channel_id", h.UpdateMM)
	}

	// Эндпоинты, доступные только модераторам
	moderator := r.Group("/")
	moderator.Use(h.AuthMiddleware, h.ModeratorMiddleware)
	{
		// Управление факторами (создание, изменение, удаление)
		moderator.POST("/channels", h.CreateChannel)
		moderator.PUT("/channels/:id", h.UpdateChannel)
		moderator.DELETE("/channels/:id", h.DeleteChannel)
		moderator.POST("/channels/:id/image", h.UploadChannelImage)

		// Управление заявками (завершение/отклонение)
		moderator.PUT("/msghistory/:id/resolve", h.ResolveMsghistory)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

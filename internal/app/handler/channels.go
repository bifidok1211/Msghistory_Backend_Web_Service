package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/channels - список каналов с фильтрацией
func (h *Handler) GetChannels(c *gin.Context) {
	title := c.Query("title")

	channels, total, err := h.Repository.ChannelsList(title)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	var channelDTOs []ds.ChannelDTO
	for _, f := range channels {
		channelDTOs = append(channelDTOs, ds.ChannelDTO{
			ID:          f.ID,
			Title:       f.Title,
			Text:        f.Text,
			Image:       f.Image,
			Subscribers: f.Subscribers,
			Status:      f.Status,
		})
	}

	c.JSON(http.StatusOK, ds.PaginatedResponse{
		Items: channelDTOs,
		Total: total,
	})
}

// GET /api/channels/:id - один канал
func (h *Handler) GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	channel, err := h.Repository.GetChannelByID(id)
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	channelDTO := ds.ChannelDTO{
		ID:          channel.ID,
		Title:       channel.Title,
		Text:        channel.Text,
		Image:       channel.Image,
		Subscribers: channel.Subscribers,
		Status:      channel.Status,
	}

	c.JSON(http.StatusOK, channelDTO)
}

// POST /api/channels - создание канала
func (h *Handler) CreateChannel(c *gin.Context) {
	var req ds.ChannelCreateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	statusValue := false

	channel := ds.Channels{
		Title:       req.Title,
		Text:        req.Text,
		Subscribers: req.Subscribers,
		Status:      &statusValue,
	}

	if err := h.Repository.CreateChannel(&channel); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	channelDTO := ds.ChannelDTO{
		ID:          channel.ID,
		Title:       channel.Title,
		Text:        channel.Text,
		Image:       channel.Image,
		Subscribers: channel.Subscribers,
		Status:      channel.Status,
	}

	c.JSON(http.StatusCreated, channelDTO)
}

// PUT /api/channels/:id - обновление канала
func (h *Handler) UpdateChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.ChannelUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	channel, err := h.Repository.UpdateChannel(uint(id), req)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	channelDTO := ds.ChannelDTO{
		ID:          channel.ID,
		Title:       channel.Title,
		Text:        channel.Text,
		Image:       channel.Image,
		Subscribers: channel.Subscribers,
		Status:      channel.Status,
	}

	c.JSON(http.StatusOK, channelDTO)
}

// DELETE /api/channels/:id - удаление канала
func (h *Handler) DeleteChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteChannel(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Канал удален",
	})
}

// POST /api/msghistory/draft/channels/:channel_id - добавление канала в черновик
func (h *Handler) AddChannelToDraft(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("channel_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.AddChannelToDraft(userID, uint(channelID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Черновик создан. Канал добавлен в черновик.",
	})
}

// POST /api/channels/:id/image - загрузка изображения канала
func (h *Handler) UploadChannelImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	imageURL, err := h.Repository.UploadChannelImage(uint(id), file)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": imageURL})
}

package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetChannels godoc
// @Summary      Получить список каналов (все)
// @Description  Возвращает постраничный список каналов риска.
// @Tags         channels
// @Produce      json
// @Param        title query string false "Фильтр по названию канала"
// @Success      200 {object} ds.PaginatedResponse
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /channels [get]
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

// GetChannel godoc
// @Summary      Получить один канал по ID (все)
// @Description  Возвращает детальную информацию о канале риска.
// @Tags         channels
// @Produce      json
// @Param        id path int true "ID канала"
// @Success      200 {object} ds.ChannelDTO
// @Failure      404 {object} map[string]string "Фактор не найден"
// @Router       /channels/{id} [get]
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

// CreateChannel godoc
// @Summary      Создать новый канал (только модератор)
// @Description  Создает новую запись о канале риска.
// @Tags         channels
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        channelData body ds.ChannelCreateRequest true "Данные нового канала"
// @Success      201 {object} ds.ChannelDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен (не модератор)"
// @Router       /channels [post]
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
		Image:       req.Image,
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

// UpdateChannel godoc
// @Summary      Обновить канал (только модератор)
// @Description  Обновляет информацию о существующем канале риска.
// @Tags         channels
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID канала"
// @Param        updateData body ds.ChannelUpdateRequest true "Данные для обновления"
// @Success      200 {object} ds.ChannelDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /channels/{id} [put]
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

// DeleteChannel godoc
// @Summary      Удалить канал (только модератор)
// @Description  Удаляет канал риска из системы.
// @Tags         channels
// @Security     ApiKeyAuth
// @Param        id path int true "ID канала для удаления"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /channels/{id} [delete]
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

// AddChannelToDraft godoc
// @Summary      Добавить канал в черновик заявки (все)
// @Description  Находит или создает черновик заявки для текущего пользователя и добавляет в него канал.
// @Tags         channels
// @Security     ApiKeyAuth
// @Param        channel_id path int true "ID канала для добавления"
// @Success      201 {object} map[string]string "Сообщение об успехе"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /msghistory/draft/channels/{channel_id} [post]
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

// UploadChannelImage godoc
// @Summary      Загрузить изображение для канала (только модератор)
// @Description  Загружает и привязывает изображение к каналу риска.
// @Tags         channels
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID канала"
// @Param        file formData file true "Файл изображения"
// @Success      200 {object} map[string]string "URL загруженного изображения"
// @Failure      400 {object} map[string]string "Файл не предоставлен"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /channels/{id}/image [post]
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

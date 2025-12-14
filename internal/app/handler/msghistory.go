package handler

import (
	"RIP/internal/app/ds"
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GET /api/msghistory/cart - иконка корзины

// GetCartBadge godoc
// @Summary      Получить информацию для иконки корзины (авторизованный пользователь)
// @Description  Возвращает ID черновика текущего пользователя и количество каналов в нем.
// @Tags         msghistory
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} ds.CartBadgeDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /msghistory/channelscart [get]
func (h *Handler) GetCartBadge(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	draft, err := h.Repository.GetDraftMsghistory(userID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			MsghistoryID: nil,
			Count:        0,
		})
		return
	}

	fullMsghistory, err := h.Repository.GetMsghistoryWithChannels(draft.ID)
	if err != nil {
		logrus.Error("Error getting msghistory with channels:", err)
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			MsghistoryID: nil,
			Count:        0,
		})
		return
	}

	c.JSON(http.StatusOK, ds.CartBadgeDTO{
		MsghistoryID: &fullMsghistory.ID,
		Count:        len(fullMsghistory.ChannelsLink),
	})
}

// GET /api/msghistory - список заявок с фильтрацией

// ListMsghistory godoc
// @Summary      Получить список заявок (авторизованный пользователь)
// @Description  Возвращает отфильтрованный список всех сформированных заявок (кроме черновиков и удаленных).
// @Tags         msghistory
// @Produce      json
// @Security     ApiKeyAuth
// @Param        status query int false "Фильтр по статусу заявки"
// @Param        from query string false "Фильтр по дате 'от' (формат YYYY-MM-DD)"
// @Param        to query string false "Фильтр по дате 'до' (формат YYYY-MM-DD)"
// @Success      200 {array} ds.MsghistoryDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /msghistory [get]
func (h *Handler) ListMsghistory(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}
	isModerator := isUserModerator(c)

	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	msghistoryList, err := h.Repository.MsghistoryListFiltered(userID, isModerator, status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, msghistoryList)
}

// GET /api/msghistory/:id - одна заявка с услугами

// GetMsghistory godoc
// @Summary      Получить одну заявку по ID (авторизованный пользователь)
// @Description  Возвращает полную информацию о заявке, включая привязанные каналы.
// @Tags         msghistory
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Success      200 {object} ds.MsghistoryDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      404 {object} map[string]string "Заявка не найдена"
// @Router       /msghistory/{id} [get]
func (h *Handler) GetMsghistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	msghistory, err := h.Repository.GetMsghistoryWithChannels(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	var channels []ds.ChannelInMsghistoryDTO
	for _, link := range msghistory.ChannelsLink {
		channels = append(channels, ds.ChannelInMsghistoryDTO{
			ChannelID:   link.ChannelID,
			Title:       link.Channel.Title,
			Text:        link.Channel.Text,
			Image:       link.Channel.Image,
			Subscribers: link.Channel.Subscribers,
			Views:       link.Views,
			RepostLevel: link.RepostLevel,
		})
	}

	msghistoryDTO := ds.MsghistoryDTO{
		ID:             msghistory.ID,
		Status:         msghistory.Status,
		CreationDate:   msghistory.CreationDate,
		CreatorID:      msghistory.Creator.ID,
		ModeratorID:    nil,
		FormingDate:    msghistory.FormingDate,
		ComplitionDate: msghistory.ComplitionDate,
		Description:    msghistory.Description,
		Coverage:       msghistory.Coverage,
		Coefficient:    msghistory.Coefficient,
		Channels:       channels,
	}

	if msghistory.ModeratorID != nil {
		msghistoryDTO.ModeratorID = &msghistory.Moderator.ID
	}

	c.JSON(http.StatusOK, msghistoryDTO)
}

// PUT /api/msghistory/:id - изменение полей заявки

// UpdateMsghistory godoc
// @Summary      Обновить данные заявки (авторизованный пользователь)
// @Description  Позволяет пользователю обновить поля своей заявки (возраст, пол, вес, рост).
// @Tags         msghistory
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        updateData body ds.MsghistoryUpdateRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /msghistory/{id} [put]
func (h *Handler) UpdateMsghistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.MsghistoryUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateMsghistoryUserFields(uint(id), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Данные заявки обновлены",
	})
}

// PUT /api/msghistory/:id/form - сформировать заявку

// FormMsghistory godoc
// @Summary      Сформировать заявку (авторизованный пользователь)
// @Description  Переводит заявку из статуса "черновик" в "сформирована".
// @Tags         msghistory
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки (черновика)"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Не все поля заполнены"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /msghistory/{id}/form [put]
func (h *Handler) FormMsghistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.FormMsghistory(uint(id), userID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка сформирована",
	})
}

// PUT /api/msghistory/:id/resolve - завершить/отклонить заявку

// ResolveMsghistory godoc
// @Summary      Завершить или отклонить заявку (только модератор)
// @Description  Модератор завершает (с расчетом) или отклоняет заявку.
// @Tags         msghistory
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        action body ds.MsghistoryResolveRequest true "Действие: 'complete' или 'reject'"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /msghistory/{id}/resolve [put]
func (h *Handler) ResolveMsghistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.MsghistoryResolveRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}
	moderatorID := uint(userID)

	// 1. Меняем статус
	if err := h.Repository.ResolveMsghistory(uint(id), moderatorID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	// 2. Сбор данных для сервиса
	if req.Action == "complete" {
		fullMsg, err := h.Repository.GetMsghistoryFull(uint(id))

		if err == nil {
			var items []ds.ItemData

			for _, link := range fullMsg.ChannelsLink {
				item := ds.ItemData{
					ChannelID:   link.ChannelID,
					Views:       link.Views,
					RepostLevel: link.RepostLevel,
				}

				// ИСПРАВЛЕНИЕ ЗДЕСЬ:
				// Проверяем ID канала вместо сравнения с nil.
				// Если ChannelID != 0, значит GORM успешно подтянул данные канала.
				if link.Channel.ID != 0 {
					item.Subscribers = link.Channel.Subscribers
				}
				// Примечание: Если link.Channel — это пустая структура, то link.Channel.Subscribers и так будет nil,
				// так что условие можно даже опустить, но с проверкой ID надежнее.

				items = append(items, item)
			}

			calcReq := ds.AsyncCalcRequest{
				ID:    fullMsg.ID,
				Items: items,
			}

			go sendAsyncCalculation("http://localhost:8000/api/analysis/", calcReq)

		} else {
			logrus.Errorf("Failed to fetch msghistory data for async calc: %v", err)
		}
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка обработана, расчет запущен.",
	})
}

// DELETE /api/msghistory/:id - удаление заявки

// DeleteMsghistory godoc
// @Summary      Удалить заявку (авторизованный пользователь)
// @Description  Логически удаляет заявку, переводя ее в статус "удалена".
// @Tags         msghistory
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"ы
// @Router       /msghistory/{id} [delete]
func (h *Handler) DeleteMsghistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.LogicallyDeleteMsghistory(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка удалена",
	})
}

// DELETE /api/msghistory/:id/channels/:channel_id - удаление канала из заявки

// RemoveChannelFromMsghistory godoc
// @Summary      Удалить канал из заявки (авторизованный пользователь)
// @Description  Удаляет связь между заявкой и каналом.
// @Tags         m-m
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        channel_id path int true "ID канала"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /msghistory/{id}/channels/{channel_id} [delete]
func (h *Handler) RemoveChannelFromMsghistory(c *gin.Context) {
	msghistoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	channelID, err := strconv.Atoi(c.Param("channel_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.RemoveChannelFromMsghistory(uint(msghistoryID), uint(channelID)); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Канал удален из заявки",
	})
}

// PUT /api/msghistory/:id/channels/:channel_id - изменение м-м связи

// UpdateMM godoc
// @Summary      Обновить описание канала в заявке (авторизованный пользователь)
// @Description  Изменяет дополнительное описание для конкретного канала в рамках одной заявки.
// @Tags         m-m
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        channel_id path int true "ID канала"
// @Param        updateData body ds.ChannelToMsghistoryUpdateRequest true "Новое описание"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /msghistory/{id}/channels/{channel_id} [put]
func (h *Handler) UpdateMM(c *gin.Context) {
	msghistoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	channelID, err := strconv.Atoi(c.Param("channel_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.ChannelToMsghistoryUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	updateData := ds.ChannelToMsghistory{
		Views:       req.Views,
		RepostLevel: req.RepostLevel,
	}

	if err := h.Repository.UpdateMM(uint(msghistoryID), uint(channelID), updateData); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Дополнительная информация к каналу обновлена",
	})
}

// PUT /api/internal/msghistory/result
func (h *Handler) SetMsghistoryResult(c *gin.Context) {
	// 1. Простая авторизация по токену
	token := c.GetHeader("Authorization")
	expectedToken := "secret12"

	if token != expectedToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid token"})
		return
	}

	// 2. Парсим ответ от Python
	var res ds.AsyncCalcResponse
	if err := c.BindJSON(&res); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	// 3. Пишем Coverage и Coefficient в базу
	if err := h.Repository.UpdateMsghistoryResults(res.ID, res.Coverage, res.Coefficient); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "results updated"})
}

func sendAsyncCalculation(url string, data ds.AsyncCalcRequest) {
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.Errorf("Failed to send async calc request: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Async service returned non-200 status: %d", resp.StatusCode)
	}
}

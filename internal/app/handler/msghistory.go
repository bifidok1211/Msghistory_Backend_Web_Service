package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GET /api/msghistory/cart - иконка корзины
func (h *Handler) GetCartBadge(c *gin.Context) {
	draft, err := h.Repository.GetDraftMsghistory(hardcodedUserID)
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
func (h *Handler) ListMsghistory(c *gin.Context) {
	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	msghistoryList, err := h.Repository.MsghistoryListFiltered(status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, msghistoryList)
}

// GET /api/msghistory/:id - одна заявка с услугами
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
func (h *Handler) FormMsghistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.FormMsghistory(uint(id), hardcodedUserID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка сформирована",
	})
}

// PUT /api/msghistory/:id/resolve - завершить/отклонить заявку
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

	moderatorID := uint(hardcodedUserID)
	if err := h.Repository.ResolveMsghistory(uint(id), moderatorID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка обработана модератором",
	})
}

// DELETE /api/msghistory/:id - удаление заявки
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
		"message": "Дополнительная информация к фаткору обновлена",
	})
}

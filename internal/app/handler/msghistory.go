package handler

import (
	"RIP/internal/app/ds"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const hardcodedUserID = 1

func (h *Handler) AddChannelToMsghistory(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("channel_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	msghistory, err := h.Repository.GetDraftMsghistory(hardcodedUserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newMsghistory := ds.MsghistorySearching{
			CreatorID: hardcodedUserID,
			Status:    ds.StatusDraft,
		}
		if createErr := h.Repository.CreateMsghistory(&newMsghistory); createErr != nil {
			h.errorHandler(c, http.StatusInternalServerError, createErr)
			return
		}
		msghistory = &newMsghistory
	} else if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	if err = h.Repository.AddChannelToMsghistory(msghistory.ID, uint(channelID)); err != nil {
	}

	c.Redirect(http.StatusFound, "/Msghistory")
}

func (h *Handler) GetMsghistory(c *gin.Context) {
	msghistoryID, err := strconv.Atoi(c.Param("msghistory_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	msghistory, err := h.Repository.GetMsghistoryWithChannels(uint(msghistoryID))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}
	if msghistory.Status == ds.StatusDeleted {
		c.HTML(http.StatusOK, "deleted_page.html", msghistory)
		return
	}

	if len(msghistory.ChannelsLink) == 0 {
		h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty msghistory page, add channels first"))
		return
	}

	c.HTML(http.StatusOK, "msghistory.html", msghistory)
}

func (h *Handler) DeleteMsghistory(c *gin.Context) {
	msghistoryID, _ := strconv.Atoi(c.Param("msghistory_id"))

	if err := h.Repository.LogicallyDeleteMsghistory(uint(msghistoryID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/Msghistory")
}

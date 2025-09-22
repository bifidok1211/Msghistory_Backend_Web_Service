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

func (h *Handler) AddChannelToTG(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("channel_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	tg, err := h.Repository.GetDraftTG(hardcodedUserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newTG := ds.TGSearching{
			CreatorID: hardcodedUserID,
			Status:    ds.StatusDraft,
		}
		if createErr := h.Repository.CreateTG(&newTG); createErr != nil {
			h.errorHandler(c, http.StatusInternalServerError, createErr)
			return
		}
		tg = &newTG
	} else if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	if err = h.Repository.AddChannelToTG(tg.ID, uint(channelID)); err != nil {
	}

	c.Redirect(http.StatusFound, "/TG")
}

func (h *Handler) GetTG(c *gin.Context) {
	tgID, err := strconv.Atoi(c.Param("tg_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	tg, err := h.Repository.GetTGWithChannels(uint(tgID))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	if len(tg.ChannelsLink) == 0 {
		h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty tg page, add channels first"))
		return
	}

	c.HTML(http.StatusOK, "tg.html", tg)
}

func (h *Handler) DeleteTG(c *gin.Context) {
	tgID, _ := strconv.Atoi(c.Param("tg_id"))

	if err := h.Repository.LogicallyDeleteTG(uint(tgID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/TG")
}

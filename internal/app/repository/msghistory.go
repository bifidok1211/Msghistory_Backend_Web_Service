package repository

import (
	"RIP/internal/app/ds"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// GET /api/msghistory/cart - иконка корзины
func (r *Repository) GetDraftMsghistory(userID uint) (*ds.MsghistorySearching, error) {
	var msghistory ds.MsghistorySearching
	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&msghistory).Error
	if err != nil {
		return nil, err
	}
	return &msghistory, nil
}

// GET /api/msghistory/cart - иконка корзины
// GET /api/msghistory/:id - одна заявка с услугами
func (r *Repository) GetMsghistoryWithChannels(msghistoryID uint) (*ds.MsghistorySearching, error) {
	var msghistory ds.MsghistorySearching
	err := r.db.Preload("ChannelsLink.Channel").Preload("Creator").Preload("Moderator").First(&msghistory, msghistoryID).Error
	if err != nil {
		return nil, err
	}

	if msghistory.Status == ds.StatusDeleted {
		return nil, errors.New("msghistory page not found or has been deleted")
	}

	return &msghistory, nil
}

// GET /api/msghistory - список заявок с фильтрацией
func (r *Repository) MsghistoryListFiltered(userID uint, isModerator bool, status, from, to string) ([]ds.MsghistoryDTO, error) {
	var msghistoryList []ds.MsghistorySearching
	query := r.db.Preload("Creator").Preload("Moderator")

	query = query.Where("status != ? AND status != ?", ds.StatusDeleted, ds.StatusDraft)

	if !isModerator {
		query = query.Where("creator_id = ?", userID)
	}

	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	if from != "" {
		if fromTime, err := time.Parse("2006-01-02", from); err == nil {
			query = query.Where("forming_date >= ?", fromTime)
		}
	}

	if to != "" {
		if toTime, err := time.Parse("2006-01-02", to); err == nil {
			query = query.Where("forming_date <= ?", toTime)
		}
	}

	if err := query.Find(&msghistoryList).Error; err != nil {
		return nil, err
	}

	var result []ds.MsghistoryDTO
	for _, msghistory := range msghistoryList {
		dto := ds.MsghistoryDTO{
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
		}

		if msghistory.ModeratorID != nil {
			dto.ModeratorID = &msghistory.Moderator.ID
		}
		result = append(result, dto)
	}
	return result, nil
}

// PUT /api/msghistory/:id - изменение полей заявки
func (r *Repository) UpdateMsghistoryUserFields(id uint, req ds.MsghistoryUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&ds.MsghistorySearching{}).Where("id = ?", id).Updates(updates).Error
}

// PUT /api/msghistory/:id/form - сформировать заявку
func (r *Repository) FormMsghistory(id uint, creatorID uint) error {
	var msghistory ds.MsghistorySearching
	if err := r.db.First(&msghistory, id).Error; err != nil {
		return err
	}

	if msghistory.CreatorID != creatorID {
		return errors.New("only creator can form msghistory")
	}

	if msghistory.Status != ds.StatusDraft {
		return errors.New("only draft msghistory can be formed")
	}

	if msghistory.Description == nil {
		return errors.New("description are required")
	}

	now := time.Now()
	return r.db.Model(&msghistory).Updates(map[string]interface{}{
		"status":       ds.StatusFormed,
		"forming_date": now,
	}).Error
}

// PUT /api/msghistory/:id/resolve - завершить/отклонить заявку
func (r *Repository) ResolveMsghistory(id uint, moderatorID uint, action string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		var msghistory ds.MsghistorySearching
		if err := tx.Preload("ChannelsLink.Channel").First(&msghistory, id).Error; err != nil {
			return err
		}

		if msghistory.Status != ds.StatusFormed {
			return errors.New("only formed msghistory can be resolved")
		}

		now := time.Now()
		updates := map[string]interface{}{
			"moderator_id":    moderatorID,
			"complition_date": now,
		}

		switch action {
		case "complete":
			{
				updates["status"] = ds.StatusCompleted
				coverage, coefficient := r.postAnalysis(msghistory)
				updates["coverage"] = coverage
				updates["coefficient"] = coefficient
			}
		case "reject":
			{
				updates["status"] = ds.StatusRejected
			}
		default:
			{
				return errors.New("invalid action, must be 'complete' or 'reject'")
			}
		}

		if err := tx.Model(&msghistory).Updates(updates).Error; err != nil {
			return err
		}

		var channelIDs []uint
		for _, link := range msghistory.ChannelsLink {
			channelIDs = append(channelIDs, link.ChannelID)
		}

		if len(channelIDs) > 0 {
			if err := tx.Model(&ds.Channels{}).Where("id IN ?", channelIDs).Update("status", false).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Функция расчета
// postAnalysis считает:
// 1) Coverage% = (∑ min(views_i, subscribers_i)) / (∑ subscribers_i уникальных каналов) * 100
// 2) R = (∑ min(views_i, subscribers_i) на уровнях repostLevel>0) / (∑ min(views_i, subscribers_i) на уровне 0)
//
// Если у канала нет subscribers (nil), ограничение не применяется.
// Если views nil — считаем 0. Если repostLevel nil — считаем как корень (0).
func (r *Repository) postAnalysis(msghistory ds.MsghistorySearching) (coveragePercent float64, retentionR float64) {
	var totalEffViews float64    // ∑ min(views, subscribers)
	var totalSubscribers float64 // ∑ subscribers по уникальным каналам

	var rootEffViews float64   // уровень 0
	var repostEffViews float64 // уровни > 0

	seen := make(map[uint]struct{}) // чтобы не суммировать подписчиков одного канала дважды

	for _, link := range msghistory.ChannelsLink {
		// --- извлекаем просмотры (может быть nil)
		var viewsF float64
		if link.Views != nil {
			viewsF = float64(*link.Views)
		}

		// --- извлекаем подписчиков (может быть nil)
		var subsF float64
		if link.Channel.Subscribers != nil {
			subsF = float64(*link.Channel.Subscribers)
		}

		// --- эффективные просмотры: min(views, subscribers) если subs известны
		effViews := viewsF
		if link.Channel.Subscribers != nil && subsF < effViews {
			effViews = subsF
		}

		totalEffViews += effViews

		// --- суммируем подписчиков по уникальным каналам
		if _, ok := seen[link.ChannelID]; !ok {
			if link.Channel.Subscribers != nil {
				totalSubscribers += subsF
			}
			seen[link.ChannelID] = struct{}{}
		}

		// --- распределяем по уровням для R
		level := uint(0)
		if link.RepostLevel != nil {
			level = *link.RepostLevel
		}
		if level == 0 {
			rootEffViews += effViews
		} else {
			repostEffViews += effViews
		}
	}

	// Coverage %
	if totalSubscribers > 0 {
		coveragePercent = (totalEffViews / totalSubscribers) * 100.0
	} else {
		coveragePercent = 0
	}

	// R = репосты / корень
	if rootEffViews > 0 {
		retentionR = repostEffViews / rootEffViews
	} else {
		retentionR = 0
	}

	return coveragePercent, retentionR
}

// DELETE /api/msghistory/:id - удаление заявки
func (r *Repository) LogicallyDeleteMsghistory(msghistoryID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var msghistory ds.MsghistorySearching

		if err := tx.Preload("ChannelsLink").First(&msghistory, msghistoryID).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":       ds.StatusDeleted,
			"forming_date": time.Now(),
		}

		if err := tx.Model(&ds.MsghistorySearching{}).Where("id = ?", msghistoryID).Updates(updates).Error; err != nil {
			return err
		}

		var channelIDs []uint
		for _, link := range msghistory.ChannelsLink {
			channelIDs = append(channelIDs, link.ChannelID)
		}

		if len(channelIDs) > 0 {
			if err := tx.Model(&ds.Channels{}).Where("id IN ?", channelIDs).Update("status", false).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DELETE /api/msghistory/:id/channels/:channel_id - удаление канала из заявки
func (r *Repository) RemoveChannelFromMsghistory(msghistoryID, channelID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		result := tx.Where("msghistory_id = ? AND channel_id = ?", msghistoryID, channelID).Delete(&ds.ChannelToMsghistory{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("channel not found in this msghistory")
		}

		if err := tx.Model(&ds.Channels{}).Where("id = ?", channelID).Update("status", false).Error; err != nil {
			return err
		}

		var remainingCount int64
		if err := tx.Model(&ds.ChannelToMsghistory{}).Where("msghistory_id = ?", msghistoryID).Count(&remainingCount).Error; err != nil {
			return err
		}

		if remainingCount == 0 {
			updates := map[string]interface{}{
				"status":       ds.StatusDeleted,
				"forming_date": time.Now(),
			}
			if err := tx.Model(&ds.MsghistorySearching{}).Where("id = ?", msghistoryID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PUT /api/msghistory/:id/channels/:channel_id - изменение м-м связи
func (r *Repository) UpdateMM(msghistoryID, channelID uint, updateData ds.ChannelToMsghistory) error {
	var link ds.ChannelToMsghistory
	if err := r.db.Where("msghistory_id = ? AND channel_id = ?", msghistoryID, channelID).First(&link).Error; err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if updateData.Views != nil {
		updates["views"] = *updateData.Views
	}
	if updateData.RepostLevel != nil {
		updates["repost_level"] = *updateData.RepostLevel
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&link).Updates(updates).Error
}

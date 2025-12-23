package repository

import (
	"RIP/internal/app/ds"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GET /api/channels - список каналов с фильтрацией
func (r *Repository) ChannelsList(title string) ([]ds.Channels, int64, error) {
	var channels []ds.Channels
	var total int64

	query := r.db.Model(&ds.Channels{})
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	channelsQuery := query.Order("id asc")
	if err := channelsQuery.Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// GET /api/channels/:id - один канал
func (r *Repository) GetChannelByID(id int) (*ds.Channels, error) {
	var channel ds.Channels
	err := r.db.First(&channel, id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// POST /api/channels - создание канала
func (r *Repository) CreateChannel(channel *ds.Channels) error {
	return r.db.Create(channel).Error
}

// PUT /api/channels/:id - обновление канала
func (r *Repository) UpdateChannel(id uint, req ds.ChannelUpdateRequest) (*ds.Channels, error) {
	var channel ds.Channels
	if err := r.db.First(&channel, id).Error; err != nil {
		return nil, err
	}

	if req.Title != nil {
		channel.Title = *req.Title
	}
	if req.Text != nil {
		channel.Text = *req.Text
	}
	if req.Subscribers != nil {
		channel.Subscribers = req.Subscribers
	}
	if req.Image != nil {
		channel.Image = req.Image
	}

	if err := r.db.Save(&channel).Error; err != nil {
		return nil, err
	}

	return &channel, nil
}

// DELETE /api/channels/:id - удаление канала
func (r *Repository) DeleteChannel(id uint) error {
	var channel ds.Channels
	var imageURLToDelete string

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&channel, id).Error; err != nil {
			return err
		}
		if channel.Image != nil {
			imageURLToDelete = *channel.Image
		}
		if err := tx.Delete(&ds.Channels{}, id).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	if imageURLToDelete != "" {
		parsedURL, err := url.Parse(imageURLToDelete)
		if err != nil {
			log.Printf("ERROR: could not parse image URL for deletion: %v", err)
			return nil
		}

		objectName := strings.TrimPrefix(parsedURL.Path, fmt.Sprintf("/%s/", r.bucketName))

		err = r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("ERROR: failed to delete object '%s' from MinIO: %v", objectName, err)
		}
	}

	return nil
}

// POST /api/msghistory/draft/channels/:channel_id - добавление канала в черновик
func (r *Repository) AddChannelToDraft(userID, channelID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var msghistory ds.MsghistorySearching
		err := tx.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&msghistory).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newMsghistory := ds.MsghistorySearching{
					CreatorID:    userID,
					Status:       ds.StatusDraft,
					CreationDate: time.Now(),
				}
				if err := tx.Create(&newMsghistory).Error; err != nil {
					return fmt.Errorf("failed to create draft msghistory: %w", err)
				}
				msghistory = newMsghistory
			} else {
				return err
			}
		}

		var count int64
		tx.Model(&ds.ChannelToMsghistory{}).Where("msghistory_id = ? AND channel_id = ?", msghistory.ID, channelID).Count(&count)
		if count > 0 {
			return errors.New("channel already in msghistory")
		}

		link := ds.ChannelToMsghistory{
			MsghistoryID: msghistory.ID,
			ChannelID:    channelID,
		}

		if err := tx.Create(&link).Error; err != nil {
			return fmt.Errorf("failed to add channel to msghistory: %w", err)
		}

		if err := tx.Model(&ds.Channels{}).Where("id = ?", channelID).Update("status", true).Error; err != nil {
			return fmt.Errorf("failed to update channel status: %w", err)
		}
		return nil
	})
}

// POST /api/channels/:id/image - загрузка изображения канала
func (r *Repository) UploadChannelImage(channelID uint, fileHeader *multipart.FileHeader) (string, error) {
	var finalImageURL string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var channel ds.Channels
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&channel, channelID).Error; err != nil {
			return fmt.Errorf("channel with id %d not found: %w", channelID, err)
		}

		const imagePathPrefix = "Images/"

		if channel.Image != nil && *channel.Image != "" {
			oldImageURL, err := url.Parse(*channel.Image)
			if err == nil {
				oldObjectName := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.bucketName))
				r.minioClient.RemoveObject(context.Background(), r.bucketName, oldObjectName, minio.RemoveObjectOptions{})
			}
		}

		fileName := filepath.Base(fileHeader.Filename)
		objectName := imagePathPrefix + fileName

		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = r.minioClient.PutObject(context.Background(), r.bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		})

		if err != nil {
			return fmt.Errorf("failed to upload to minio: %w", err)
		}

		imageURL := fmt.Sprintf("http://%s/%s/%s", r.minioEndpoint, r.bucketName, objectName)

		if err := tx.Model(&channel).Update("image", imageURL).Error; err != nil {
			return fmt.Errorf("failed to update channel image url in db: %w", err)
		}

		finalImageURL = imageURL
		return nil
	})
	if err != nil {
		return "", err
	}
	return finalImageURL, nil
}

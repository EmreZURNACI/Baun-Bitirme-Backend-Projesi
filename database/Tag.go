package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	Uuid      string     `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string     `json:"name" gorm:"type:varchar(255)" validate:"required"`
	Question  []Question `gorm:"many2many:questions_tags"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (t *TagHandler) CreateTable() error {
	migrator := t.Db.Migrator()
	if !migrator.HasTable(&Tag{}) {
		err := migrator.CreateTable(&Tag{})
		if err != nil {
			return err
		}
	}
	return nil
}
func (q *TagHandler) GetTags() ([]Tag, error) {
	var count int64
	if err := q.Db.Model(&Tag{}).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("Veri tabanında tag sayısı alınırken hata oluştu: %v", err)
	}

	if count == 0 {
		return nil, fmt.Errorf("Veri tabanında hiç tag bulunmamaktadır.")
	}

	var tags []Tag
	err := q.Db.Model(&Tag{}).Find(&tags).Error

	if err != nil {
		return nil, fmt.Errorf("Tagler getirilirken hata oluştu: %v", err)
	}

	return tags, nil
}
func (q *TagHandler) AddTag(tag Tag) error {
	tx := q.Db.Begin()
	if tx.Error != nil {
		return errors.New("Tx başlatılamadı")
	}

	tag.Name = strings.ToLower(tag.Name)

	var t Tag
	err := tx.Where("name = ?", tag.Name).First(&t).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return fmt.Errorf("Tag arama hatası: %w", err)
	}

	if t.Name != "" {
		tx.Rollback()
		return errors.New("Bu tag zaten mevcut")
	}

	if err := tx.Create(&tag).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Tag oluşturulurken hata: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Commit sırasında hata: %w", err)
	}

	return nil
}

func (q *TagHandler) DeleteTag(tag Tag) error {
	tx := q.Db.Begin()

	if tx.Error != nil {
		tx.Rollback()
		return errors.New("Tx başlatılamadı")
	}

	if err := tx.Model(&Tag{}).Where("uuid = ?", tag.Uuid).Delete(&Tag{}).Error; err != nil {
		tx.Rollback()
		return errors.New("Sorgu çalıştırılırken hata ile karşılaşıldı.")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Commit sırasında hata: %w", err)
	}

	return nil
}
func (q *TagHandler) IsTagExists(tag_uuid string) (Tag, error) {

	var tag Tag

	result := q.Db.Table("tags").Where("uuid = ?", tag_uuid).First(&tag)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Tag{}, errors.New("İlgili uuid'ye ait tag bulunmamaktadır.")
		}
		return Tag{}, errors.New("Sorgu çalıştırılırken hata ile karşılaşıldı.")

	}

	return tag, nil

}

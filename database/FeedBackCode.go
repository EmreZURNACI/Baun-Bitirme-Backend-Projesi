package database

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Code struct {
	Uuid      string    `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null"`
	Code      string    `json:"code" gorm:"type:varchar(6)"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (c *CodeHandler) CreateTable() error {
	migrator := c.Db.Migrator()

	if !migrator.HasTable(&Code{}) {
		err := migrator.CreateTable(&Code{})
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *CodeHandler) SaveCode(email string, code string) error {
	tx := c.Db.Begin()

	if tx.Error != nil {
		return fmt.Errorf("Tx başlatılamadı. %w", tx.Error)
	}

	expiresAt := time.Now().Add(3 * time.Minute)

	newCode := &Code{
		Code:      code,
		Email:     email,
		ExpiresAt: expiresAt,
	}

	res := tx.Create(newCode)
	if res.Error != nil {
		return fmt.Errorf("Kod veri tabanına kaydedilemedi. %w", res.Error)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Tx commit edilemedi. %w", err)
	}

	return nil

}
func (c *CodeHandler) VerifyCode(email string, input string) error {

	var code string
	res := c.Db.Model(&Code{}).Select("codes.code").
		Where("codes.email = ? AND codes.expires_at - CURRENT_TIMESTAMP >= INTERVAL '0 minutes' AND codes.expires_at - CURRENT_TIMESTAMP <= INTERVAL '3 minutes'", email).
		Order("expires_at DESC").
		Scan(&code)

	if res.Error != nil {
		return fmt.Errorf("Sorgu çalıştırılırken hata ile karşılaşıldı. %w", res.Error)
	}

	if input != code {
		return errors.New("Girdiğiniz kod hatalıdır.")
	}

	return nil

}

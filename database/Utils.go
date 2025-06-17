package database

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

func Encrypted(pas []byte) string {
	dest := make([]byte, base64.StdEncoding.EncodedLen(len(pas)))
	base64.StdEncoding.Encode(dest, pas)
	return string(dest)
}
func EncryptedFileName(file *multipart.FileHeader) string {

	timestamp := time.Now().Unix()

	hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%s", timestamp, file.Filename)))

	encodedHash := base64.RawURLEncoding.EncodeToString(hash[:])

	return fmt.Sprintf("%s.%s", []byte(encodedHash), strings.Split(file.Header.Get("Content-Type"), "/")[1])
}
func SaveImagesNames(file *multipart.Form) ([]string, []string) {

	allowedTypes := map[string]bool{"image/jpeg": true, "image/png": true, "image/jpg": true}
	var errors []string
	var imagePaths []string

	for _, fileHeaders := range file.File {
		for _, fileHeader := range fileHeaders {

			if fileHeader.Size > 3*1024*1024 {
				errors = append(errors, fmt.Sprintf("%s dosyası 3 MB sınırını aşıyor.", fileHeader.Filename))
				continue
			}

			if _, exists := allowedTypes[fileHeader.Header.Get("Content-Type")]; !exists {
				errors = append(errors, fmt.Sprintf("%s istenilen formatta değil (sadece JPEG, JPG veya PNG desteklenir).", fileHeader.Filename))
				continue
			}

			imagePaths = append(imagePaths, EncryptedFileName(fileHeader))
		}
	}

	if len(errors) > 0 {
		return nil, errors
	}

	return imagePaths, nil
}
func SaveImages(form *multipart.Form) error {
	for _, fileHeaders := range form.File {
		for _, fileHeader := range fileHeaders {
			if err := SaveImage(fileHeader); err != nil {
				return fmt.Errorf("%s adlı dosya kaydedilemedi: %w", fileHeader.Filename, err)
			}
		}
	}
	return nil
}
func SaveImage(fileHeader *multipart.FileHeader) error {
	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer src.Close()

	filePath := fmt.Sprintf("/app/images/%s", EncryptedFileName(fileHeader))
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("dosya oluşturulamadı: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("dosya yazılamadı: %w", err)
	}

	return nil
}
func ConfigureTables(db *gorm.DB) error {

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	db.Exec("CREATE TYPE role AS ENUM ('admin', 'user')")

	UserHandler := GetUserHandler(db)
	GetQuestionHandler := GetQuestionHandler(db)
	GetCommentHandler := GetCommentHandler(db)
	GetCodeHandler := GetCodeHandler(db)
	GetTagHandler := GetTagHandler(db)
	if err := UserHandler.CreateTable(); err != nil {
		return err
	}
	if err := GetQuestionHandler.CreateTable(); err != nil {
		return err
	}
	if err := GetCommentHandler.CreateTable(); err != nil {
		return err
	}
	if err := GetCodeHandler.CreateTable(); err != nil {
		return err
	}
	if err := GetTagHandler.CreateTable(); err != nil {
		return err
	}

	tx := db.Exec(`CREATE TABLE IF NOT EXISTS questions_tags (
		question_uuid UUID NOT NULL,
		tag_uuid UUID NOT NULL,
		PRIMARY KEY (question_uuid, tag_uuid),
		FOREIGN KEY (question_uuid) REFERENCES questions(uuid) ON DELETE CASCADE,
		FOREIGN KEY (tag_uuid) REFERENCES tags(uuid) ON DELETE CASCADE
	);`)
	if tx.Error != nil {
		return tx.Error
	}
	return nil

}
func GetSixMonths() []string {
	dates := []string{}

	monthRange := 6
	currentTime := time.Now()

	for i := 0; i < monthRange; i++ {
		month := int(currentTime.Month()) - i
		year := currentTime.Year()

		if month <= 0 {
			month += 12
			year -= 1
		}

		dates = append(dates, fmt.Sprintf("%d-%02d", year, month))
	}

	return dates
}

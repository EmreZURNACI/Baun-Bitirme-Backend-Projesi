package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func CheckValidationErr(data any) []string {

	validate := validator.New()

	if err := validate.Struct(data); err != nil {
		validationErrs := err.(validator.ValidationErrors)

		errors := []string{}

		for _, fieldErr := range validationErrs {

			if strings.Split(fieldErr.Error(), ":")[2] == "Field validation for 'Email' failed on the 'email' tag" || fieldErr.Tag() == "uuid" {
				errors = append(errors, fmt.Sprintf("Geçerli bir '%s' giriniz.", fieldErr.Field()))
			} else {
				errors = append(errors, fmt.Sprintf("'%s' alanı boş bırakılamaz.", fieldErr.Field()))
			}

		}
		return errors
	}
	return nil
}


type StandartModel struct {
	Status     bool `json:"status"`
	StatusCode int  `json:"statusCode"`
	Message    any  `json:"message"`
	Data       any  `json:"data,omitempty"`
}

func ResponseModel(c *fiber.Ctx, status bool, statusCode int, message any, data any) error {
	return c.Status(statusCode).JSON(&StandartModel{
		Status:     status,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

func EncryptedFileName(file *multipart.FileHeader) string {
	timestamp := time.Now().Unix()

	encodedTimestamp := base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", timestamp)))

	return fmt.Sprintf("%s.%s", []byte(encodedTimestamp), strings.Split(file.Header.Get("Content-Type"), "/")[1])
}
func CreateToken(mail, user_uuid string) (string, error) {

	claims := jwt.MapClaims{
		"data": map[string]string{"Mail": mail, "User_uuid": user_uuid},
		"exp":  time.Now().Add(time.Hour * 2).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := []byte(os.Getenv("SECRET_KEY"))
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
func ValidateToken(tokenString string) (map[string]string, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		return nil, errors.New("Token bulunamadı")
	}

	if !token.Valid {
		return nil, fmt.Errorf("Geçersiz token")
	}

	dataMap, ok := claims["data"].(map[string]any)
	if !ok {
		return nil, errors.New("Token içinde 'data' bulunamadı")
	}

	result := make(map[string]string)
	for k, v := range dataMap {
		if strVal, ok := v.(string); ok {
			result[k] = strVal
		} else {
			return nil, fmt.Errorf("Veri türü uyumsuz: %v", k)
		}
	}

	return result, nil
}

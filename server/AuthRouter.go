package server

import (
	"bitirme/database"
	"bitirme/mail"
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (r *RouterHandler) AuthServer() {
	userHandler := database.GetUserHandler(r.Db)
	codeHandler := database.GetCodeHandler(r.Db)

	authRouter := r.Server.Group("/auth")

	authRouter.Get("/autologin", func(c *fiber.Ctx) error {
		var token string = c.Cookies("token")
		data, err := ValidateToken(token)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Giriş başarılı.", map[string]database.User{"user": user})
	})
	authRouter.Post("/login-with-tel", func(c *fiber.Ctx) error {

		type User struct {
			Tel      string `json:"tel" validate:"required"`
			Password string `json:"password" validate:"required"`
		}

		var u User
		if err := c.BodyParser(&u); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if errors := CheckValidationErr(u); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		user, err := userHandler.IsUserExists(u.Tel)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		message, err := userHandler.LoginWithTel(u.Tel, u.Password)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		token, err := CreateToken(user.Email, user.Uuid)
		if err != nil {
			return ResponseModel(c, false, 400, "Token yaratılamadı.", nil)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  time.Now().Add(2 * time.Hour),
			HTTPOnly: true,
			Secure:   false,                       // HTTP olduğu için false olmalı
			SameSite: fiber.CookieSameSiteLaxMode, // Lax en güvenli ve uyumlu seçimdir HTTP için
		})

		return ResponseModel(c, true, 200, message, map[string]interface{}{"token": token, "user": user})
	})
	authRouter.Post("/login-with-email", func(c *fiber.Ctx) error {

		type User struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required"`
		}

		var u User
		if err := c.BodyParser(&u); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if errors := CheckValidationErr(u); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		user, err := userHandler.IsUserExists(u.Email)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		message, err := userHandler.LoginWithEmail(u.Email, u.Password)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		token, err := CreateToken(user.Email, user.Uuid)
		if err != nil {
			return ResponseModel(c, false, 400, "Token yaratılamadı.", nil)
		}
		c.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  time.Now().Add(2 * time.Hour),
			HTTPOnly: true,
			Secure:   false,                       // HTTP olduğu için false olmalı
			SameSite: fiber.CookieSameSiteLaxMode, // Lax en güvenli ve uyumlu seçimdir HTTP için
		})

		return ResponseModel(c, true, 200, message, map[string]interface{}{"token": token, "user": user})

	})
	authRouter.Post("/signup", func(c *fiber.Ctx) error {

		var u database.User
		if err := c.BodyParser(&u); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if errors := CheckValidationErr(u); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		if err := userHandler.SignUp(u); err != nil || len(err) != 0 {
			return ResponseModel(c, false, 400, err, nil)
		}

		return ResponseModel(c, true, 200, "Hesabınız başarıyla oluşturuldu", nil)
	})
	authRouter.Post("/forgot_password", func(c *fiber.Ctx) error {

		type user struct {
			Email string `json:"email" validate:"required,email"`
		}
		var u user

		if err := c.BodyParser(&u); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if errors := CheckValidationErr(u); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		if _, err := userHandler.IsUserExists(u.Email); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		code := fmt.Sprintf("%06d", rand.Intn(900000)+100000)

		err := mail.SendEmail(u.Email, code)
		err2 := codeHandler.SaveCode(u.Email, code)

		if err != nil || err2 != nil {
			fmt.Println(err.Error())
			fmt.Println(err2.Error())
			return ResponseModel(c, false, 400, fmt.Sprintf("%s %s", err.Error(), err2.Error()), nil)
		}

		return ResponseModel(c, true, 200, fmt.Sprintf("Mail %s adresine başarıyla gönderildi.", u.Email), nil)
	})
	authRouter.Post("/password_reset", func(c *fiber.Ctx) error {

		type Input struct {
			Code  string `json:"code" validate:"required,max=6,min=6"`
			Email string `json:"email" validate:"required,email"`
		}
		var i Input

		if err := c.BodyParser(&i); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if errors := CheckValidationErr(i); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		if _, err := userHandler.IsUserExists(i.Email); err != nil {
			return ResponseModel(c, false, 400, err, nil)
		}

		if err := codeHandler.VerifyCode(i.Email, i.Code); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Doğrulama kodu onaylanmıştır.", nil)
	})
	authRouter.Post("/password_reset/verify", func(c *fiber.Ctx) error {

		type Input struct {
			Email      string `json:"email" validate:"email,required"`
			Password   string `json:"password" validate:"required"`
			Repassword string `json:"repassword" validate:"required"`
		}
		var i Input
		if err := c.BodyParser(&i); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if errors := CheckValidationErr(i); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		if _, err := userHandler.IsUserExists(i.Email); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		if i.Password != i.Repassword {
			return ResponseModel(c, false, 400, "Parolalar eşleşmiyor.", nil)
		}

		if err := userHandler.ResetPassword(
			i.Email,
			i.Password,
			i.Repassword); err != nil {
			return ResponseModel(c, false, 400, err, nil)
		}

		return ResponseModel(c, true, 200, "Şifreniz başarıyla değişmiştir.", nil)
	})
	authRouter.Post("/logout", func(c *fiber.Ctx) error {

		token := c.Cookies("token")
		if token == "" {
			return ResponseModel(c, false, 400, "Token bulunmamaktadır.", nil)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  time.Now().Add(-2 * time.Hour),
			HTTPOnly: true,
			Secure:   false,                       // HTTP olduğu için false olmalı
			SameSite: fiber.CookieSameSiteLaxMode, // Lax en güvenli ve uyumlu seçimdir HTTP için
		})

		return ResponseModel(c, true, 200, "Hesabınızdan başarıyla çıkış yapıldı.", nil)
	})
}

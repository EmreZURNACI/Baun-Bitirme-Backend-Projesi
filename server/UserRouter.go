package server

import (
	"bitirme/database"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func (r *RouterHandler) UserServer() {

	userHandler := database.GetUserHandler(r.Db)

	userRouter := r.Server.Group("/user")

	userRouter.Get("/", func(c *fiber.Ctx) error {

		users, err := userHandler.GetUsers()
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, strconv.Itoa(len(users))+" adet kullanıcı başarıyla getirildi.", users)
	})
	userRouter.Get("/:id", func(c *fiber.Ctx) error {

		id := c.Params("id")

		type Input struct {
			User_uuid string `json:"user_uuid" validate:"required,uuid"`
		}

		input := Input{User_uuid: id}
		if err := CheckValidationErr(input); err != nil {
			return ResponseModel(c, false, 400, err, nil)
		}

		// Kullanıcıyı getir
		user, err := userHandler.GetUser(id)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Kullanıcı başarıyla getirildi.", user)
	})
	userRouter.Put("/load-avatar", func(c *fiber.Ctx) error {

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, "Yetkisiz Giriş", nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		// file , json olmadıgı için BodyParser ile alınamadıgından form ile almak gerekir.

		errors := []string{}

		file, err := c.FormFile("avatar")
		if err != nil {
			errors = append(errors, "Dosya yüklenemedi veya 'avatar' key'i eksik.")
		}

		if len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		message, err := userHandler.LoadAvatar(user.Uuid, file)
		if err != nil {
			return ResponseModel(c, false, 400, fmt.Sprintf("Avatar yüklenirken hata oluştu: %s", err.Error()), nil)
		}

		return ResponseModel(c, true, 200, message, nil)
	})
	userRouter.Put("/delete-avatar", func(c *fiber.Ctx) error {
		//PUT isteklerinde veri byte dizisi olarak gelmektedir.

		c.Accepts("text/plain", "application/json")

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, "Yetkisiz Giriş", nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		if err := userHandler.DeleteAvatar(user.Uuid); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Avatar başarıyla kaldırıldı.", nil)

	})
	userRouter.Put("/update", func(c *fiber.Ctx) error {

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		type Input struct {
			Name       string `json:"name"`
			Lastname   string `json:"lastname"`
			Nickname   string `json:"nickname"`
			Website    string `json:"website"`
			About      string `json:"about"`
			Password   string `json:"password"`
			Repassword string `json:"repassword"`
			Email      string `json:"email"`
			Tel        string `json:"tel"`
		}

		var i Input
		if err := c.BodyParser(&i); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if i.Password != i.Repassword {
			return ResponseModel(c, false, 400, "Şifreler eşleşmiyor.", nil)
		}

		_user := database.User{
			Uuid:     user.Uuid,
			Name:     i.Name,
			Lastname: i.Lastname,
			Nickname: i.Nickname,
			Website:  i.Website,
			About:    i.About,
			Password: i.Password,
			Email:    i.Email,
			Tel:      i.Tel,
		}

		if err := userHandler.UpdateUser(_user); err != nil || len(err) > 0 {
			return ResponseModel(c, false, 400, err, nil)
		}

		return ResponseModel(c, false, 400, "Kullanıcı verileri başarıyla güncellendi.", nil)
	})
}

package server

import (
	"bitirme/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (r *RouterHandler) AdminServer() {
	adminHandler := database.GetAdminHandler(r.Db)

	userHandler := database.GetUserHandler(r.Db)

	tagHandler := database.GetTagHandler(r.Db)

	adminRouter := r.Server.Group("/admin")

	adminRouter.Use(func(c *fiber.Ctx) error {
		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])

		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		if user.Role != "admin" {
			return ResponseModel(c, false, 400, "Yetkiniz bulunmamaktadır.", nil)
		}

		return c.Next()
	})

	adminRouter.Get("/get-statics", func(c *fiber.Ctx) error {
		statics, err := adminHandler.GetAllStatics()
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}
		return ResponseModel(c, true, 200, "Veriler başarıyla getirildi.", statics)
	})

	adminRouter.Delete("/:id/delete-user", func(c *fiber.Ctx) error {

		uuid := c.Params("id")

		type input struct {
			Uuid string `json:"uuid" validate:"required,uuid"`
		}

		var i input
		i.Uuid = uuid

		if err := c.QueryParser(&i); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if err := CheckValidationErr(i); len(err) > 0 {
			return ResponseModel(c, false, 400, err, nil)
		}

		user, err := userHandler.IsUserExistsByUuid(uuid)
		if err != nil {
			return ResponseModel(c, false, 404, "Bu uuid'ye ait kullanıcı bulunmamaktadır.", nil)
		}

		if err := adminHandler.DeleteUser(user.Uuid); err != nil {
			return ResponseModel(c, false, 404, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, uuid+"'li kullanıcı başarıyla silindi.", nil)
	})

	adminRouter.Post("/add-tag", func(c *fiber.Ctx) error {

		var t database.Tag

		if err := c.BodyParser(&t); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz Body", nil)
		}

		if errors := CheckValidationErr(t); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		err := tagHandler.AddTag(database.Tag{
			Name: t.Name,
		})

		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, t.Name+" tag'ı başarıyla eklendi.", nil)
	})

	adminRouter.Delete("/:id/delete-tag", func(c *fiber.Ctx) error {

		var tagUUID string = c.Params("id")

		if _, err := uuid.Parse(tagUUID); err != nil {
			return ResponseModel(c, false, 400, "Geçerli uuid giriniz.", nil)
		}

		tag, err := tagHandler.IsTagExists(tagUUID)

		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		if err := tagHandler.DeleteTag(tag); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, tag.Name+" başarıyla silindi.", nil)
	})

	adminRouter.Get("/six-months-data", func(c *fiber.Ctx) error {
		datas, err := adminHandler.GetSixMonthsData()
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}
		return ResponseModel(c, true, 200, "6 aylık veriler başarıyla getirildi", datas)

	})

}

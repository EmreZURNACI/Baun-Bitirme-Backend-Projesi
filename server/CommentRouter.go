package server

import (
	"bitirme/database"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
)

func (r *RouterHandler) CommentRouter() {

	commentHandler := database.GetCommentHandler(r.Db)
	questionHandler := database.GetQuestionHandler(r.Db)
	userHandler := database.GetUserHandler(r.Db)

	commentRouter := r.Server.Group("/comment")

	isAuthorized := func(c *fiber.Ctx) error {
		token := c.Cookies("token")
		if _, err := ValidateToken(token); err != nil {
			return ResponseModel(c, false, 401, "Yetkisiz giriş", nil)
		}
		return c.Next()
	}

	commentRouter.Post("/:id/add-comment", isAuthorized, func(c *fiber.Ctx) error {

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
			Question_uuid string          `json:"question_uuid" validate:"required,uuid"`
			Comment       string          `json:"comment" validate:"required"`
			Form          *multipart.Form `json:"form" form:"form" validate:"omitempty"`
		}

		var input Input

		// Dosyaları al
		form, err := c.MultipartForm()
		if err == nil {
			input.Form = form
		}
		input.Question_uuid = c.Params("id")
		input.Comment = c.FormValue("comment")

		if errors := CheckValidationErr(input); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		if _, err := questionHandler.IsQuestionExistByUuid(input.Question_uuid); err != nil {
			return ResponseModel(c, false, 404, err.Error(), nil)
		}

		var fileNames []string
		if input.Form != nil {
			// Dosya isimlerini kaydet
			var errors []string
			fileNames, errors = database.SaveImagesNames(input.Form)
			if errors != nil && len(errors) > 0 {
				return ResponseModel(c, false, 400, errors, nil)
			}
		}

		// Veritabanı işlemi
		tx, err := commentHandler.AddComment(database.Comment{Question_uuid: input.Question_uuid, User_uuid: user.Uuid, Comment: input.Comment, Image: fileNames})
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		if input.Form != nil {
			// Dosyaları kaydet
			if err := database.SaveImages(form); err != nil {
				_ = tx.Rollback() // Hata durumunda rollback çalışmalı
				return ResponseModel(c, false, 400, err.Error(), nil)
			}
		}

		return ResponseModel(c, true, 200, "Yorumunuz başarıyla eklendi.", nil)

	})

	commentRouter.Put("/:id/update-comment", isAuthorized, func(c *fiber.Ctx) error {
		commentID := c.Params("id")

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
			Uuid    string `json:"comment_uuid" validate:"required,uuid"`
			Comment string `json:"comment" validate:"required"`
		}
		var i Input

		if err := c.BodyParser(&i); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		i.Uuid = commentID

		if errors := CheckValidationErr(i); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		comment := database.Comment{
			Uuid:      i.Uuid,
			User_uuid: user.Uuid,
			Comment:   i.Comment,
		}

		if err := commentHandler.UpdateComment(comment, user.Uuid); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Yorumunuz başarıyla güncellendi.", nil)
	})

	commentRouter.Delete("/:id/delete-comment", isAuthorized, func(c *fiber.Ctx) error {
		commentID := c.Params("id")

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
			Comment_uuid string `json:"comment_uuid" validate:"required,uuid"`
		}

		var i Input

		if err := c.BodyParser(&i); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz Body", nil)
		}

		i.Comment_uuid = commentID

		if errors := CheckValidationErr(i); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		if _, err := commentHandler.IsCommentExistByUuid(i.Comment_uuid); err != nil {
			return ResponseModel(c, false, 404, "Yorum bulunamadı", nil)
		}

		if err := commentHandler.DeleteComment(i.Comment_uuid, user.Uuid); err != nil {
			return ResponseModel(c, false, 403, "Yorumu silme yetkiniz yok", nil)
		}

		return ResponseModel(c, true, 200, "Yorumunuz başarıyla silindi.", nil)
	})

	commentRouter.Get("/:id/comments", func(c *fiber.Ctx) error {
		question_uuid := c.Params("id")

		type Input struct {
			Question_uuid string `json:"question_uuid" validate:"required,uuid"`
		}
		i := Input{Question_uuid: question_uuid}

		if errors := CheckValidationErr(i); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		comments, err := commentHandler.GetComments(i.Question_uuid)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}
		return ResponseModel(c, true, 200, "Soruya ait yorum ve bilgileri getirildi", comments)
	})

	commentRouter.Post("/:id/like", isAuthorized, func(c *fiber.Ctx) error {

		var comment_uuid string = c.Params("id")

		_, err := commentHandler.IsCommentExistByUuid(comment_uuid)

		if err != nil {
			return ResponseModel(c, false, 400, "Bu uuid'ye ait yorum bulunmamaktadır.", nil)
		}

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		if err := commentHandler.LikeComment(comment_uuid, user.Uuid); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Yoruma tepkiniz eklendi", nil)
	})

	commentRouter.Post("/:id/dislike", isAuthorized, func(c *fiber.Ctx) error {
		var comment_uuid string = c.Params("id")

		_, err := commentHandler.IsCommentExistByUuid(comment_uuid)

		if err != nil {
			return ResponseModel(c, false, 400, "Bu uuid'ye ait yorum bulunmamaktadır.", nil)
		}

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		if err := commentHandler.DislikeComment(comment_uuid, user.Uuid); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Yoruma tepkiniz eklendi", nil)
	})

	commentRouter.Get("/:id/reaction-count", func(c *fiber.Ctx) error {
		var comment_uuid string = c.Params("id")

		_, err := commentHandler.IsCommentExistByUuid(comment_uuid)

		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		reaction_count, err := commentHandler.ReactionCount(comment_uuid)

		if err != nil || len(reaction_count) != 2 {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, false, 400, "Yorum reaksiyon adetleri başarıyla getirildi.",
			map[string]int64{"like_count": reaction_count[0], "dislike_count": reaction_count[1]})
	})

}

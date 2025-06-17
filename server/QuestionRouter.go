package server

import (
	"bitirme/database"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (r *RouterHandler) QuestionServer() {

	questionHandler := database.GetQuestionHandler(r.Db)
	userHandler := database.GetUserHandler(r.Db)

	questionRouter := r.Server.Group("/question")

	questionRouter.Get("/questions", func(c *fiber.Ctx) error {

		type Cond struct {
			Tags   string `json:"tags"`
			Sort   string `json:"sort"`
			Limit  int32  `json:"limit"`
			Offset int32  `json:"offset"`
		}

		var cond Cond
		if err := c.QueryParser(&cond); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		questions, err := questionHandler.GetQuestions(cond.Tags, cond.Sort, cond.Limit, cond.Offset)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Sorular başarıyla getirildi", questions)

	})

	questionRouter.Get("/:id", func(c *fiber.Ctx) error {

		var params string = c.Params("id")
		type Input struct {
			Uuid string `json:"uuid" validate:"required,uuid"`
		}
		i := Input{Uuid: params}

		if err := CheckValidationErr(i); len(err) > 0 {
			return ResponseModel(c, false, 400, err, nil)
		}

		affectedRows, err := questionHandler.IncreaseQuestionsView(params)
		if affectedRows == 0 || err != nil {
			log.Fatalf("%s id'li sorunun görüntülenme sayısı düzenlenemedi", params)
		}

		questions, err := questionHandler.GetQuestion(params)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Soru başarıyla getirildi.", questions)

	})

	questionRouter.Post("/create", func(c *fiber.Ctx) error {

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, "Yetkisiz Giriş", nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		type Input struct {
			Header  string          `json:"header" form:"header" validate:"required"`
			Content string          `json:"content" form:"content" validate:"required"`
			Tags    []database.Tag  `json:"tags" form:"tags"`
			Form    *multipart.Form `json:"form" form:"form" validate:"omitempty"`
		}

		var input Input

		// Dosyaları al
		form, err := c.MultipartForm()
		if err == nil {
			input.Form = form
		}
		input.Header = c.FormValue("header")
		input.Content = c.FormValue("content")

		var tags []database.Tag
		for _, value := range strings.Split(c.FormValue("tags"), ",") {
			tags = append(tags, database.Tag{Uuid: value})
		}

		if errors := CheckValidationErr(input); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
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

		if err := questionHandler.CreateQuestion(database.Question{
			Header:    input.Header,
			Content:   input.Content,
			Image:     fileNames,
			Tags:      tags,
			User_uuid: user.Uuid,
		}); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		if input.Form != nil {
			// Dosyaları kaydet
			if err := database.SaveImages(form); err != nil {
				return ResponseModel(c, false, 400, err.Error(), nil)
			}
		}

		return ResponseModel(c, true, 200, "Soru başarıyla oluşturuldu.", nil)
	})

	questionRouter.Get("by/:user_id/", func(c *fiber.Ctx) error {

		params := c.Params("user_id")

		type Input struct {
			User_uuid string `json:"uuid" validate:"required,uuid"`
		}
		i := Input{User_uuid: params}

		if err := CheckValidationErr(i); len(err) > 0 {
			return ResponseModel(c, false, 400, err, nil)
		}

		questions, err := questionHandler.GetQuestionsByUser(i.User_uuid)
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, strconv.Itoa(len(questions))+" Adet soru başarıyla getirildi.", questions)
	})

	questionRouter.Put("/:id/update", func(c *fiber.Ctx) error {
		questionID := c.Params("id")

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, "Yetkisiz Giriş", nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		question, err := questionHandler.GetQuestion(questionID)
		if err != nil {
			return ResponseModel(c, false, 400, "Soru bulunamadı: "+err.Error(), nil)
		}

		if !(user.Role == "admin" || question.User_uuid == user.Uuid) {
			return ResponseModel(c, false, 400, "Bu soruyu düzenlemek için yetkin bulunmamaktadır.", nil)
		}

		var q database.Question
		q.Uuid = questionID

		if err := c.BodyParser(&q); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if errors := CheckValidationErr(q); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		if err := questionHandler.UpdateQuestion(q); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Soru başarıyla güncellendi.", nil)
	})

	questionRouter.Delete("/:id/delete", func(c *fiber.Ctx) error {

		params := c.Params("id")

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, "Yetkisiz Giriş", nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		type Input struct {
			Question_uuid string `json:"question_uuid" validate:"required,uuid"`
		}

		var i Input
		i.Question_uuid = params

		if err := c.BodyParser(&i); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz Body Yapısı", nil)
		}

		if err := CheckValidationErr(i); len(err) > 0 {
			return ResponseModel(c, false, 400, err, nil)
		}

		if err := questionHandler.DeleteQuestion(i.Question_uuid, user.Uuid); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Soru başarıyla silindi.", nil)

	})

	questionRouter.Delete("/delete-images", func(c *fiber.Ctx) error {

		var token string = c.Cookies("token")

		data, err := ValidateToken(token)

		if err != nil {
			return ResponseModel(c, false, 400, "Yetkisiz Giriş", nil)
		}

		user, err := userHandler.IsUserExistsByUuid(data["User_uuid"])
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		type Input struct {
			Question_uuid string   `json:"question_uuid" validate:"required,uuid"`
			Image_list    []string `json:"image_list" validate:"required,max=10,min=1"`
		}

		var i Input
		if err := c.BodyParser(&i); err != nil {
			return ResponseModel(c, false, 400, "Geçersiz body yapısı.", nil)
		}

		if errors := CheckValidationErr(i); len(errors) > 0 {
			return ResponseModel(c, false, 400, errors, nil)
		}

		if err := questionHandler.DeleteImages(i.Question_uuid, user.Uuid, i.Image_list); err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "İlgili resimler başarıyla temizlendi.", nil)
	})

}

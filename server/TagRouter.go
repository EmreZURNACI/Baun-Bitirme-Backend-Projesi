package server

import (
	"bitirme/database"

	"github.com/gofiber/fiber/v2"
)

func (r *RouterHandler) TagServer() {

	tagHandler := database.GetTagHandler(r.Db)
	tagRouter := r.Server.Group("/tags")

	tagRouter.Get("/", func(c *fiber.Ctx) error {
		tags, err := tagHandler.GetTags()
		if err != nil {
			return ResponseModel(c, false, 400, err.Error(), nil)
		}

		return ResponseModel(c, true, 200, "Tagler başarıyla getirildi", tags)
	})
}

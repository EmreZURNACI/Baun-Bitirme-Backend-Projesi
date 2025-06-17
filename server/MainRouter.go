package server

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

func Server(db *gorm.DB) {

	server := fiber.New()

	server.Use(cors.New(cors.Config{
	 AllowOrigins:     "http://45.147.46.57",
	 AllowMethods:     "GET,PUT,POST,OPTIONS,DELETE",
	 AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	 AllowCredentials: true,
	}))
	
	// Prometheus middleware'i
	server.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	server.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello Stack Overflow")
	})

	server.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Router handler
	routerHandler := GetRouter(db, server)

	routerHandler.AuthServer()
	routerHandler.CommentRouter()
	routerHandler.UserServer()
	routerHandler.AdminServer()
	routerHandler.QuestionServer()
	routerHandler.TagServer()
	routerHandler.ImageRouter()

	log.Println("Server 8085 portundan ayağa kaldırılıyor.")

	// Server'ı başlat
	if err := server.Listen(":8085"); err != nil {
		log.Println("Server 8085 portundan ayağa kaldırılırken hata ile karşılaşıldı.")
	}
}

type RouterHandler struct {
	Server *fiber.App
	Db     *gorm.DB
}

func GetRouter(db *gorm.DB, server *fiber.App) *RouterHandler {
	return &RouterHandler{
		Server: server,
		Db:     db,
	}
}

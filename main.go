package main

import (
	"bitirme/database"
	"bitirme/server"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load("./.env"); err != nil {
		log.Printf("ENV dosyası yüklenmedi %s", err.Error())
		return
	}
}

func main() {

	db, err := database.Connection()
	if err != nil {
		log.Println(err)
		return
	}

	// if err := database.ConfigureTables(db); err != nil {
	// 	log.Printf("Veri tabanı konfigüre edileMEdi %s", err.Error())
	// 	return
	// }
	// log.Println("Veri tabanı konfigüre edildi")
	log.Printf("Veri tabanı bağlantısı %s portundan sağlandı.\n", os.Getenv("PORT"))

	server.Server(db)

}

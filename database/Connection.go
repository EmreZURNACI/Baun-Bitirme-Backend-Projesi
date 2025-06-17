package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connection() (*gorm.DB, error) {
	var dsn string = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("DBNAME"))
	con, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Veri tabanı bilgileri eksik veya hatalı : %v", err)
	}
	err = con.Ping()
	if err != nil {
		return nil, fmt.Errorf("Veri tabanıyla bağlantı kurulamadı : %v", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: con,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Gorm ile bağlantı kurulamadı : %v", err)
	}
	return db, nil
}

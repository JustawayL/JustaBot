package main

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

//InitDB initialize the database
func InitDB() (err error) {
	db, err := gorm.Open("sqlite3", "SiluetasData.db")

	db.AutoMigrate(&Server{}, &Player{}, &Message{}, Semaphore{})

	var count int
	db.Table("semaphores").Count(&count)

	if count == 0 {
		log.Println("Adding semaphore")
		db, err := ConnectDB()
		if err != nil {
			log.Println(err)
		}
		sf := Semaphore{
			ID: 123456789,
		}
		db.Create(&sf)
	}

	return err
}

//ConnectDB is a handle that retrieves *gorm.DB
func ConnectDB() (dbout *gorm.DB, err error) {
	db, err := gorm.Open("sqlite3", "SiluetasData.db")

	return db, err
}

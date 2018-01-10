package main

import (
	"time"
)

//Server
type Server struct {
	ID        uint `gorm:"primary_key"`
	Name      string
	Players   []Player `gorm:"ForeignKey:ServerID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

//Message
type Message struct {
	ID        string `gorm:"primary_key"`
	URL       string
	FileName  string
	Autor     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Semaphore struct {
	ID     uint `gorm:"primary_key"`
	Player string
}

//Player model definition
type Player struct {
	ID        string `gorm:"primary_key"`
	Name      string
	Score     int  `gorm:"default:0"`
	MaxStreak uint `gorm:"default:0"`
	ServerID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

//ServerConf struct that contains needed information JustaBot
type ServerConf struct {
	Token    string
	Channel  string
	Rol      string
	Admin    string
	FileSize int
	Prefix   string
}

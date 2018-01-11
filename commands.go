package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

type command interface {
	execute(s *discordgo.Session, m *discordgo.MessageCreate)
}

type topCommand struct{}

func (p *topCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	var jugadores []Player
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	db.Order("score desc").Find(&jugadores)
	s.ChannelMessageSend(m.ChannelID, parseTop(jugadores, m.Author.ID))
}

type turnCommand struct{}

func (p *turnCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	var param = strings.Split(m.Content, " ")
	if len(param) == 1 {
		showTurn(s, m)
	}

	if len(param) == 2 && len(m.Mentions) != 0 {

		passTurn(s, m)
	}
}

func showTurn(s *discordgo.Session, m *discordgo.MessageCreate) {
	var sf Semaphore
	var activo Player
	db, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	// Get all matched records
	db.Find(&sf, sfid)
	if sf.Player != "" {
		db.Find(&activo, sf.Player)
		s.ChannelMessageSend(m.ChannelID, "El turno actual pertenece a "+activo.Name+", por favor respetar el turno")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "No se ha asignado turno")
}

func passTurn(s *discordgo.Session, m *discordgo.MessageCreate) {
	//Trim mention to id
	id := m.Mentions[0].ID
	if !isLegal(s, m) {
		s.ChannelMessageSend(m.ChannelID, "Nice try")
		return
	}
	db, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	// Update multiple attributes with `map`, will only update those changed fields
	db.Model(&Semaphore{ID: sfid}).Update("player", id)
	removeFile()
	s.ChannelMessageSend(m.ChannelID, "Has cedido el turno a @"+m.Mentions[0].Username)
	return
}

func getTurn() Player {
	var sf Semaphore
	var activo Player
	db, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	// Get all matched records
	db.Find(&sf, sfid)
	if sf.Player != "" {
		db.Find(&activo, sf.Player)
	}
	return activo
}

type playCommand struct{}

func (p *playCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	// Find the guild for that channel.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find guild.
		return
	}
	//Look for the bot in that guild's current voice states
	for _, vs := range g.VoiceStates {
		if vs.UserID == s.State.User.ID {
			s.ChannelMessageSend(m.ChannelID, "El bot esta en el canal de voz")
			return
		}
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			playSound(s, g.ID, vs.ChannelID)
			if err != nil {
				fmt.Println("Error playing sound:", err)
			}

			return
		}
	}
}

type stopCommand struct{}

func (p *stopCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	// Find the guild for that channel.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find guild.
		return
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			stopSound(s, g.ID, vs.ChannelID)
			if err != nil {
				fmt.Println("Error playing sound:", err)
			}

			return
		}
	}
}

type jugarCommand struct{}

func (p *jugarCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	var autor Player
	var param = strings.Split(m.Content, " ")
	if len(param) == 1 {
		db, err := ConnectDB()
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()
		db.First(&autor, m.Author.ID)
		if autor.ID == m.Author.ID {
			s.ChannelMessageSend(m.ChannelID, "Jugador ya existe")
		} else {
			db.Create(&Player{ID: m.Author.ID, Name: m.Author.Username, Score: 0})
			s.ChannelMessageSend(m.ChannelID, "Jugador registrado, buena suerte")
		}
	}
}

type initCommand struct{}

func (p *initCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Attachments) == 0 && !isPrivateDM(s, m) {
		return
	}
	if !isLegal(s, m) {
		pm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			println(err)
		}
		s.ChannelMessageSend(pm.ID, "Nice try")
		return
	}
	if !ValidateFile(m.Attachments[0].URL) {
		pm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			println(err)
		}
		s.ChannelMessageSend(pm.ID, "Archivo adjunto no valido, por favor revisar")
		return
	}
	names := strings.Split(m.Attachments[0].URL, "/")
	println(names[len(names)-1])
	var sf Semaphore
	db, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	db.First(&sf)
	if sf.Player == m.Author.ID {

		db.Create(&Message{ID: m.ID, URL: m.Attachments[0].URL, FileName: names[len(names)-1], Autor: m.Author.Username, State: 1})
		dlAudio(m.Attachments[0].URL, names[len(names)-1])
		s.ChannelMessageSend(m.ChannelID, "Lijto")
		s.ChannelMessageSend(Config.Channel, "Nuevo Audio <@"+Config.Rol+">")
		return
	}
}

type scoreCommand struct{}

func (p *scoreCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	var parametro Player
	var param = strings.Split(m.Content, " ")

	if len(param) == 2 {
		//Trim mention to id
		id := strings.TrimFunc(param[1], func(c rune) bool {
			return !unicode.IsNumber(c)
		})
		if !isLegal(s, m) {
			s.ChannelMessageSend(m.ChannelID, "Nice try")
			return
		}
		db, err := ConnectDB()
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()
		// Update multiple attributes with `map`, will only update those changed fields
		db.First(&parametro, id)
		db.Model(&Semaphore{ID: sfid}).Update("player", id)
		db.Model(&parametro).UpdateColumns(Player{Score: parametro.Score + 1})
		removeFile()
		s.ChannelMessageSend(m.ChannelID, "Congrats "+param[1]+", has ganado esta ronda, ahora es tu turno!")
	}
}

type rulesCommand struct{}

func (p *rulesCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	pm, err := s.UserChannelCreate(m.Author.ID)
	content, err := ioutil.ReadFile("Comandos.txt")
	if err != nil {
		log.Println(err)
	}
	msg := string(content[:])
	s.ChannelMessageSend(pm.ID, msg)
}

type nabCommand struct{}

func (p *nabCommand) execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	var player Player
	var id string
	if !isLegal(s, m) {
		s.ChannelMessageSend(m.ChannelID, "Nice try")
		return
	}
	if len(m.Mentions) == 0 {
		return
	}
	id = m.Mentions[0].ID
	db, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	// Update multiple attributes with `map`, will only update those changed fields
	db.First(&player, id)
	db.Model(&player).UpdateColumns(Player{Score: player.Score - 1})
}

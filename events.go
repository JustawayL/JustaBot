package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

var T = time.NewTimer(0)

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	//Check if channel is private or DM
	/*channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Println(err)
	}
	*/

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	//Returns false if the request doesn't follow the rules
	isLegal := func(parametro *Player, autor *Player, id string) bool {
		var sf Semaphore
		aux, err := s.User(id)
		db, err := ConnectDB()
		if err != nil {
			log.Println(err)
		}
		defer db.Close()
		db.First(parametro, id)
		db.First(autor, m.Author.ID)
		db.Where("player = ?", m.Author.ID).First(&sf)
		if parametro.ID == "" || autor.ID == "" || aux == nil {
			return false
		}
		if autor.ID == Config.Admin {
			return true
		}
		if autor.ID == parametro.ID || sf.ID == 0 {
			return false
		}
		return true
	}
	/*
		if strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix+"pistas") && channel.Type == 1 {
			var sf Semaphore
			var param = strings.Split(m.Content, " ")
			var lastMsg Message
			if len(param) == 2 {
				db, err := ConnectDB()
				if err != nil {
					fmt.Println(err)
				}
				defer db.Close()
				// Get all matched records
				db.Find(&sf, 123456789)
				db.Last(&lastMsg)
				if sf.Player == m.Author.ID {
					pistas := strings.Split(param[1], ",")
					for i := 0; i < len(pistas); i++ {
						timer1 := time.NewTimer(time.Second * time.Duration(30))
						<-timer1.C
						s.ChannelMessageSend(Config.Channel, "Lanzando pista: "+pistas[i])
						msg, _ := s.ChannelMessage(Config.Channel, lastMsg.ID)
						newMsg := strings.Join([]string{msg.Content, "\n", pistas[i]}, "")
						s.ChannelMessageEdit(Config.Channel, lastMsg.ID, newMsg)
					}
					return
				}

			}

		}
	*/

	//Check if the channel is right
	if m.ChannelID != Config.Channel {
		return
	}
	/*
		// If the message is "ping" reply with "Pong!"
		if m.Content == "ping" {

			s.ChannelMessageSend(m.ChannelID, "Pong!")
		}

		// If the message is "pong" reply with "Ping!"
		if m.Content == "pong" {
			s.ChannelMessageSend(m.ChannelID, "Ping!")
		}
	*/
	if strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix+"timer") {
		var second = strings.Split(m.Content, " ")
		if len(second) != 1 {
			s.ChannelMessageSend(m.ChannelID, "Ok, <@"+m.Author.ID+">, timer por "+second[1]+" minutos!")

			n, err := strconv.Atoi(second[1])
			if err != nil {
				log.Printf("%s", err)
			}

			T = time.NewTimer(time.Minute * time.Duration(n))
			<-T.C
			s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">, el tiempo ha terminado!")
			return
		}

	}

	if strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix+"punto") {
		var autor Player
		var parametro Player
		var lastMsg Message
		var param = strings.Split(m.Content, " ")

		if len(param) == 2 {
			//Trim mention to id
			id := strings.TrimFunc(param[1], func(c rune) bool {
				return !unicode.IsNumber(c)
			})
			if !isLegal(&parametro, &autor, id) {
				s.ChannelMessageSend(m.ChannelID, "Nice try")
				return
			}
			db, err := ConnectDB()
			if err != nil {
				fmt.Println(err)
			}
			defer db.Close()
			// Update multiple attributes with `map`, will only update those changed fields
			db.Model(&Semaphore{ID: 123456789}).Update("player", id)
			db.Model(&parametro).UpdateColumns(Player{Score: parametro.Score + 1})
			db.Last(&lastMsg)
			s.ChannelMessageSend(m.ChannelID, "Congrats "+param[1]+", has ganado esta ronda, ahora es tu turno!")
			s.ChannelMessageUnpin(m.ChannelID, lastMsg.ID)
			T.Stop()
			return
		}

	}

	if strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix+"turno") {
		var sf Semaphore
		var activo Player
		var param = strings.Split(m.Content, " ")
		var autor Player
		var parametro Player
		if len(param) == 1 {
			db, err := ConnectDB()
			if err != nil {
				fmt.Println(err)
			}
			defer db.Close()
			// Get all matched records
			db.Find(&sf, 123456789)
			if sf.Player != "" {
				db.Find(&activo, sf.Player)
				s.ChannelMessageSend(m.ChannelID, "El turno actual pertenece a "+activo.Name+", por favor respetar el turno")
				return
			}

			s.ChannelMessageSend(m.ChannelID, "No se ha asignado turno")

		}

		if len(param) == 2 {
			//Trim mention to id
			id := strings.TrimFunc(param[1], func(c rune) bool {
				return !unicode.IsNumber(c)
			})

			if !isLegal(&parametro, &autor, id) {
				s.ChannelMessageSend(m.ChannelID, "Nice try")
				return
			}
			db, err := ConnectDB()
			if err != nil {
				fmt.Println(err)
			}
			defer db.Close()
			// Update multiple attributes with `map`, will only update those changed fields
			db.Model(&Semaphore{ID: 123456789}).Update("player", id)
			s.ChannelMessageSend(m.ChannelID, "Has cedido el turno a "+param[1])
			return
		}

	}

	if strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix+"top") {
		var jugadores []Player
		var param = strings.Split(m.Content, " ")
		var autor Player
		var posAutor int
		if len(param) == 1 {
			db, err := ConnectDB()
			if err != nil {
				log.Println(err)
			}
			defer db.Close()
			db.Order("score desc").Find(&jugadores)
			i := 0
			msg := ""
			for ; i < len(jugadores); i++ {
				if i < 5 {
					msg = strings.Join([]string{msg, "[", strconv.Itoa(i + 1), "] ", "Jugador: ", jugadores[i].Name, " Puntaje: ", strconv.Itoa(jugadores[i].Score), "\n"}, "")
				}
				if jugadores[i].ID == m.Author.ID {
					autor = jugadores[i]
					posAutor = i + 1
				}

			}
			msg = strings.Join([]string{msg, "------------------------------\n", "[", strconv.Itoa(posAutor), "] ", "Jugador: ", autor.Name, " Puntaje: ", strconv.Itoa(autor.Score)}, "")
			msg = strings.Join([]string{"```", msg, "```"}, "")
			s.ChannelMessageSend(m.ChannelID, msg)
			return
		}

	}

	if strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix+"init") {
		var sf Semaphore

		db, err := ConnectDB()
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()
		db.First(&sf)
		if sf.Player == m.Author.ID {

			db.Create(&Message{ID: m.ID})
			s.ChannelMessageSend(m.ChannelID, "Nueva Silueta <@"+Config.Rol+">")
			s.ChannelMessagePin(m.ChannelID, m.ID)
			return
		}
	}

	if strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix+"jugar") {
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
				return
			} else {
				db.Create(&Player{ID: m.Author.ID, Name: m.Author.Username})
				s.ChannelMessageSend(m.ChannelID, "Jugador creado, buena suerte")
				return
			}
		}

	}

	if strings.HasPrefix(strings.ToLower(m.Content), Config.Prefix+"holi") {
		pm, err := s.UserChannelCreate(m.Author.ID)
		content, err := ioutil.ReadFile("Comandos.txt")
		if err != nil {
			log.Println(err)
		}
		msg := string(content[:])
		s.ChannelMessageSend(pm.ID, msg)
	}

}

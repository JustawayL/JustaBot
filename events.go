package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Register commands
	Channelcommands := map[string]command{
		"top":    &topCommand{},
		"turno":  &turnCommand{},
		"play":   &playCommand{},
		"jugar":  &jugarCommand{},
		"stop":   &stopCommand{},
		"punto":  &scoreCommand{},
		"reglas": &rulesCommand{},
		"nab":    &nabCommand{},
	}

	DmCommands := map[string]command{
		"init": &initCommand{},
	}

	//Check if the channel is right
	if isPrivateDM(s, m) {
		executeCommand(DmCommands, s, m)
	}
	if m.ChannelID != Config.Channel {
		return
	}
	executeCommand(Channelcommands, s, m)

}
func executeCommand(commands map[string]command, s *discordgo.Session, m *discordgo.MessageCreate) {
	var name string
	var param = strings.Split(m.Content, " ")
	if len(param) != 0 {
		name = param[0]
	}
	if command := commands[name]; command == nil {

	} else {
		command.execute(s, m)
	}
}

// parseTop format the list of players and their scores
func parseTop(jugadores []Player, autorID string) string {
	msg := ""
	var autorP Player
	var posAutor int
	for i := 0; i < len(jugadores); i++ {
		if i < 5 {
			msg = strings.Join([]string{msg, "[", strconv.Itoa(i + 1), "] ", "Jugador: ", jugadores[i].Name, " Puntaje: ", strconv.Itoa(jugadores[i].Score), "\n"}, "")
		}
		if jugadores[i].ID == autorID {
			autorP = jugadores[i]
			posAutor = i + 1
		}

	}
	msg = strings.Join([]string{msg, "------------------------------\n", "[", strconv.Itoa(posAutor), "] ", "Jugador: ", autorP.Name, " Puntaje: ", strconv.Itoa(autorP.Score)}, "")
	msg = strings.Join([]string{"```", msg, "```"}, "")

	return msg
}

//Returns false if the request doesn't follow the rules
func isLegal(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	var sf Semaphore
	var jugador, autor Player
	var id string
	if len(m.Mentions) == 1 {
		id = m.Mentions[0].ID
		usuario, err := s.User(id)
		db, err := ConnectDB()
		if err != nil {
			log.Println(err)
		}
		defer db.Close()
		db.First(&jugador, id)
		db.First(&autor, m.Author.ID)
		db.Where("player = ?", m.Author.ID).First(&sf)
		if jugador.ID == "" || autor.ID == "" || usuario == nil {
			return false
		}
		if autor.ID == Config.Admin {
			return true
		}
		if autor.ID == jugador.ID || sf.ID == 0 {
			return false
		}
		return true

	}
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	db.First(&autor, m.Author.ID)
	db.Where("player = ?", m.Author.ID).First(&sf)
	if sf.ID == 0 {
		return false
	}

	return true
}

//dlAudio download the audiofiles to play
func dlAudio(url string, name string) (err error) {
	if !ValidateFile(url) {
		return
	}

	// Create the file
	out, err := os.Create(name)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		// handle error
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

//playSound plays the sound of the actual turn
func playSound(s *discordgo.Session, GuildID string, ChannelID string) {
	var lastMsg Message
	dgv, err := s.ChannelVoiceJoin(GuildID, ChannelID, false, true)
	if err != nil {
		fmt.Println(err)
	}

	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	db.Last(&lastMsg)

	dgvoice.PlayAudioFile(dgv, lastMsg.FileName, make(chan bool))
	defer dgv.Disconnect()
	defer dgv.Close()
}

//stopSound stops and disconnect the bot from the voice channel
func stopSound(s *discordgo.Session, GuildID string, ChannelID string) {
	dgv, err := s.ChannelVoiceJoin(GuildID, ChannelID, false, true)
	if err != nil {
		fmt.Println(err)

	}
	defer dgv.Disconnect()
	defer dgv.Close()
}

func ValidateFile(url string) bool {
	validExt := []string{"mp3", "wav", "opus", "acc"}
	parts := strings.Split(strings.ToLower(url), "/")
	filenameData := strings.Split(parts[len(parts)-1], ".")
	extension := filenameData[len(filenameData)-1]

	resp, err := http.Get(url)
	if err != nil {
		// handle error
		println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if len(body) > Config.FileSize {
		return false
	}

	return index(validExt, extension) >= 0

}

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func isPrivateDM(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Println(err)
	}
	return channel.Type == 1
}

func removeFile() {
	var lastMsg Message

	db, err := ConnectDB()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	// Update multiple attributes with `map`, will only update those changed fields
	db.Last(&lastMsg)
	if lastMsg.State == 0 {
		return
	}
	err = os.Remove(lastMsg.FileName)
	if err != nil {
		println("Couldn't remove the file " + lastMsg.FileName)
	}
	db.Model(&lastMsg).UpdateColumns(Message{State: 0})
}

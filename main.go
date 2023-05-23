package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var Token string
var downloads []string

func userfile(user string) bool {
	filename := "data/users.txt"
	targetString := user

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, targetString) {
			fmt.Println("Match found:", line)
			found = true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return found
}

func download_attachments() {
	for {
		if downloads != nil {
			fmt.Println(downloads[0])
			parts := strings.SplitN(downloads[0], "*", 2)

			at_url := parts[0]
			fp := parts[1]

			file, err := os.Create(fp)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			// Download the attachment using its URL.
			resp, err := http.Get(at_url)
			if err != nil {
				fmt.Println("Error downloading attachment:", err)
				return
			}
			defer resp.Body.Close()

			// Save the attachment contents to the file.
			_, err = io.Copy(file, resp.Body)
			if err != nil {
				fmt.Println("Error saving attachment:", err)
				return
			}

			fmt.Println("Attachment downloaded:", fp)
			fmt.Println()

			if len(downloads) > 1 {
				downloads = downloads[1:]
			} else {
				downloads = nil
			}

		}
	}
}

func main() {

	go download_attachments()

	token_read, err := ioutil.ReadFile("token.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Convert the byte slice to a string.
	Token = string(token_read)

	// Create a new Discord session using the provided bot token.
	bot, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	bot.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	bot.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = bot.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	bot.Close()
}

func rand_16() int {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate a random 16-digit number
	min := 1000000000000000
	max := 9999999999999999
	randomNumber := rand.Intn(max-min+1) + min

	return randomNumber
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if message.Author.ID == session.State.User.ID {
		return
	}

	fmt.Println(message.Content)

	user := message.Author.Username + "//" + message.Author.ID
	user_folder := "data/" + message.Author.Username + "-" + message.Author.ID
	user_file := "data/" + message.Author.Username + "-" + message.Author.ID + "/messages.txt"
	user_folder_file := "data/" + message.Author.Username + "-" + message.Author.ID + "/files"

	if !userfile(user) {
		os.MkdirAll(user_folder, os.ModePerm)
		os.MkdirAll(user_folder_file, os.ModePerm)
		os.Create(user_file)

		file, err := os.OpenFile("data/users.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			// Handle error
			log.Fatal(err)
		}
		defer file.Close()

		_, err = file.WriteString((user + "\n"))
		if err != nil {
			// Handle error
			log.Fatal(err)
		}
	}

	file, err := os.OpenFile(user_file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		// Handle error
		log.Fatal(err)
	}
	defer file.Close()

	text := "############################################" + "\n" + "Time: " + strconv.FormatInt(time.Now().Unix(), 10) + "\n" + "Guild: " + message.GuildID + "\n" + "-----Begin Message-----" + "\n" + message.Content + "\n" + "-----End of Message-----" + "\n"
	_, err = file.WriteString(text)
	if err != nil {
		// Handle error
		log.Fatal(err)
	}

	if len(message.Attachments) != 0 {
		fmt.Println("Attachment found")
		fmt.Println(message.Attachments)
		//session.ChannelMessageSend(message.ChannelID, "Attachment found")

		for _, attachment := range message.Attachments {
			// Create a file to save the attachment.

			guild, err := session.Guild(message.GuildID)
			if err != nil {
				return
			}

			serverName := guild.Name

			fp := user_folder_file + "/" + "-arcly_begin-" + serverName + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "-arcly_end-" + attachment.Filename
			dowat := attachment.URL + "*" + fp
			downloads = append(downloads, dowat)
		}
	}
}

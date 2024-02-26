package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jakic12/spar-discord-bot/scraping"
	"github.com/joho/godotenv"
)

type TWeekday struct {
	Weekday time.Weekday
	Aliases []string
}

const data_path = "data/"

var week_day_translations = []TWeekday{
	{
		Weekday: time.Monday,
		Aliases: []string{
			"pondeljek",
			"pondelk",
			"monday",
		},
	},
	{
		Weekday: time.Tuesday,
		Aliases: []string{
			"torek",
			"tork",
			"tuesday",
		},
	},
	{
		Weekday: time.Wednesday,
		Aliases: []string{
			"sreda",
			"sredo",
			"srjda",
			"srjdo",
			"wednesday",
		},
	},
	{
		Weekday: time.Thursday,
		Aliases: []string{
			"cetrtek",
			"četrtek",
			"thursday",
		},
	},
	{
		Weekday: time.Friday,
		Aliases: []string{
			"petek",
			"petk",
			"friday",
		},
	},
	{
		Weekday: time.Saturday,
		Aliases: []string{
			"sobota",
			"soboto",
			"saturday",
		},
	},
}

func main() {
	godotenv.Load()

	year, week := time.Now().Local().ISOWeek()

	scraping.GetSubdividedMenus(data_path, year, week)

	api_key := os.Getenv("API_KEY")
	dg, err := discordgo.New("Bot " + api_key)
	if err != nil {
		panic(err)
	}

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the message contains the keyword
	if strings.Contains(strings.ToLower(m.Content), "spar") {
		date := time.Now().Local()
		this_week := true
		date_str := "today"

		if strings.Contains(strings.ToLower(m.Content), "drugi teden") || strings.Contains(strings.ToLower(m.Content), "next week") || strings.Contains(strings.ToLower(m.Content), "drug teden") || strings.Contains(strings.ToLower(m.Content), "drugi tedn") || strings.Contains(strings.ToLower(m.Content), "drug tedn") {
			date = date.AddDate(0, 0, 8-int(date.Weekday()))
			date_str = ""
			this_week = false
		}

		year, week_idx := date.ISOWeek()
		scraping.GetSubdividedMenus(data_path, year, week_idx)

		if this_week && (strings.Contains(strings.ToLower(m.Content), "tomorrow") || strings.Contains(strings.ToLower(m.Content), "jutri") || strings.Contains(strings.ToLower(m.Content), "jutre")) {
			date = date.AddDate(0, 0, 1)
			date_str = "tomorrow"
		} else {
			for _, tWeek := range week_day_translations {
				if strContainsOne(strings.ToLower(m.Content), tWeek.Aliases) {
					date_str = tWeek.Weekday.String()
					date = date.AddDate(0, 0, (int(tWeek.Weekday) - int(date.Weekday())))
					break
				}
			}
		}

		// Read the image file
		imagePath := scraping.GetImagePathFromDate(data_path, date)
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			fmt.Println("Error reading image file: ", err)
			return
		}

		// Encode the image data to base64
		//encodedImage := base64.StdEncoding.EncodeToString(imageData)

		//fmt.Println(encodedImage)

		suffix := ""
		if !this_week {
			if len(date_str) == 0 {
				date_str = "next week"
			} else {
				suffix = " next week"
			}
		}

		// Send the image as a message attachment
		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content: fmt.Sprintf("Since you asked, here is %ss menu%s:", date_str, suffix),
			Files: []*discordgo.File{
				{
					Name:   "menu.jpg",
					Reader: bytes.NewReader(imageData),
				},
			},
		})
		if err != nil {
			fmt.Println("Error sending message: ", err)
		}
	}
}

func strContainsOne(str string, strs []string) bool {
	for _, s := range strs {
		if strings.Contains(str, s) {
			return true
		}
	}
	return false
}

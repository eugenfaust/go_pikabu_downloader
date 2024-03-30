package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}
}

func echo(b *gotgbot.Bot, ctx *ext.Context) error {
	if !strings.Contains(ctx.Message.Text, "https://pikabu.ru/") {
		return nil
	}
	startTime := time.Now()
	re, _ := regexp.Compile("[0-9]+")	
	linkID := re.FindString(ctx.Message.Text)
	if linkID == "" {
		b.SendMessage(ctx.Message.Chat.Id, "Неверный формат ссылки или такого поста не существует", &gotgbot.SendMessageOpts{})
		return nil
	}
	pikabuID, err := strconv.Atoi(linkID)
	if err != nil {
		b.SendMessage(ctx.Message.Chat.Id, "Неверный формат ссылки или такого поста не существует", &gotgbot.SendMessageOpts{})
		return nil
	}
	initMsg, err := b.SendMessage(ctx.Message.Chat.Id, "Начинаю загрузку...", &gotgbot.SendMessageOpts{})
	if ctx.Message.Chat.Id != ctx.Message.From.Id {
		b.DeleteMessage(ctx.Message.Chat.Id, ctx.Message.MessageId, &gotgbot.DeleteMessageOpts{})
	}
	if err != nil {
		return fmt.Errorf("failed to send msg: %w", err)
	}
	captionText := "<a href=\"%v\">%v</a>"
	var pikabuStory Story
	contentExist := true
	err = db.Model(&Story{}).Preload("Images").Preload("Videos").First(&pikabuStory, "pikabu_id = ?", pikabuID).Error
	if err != nil {
		contentExist = false
		pikabuStory, err = parsePage(ctx.Message.Text, pikabuID)
		if err != nil {
			b.SendMessage(ctx.Message.Chat.Id, "Бот временно неспособен обработать данный пост. Попробуйте немного позже...", &gotgbot.SendMessageOpts{})
			return fmt.Errorf("failed to handle pikabu Url: %w", err)
		}
	}
	err = sendImages(captionText, &pikabuStory, b, ctx)
	if err != nil {
		return fmt.Errorf("failed to echo message: %w", err)
	}
	err = sendVideos(captionText, pikabuStory, b, ctx)
	if err != nil {
		return fmt.Errorf("failed to echo message: %w", err)
	}
	if !contentExist {
		db.Create(&pikabuStory)
	}
	b.SendMessage(ctx.Message.Chat.Id, fmt.Sprintf("Загрузка заняла %.5v ms", time.Since(startTime)), &gotgbot.SendMessageOpts{})
	_, err = initMsg.Delete(b, &gotgbot.DeleteMessageOpts{})
	return err
}

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		panic("Database DSN environment variable is empty")
	}
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database")
	}
	db.AutoMigrate(&Story{})
	db.AutoMigrate(&Video{})
	db.AutoMigrate(&Image{})
	token := os.Getenv("TOKEN")
	if token == "" {
		panic("TOKEN environment variable is empty")
	}
	b, err := gotgbot.NewBot(token,
		&gotgbot.BotOpts{
			BotClient: &gotgbot.BaseBotClient{
				Client: http.Client{},
				DefaultRequestOpts: &gotgbot.RequestOpts{
					Timeout: gotgbot.DefaultTimeout,
					APIURL:  gotgbot.DefaultAPIURL,
				},
			},
		})
	if err != nil {
		panic("Failed to create bot: " + err.Error())
	}
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)
	// Add echo handler to reply to all text messages.
	dispatcher.AddHandler(handlers.NewMessage(message.Text, echo))

	// Start receiving updates.
	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 60,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", b.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

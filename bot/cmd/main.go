package main

import (
	"database/sql"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/api"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	"github.com/BulizhnikGames/subbot/bot/internal/commands"
	"github.com/BulizhnikGames/subbot/bot/internal/commands/middleware"
	"github.com/BulizhnikGames/subbot/bot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"log"
	"time"
)

// TODO: add /help
// TODO: improve error handling
// TODO: rate limit users
// TODO: improve error handling from fetcher's API

func main() {
	config.Load()
	cfg := config.Get()

	middleware.Init()

	tgBotApi, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Error initializing bot: %v", err)
	}
	tgBotApi.Debug = false

	dbConn, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer dbConn.Close()

	dbOrm := orm.New(dbConn)

	Bot := bot.Init(tgBotApi, dbOrm, 4*time.Second, 30*time.Minute)

	Bot.RegisterCommand(
		"list",
		middleware.GroupOnly(
			commands.List(dbOrm)),
	)
	Bot.RegisterCommand(
		"sub",
		middleware.GroupOnly(
			middleware.AdminOnly(
				commands.SubNext(dbOrm))),
	)
	Bot.RegisterCommand(
		"del",
		middleware.GroupOnly(
			middleware.AdminOnly(
				commands.DelNext(dbOrm))),
	)
	Bot.RegisterCommand(
		"",
		middleware.GetUsersNext(),
	)

	Bot.RegisterCallback(
		"del",
		middleware.GetUsersNext(),
	)

	go func() {
		Bot.Run()
	}()

	botApi := api.Init(dbOrm, cfg.Port)
	err = botApi.Run()
	if err != nil {
		log.Fatalf("Error running bot's api: %v", err)
	}
}

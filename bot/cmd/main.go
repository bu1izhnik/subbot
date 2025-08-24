package main

import (
	"database/sql"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/api"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	"github.com/BulizhnikGames/subbot/bot/internal/commands"
	"github.com/BulizhnikGames/subbot/bot/internal/commands/middleware"
	"github.com/BulizhnikGames/subbot/bot/internal/config"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

// TODO: improve error handling
// TODO: handle photo edits and edits on posts with reply markup

func main() {
	config.Load()
	cfg := config.Get()

	middleware.Init()

	tgBotApi, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Error initializing bot: %v", err)
	}
	//tgBotApi.Debug = true

	dbConn, err := sql.Open("postgres", cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer dbConn.Close()

	dbOrm := orm.New(dbConn)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Url,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DBid,
	})

	Bot := bot.Init(
		tgBotApi,
		dbOrm,
		4*time.Second,
		cfg.RateLimitConfig,
		2*time.Second,
	)

	botRateLimit := func(next tools.Command) tools.Command {
		return middleware.CheckRateLimit(Bot, next)
	}
	rateLimitAndAdmin := middleware.CreateStack(botRateLimit, middleware.AdminOnly)

	Bot.RegisterCommand(
		"list",
		botRateLimit(
			middleware.GroupOnly(
				commands.List(dbOrm),
			),
		),
	)
	Bot.RegisterCommand(
		"sub",
		rateLimitAndAdmin(commands.SubInit(redisClient)),
	)
	Bot.RegisterCommand(
		"del",
		rateLimitAndAdmin(commands.DelInit(dbOrm)),
	)
	Bot.RegisterCommand(
		"help",
		botRateLimit(commands.Help),
	)
	Bot.RegisterCommand(
		"start",
		botRateLimit(commands.Start),
	)
	Bot.RegisterCommand(
		"",
		middleware.GetUsersNext(Bot, redisClient),
	)

	Bot.RegisterCallback(
		"del",
		botRateLimit(commands.Del(dbOrm)),
	)

	middleware.RegisterCommand(
		"sub",
		botRateLimit(commands.Sub(dbOrm)),
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

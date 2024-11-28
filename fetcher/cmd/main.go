package main

import (
	"github.com/BulizhnikGames/subbot/fetcher/internal/api"
	"github.com/BulizhnikGames/subbot/fetcher/internal/config"
	"github.com/BulizhnikGames/subbot/fetcher/internal/fetcher"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

func main() {
	config.Load()
	cfg := config.Get()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Url,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DBid,
	})

	f, err := fetcher.Init(redisClient, cfg.APIID, cfg.APIHash, cfg.BotUsername, 2*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err = f.Run(cfg.Phone, cfg.Password, cfg.APIURL, cfg.IP, cfg.Port)
		log.Fatalf("Error running fetcher: %v", err)
	}()

	fetcherApi := api.Init(f, cfg.Port)
	err = fetcherApi.Run()
	if err != nil {
		log.Fatalf("Error running fetcher's api: %v", err)
	}
}

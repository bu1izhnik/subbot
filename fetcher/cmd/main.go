package main

import (
	"github.com/BulizhnikGames/subbot/fetcher/internal/api"
	"github.com/BulizhnikGames/subbot/fetcher/internal/config"
	"github.com/BulizhnikGames/subbot/fetcher/internal/fetcher"
	"log"
)

func main() {
	config.Load()
	cfg := config.Get()

	f, err := fetcher.Init(cfg.APIID, cfg.APIHash)
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

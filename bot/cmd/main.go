package main

import "github.com/BulizhnikGames/subbot/bot/internal/config"

func main() {
	config.Load()
	cfg := config.Get()
}

package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

const SUB_LIMIT = 10

type Config struct {
	BotToken string
	DBURL    string
	Port     string
}

func Load() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
}

func Get() Config {
	c := Config{}

	BotToken := os.Getenv("BOT_TOKEN")
	if BotToken == "" {
		log.Fatal("Bot token not found in .env")
	}

	c.DBURL = os.Getenv("DB_URL")
	if c.DBURL == "" {
		log.Fatal("DB URL not found in .env")
	}

	c.Port = os.Getenv("Port")
	if c.Port == "" {
		log.Fatal("Port not found in .env")
	}

	return c
}

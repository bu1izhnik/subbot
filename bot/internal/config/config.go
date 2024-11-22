package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

const SubLimit = 6

type RateLimitConfig struct {
	// Specified in seconds
	RateLimitTime int64
	// Specified in seconds
	RateLimitCheckInterval int64
	RateLimitMaxMessages   int64
}

type Config struct {
	BotToken string
	DBURL    string
	Port     string
	RateLimitConfig
}

func Load() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
}

func Get() Config {
	c := Config{}

	c.BotToken = os.Getenv("BOT_TOKEN")
	if c.BotToken == "" {
		log.Fatal("Bot token not found in .env")
	}

	c.DBURL = os.Getenv("DB_URL")
	if c.DBURL == "" {
		log.Fatal("DB URL not found in .env")
	}

	c.Port = os.Getenv("PORT")
	if c.Port == "" {
		log.Fatal("Port not found in .env")
	}

	rateLimitTimeStr := os.Getenv("RATE_LIMIT_TIME")
	var err error
	c.RateLimitTime, err = strconv.ParseInt(rateLimitTimeStr, 10, 64)
	if err != nil {
		log.Fatal("Rate limit time not found in .env")
	}

	rateLimitCheckIntervalStr := os.Getenv("RATE_LIMIT_CHECK_INTERVAL")
	c.RateLimitCheckInterval, err = strconv.ParseInt(rateLimitCheckIntervalStr, 10, 64)
	if err != nil {
		log.Fatal("Rate limit check interval not found in .env")
	}

	rateLimitMaxMessagesStr := os.Getenv("RATE_LIMIT_MAX_MESSAGES")
	c.RateLimitMaxMessages, err = strconv.ParseInt(rateLimitMaxMessagesStr, 10, 64)
	if err != nil {
		log.Fatal("Rate limit max messages not found in .env")
	}

	return c
}

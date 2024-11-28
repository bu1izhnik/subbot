package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type RedisConfig struct {
	Url      string
	DBid     int
	Username string
	Password string
}

type Config struct {
	BotUsername string
	APIURL      string
	IP          string
	Port        string
	Phone       string
	Password    string
	APIID       int
	APIHash     string
	Redis       RedisConfig
}

func Load() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
}

func Get() Config {
	c := Config{}
	var err error

	c.BotUsername = os.Getenv("BOT_USERNAME")
	if c.BotUsername == "" {
		log.Fatal("BOT_USERNAME not found in .env")
	}

	c.Redis.Url = os.Getenv("REDIS_URL")
	if c.Redis.Url == "" {
		log.Fatal("Redis url not found in .env")
	}

	dbIDStr := os.Getenv("REDIS_DB_ID")
	if dbIDStr == "" {
		log.Fatal("Redis db id not found in .env")
	}
	if c.Redis.DBid, err = strconv.Atoi(dbIDStr); err != nil {
		log.Fatalf("Error parsing redis db id to int: %v", err)
	}

	c.Redis.Username = os.Getenv("REDIS_USERNAME")
	c.Redis.Password = os.Getenv("REDIS_PASSWORD")

	c.APIURL = os.Getenv("API_URL")
	if c.APIURL == "" {
		log.Fatal("API URL not found in .env")
	}

	c.IP = os.Getenv("IP")
	if c.IP == "" {
		log.Fatal("IP not found in .env")
	}

	c.Port = os.Getenv("PORT")
	if c.Port == "" {
		log.Fatal("Port not found in .env")
	}

	c.Phone = os.Getenv("PHONE")
	if c.Phone == "" {
		log.Fatal("Phone not found in .env")
	}

	c.Password = os.Getenv("PASSWORD")

	apiIDStr := os.Getenv("API_ID")
	if apiIDStr == "" {
		log.Fatal("API ID not found in .env")
	}
	if c.APIID, err = strconv.Atoi(apiIDStr); err != nil {
		log.Fatalf("Error parsing API ID to int: %v", err)
	}

	c.APIHash = os.Getenv("API_HASH")
	if c.APIHash == "" {
		log.Fatal("API hash not found in .env")
	}

	return c
}

package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
	"strconv"
)

type (
	Config struct {
		HttpPort  int
		LomsApi   string
		RedisAddr string
		RedisTTL  int
	}
)

func NewConfig() Config {
	loadEnv()

	ttl, err := strconv.Atoi(os.Getenv("CART_REDIS_TTL_SEC"))

	if err != nil {
		log.Fatal("Failed to parse .env file: " + err.Error())
	}

	return Config{
		HttpPort:  parsePort("CART_APP_HTTP_PORT"),
		LomsApi:   os.Getenv("CART_APP_LOMS_API"),
		RedisAddr: os.Getenv("CART_REDIS_ADDR"),
		RedisTTL:  ttl,
	}
}

const projectDirName = "cart"

// https://github.com/joho/godotenv/issues/43
func loadEnv() {
	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal("Failed to parse .env file: " + err.Error())
	}

	rootPath := re.Find([]byte(cwd))
	err = godotenv.Load(string(rootPath) + `/.env`)

	if err != nil {
		log.Fatal("Failed to parse .env file: " + err.Error())
	}
}

func parsePort(flagName string) int {
	port, err := strconv.Atoi(os.Getenv(flagName))

	if err != nil {
		log.Fatal("Failed to parse " + flagName)
	}

	return port
}

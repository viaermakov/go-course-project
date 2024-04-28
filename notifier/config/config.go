package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
)

type (
	Config struct {
		App      AppConfig
		Kafka    KafkaConfig
		Consumer ConsumerConfig
	}

	AppConfig struct {
	}
	KafkaConfig struct {
		Brokers []string
	}
	ConsumerConfig struct {
		Topic     string
		GroupName string
	}
)

func NewConfig() Config {
	loadEnv()

	return Config{
		App: AppConfig{},
		Kafka: KafkaConfig{
			Brokers: []string{
				os.Getenv("KAFKA_BOOTSTRAP_SERVER"),
			},
		},
		Consumer: ConsumerConfig{
			Topic:     os.Getenv("KAFKA_TOPIC"),
			GroupName: os.Getenv("KAFKA_CONSUMER_GROUP_NAME"),
		},
	}
}

const projectDirName = "notifier"

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

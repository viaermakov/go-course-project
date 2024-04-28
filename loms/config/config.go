package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
	"route256.ozon.ru/project/loms/internals/infra/kafka"
	"strconv"
)

type (
	Config struct {
		App      AppConfig
		Kafka    kafka.Config
		Producer ProducerConfig
	}
	DbConnection struct {
		Primary   string
		Secondary []string
	}
	AppConfig struct {
		DbConnections   []DbConnection
		StocksDbConnStr string
		TestDbUser      string
		TestDbPassword  string
		TestDbDatabase  string
		HttpPort        int
		GrpcPort        int
	}
	ProducerConfig struct {
		Topic       string
		IntervalSec int
	}
)

func NewConfig() Config {
	loadEnv()

	return Config{
		App: AppConfig{
			HttpPort:        parsePort("LOMS_APP_HTTP_PORT"),
			GrpcPort:        parsePort("LOMS_APP_GRPC_PORT"),
			StocksDbConnStr: os.Getenv("STOCKS_POSTGRES_URL"),
			DbConnections:   parseDbConnsStr(),
			TestDbUser:      os.Getenv("POSTGRES_TEST_USER"),
			TestDbPassword:  os.Getenv("POSTGRES_TEST_PASSWORD"),
			TestDbDatabase:  os.Getenv("POSTGRESQL_DATABASE"),
		},
		Kafka: kafka.Config{
			Brokers: []string{
				os.Getenv("KAFKA_BOOTSTRAP_SERVER"),
			},
		},
		Producer: ProducerConfig{
			Topic:       os.Getenv("KAFKA_TOPIC"),
			IntervalSec: 2,
		},
	}
}

const projectDirName = "loms"

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

func parseDbConnsStr() []DbConnection {
	return []DbConnection{
		{
			Primary:   os.Getenv("POSTGRES_URL_1"),
			Secondary: []string{},
		},
		{
			Primary:   os.Getenv("POSTGRES_URL_2"),
			Secondary: []string{},
		},
	}
}

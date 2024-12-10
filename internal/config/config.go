package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
}

func NewConfig() (*Config, error) {
	envPath := filepath.Join(".env")

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		log.Fatalf(".env file is missing in the project root")
	}

	err := godotenv.Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file: %v", err)
	}

	cfg := &Config{
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
	}

	return cfg, nil
}

func (c *Config) GetPostgresConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDB)
}

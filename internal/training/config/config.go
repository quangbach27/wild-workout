package config

import (
	"fmt"
	"os"
)

type Config struct {
	App      App
	Database Database
}

type App struct {
	HTTPAddress        string
	TrainerGRPCAddress string
	UserGRPCAddress    string
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	DSN      string
}

func New() *Config {
	return &Config{
		App:      newApp(),
		Database: newDatabase(),
	}
}

func newApp() App {
	return App{
		HTTPAddress:        getEnv("HTTP_ADDRESS", ":4000"),
		TrainerGRPCAddress: getEnv("TRAINER_GRPC_ADDRESS", "trainer:4001"),
		UserGRPCAddress:    getEnv("USER_GRPC_ADDRESS", "user:4001"),
	}
}

func newDatabase() Database {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "user")
	password := getEnv("DB_PASSWORD", "password")
	name := getEnv("DB_NAME", "wild-workout")
	sslMode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, name, sslMode,
	)

	return Database{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
		SSLMode:  sslMode,
		DSN:      dsn,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

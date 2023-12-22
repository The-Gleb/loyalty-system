package main

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type Config struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	LogLevel       string
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	SignKey        string
}

type ConfigBuilder struct {
	config Config
}

func (b ConfigBuilder) SetRunAddres(address string) ConfigBuilder {
	b.config.RunAddress = address
	return b
}
func (b ConfigBuilder) SetLogLevel(level string) ConfigBuilder {
	b.config.LogLevel = level
	return b
}
func (b ConfigBuilder) SetAccrualAddress(address string) ConfigBuilder {
	b.config.AccrualAddress = address
	return b
}
func (b ConfigBuilder) SetDatabaseURI(URI string) ConfigBuilder {
	b.config.DatabaseURI = URI
	return b
}
func (b ConfigBuilder) SetSignKey(key string) ConfigBuilder {
	b.config.SignKey = key
	return b
}

func NewConfigFromFlags() Config {
	flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var runAddress string
	flag.StringVar(&runAddress, "a", ":8080", "address and port to run server")

	var accrualAddress string
	flag.StringVar(&accrualAddress, "f", "/tmp/metrics-db.json", "path to file to store metrics")

	var databaseURI string
	flag.StringVar(&databaseURI, "d", "",
		"info to connect to database, host=host port=port user=myuser password=xxxx dbname=mydb sslmode=disable",
	)

	var logLevel string
	flag.StringVar(&logLevel, "l", "info", "log level")

	var key string
	flag.StringVar(&key, "k", "secret_key", "key for signing")

	flag.Parse()

	var builder ConfigBuilder
	log.Printf("ENV ADDRESS %v", os.Getenv("ADDRESS"))

	builder = builder.SetRunAddres(runAddress).
		SetLogLevel(logLevel).
		SetAccrualAddress(accrualAddress).
		SetDatabaseURI(databaseURI).
		SetSignKey(key)

	env.Parse(&builder.config)

	return builder.config
}

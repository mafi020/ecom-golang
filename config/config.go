package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		ENV      string `mapstructure:"ENV"`
		HTTPPort string `mapstructure:"HTTP_PORT"`
		GRPCPort string `mapstructure:"GRPC_PORT"`
	} `mapstructure:",squash"` // Allows it to find PORT at the top level of .env

	Postgres struct {
		Host     string `mapstructure:"PG_HOST"`
		Port     int    `mapstructure:"PG_PORT"`
		User     string `mapstructure:"PG_USER"`
		Password string `mapstructure:"PG_PASSWORD"`
		DBName   string `mapstructure:"PG_DB_NAME"`
	} `mapstructure:",squash"`

	JWT struct {
		Secret            string `mapstructure:"JWT_SECRET"`
		Expiration        string `mapstructure:"JWT_EXPIRATION"`
		RefreshExpiration string `mapstructure:"JWT_REFRESH_EXPIRATION"`
	} `mapstructure:",squash"`
}

func LoadConfig() *Config {
	// Point directly to your .env file
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	var cfg Config
	// Unmarshal uses the mapstructure tags
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}

	return &cfg
}

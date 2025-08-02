package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the full application configuration loaded from environment or .env file.
type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	JWT      JWTConfig
	Mongo    MongoConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port string
}

// PostgresConfig defines connection details for PostgresSQL.
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// MongoConfig defines connection details for MongoDB.
type MongoConfig struct {
	URI  string
	Name string
}

// JWTConfig contains secrets and token settings for JWT authentication.
type JWTConfig struct {
	Secret         string
	ExpirationHour int
	Issuer         string
}

// AppConfig is the globally accessible configuration instance.
var AppConfig *Config

// LoadConfig reads configuration from .env file or environment variables
// and populates the global AppConfig. It also applies default values.
func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("POSTGRES_SSLMODE", "disable")
	viper.SetDefault("JWT_EXPIRATION_HOUR", 24)

	if err := viper.ReadInConfig(); err == nil {
		log.Println("Loaded configuration from .env")
	} else {
		log.Println(".env not found. Falling back to environment variables")
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Postgres: PostgresConfig{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetString("POSTGRES_PORT"),
			User:     viper.GetString("POSTGRES_USER"),
			Password: viper.GetString("POSTGRES_PASSWORD"),
			DBName:   viper.GetString("POSTGRES_DB"),
			SSLMode:  viper.GetString("POSTGRES_SSLMODE"),
		},
		Mongo: MongoConfig{
			URI:  viper.GetString("MONGO_URI"),
			Name: viper.GetString("MONGO_DB"),
		},
		JWT: JWTConfig{
			Secret:         viper.GetString("JWT_SECRET"),
			ExpirationHour: viper.GetInt("JWT_EXPIRATION_HOUR"),
			Issuer:         viper.GetString("JWT_ISSUER"),
		},
	}

	validateConfig()
}

// validateConfig ensures required variables are present and logs warnings if not.
func validateConfig() {
	var missing []string

	if AppConfig.Server.Port == "" {
		missing = append(missing, "SERVER_PORT")
	}
	if AppConfig.Postgres.Host == "" {
		missing = append(missing, "POSTGRES_HOST")
	}
	if AppConfig.JWT.Secret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		log.Printf("Missing required environment variables: %v\n", missing)
	}
}

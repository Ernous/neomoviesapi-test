package config

import (
	"log"
	"os"
)

type Config struct {
	MongoURI        string
	MongoDBName     string
	TMDBAccessToken string
	JWTSecret       string
	Port            string
	BaseURL         string
	NodeEnv         string
	GmailUser       string
	GmailPassword   string
	LumexURL        string
	AllohaToken     string
	RedAPIBaseURL   string
	RedAPIKey       string
	GoogleClientID  string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendURL        string
}

func New() *Config {
	mongoURI := getMongoURI()

	return &Config{
		MongoURI:        mongoURI,
		MongoDBName:     getEnv(EnvMongoDBName, DefaultMongoDBName),
		TMDBAccessToken: getEnv(EnvTMDBAccessToken, ""),
		JWTSecret:       getEnv(EnvJWTSecret, DefaultJWTSecret),
		Port:            getEnv(EnvPort, DefaultPort),
		BaseURL:         getEnv(EnvBaseURL, DefaultBaseURL),
		NodeEnv:         getEnv(EnvNodeEnv, DefaultNodeEnv),
		GmailUser:       getEnv(EnvGmailUser, ""),
		GmailPassword:   getEnv(EnvGmailPassword, ""),
		LumexURL:        getEnv(EnvLumexURL, ""),
		AllohaToken:     getEnv(EnvAllohaToken, ""),
		RedAPIBaseURL:   getEnv(EnvRedAPIBaseURL, DefaultRedAPIBase),
		RedAPIKey:       getEnv(EnvRedAPIKey, ""),
		GoogleClientID:  getEnv(EnvGoogleClientID, ""),
		GoogleClientSecret: getEnv(EnvGoogleClientSecret, ""),
		GoogleRedirectURL:  getEnv(EnvGoogleRedirectURL, ""),
		FrontendURL:        getEnv(EnvFrontendURL, ""),
	}
}

func getMongoURI() string {
	for _, envVar := range []string{"MONGO_URI", "MONGODB_URI", "DATABASE_URL", "MONGO_URL"} {
		if value := os.Getenv(envVar); value != "" {
			log.Printf("DEBUG: Using %s for MongoDB connection", envVar)
			return value
		}
	}
	log.Printf("DEBUG: No MongoDB URI environment variable found")
	return ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
package config

import "os"

type Config struct {
	DBPath string
	Port   string
}

func Load() Config {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./candidates.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{DBPath: dbPath, Port: port}
}

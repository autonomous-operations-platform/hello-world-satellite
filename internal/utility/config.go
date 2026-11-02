package utility

import (
	"log/slog"
	"os"
)

type Config struct {
	DorisAddr     string
	DorisUser     string
	DorisPassword string
	DorisDB       string
	OtlpAddr      string
	LogFilePath   string
	LogLevel      slog.Level
}

func LoadConfig() (Config, error) {
	return Config{
		DorisAddr:     getenv("DORIS_ADDR", "localhost:9030"),
		DorisUser:     getenv("DORIS_USER", "root"),
		DorisPassword: getenv("DORIS_PASSWORD", ""),
		DorisDB:       getenv("DORIS_DB", "aos"),
		OtlpAddr:      getenv("OTLP_ADDR", ":8080"),
		LogFilePath:   getenv("LOG_FILE_PATH", "satellite.log"),
		LogLevel:      parseLevel(getenv("LOG_LEVEL", "info")),
	}, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

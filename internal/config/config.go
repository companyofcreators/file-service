package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddress   string        `env:"HTTP_ADDRESS" env-default:":8086"`
	DBDSN         string        `env:"DB_DSN" env-required:"true"`
	MinioEndpoint string        `env:"MINIO_ENDPOINT" env-default:"localhost:9000"`
	MinioAccessKey string       `env:"MINIO_ACCESS_KEY" env-default:"minioadmin"`
	MinioSecretKey string       `env:"MINIO_SECRET_KEY" env-default:"minioadmin"`
	MinioBucket    string        `env:"MINIO_BUCKET" env-default:"diploma-files"`
	MinioUseSSL    bool          `env:"MINIO_USE_SSL" env-default:"false"`
	MaxFileSize    int64         `env:"MAX_FILE_SIZE" env-default:"10485760"`
	MaxVideoSize   int64         `env:"MAX_VIDEO_SIZE" env-default:"52428800"`
	PresignedTTL   time.Duration `env:"PRESIGNED_TTL" env-default:"15m"`
	ThumbnailTTL   time.Duration `env:"THUMBNAIL_TTL" env-default:"1h"`
	KafkaBrokers   []string      `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	LogLevel       string        `env:"LOG_LEVEL" env-default:"info"`
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

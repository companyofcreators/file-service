package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/companyofcreators/file-service/internal/application/file"
	"github.com/companyofcreators/file-service/internal/config"
	"github.com/companyofcreators/file-service/internal/infrastructure/db"
	"github.com/companyofcreators/file-service/internal/infrastructure/kafka"
	"github.com/companyofcreators/file-service/internal/infrastructure/storage"
	httphandler "github.com/companyofcreators/file-service/internal/interfaces/http"
	"github.com/companyofcreators/file-service/internal/pkg"
	"github.com/companyofcreators/file-service/pkg/header_auth"
)

type Container struct {
	Config  *config.Config
	Logger  *slog.Logger
	Handler *httphandler.FileHandler

	fileRepo *db.FileRepository
	minio    *storage.MinioClient
	kafka    *kafka.Producer
}

func NewContainer() (*Container, error) {
	cfg := config.MustLoad()

	logger := pkg.NewLogger(cfg.LogLevel)

	// Database
	postgresDB, err := db.NewPostgresDB(cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("connected to PostgreSQL")

	fileRepo := db.NewFileRepository(postgresDB)

	// MinIO
	minioClient, err := storage.NewMinioClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Ensure bucket exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := minioClient.EnsureBucket(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure MinIO bucket exists: %w", err)
	}
	logger.Info("MinIO bucket ensured", slog.String("bucket", cfg.MinioBucket))

	// Kafka producer
	kafkaProd := kafka.NewProducer(cfg.KafkaBrokers)
	logger.Info("Kafka producer initialized", slog.Any("brokers", cfg.KafkaBrokers))

	// Use cases
	uploadUC := file.NewUploadUseCase(fileRepo, minioClient, kafkaProd, cfg.MaxFileSize, cfg.MaxVideoSize, cfg.MinioBucket)
	downloadUC := file.NewDownloadUseCase(fileRepo, minioClient, cfg.PresignedTTL)
	getFileUC := file.NewGetFileUseCase(fileRepo)
	listFilesUC := file.NewListFilesUseCase(fileRepo)
	deleteUC := file.NewDeleteUseCase(fileRepo, minioClient, kafkaProd)

	// User client for role checks
	userClientSigner := header_auth.NewHeaderSigner(cfg.HeaderHMACKey)
	userClient := NewUserClient(cfg.UserServiceURL, userClientSigner, logger)

	// HTTP handler
	httpScheme := "http"
	if cfg.MinioUseSSL {
		httpScheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", httpScheme, cfg.HTTPAddress)
	if cfg.HTTPAddress[0] == ':' {
		baseURL = ""
	} else {
		baseURL = fmt.Sprintf("%s://%s", httpScheme, cfg.HTTPAddress)
	}

	handler := httphandler.NewFileHandler(
		uploadUC,
		downloadUC,
		getFileUC,
		listFilesUC,
		deleteUC,
		userClient,
		logger,
		baseURL,
	)

	return &Container{
		Config:   cfg,
		Logger:   logger,
		Handler:  handler,
		fileRepo: fileRepo,
		minio:    minioClient,
		kafka:    kafkaProd,
	}, nil
}

func (c *Container) Shutdown() {
	if c.kafka != nil {
		if err := c.kafka.Close(); err != nil {
			c.Logger.Error("failed to close Kafka producer", slog.String("error", err.Error()))
		}
	}
	c.Logger.Info("file-service shutdown complete")
}

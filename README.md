# File Service

Manages file uploads and downloads using MinIO S3-compatible object storage. Services store only `file_id`, not the actual file content.

## Architecture

Clean architecture with four layers:
- **Domain** -- entities, repository interfaces, storage interface, errors
- **Application** -- use cases: upload, download, delete, get file, list files
- **Infrastructure** -- PostgreSQL repository, MinIO client, Kafka producer
- **Interfaces** -- HTTP handlers, DTOs, router

## Configuration

Copy `.env.example` to `.env` and adjust values:

| Variable | Default | Description |
|---|---|---|
| HTTP_ADDRESS | :8086 | Server listen address |
| DB_DSN | required | PostgreSQL connection string |
| MINIO_ENDPOINT | localhost:9000 | MinIO endpoint |
| MINIO_ACCESS_KEY | minioadmin | MinIO access key |
| MINIO_SECRET_KEY | minioadmin | MinIO secret key |
| MINIO_BUCKET | diploma-files | MinIO bucket name |
| MINIO_USE_SSL | false | Use SSL for MinIO |
| MAX_FILE_SIZE | 10485760 | Max file size in bytes (10MB) |
| MAX_VIDEO_SIZE | 52428800 | Max video size in bytes (50MB) |
| PRESIGNED_TTL | 15m | Presigned URL time-to-live |
| KAFKA_BROKERS | localhost:9092 | Kafka broker addresses |
| LOG_LEVEL | info | Logging level |

## API Endpoints

### Upload file
```
POST /internal/files/upload
Headers:
  X-User-ID: <uuid>
  Content-Type: multipart/form-data
Form fields:
  file: binary file
  type: avatar|order_attachment|chat_attachment|document
```

### Get file metadata
```
GET /internal/files/{id}
```

### Download file (presigned URL)
```
GET /internal/files/{id}/download
```

### Delete file
```
DELETE /internal/files/{id}
Headers:
  X-User-ID: <uuid>
  X-User-Role: user|admin  (optional, admin can delete any)
```

### List files by owner
```
GET /internal/files?limit=20&offset=0
Headers:
  X-User-ID: <uuid>
```

### Health check
```
GET /internal/health
```

## Allowed MIME Types

- image/jpeg
- image/png
- image/webp
- image/gif
- video/mp4
- application/pdf

## Running

```bash
# Development
go run ./cmd/api/main.go

# Docker
docker build -t file-service .
docker run --env-file .env file-service
```

## Kafka Events

- `file.uploaded` -- emitted when a file is successfully uploaded
- `file.deleted` -- emitted when a file is successfully deleted

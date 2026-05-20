package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domain "github.com/companyofcreators/file-service/internal/domain/file"
	appfile "github.com/companyofcreators/file-service/internal/application/file"
)

type FileHandler struct {
	uploadUC    *appfile.UploadUseCase
	downloadUC  *appfile.DownloadUseCase
	getFileUC   *appfile.GetFileUseCase
	listFilesUC *appfile.ListFilesUseCase
	deleteUC    *appfile.DeleteUseCase
	logger      *slog.Logger
	baseURL     string
}

func NewFileHandler(
	uploadUC *appfile.UploadUseCase,
	downloadUC *appfile.DownloadUseCase,
	getFileUC *appfile.GetFileUseCase,
	listFilesUC *appfile.ListFilesUseCase,
	deleteUC *appfile.DeleteUseCase,
	logger *slog.Logger,
	baseURL string,
) *FileHandler {
	return &FileHandler{
		uploadUC:    uploadUC,
		downloadUC:  downloadUC,
		getFileUC:   getFileUC,
		listFilesUC: listFilesUC,
		deleteUC:    deleteUC,
		logger:      logger,
		baseURL:     baseURL,
	}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with max 50MB buffer
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.writeError(w, http.StatusBadRequest, "не удалось обработать multipart-форму: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "отсутствует поле file: "+err.Error())
		return
	}
	defer file.Close()

	// Read MIME type from header
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Read file type parameter
	fileTypeStr := r.FormValue("type")
	if fileTypeStr == "" {
		fileTypeStr = string(domain.FileTypeDocument)
	}
	fileType, err := domain.ParseFileType(fileTypeStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недопустимый тип файла")
		return
	}

	// Extract owner ID from request header (set by API Gateway)
	ownerIDStr := r.Header.Get("X-User-ID")
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "недействительный или отсутствующий ID пользователя")
		return
	}

	input := appfile.UploadInput{
		Reader:   file,
		FileName: header.Filename,
		MimeType: mimeType,
		Size:     header.Size,
		FileType: fileType,
		OwnerID:  ownerID,
	}

	output, err := h.uploadUC.Execute(r.Context(), input)
	if err != nil {
		h.logger.Error("upload failed", slog.String("error", err.Error()))

		switch {
		case errors.Is(err, domain.ErrInvalidMimeType):
			h.writeError(w, http.StatusBadRequest, "недопустимый тип файла (MIME)")
		case errors.Is(err, domain.ErrFileTooLarge):
			h.writeError(w, http.StatusRequestEntityTooLarge, "файл слишком большой")
		case errors.Is(err, domain.ErrEmptyFile):
			h.writeError(w, http.StatusBadRequest, "файл пуст")
		default:
			h.writeError(w, http.StatusInternalServerError, "не удалось загрузить файл")
		}
		return
	}

	resp := UploadResponseDTO{
		Success: true,
		Data:    fileToResponseDTO(output.File, h.baseURL),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	fileID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID файла")
		return
	}

	f, err := h.getFileUC.Execute(r.Context(), appfile.GetFileInput{FileID: fileID})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, "файл не найден")
			return
		}
		h.logger.Error("get file failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "не удалось получить файл")
		return
	}

	// Authorization: allow only if caller is the file owner or an admin.
	requesterIDStr := r.Header.Get("X-User-Id")
	if requesterIDStr != "" {
		requesterID, err := uuid.Parse(requesterIDStr)
		if err == nil {
			isAdmin := r.Header.Get("X-User-Role") == "admin"
			if requesterID != f.OwnerID && !isAdmin {
				h.writeError(w, http.StatusForbidden, "доступ запрещён")
				return
			}
		}
	}

	resp := map[string]interface{}{
		"success": true,
		"data":    fileToDetailDTO(f),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	fileID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID файла")
		return
	}

	output, err := h.downloadUC.Execute(r.Context(), appfile.DownloadInput{FileID: fileID})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, "файл не найден")
			return
		}
		h.logger.Error("download failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "не удалось сгенерировать ссылку для скачивания")
		return
	}

	resp := DownloadResponseDTO{
		Success: true,
		Data: &DownloadDataDTO{
			FileID:       output.File.ID,
			PresignedURL: output.PresignedURL,
			ThumbnailURL: output.ThumbnailURL,
			MimeType:     output.File.MimeType,
			Size:         output.File.Size,
		},
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	fileID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID файла")
		return
	}

	// Extract requester ID from header
	requesterIDStr := r.Header.Get("X-User-ID")
	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "недействительный или отсутствующий ID пользователя")
		return
	}

	// Check if requester is admin
	isAdmin := r.Header.Get("X-User-Role") == "admin"

	err = h.deleteUC.Execute(r.Context(), appfile.DeleteInput{
		FileID:      fileID,
		RequesterID: requesterID,
		IsAdmin:     isAdmin,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			h.writeError(w, http.StatusNotFound, "файл не найден")
		case errors.Is(err, domain.ErrForbidden):
			h.writeError(w, http.StatusForbidden, "доступ запрещён")
		default:
			h.logger.Error("delete failed", slog.String("error", err.Error()))
			h.writeError(w, http.StatusInternalServerError, "не удалось удалить файл")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "файл успешно удалён",
	})
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := r.Header.Get("X-User-ID")
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "недействительный или отсутствующий ID пользователя")
		return
	}

	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	output, err := h.listFilesUC.Execute(r.Context(), appfile.ListFilesInput{
		OwnerID: ownerID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		h.logger.Error("list files failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "не удалось получить список файлов")
		return
	}

	var detailDTOs []*FileDetailDTO
	for _, f := range output.Files {
		detailDTOs = append(detailDTOs, fileToDetailDTO(f))
	}

	resp := ListFilesResponseDTO{
		Success: true,
		Data: &ListFilesDataDTO{
			Files:  detailDTOs,
			Total:  output.Total,
			Limit:  limit,
			Offset: offset,
		},
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *FileHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "file-service",
	})
}

func (h *FileHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", slog.String("error", err.Error()))
	}
}

func (h *FileHandler) writeError(w http.ResponseWriter, status int, message string) {
	resp := ErrorResponseDTO{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode error response", slog.String("error", err.Error()))
	}
}

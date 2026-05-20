package file

import "errors"

var (
	ErrNotFound        = errors.New("файл не найден")
	ErrInvalidFileType = errors.New("недопустимый тип файла")
	ErrForbidden       = errors.New("доступ запрещён: вы не владелец файла")
	ErrFileTooLarge    = errors.New("файл превышает максимально допустимый размер")
	ErrInvalidMimeType = errors.New("недопустимый MIME-тип файла")
	ErrEmptyFile       = errors.New("файл пуст")
	ErrUploadFailed    = errors.New("не удалось загрузить файл в хранилище")
	ErrDeleteFailed    = errors.New("не удалось удалить файл из хранилища")
)

package submission

import (
	"path/filepath"
	"strings"
)

// WorkFile представляет метаданные сданного файла работы (неизменяемый Value Object).
type WorkFile struct {
	name      string
	size      int64
	storageID string
}

// NewWorkFile создаёт файл работы с валидацией инвариантов (Правило 4).
func NewWorkFile(name string, size int64, storageID string) (*WorkFile, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidFileName
	}
	if size <= 0 {
		return nil, ErrInvalidFileSize
	}
	storageID = strings.TrimSpace(storageID)
	if storageID == "" {
		return nil, ErrInvalidStorageID
	}
	return &WorkFile{
		name:      name,
		size:      size,
		storageID: storageID,
	}, nil
}

// Name возвращает имя файла.
func (f *WorkFile) Name() string {
	return f.name
}

// Size возвращает размер файла в байтах.
func (f *WorkFile) Size() int64 {
	return f.size
}

// StorageID возвращает идентификатор сохранённого содержимого в хранилище.
func (f *WorkFile) StorageID() string {
	return f.storageID
}

// Extension возвращает нормализованное расширение файла в нижнем регистре с точкой (например, ".pdf").
func (f *WorkFile) Extension() string {
	return strings.ToLower(filepath.Ext(f.name))
}

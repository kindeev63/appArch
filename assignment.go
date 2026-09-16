package submission

import (
	"slices"
	"strings"
	"time"
)

// AssignmentStatus задаёт статус публикации задания.
type AssignmentStatus string

const (
	AssignmentStatusDraft     AssignmentStatus = "Draft"
	AssignmentStatusPublished AssignmentStatus = "Published"
)

// Assignment представляет учебное задание курса.
type Assignment struct {
	id                string
	courseID          string
	teacherID         string
	title             string
	deadline          time.Time
	status            AssignmentStatus
	allowedExtensions []string
	maxFileSize       int64
}

// NewAssignment создаёт задание в статусе Draft с проверкой инвариантов.
func NewAssignment(
	id string,
	courseID string,
	teacherID string,
	title string,
	deadline time.Time,
	allowedExtensions []string,
	maxFileSize int64,
) (*Assignment, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidID
	}
	courseID = strings.TrimSpace(courseID)
	if courseID == "" {
		return nil, ErrInvalidCourseID
	}
	teacherID = strings.TrimSpace(teacherID)
	if teacherID == "" {
		return nil, ErrInvalidID
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrInvalidTitle
	}
	if maxFileSize <= 0 {
		return nil, ErrInvalidMaxFileSize
	}

	normalizedExts := make([]string, 0, len(allowedExtensions))
	for _, ext := range allowedExtensions {
		trimmed := strings.ToLower(strings.TrimSpace(ext))
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, ".") {
			trimmed = "." + trimmed
		}
		if !slices.Contains(normalizedExts, trimmed) {
			normalizedExts = append(normalizedExts, trimmed)
		}
	}

	if len(normalizedExts) == 0 {
		return nil, ErrNoAllowedExtensions
	}

	return &Assignment{
		id:                id,
		courseID:          courseID,
		teacherID:         teacherID,
		title:             title,
		deadline:          deadline,
		status:            AssignmentStatusDraft,
		allowedExtensions: normalizedExts,
		maxFileSize:       maxFileSize,
	}, nil
}

// ID возвращает неизменяемый идентификатор задания (Правило 1).
func (a *Assignment) ID() string {
	return a.id
}

// CourseID возвращает идентификатор курса, которому принадлежит задание.
func (a *Assignment) CourseID() string {
	return a.courseID
}

// TeacherID возвращает идентификатор преподавателя-создателя.
func (a *Assignment) TeacherID() string {
	return a.teacherID
}

// Title возвращает название задания.
func (a *Assignment) Title() string {
	return a.title
}

// Deadline возвращает установленный дедлайн задания.
func (a *Assignment) Deadline() time.Time {
	return a.deadline
}

// Status возвращает текущий статус задания (Draft или Published).
func (a *Assignment) Status() AssignmentStatus {
	return a.status
}

// AllowedExtensions возвращает защищённую копию списка разрешённых расширений.
func (a *Assignment) AllowedExtensions() []string {
	exts := make([]string, len(a.allowedExtensions))
	copy(exts, a.allowedExtensions)
	return exts
}

// MaxFileSize возвращает максимальный разрешённый размер файла в байтах.
func (a *Assignment) MaxFileSize() int64 {
	return a.maxFileSize
}

// Publish переводит задание в статус опубликованного.
func (a *Assignment) Publish() error {
	if a.status == AssignmentStatusPublished {
		return ErrAlreadyPublished
	}
	a.status = AssignmentStatusPublished
	return nil
}

// IsExtensionAllowed проверяет, допустимо ли данное расширение файла.
func (a *Assignment) IsExtensionAllowed(ext string) bool {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return false
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return slices.Contains(a.allowedExtensions, ext)
}

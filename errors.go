package submission

import "errors"

var (
	// Ошибки валидации идентификаторов и базовых полей
	ErrInvalidID          = errors.New("id cannot be empty")
	ErrInvalidName        = errors.New("name cannot be empty")
	ErrInvalidTitle       = errors.New("title cannot be empty")
	ErrInvalidCourseID    = errors.New("course id cannot be empty")
	ErrCourseAlreadyAdded = errors.New("course already exists in the list")
	ErrCourseNotFound     = errors.New("course not found in the list")

	// Ошибки задания (Assignment)
	ErrNoAllowedExtensions    = errors.New("at least one allowed extension must be specified")
	ErrInvalidMaxFileSize     = errors.New("max file size must be greater than zero")
	ErrAssignmentNotPublished = errors.New("assignment is not published")
	ErrAlreadyPublished       = errors.New("assignment is already published")

	// Ошибки файла (WorkFile)
	ErrFileRequired     = errors.New("work file is required")
	ErrInvalidFileName  = errors.New("file name cannot be empty")
	ErrInvalidFileSize  = errors.New("file size must be greater than zero")
	ErrInvalidStorageID = errors.New("storage id cannot be empty")

	// Ошибки отправки (Submission)
	ErrStudentNotEnrolled     = errors.New("student is not enrolled in the assignment's course")
	ErrDeadlinePassed         = errors.New("submission deadline has passed")
	ErrFileSizeExceeded       = errors.New("file size exceeds assignment limit")
	ErrExtensionNotAllowed    = errors.New("file extension is not allowed")
	ErrAlreadySubmitted       = errors.New("submission has already been submitted; use ReplaceFile to update")
	ErrSubmissionNotSubmitted = errors.New("submission has not been submitted yet")
	ErrStudentMismatch        = errors.New("submission does not belong to this student")
	ErrAssignmentMismatch     = errors.New("submission does not belong to this assignment")

	// Ошибки обратной связи (Feedback)
	ErrTeacherNotAssigned = errors.New("teacher is not assigned to the assignment's course")
	ErrEmptyFeedbackText  = errors.New("feedback text cannot be empty")
	ErrTeacherRequired    = errors.New("teacher cannot be nil")
	ErrStudentRequired    = errors.New("student cannot be nil")
	ErrAssignmentRequired = errors.New("assignment cannot be nil")
)

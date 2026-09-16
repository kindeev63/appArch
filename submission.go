package submission

import (
	"strings"
	"time"
)

// SubmissionStatus отражает жизненный цикл отправки работы.
type SubmissionStatus string

const (
	SubmissionStatusDraft     SubmissionStatus = "Draft"
	SubmissionStatusSubmitted SubmissionStatus = "Submitted"
	SubmissionStatusReviewed  SubmissionStatus = "Reviewed"
)

// Feedback представляет неизменяемую рецензию преподавателя.
type Feedback struct {
	teacherID string
	text      string
	createdAt time.Time
}

// TeacherID возвращает идентификатор преподавателя-автора.
func (f Feedback) TeacherID() string {
	return f.teacherID
}

// Text возвращает текст обратной связи.
func (f Feedback) Text() string {
	return f.text
}

// CreatedAt возвращает время оставления обратной связи.
func (f Feedback) CreatedAt() time.Time {
	return f.createdAt
}

// Submission связывает студента и задание, хранит принятую работу, время подтверждения и статус.
type Submission struct {
	id            string
	assignmentID  string
	studentID     string
	file          *WorkFile
	submittedAt   time.Time
	status        SubmissionStatus
	feedbackItems []Feedback
}

// NewSubmission создаёт черновик отправки с фиксацией неизменяемых связей (Правило 1).
func NewSubmission(id, assignmentID, studentID string) (*Submission, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidID
	}
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return nil, ErrInvalidID
	}
	studentID = strings.TrimSpace(studentID)
	if studentID == "" {
		return nil, ErrInvalidID
	}

	return &Submission{
		id:            id,
		assignmentID:  assignmentID,
		studentID:     studentID,
		status:        SubmissionStatusDraft,
		feedbackItems: make([]Feedback, 0),
	}, nil
}

// ID возвращает неизменяемый идентификатор отправки (Правило 1).
func (s *Submission) ID() string {
	return s.id
}

// AssignmentID возвращает идентификатор связанного задания (Правило 1).
func (s *Submission) AssignmentID() string {
	return s.assignmentID
}

// StudentID возвращает идентификатор связанного студента (Правило 1).
func (s *Submission) StudentID() string {
	return s.studentID
}

// File возвращает принятый файл работы или nil, если работа ещё не отправлена.
func (s *Submission) File() *WorkFile {
	return s.file
}

// SubmittedAt возвращает время подтверждения отправки (Правило 5).
func (s *Submission) SubmittedAt() time.Time {
	return s.submittedAt
}

// Status возвращает текущий статус отправки.
func (s *Submission) Status() SubmissionStatus {
	return s.status
}

// FeedbackItems возвращает защищённую копию списка рецензий.
func (s *Submission) FeedbackItems() []Feedback {
	items := make([]Feedback, len(s.feedbackItems))
	copy(items, s.feedbackItems)
	return items
}

// Submit выполняет первичную отправку работы с проверкой всех инвариантов.
// В случае ошибки состояние объекта остаётся неизменным (Правило 10).
func (s *Submission) Submit(
	assignment *Assignment,
	student *Student,
	file *WorkFile,
	submittedAt time.Time,
) error {
	// Проверка принадлежности отправки
	if assignment == nil {
		return ErrAssignmentRequired
	}
	if student == nil {
		return ErrStudentRequired
	}
	if assignment.ID() != s.assignmentID {
		return ErrAssignmentMismatch
	}
	if student.ID() != s.studentID {
		return ErrStudentMismatch
	}

	// Повторная отправка через Submit запрещена, используйте ReplaceFile (Правило 6)
	if s.status != SubmissionStatusDraft {
		return ErrAlreadySubmitted
	}

	// Правило 2: Задание должно быть опубликовано
	if assignment.Status() != AssignmentStatusPublished {
		return ErrAssignmentNotPublished
	}

	// Правило 2: Студент должен быть записан на курс задания
	if !student.IsEnrolled(assignment.CourseID()) {
		return ErrStudentNotEnrolled
	}

	// Правило 3: Время отправки не позже дедлайна
	if submittedAt.After(assignment.Deadline()) {
		return ErrDeadlinePassed
	}

	// Правило 4: Проверка файла
	if err := validateFileForAssignment(file, assignment); err != nil {
		return err
	}

	// Правило 5 и 10: Атомарное изменение состояния только после всех проверок
	s.file = file
	s.submittedAt = submittedAt
	s.status = SubmissionStatusSubmitted
	return nil
}

// ReplaceFile заменяет текущую работу студента до наступления дедлайна (Правила 6, 7).
// В случае ошибки состояние объекта остаётся неизменным (Правило 10).
func (s *Submission) ReplaceFile(
	assignment *Assignment,
	student *Student,
	newFile *WorkFile,
	replacedAt time.Time,
) error {
	if assignment == nil {
		return ErrAssignmentRequired
	}
	if student == nil {
		return ErrStudentRequired
	}

	// Правило 7: Заменить работу может только тот студент и для того же задания
	if student.ID() != s.studentID {
		return ErrStudentMismatch
	}
	if assignment.ID() != s.assignmentID {
		return ErrAssignmentMismatch
	}

	// Замена возможна только для ранее отправленной работы
	if s.status == SubmissionStatusDraft {
		return ErrSubmissionNotSubmitted
	}

	// Правило 6: Замена возможна только до наступления дедлайна
	if replacedAt.After(assignment.Deadline()) {
		return ErrDeadlinePassed
	}

	// Проверка нового файла (Правило 4)
	if err := validateFileForAssignment(newFile, assignment); err != nil {
		return err
	}

	// Правило 6, 10: Успешная атомарная замена файла и обновление времени подтверждения
	s.file = newFile
	s.submittedAt = replacedAt
	s.status = SubmissionStatusSubmitted
	return nil
}

// AddFeedback добавляет рецензию преподавателя (Правила 8, 9).
// В случае ошибки состояние объекта остаётся неизменным (Правило 10).
func (s *Submission) AddFeedback(
	assignment *Assignment,
	teacher *Teacher,
	text string,
	createdAt time.Time,
) error {
	if assignment == nil {
		return ErrAssignmentRequired
	}
	if teacher == nil {
		return ErrTeacherRequired
	}

	if assignment.ID() != s.assignmentID {
		return ErrAssignmentMismatch
	}

	// Рецензия возможна только на сданную работу
	if s.status == SubmissionStatusDraft {
		return ErrSubmissionNotSubmitted
	}

	// Правило 8: Преподаватель должен быть назначен на курс задания
	if !teacher.IsAssignedTo(assignment.CourseID()) {
		return ErrTeacherNotAssigned
	}

	// Правило 9: Текст обратной связи не может быть пустым
	text = strings.TrimSpace(text)
	if text == "" {
		return ErrEmptyFeedbackText
	}

	// Правило 10: Атомарное обновление списка рецензий и статуса
	s.feedbackItems = append(s.feedbackItems, Feedback{
		teacherID: teacher.ID(),
		text:      text,
		createdAt: createdAt,
	})
	s.status = SubmissionStatusReviewed
	return nil
}

// validateFileForAssignment проверяет файл на соответствие правилу 4.
func validateFileForAssignment(file *WorkFile, assignment *Assignment) error {
	if file == nil {
		return ErrFileRequired
	}
	if file.Size() <= 0 {
		return ErrInvalidFileSize
	}
	if file.Size() > assignment.MaxFileSize() {
		return ErrFileSizeExceeded
	}
	if !assignment.IsExtensionAllowed(file.Extension()) {
		return ErrExtensionNotAllowed
	}
	return nil
}

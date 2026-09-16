package submission

import (
	"errors"
	"testing"
	"time"
)

func TestDomainRules(t *testing.T) {
	deadline := time.Date(2026, 9, 20, 18, 0, 0, 0, time.UTC)
	courseID := "course-arch-101"
	studentID := "student-001"
	teacherID := "teacher-001"
	assignmentID := "assignment-001"

	setupValidDomain := func(t *testing.T) (*Course, *Student, *Teacher, *Assignment, *WorkFile) {
		t.Helper()

		course, err := NewCourse(courseID, "Архитектура приложений")
		if err != nil {
			t.Fatalf("unexpected course error: %v", err)
		}

		student, err := NewStudent(studentID, "Анна")
		if err != nil {
			t.Fatalf("unexpected student error: %v", err)
		}
		if err := student.Enroll(courseID); err != nil {
			t.Fatalf("failed to enroll student: %v", err)
		}

		teacher, err := NewTeacher(teacherID, "Ирина")
		if err != nil {
			t.Fatalf("unexpected teacher error: %v", err)
		}
		if err := teacher.AssignCourse(courseID); err != nil {
			t.Fatalf("failed to assign teacher: %v", err)
		}

		assignment, err := NewAssignment(
			assignmentID,
			courseID,
			teacherID,
			"Архитектурное решение",
			deadline,
			[]string{".pdf", ".zip"},
			5_000_000,
		)
		if err != nil {
			t.Fatalf("unexpected assignment error: %v", err)
		}
		if err := assignment.Publish(); err != nil {
			t.Fatalf("failed to publish assignment: %v", err)
		}

		file, err := NewWorkFile("solution.pdf", 120_000, "blob-storage-001")
		if err != nil {
			t.Fatalf("unexpected file error: %v", err)
		}

		return course, student, teacher, assignment, file
	}

	t.Run("Правило 1: Идентификаторы неизменяемы и валидируются при создании", func(t *testing.T) {
		_, _, _, assignment, _ := setupValidDomain(t)

		if assignment.ID() != assignmentID {
			t.Errorf("expected assignment ID %s, got %s", assignmentID, assignment.ID())
		}
		// У структур нет публичных сеттеров для ID.
		// Проверяем валидацию пустых ID при создании:
		if _, err := NewStudent("   ", "Иван"); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
		if _, err := NewStudent("s1", "   "); !errors.Is(err, ErrInvalidName) {
			t.Errorf("expected ErrInvalidName, got %v", err)
		}
		if _, err := NewTeacher("", "Иван"); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
		if _, err := NewTeacher("t1", "   "); !errors.Is(err, ErrInvalidName) {
			t.Errorf("expected ErrInvalidName, got %v", err)
		}
		if _, err := NewCourse("", "Title"); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
		if _, err := NewCourse("c1", "   "); !errors.Is(err, ErrInvalidTitle) {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
		if _, err := NewSubmission("", assignmentID, studentID); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
		if _, err := NewSubmission("sub-1", "", studentID); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
		if _, err := NewSubmission("sub-1", assignmentID, ""); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("Правило 2: Задание должно быть опубликовано", func(t *testing.T) {
		_, student, _, _, file := setupValidDomain(t)

		draftAssignment, err := NewAssignment("draft-1", courseID, teacherID, "Draft", deadline, []string{".pdf"}, 1000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sub, _ := NewSubmission("sub-1", draftAssignment.ID(), student.ID())
		submitTime := deadline.Add(-1 * time.Hour)
		err = sub.Submit(draftAssignment, student, file, submitTime)

		if !errors.Is(err, ErrAssignmentNotPublished) {
			t.Errorf("expected ErrAssignmentNotPublished, got %v", err)
		}
		if sub.Status() != SubmissionStatusDraft {
			t.Errorf("expected status Draft, got %v", sub.Status())
		}
	})

	t.Run("Правило 2: Студент должен быть записан на курс задания", func(t *testing.T) {
		_, _, _, assignment, file := setupValidDomain(t)

		unregisteredStudent, err := NewStudent("student-999", "Борис")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sub, _ := NewSubmission("sub-2", assignment.ID(), unregisteredStudent.ID())
		err = sub.Submit(assignment, unregisteredStudent, file, deadline.Add(-10*time.Minute))

		if !errors.Is(err, ErrStudentNotEnrolled) {
			t.Errorf("expected ErrStudentNotEnrolled, got %v", err)
		}
		if sub.Status() != SubmissionStatusDraft {
			t.Errorf("expected status Draft, got %v", sub.Status())
		}
	})

	t.Run("Правило 3: Дедлайн - сдача вовремя принимается, сдача с опозданием отклоняется", func(t *testing.T) {
		_, student, _, assignment, file := setupValidDomain(t)

		// Точно в дедлайн (не позже) - должно быть принято
		subOnTime, _ := NewSubmission("sub-3", assignment.ID(), student.ID())
		if err := subOnTime.Submit(assignment, student, file, deadline); err != nil {
			t.Errorf("submission exactly at deadline should be accepted, got error: %v", err)
		}
		if subOnTime.Status() != SubmissionStatusSubmitted {
			t.Errorf("expected status Submitted, got %v", subOnTime.Status())
		}

		// На 1 секунду позже дедлайна - отказ
		subLate, _ := NewSubmission("sub-4", assignment.ID(), student.ID())
		lateTime := deadline.Add(1 * time.Second)
		err := subLate.Submit(assignment, student, file, lateTime)
		if !errors.Is(err, ErrDeadlinePassed) {
			t.Errorf("expected ErrDeadlinePassed, got %v", err)
		}
		if subLate.Status() != SubmissionStatusDraft {
			t.Errorf("expected status to remain Draft, got %v", subLate.Status())
		}
	})

	t.Run("Правило 4: Проверка файла (размер и расширение)", func(t *testing.T) {
		_, student, _, assignment, _ := setupValidDomain(t)

		// 1. Файл nil
		sub, _ := NewSubmission("sub-file-nil", assignment.ID(), student.ID())
		if err := sub.Submit(assignment, student, nil, deadline.Add(-1*time.Hour)); !errors.Is(err, ErrFileRequired) {
			t.Errorf("expected ErrFileRequired, got %v", err)
		}

		// 2. Превышение максимального размера
		oversizedFile, err := NewWorkFile("big.pdf", 6_000_000, "blob-big")
		if err != nil {
			t.Fatalf("unexpected file error: %v", err)
		}
		sub2, _ := NewSubmission("sub-file-big", assignment.ID(), student.ID())
		if err := sub2.Submit(assignment, student, oversizedFile, deadline.Add(-1*time.Hour)); !errors.Is(err, ErrFileSizeExceeded) {
			t.Errorf("expected ErrFileSizeExceeded, got %v", err)
		}

		// 3. Недопустимое расширение
		disallowedFile, err := NewWorkFile("virus.exe", 1000, "blob-exe")
		if err != nil {
			t.Fatalf("unexpected file error: %v", err)
		}
		sub3, _ := NewSubmission("sub-file-ext", assignment.ID(), student.ID())
		if err := sub3.Submit(assignment, student, disallowedFile, deadline.Add(-1*time.Hour)); !errors.Is(err, ErrExtensionNotAllowed) {
			t.Errorf("expected ErrExtensionNotAllowed, got %v", err)
		}

		// 4. Расширение в верхнем регистре должно распознаваться
		capsFile, err := NewWorkFile("report.PDF", 1000, "blob-caps")
		if err != nil {
			t.Fatalf("unexpected file error: %v", err)
		}
		sub4, _ := NewSubmission("sub-file-caps", assignment.ID(), student.ID())
		if err := sub4.Submit(assignment, student, capsFile, deadline.Add(-1*time.Hour)); err != nil {
			t.Errorf("uppercase extension should be accepted, got: %v", err)
		}
	})

	t.Run("Правило 5: Время подтверждения задаётся успешной операцией отправки", func(t *testing.T) {
		_, student, _, assignment, file := setupValidDomain(t)

		sub, _ := NewSubmission("sub-5", assignment.ID(), student.ID())
		if !sub.SubmittedAt().IsZero() {
			t.Errorf("expected zero SubmittedAt before submit, got %v", sub.SubmittedAt())
		}

		submitTime := deadline.Add(-2 * time.Hour)
		if err := sub.Submit(assignment, student, file, submitTime); err != nil {
			t.Fatalf("submit failed: %v", err)
		}

		if !sub.SubmittedAt().Equal(submitTime) {
			t.Errorf("expected SubmittedAt %v, got %v", submitTime, sub.SubmittedAt())
		}
	})

	t.Run("Правила 6 и 7: Замена работы до дедлайна и проверка авторства", func(t *testing.T) {
		_, student, _, assignment, file1 := setupValidDomain(t)

		sub, _ := NewSubmission("sub-replace", assignment.ID(), student.ID())
		submitTime := deadline.Add(-5 * time.Hour)
		if err := sub.Submit(assignment, student, file1, submitTime); err != nil {
			t.Fatalf("initial submit failed: %v", err)
		}

		file2, err := NewWorkFile("solution_v2.pdf", 200_000, "blob-002")
		if err != nil {
			t.Fatalf("file creation failed: %v", err)
		}

		// Попытка повторного вызова Submit вместо ReplaceFile
		if err := sub.Submit(assignment, student, file2, deadline.Add(-4*time.Hour)); !errors.Is(err, ErrAlreadySubmitted) {
			t.Errorf("expected ErrAlreadySubmitted on second Submit, got %v", err)
		}

		// Попытка замены другим студентом (Правило 7)
		otherStudent, _ := NewStudent("student-impostor", "Хакер")
		_ = otherStudent.Enroll(courseID)
		err = sub.ReplaceFile(assignment, otherStudent, file2, deadline.Add(-4*time.Hour))
		if !errors.Is(err, ErrStudentMismatch) {
			t.Errorf("expected ErrStudentMismatch, got %v", err)
		}
		if sub.File().Name() != "solution.pdf" {
			t.Errorf("file should not be replaced by another student")
		}

		// Попытка замены с несовпадающим заданием (Правило 7)
		otherAssignment, _ := NewAssignment("assign-2", courseID, teacherID, "A2", deadline, []string{".pdf"}, 5000)
		_ = otherAssignment.Publish()
		err = sub.ReplaceFile(otherAssignment, student, file2, deadline.Add(-4*time.Hour))
		if !errors.Is(err, ErrAssignmentMismatch) {
			t.Errorf("expected ErrAssignmentMismatch, got %v", err)
		}

		// Попытка замены после дедлайна (Правило 6)
		lateReplaceTime := deadline.Add(10 * time.Minute)
		err = sub.ReplaceFile(assignment, student, file2, lateReplaceTime)
		if !errors.Is(err, ErrDeadlinePassed) {
			t.Errorf("expected ErrDeadlinePassed, got %v", err)
		}
		if sub.File().Name() != "solution.pdf" {
			t.Errorf("file should not be replaced after deadline")
		}

		// Успешная замена до дедлайна (Правило 6)
		validReplaceTime := deadline.Add(-1 * time.Hour)
		err = sub.ReplaceFile(assignment, student, file2, validReplaceTime)
		if err != nil {
			t.Fatalf("valid replace failed: %v", err)
		}
		if sub.File().Name() != "solution_v2.pdf" {
			t.Errorf("expected new file solution_v2.pdf, got %s", sub.File().Name())
		}
		if !sub.SubmittedAt().Equal(validReplaceTime) {
			t.Errorf("expected updated SubmittedAt %v, got %v", validReplaceTime, sub.SubmittedAt())
		}
	})

	t.Run("Правило 8: Обратную связь может оставить только преподаватель курса задания", func(t *testing.T) {
		_, student, teacher, assignment, file := setupValidDomain(t)

		sub, _ := NewSubmission("sub-feedback", assignment.ID(), student.ID())
		_ = sub.Submit(assignment, student, file, deadline.Add(-1*time.Hour))

		// Преподаватель другого курса
		otherTeacher, err := NewTeacher("teacher-other", "Петр")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = otherTeacher.AssignCourse("another-course-id")

		err = sub.AddFeedback(assignment, otherTeacher, "Хорошая работа", deadline.Add(2*time.Hour))
		if !errors.Is(err, ErrTeacherNotAssigned) {
			t.Errorf("expected ErrTeacherNotAssigned, got %v", err)
		}
		if len(sub.FeedbackItems()) != 0 {
			t.Errorf("feedback should not be added by unauthorized teacher")
		}

		// Преподаватель курса оставляет рецензию
		createdAt := deadline.Add(2 * time.Hour)
		err = sub.AddFeedback(assignment, teacher, "Отличное решение", createdAt)
		if err != nil {
			t.Fatalf("authorized feedback failed: %v", err)
		}
		if len(sub.FeedbackItems()) != 1 {
			t.Fatalf("expected 1 feedback item, got %d", len(sub.FeedbackItems()))
		}
		fb := sub.FeedbackItems()[0]
		if fb.Text() != "Отличное решение" {
			t.Errorf("expected text 'Отличное решение', got %s", fb.Text())
		}
		if fb.TeacherID() != teacher.ID() {
			t.Errorf("expected teacher id %s, got %s", teacher.ID(), fb.TeacherID())
		}
		if !fb.CreatedAt().Equal(createdAt) {
			t.Errorf("expected createdAt %v, got %v", createdAt, fb.CreatedAt())
		}
		if sub.Status() != SubmissionStatusReviewed {
			t.Errorf("expected status Reviewed, got %s", sub.Status())
		}
	})

	t.Run("Правило 9: Текст обратной связи не может быть пустым", func(t *testing.T) {
		_, student, teacher, assignment, file := setupValidDomain(t)

		sub, _ := NewSubmission("sub-empty-fb", assignment.ID(), student.ID())
		_ = sub.Submit(assignment, student, file, deadline.Add(-1*time.Hour))

		// Пустой текст
		err := sub.AddFeedback(assignment, teacher, "   ", deadline.Add(1*time.Hour))
		if !errors.Is(err, ErrEmptyFeedbackText) {
			t.Errorf("expected ErrEmptyFeedbackText, got %v", err)
		}
		if len(sub.FeedbackItems()) != 0 {
			t.Errorf("expected 0 feedback items after failure, got %d", len(sub.FeedbackItems()))
		}
	})

	t.Run("Правило 10: Недопустимая операция не меняет состояние частично", func(t *testing.T) {
		_, student, teacher, assignment, file := setupValidDomain(t)

		sub, _ := NewSubmission("sub-atomic", assignment.ID(), student.ID())

		// Попытка отправить недопустимый файл (размер превышен)
		bigFile, _ := NewWorkFile("too_large.pdf", 10_000_000, "blob-99")
		err := sub.Submit(assignment, student, bigFile, deadline.Add(-1*time.Hour))
		if err == nil {
			t.Fatalf("expected error on oversized file, got nil")
		}

		// Проверяем, что ни одно свойство Submission не изменилось
		if sub.File() != nil {
			t.Errorf("expected File to remain nil, got %v", sub.File())
		}
		if !sub.SubmittedAt().IsZero() {
			t.Errorf("expected SubmittedAt to remain zero, got %v", sub.SubmittedAt())
		}
		if sub.Status() != SubmissionStatusDraft {
			t.Errorf("expected status to remain Draft, got %s", sub.Status())
		}

		// Теперь успешный Submit
		if err := sub.Submit(assignment, student, file, deadline.Add(-1*time.Hour)); err != nil {
			t.Fatalf("unexpected submit error: %v", err)
		}

		// Попытка добавить пустой фидбек
		err = sub.AddFeedback(assignment, teacher, "  ", deadline.Add(1*time.Hour))
		if err == nil {
			t.Fatalf("expected error on empty feedback")
		}

		// Статус должен остаться Submitted, а не перейти в Reviewed, фидбек не должен добавиться
		if sub.Status() != SubmissionStatusSubmitted {
			t.Errorf("expected status to remain Submitted, got %s", sub.Status())
		}
		if len(sub.FeedbackItems()) != 0 {
			t.Errorf("expected 0 feedback items, got %d", len(sub.FeedbackItems()))
		}
	})

	t.Run("Дополнительные инварианты сущностей и методов", func(t *testing.T) {
		// Course геттеры и смена названия
		c, err := NewCourse("c-test", "Initial Title")
		if err != nil {
			t.Fatalf("course error: %v", err)
		}
		if c.ID() != "c-test" || c.Title() != "Initial Title" {
			t.Errorf("course getter mismatch")
		}
		if err := c.ChangeTitle("Updated Title"); err != nil || c.Title() != "Updated Title" {
			t.Errorf("failed to change title: %v", err)
		}
		if err := c.ChangeTitle("  "); !errors.Is(err, ErrInvalidTitle) {
			t.Errorf("expected ErrInvalidTitle on empty title, got %v", err)
		}

		// Student геттеры и методы
		s, _ := NewStudent("s-test", "Initial Name")
		if s.Name() != "Initial Name" {
			t.Errorf("student getter mismatch")
		}
		if err := s.ChangeName("Updated Name"); err != nil || s.Name() != "Updated Name" {
			t.Errorf("failed to change student name")
		}
		if err := s.ChangeName("  "); !errors.Is(err, ErrInvalidName) {
			t.Errorf("expected ErrInvalidName, got %v", err)
		}
		if err := s.Enroll("  "); !errors.Is(err, ErrInvalidCourseID) {
			t.Errorf("expected ErrInvalidCourseID, got %v", err)
		}
		if err := s.Unenroll("  "); !errors.Is(err, ErrInvalidCourseID) {
			t.Errorf("expected ErrInvalidCourseID, got %v", err)
		}

		// Teacher геттеры и методы
		tch, _ := NewTeacher("t-test", "Initial Teacher")
		if tch.Name() != "Initial Teacher" {
			t.Errorf("teacher getter mismatch")
		}
		if err := tch.ChangeName("New Teacher Name"); err != nil || tch.Name() != "New Teacher Name" {
			t.Errorf("failed to change teacher name")
		}
		if err := tch.ChangeName("  "); !errors.Is(err, ErrInvalidName) {
			t.Errorf("expected ErrInvalidName, got %v", err)
		}
		if err := tch.AssignCourse("  "); !errors.Is(err, ErrInvalidCourseID) {
			t.Errorf("expected ErrInvalidCourseID, got %v", err)
		}
		_ = tch.AssignCourse("c-tch")
		if err := tch.AssignCourse("c-tch"); !errors.Is(err, ErrCourseAlreadyAdded) {
			t.Errorf("expected ErrCourseAlreadyAdded, got %v", err)
		}
		if tch.CourseIDs()[0] != "c-tch" {
			t.Errorf("expected course c-tch in teacher courses")
		}
		if err := tch.RemoveCourse("c-tch"); err != nil || tch.IsAssignedTo("c-tch") {
			t.Errorf("failed to remove course from teacher")
		}
		if err := tch.RemoveCourse("c-none"); !errors.Is(err, ErrCourseNotFound) {
			t.Errorf("expected ErrCourseNotFound, got %v", err)
		}
		if err := tch.RemoveCourse("  "); !errors.Is(err, ErrInvalidCourseID) {
			t.Errorf("expected ErrInvalidCourseID, got %v", err)
		}

		// WorkFile валидация
		if _, err := NewWorkFile("", 100, "id"); !errors.Is(err, ErrInvalidFileName) {
			t.Errorf("expected ErrInvalidFileName, got %v", err)
		}
		if _, err := NewWorkFile("a.pdf", 0, "id"); !errors.Is(err, ErrInvalidFileSize) {
			t.Errorf("expected ErrInvalidFileSize, got %v", err)
		}
		if _, err := NewWorkFile("a.pdf", -10, "id"); !errors.Is(err, ErrInvalidFileSize) {
			t.Errorf("expected ErrInvalidFileSize, got %v", err)
		}
		if _, err := NewWorkFile("a.pdf", 100, ""); !errors.Is(err, ErrInvalidStorageID) {
			t.Errorf("expected ErrInvalidStorageID, got %v", err)
		}
		wf, _ := NewWorkFile("doc.pdf", 500, "storage-1")
		if wf.StorageID() != "storage-1" || wf.Size() != 500 {
			t.Errorf("workfile getters failed")
		}

		// Assignment валидация и геттеры
		if _, err := NewAssignment("", "c", "t", "T", deadline, []string{".pdf"}, 1000); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected ErrInvalidID")
		}
		if _, err := NewAssignment("a", "", "t", "T", deadline, []string{".pdf"}, 1000); !errors.Is(err, ErrInvalidCourseID) {
			t.Errorf("expected ErrInvalidCourseID")
		}
		if _, err := NewAssignment("a", "c", "", "T", deadline, []string{".pdf"}, 1000); !errors.Is(err, ErrInvalidID) {
			t.Errorf("expected ErrInvalidID for teacher")
		}
		if _, err := NewAssignment("a", "c", "t", "", deadline, []string{".pdf"}, 1000); !errors.Is(err, ErrInvalidTitle) {
			t.Errorf("expected ErrInvalidTitle")
		}
		if _, err := NewAssignment("a", "c", "t", "T", deadline, []string{}, 1000); !errors.Is(err, ErrNoAllowedExtensions) {
			t.Errorf("expected ErrNoAllowedExtensions, got %v", err)
		}
		if _, err := NewAssignment("a", "c", "t", "T", deadline, []string{".pdf"}, 0); !errors.Is(err, ErrInvalidMaxFileSize) {
			t.Errorf("expected ErrInvalidMaxFileSize, got %v", err)
		}
		a, _ := NewAssignment("a", "c", "t", "Title", deadline, []string{"pdf", "ZIP"}, 1000)
		if a.TeacherID() != "t" || a.Title() != "Title" || a.MaxFileSize() != 1000 {
			t.Errorf("assignment getters failed")
		}
		if len(a.AllowedExtensions()) != 2 {
			t.Errorf("expected 2 normalized extensions, got %v", a.AllowedExtensions())
		}
		if a.IsExtensionAllowed("") || a.IsExtensionAllowed(".exe") {
			t.Errorf("invalid extension allowed")
		}
		if !a.IsExtensionAllowed("pdf") || !a.IsExtensionAllowed(".zip") {
			t.Errorf("valid extension was rejected")
		}
		_ = a.Publish()
		if err := a.Publish(); !errors.Is(err, ErrAlreadyPublished) {
			t.Errorf("expected ErrAlreadyPublished, got %v", err)
		}

		// Submission валидация параметров Submit / ReplaceFile / AddFeedback на nil и некорректные связи
		sub, _ := NewSubmission("sub-edge", a.ID(), s.ID())
		if sub.ID() != "sub-edge" || sub.AssignmentID() != a.ID() || sub.StudentID() != s.ID() {
			t.Errorf("submission getters failed")
		}
		// ReplaceFile на черновике
		if err := sub.ReplaceFile(a, s, wf, deadline); !errors.Is(err, ErrSubmissionNotSubmitted) {
			t.Errorf("expected ErrSubmissionNotSubmitted on replace draft, got %v", err)
		}
		// AddFeedback на черновике
		if err := sub.AddFeedback(a, tch, "text", deadline); !errors.Is(err, ErrSubmissionNotSubmitted) {
			t.Errorf("expected ErrSubmissionNotSubmitted on feedback to draft, got %v", err)
		}
		// nil параметры
		if err := sub.Submit(nil, s, wf, deadline); !errors.Is(err, ErrAssignmentRequired) {
			t.Errorf("expected ErrAssignmentRequired, got %v", err)
		}
		if err := sub.Submit(a, nil, wf, deadline); !errors.Is(err, ErrStudentRequired) {
			t.Errorf("expected ErrStudentRequired, got %v", err)
		}
		if err := sub.ReplaceFile(nil, s, wf, deadline); !errors.Is(err, ErrAssignmentRequired) {
			t.Errorf("expected ErrAssignmentRequired, got %v", err)
		}
		if err := sub.ReplaceFile(a, nil, wf, deadline); !errors.Is(err, ErrStudentRequired) {
			t.Errorf("expected ErrStudentRequired, got %v", err)
		}
		if err := sub.AddFeedback(nil, tch, "txt", deadline); !errors.Is(err, ErrAssignmentRequired) {
			t.Errorf("expected ErrAssignmentRequired, got %v", err)
		}
		if err := sub.AddFeedback(a, nil, "txt", deadline); !errors.Is(err, ErrTeacherRequired) {
			t.Errorf("expected ErrTeacherRequired, got %v", err)
		}

		// Несовпадение assignment id в Submit
		mismatchSub, _ := NewSubmission("sub-m", "a-actual", s.ID())
		wrongAssign, _ := NewAssignment("a-wrong", "c", "t", "T", deadline, []string{".pdf"}, 1000)
		if err := mismatchSub.Submit(wrongAssign, s, wf, deadline); !errors.Is(err, ErrAssignmentMismatch) {
			t.Errorf("expected ErrAssignmentMismatch, got %v", err)
		}
		// Несовпадение student id в Submit
		wrongStudent, _ := NewStudent("s-wrong", "Wrong")
		actualAssign, _ := NewAssignment("a-actual", "c", "t", "T", deadline, []string{".pdf"}, 1000)
		if err := mismatchSub.Submit(actualAssign, wrongStudent, wf, deadline); !errors.Is(err, ErrStudentMismatch) {
			t.Errorf("expected ErrStudentMismatch, got %v", err)
		}
	})
}

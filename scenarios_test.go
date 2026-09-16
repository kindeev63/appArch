package submission

import (
	"errors"
	"testing"
	"time"
)

// TestScenarios проверяет 11 согласованных сценариев предметной области.
// Для каждого сценария фиксируются секции:
// Дано: предусловия и исходное состояние
// Когда: вызываемая публичная операция
// Тогда: результат и наблюдаемое итоговое состояние (включая проверку неизменности состояния при отказе - Правило 10)
func TestScenarios(t *testing.T) {
	// Базовые тестовые данные
	fixedDeadline := time.Date(2026, 10, 15, 18, 0, 0, 0, time.UTC)
	courseID := "course-arch-101"
	otherCourseID := "course-history-202"

	// Вспомогательная функция создания валидного окружения
	setupBaseEnvironment := func(t *testing.T) (*Course, *Teacher, *Assignment, *Student) {
		t.Helper()

		course, err := NewCourse(courseID, "Архитектура приложений")
		if err != nil {
			t.Fatalf("failed to create course: %v", err)
		}

		teacher, err := NewTeacher("teacher-101", "Алексей Преподаватель")
		if err != nil {
			t.Fatalf("failed to create teacher: %v", err)
		}
		if err := teacher.AssignCourse(course.ID()); err != nil {
			t.Fatalf("failed to assign course to teacher: %v", err)
		}

		assignment, err := NewAssignment(
			"assign-101",
			course.ID(),
			teacher.ID(),
			"Лабораторная работа 1",
			fixedDeadline,
			[]string{".pdf", ".zip"},
			5_000_000, // 5 MB
		)
		if err != nil {
			t.Fatalf("failed to create assignment: %v", err)
		}
		if err := assignment.Publish(); err != nil {
			t.Fatalf("failed to publish assignment: %v", err)
		}

		student, err := NewStudent("student-001", "Иван Иванов")
		if err != nil {
			t.Fatalf("failed to create student: %v", err)
		}
		if err := student.Enroll(course.ID()); err != nil {
			t.Fatalf("failed to enroll student to course: %v", err)
		}

		return course, teacher, assignment, student
	}

	// 1. Создание студента
	t.Run("Сценарий 1: Создание студента", func(t *testing.T) {
		// Дано: корректный идентификатор "student-001" и имя "Иван Иванов"
		studentID := "student-001"
		studentName := "Иван Иванов"

		// Когда: создаётся студент через фабричный метод NewStudent
		student, err := NewStudent(studentID, studentName)

		// Тогда: студент успешно создан, ID и имя соответствуют переданным значениям,
		// список курсов пуст, студент не записан ни на один курс
		if err != nil {
			t.Fatalf("unexpected error creating student: %v", err)
		}
		if student.ID() != studentID {
			t.Errorf("expected student ID %q, got %q", studentID, student.ID())
		}
		if student.Name() != studentName {
			t.Errorf("expected student name %q, got %q", studentName, student.Name())
		}
		if len(student.CourseIDs()) != 0 {
			t.Errorf("expected 0 courses for newly created student, got %d", len(student.CourseIDs()))
		}
		if student.IsEnrolled("any-course") {
			t.Errorf("student should not be enrolled in any course upon creation")
		}
	})

	// 2. Создание учителя
	t.Run("Сценарий 2: Создание учителя", func(t *testing.T) {
		// Дано: корректный идентификатор "teacher-101" и имя "Алексей Преподаватель"
		teacherID := "teacher-101"
		teacherName := "Алексей Преподаватель"

		// Когда: создаётся преподаватель через фабричный метод NewTeacher
		teacher, err := NewTeacher(teacherID, teacherName)

		// Тогда: учитель успешно создан, ID и имя соответствуют переданным значениям,
		// список назначенных курсов пуст
		if err != nil {
			t.Fatalf("unexpected error creating teacher: %v", err)
		}
		if teacher.ID() != teacherID {
			t.Errorf("expected teacher ID %q, got %q", teacherID, teacher.ID())
		}
		if teacher.Name() != teacherName {
			t.Errorf("expected teacher name %q, got %q", teacherName, teacher.Name())
		}
		if len(teacher.CourseIDs()) != 0 {
			t.Errorf("expected 0 courses for newly created teacher, got %d", len(teacher.CourseIDs()))
		}
		if teacher.IsAssignedTo("any-course") {
			t.Errorf("teacher should not be assigned to any course upon creation")
		}
	})

	// 3. Создание задания учителем
	t.Run("Сценарий 3: Создание задания учителем", func(t *testing.T) {
		// Дано: преподаватель, назначенный на курс "course-arch-101"
		course, err := NewCourse(courseID, "Архитектура приложений")
		if err != nil {
			t.Fatalf("failed to create course: %v", err)
		}
		teacher, err := NewTeacher("teacher-101", "Алексей Преподаватель")
		if err != nil {
			t.Fatalf("failed to create teacher: %v", err)
		}
		if err := teacher.AssignCourse(course.ID()); err != nil {
			t.Fatalf("failed to assign course to teacher: %v", err)
		}

		assignID := "assign-301"
		assignTitle := "Проектирование микросервисов"
		allowedExts := []string{".pdf", ".zip"}
		maxFileSize := int64(10_000_000) // 10 MB

		// Когда: преподаватель создаёт задание и публикует его
		assignment, err := NewAssignment(
			assignID,
			course.ID(),
			teacher.ID(),
			assignTitle,
			fixedDeadline,
			allowedExts,
			maxFileSize,
		)
		if err != nil {
			t.Fatalf("unexpected error creating assignment: %v", err)
		}

		// Изначальный статус должен быть Draft
		if assignment.Status() != AssignmentStatusDraft {
			t.Errorf("expected initial assignment status Draft, got %s", assignment.Status())
		}

		// Публикация задания
		if err := assignment.Publish(); err != nil {
			t.Fatalf("unexpected error publishing assignment: %v", err)
		}

		// Тогда: задание успешно создано и опубликовано, атрибуты соответствуют переданным
		if assignment.ID() != assignID {
			t.Errorf("expected assignment ID %q, got %q", assignID, assignment.ID())
		}
		if assignment.CourseID() != course.ID() {
			t.Errorf("expected course ID %q, got %q", course.ID(), assignment.CourseID())
		}
		if assignment.TeacherID() != teacher.ID() {
			t.Errorf("expected teacher ID %q, got %q", teacher.ID(), assignment.TeacherID())
		}
		if assignment.Title() != assignTitle {
			t.Errorf("expected title %q, got %q", assignTitle, assignment.Title())
		}
		if !assignment.Deadline().Equal(fixedDeadline) {
			t.Errorf("expected deadline %v, got %v", fixedDeadline, assignment.Deadline())
		}
		if assignment.Status() != AssignmentStatusPublished {
			t.Errorf("expected status Published, got %s", assignment.Status())
		}
		if assignment.MaxFileSize() != maxFileSize {
			t.Errorf("expected max file size %d, got %d", maxFileSize, assignment.MaxFileSize())
		}
		if !assignment.IsExtensionAllowed(".pdf") || !assignment.IsExtensionAllowed(".zip") {
			t.Errorf("expected .pdf and .zip to be allowed")
		}
	})

	// 4. Сабмит студентом файла по заданию через сабмишин
	t.Run("Сценарий 4: Сабмит студентом файла по заданию через сабмишин", func(t *testing.T) {
		// Дано: опубликованное задание, студент записан на курс, черновик отправки,
		// валидный файл (разрешённый размер и расширение), время сдачи до дедлайна
		_, _, assignment, student := setupBaseEnvironment(t)

		submission, err := NewSubmission("sub-401", assignment.ID(), student.ID())
		if err != nil {
			t.Fatalf("failed to create submission: %v", err)
		}

		file, err := NewWorkFile("report.pdf", 500_000, "blob-storage-401")
		if err != nil {
			t.Fatalf("failed to create work file: %v", err)
		}

		submitTime := fixedDeadline.Add(-2 * time.Hour) // за 2 часа до дедлайна

		// Когда: студент отправляет файл по заданию
		err = submission.Submit(assignment, student, file, submitTime)

		// Тогда: отправка принята без ошибок, статус перешёл в Submitted,
		// принятый файл сохранён, время подтверждения зафиксировано
		if err != nil {
			t.Fatalf("expected successful submit, got error: %v", err)
		}
		if submission.Status() != SubmissionStatusSubmitted {
			t.Errorf("expected status %s, got %s", SubmissionStatusSubmitted, submission.Status())
		}
		if submission.File() != file {
			t.Errorf("expected file %v, got %v", file, submission.File())
		}
		if !submission.SubmittedAt().Equal(submitTime) {
			t.Errorf("expected submittedAt %v, got %v", submitTime, submission.SubmittedAt())
		}
	})

	// 5. Сабмит вторым студентом файла за первого студента
	t.Run("Сценарий 5: Сабмит вторым студентом файла за первого студента", func(t *testing.T) {
		// Дано: опубликованное задание, черновик отправки создан для студента №1 ("student-001").
		// Второй студент ("student-002") также записан на курс.
		_, _, assignment, student1 := setupBaseEnvironment(t)

		student2, err := NewStudent("student-002", "Петр Петров")
		if err != nil {
			t.Fatalf("failed to create student 2: %v", err)
		}
		if err := student2.Enroll(assignment.CourseID()); err != nil {
			t.Fatalf("failed to enroll student 2: %v", err)
		}

		submission, err := NewSubmission("sub-501", assignment.ID(), student1.ID())
		if err != nil {
			t.Fatalf("failed to create submission: %v", err)
		}

		file, err := NewWorkFile("hacked.pdf", 300_000, "blob-501")
		if err != nil {
			t.Fatalf("failed to create work file: %v", err)
		}

		submitTime := fixedDeadline.Add(-1 * time.Hour)

		// Когда: второй студент пытается отправить файл в чужую отправку
		err = submission.Submit(assignment, student2, file, submitTime)

		// Тогда: отказ с ErrStudentMismatch.
		// Состояние отправки не изменилось: статус Draft, файл nil, время нулевое (Правило 10)
		if !errors.Is(err, ErrStudentMismatch) {
			t.Errorf("expected error ErrStudentMismatch, got: %v", err)
		}
		if submission.Status() != SubmissionStatusDraft {
			t.Errorf("expected status to remain %s, got %s", SubmissionStatusDraft, submission.Status())
		}
		if submission.File() != nil {
			t.Errorf("expected file to remain nil, got %v", submission.File())
		}
		if !submission.SubmittedAt().IsZero() {
			t.Errorf("expected submittedAt to remain zero, got %v", submission.SubmittedAt())
		}
	})

	// 6. Сабмит вне дедлайна
	t.Run("Сценарий 6: Сабмит вне дедлайна", func(t *testing.T) {
		// Дано: опубликованное задание с дедлайном 18:00, записанный студент, черновик отправки, валидный файл.
		// Время отправки строго после дедлайна (+1 секунда)
		_, _, assignment, student := setupBaseEnvironment(t)

		submission, err := NewSubmission("sub-601", assignment.ID(), student.ID())
		if err != nil {
			t.Fatalf("failed to create submission: %v", err)
		}

		file, err := NewWorkFile("late_solution.pdf", 400_000, "blob-601")
		if err != nil {
			t.Fatalf("failed to create work file: %v", err)
		}

		lateTime := fixedDeadline.Add(1 * time.Second)

		// Когда: студент отправляет работу после дедлайна
		err = submission.Submit(assignment, student, file, lateTime)

		// Тогда: отказ с ErrDeadlinePassed.
		// Состояние отправки не изменилось: статус Draft, файл nil, время нулевое (Правило 10)
		if !errors.Is(err, ErrDeadlinePassed) {
			t.Errorf("expected error ErrDeadlinePassed, got: %v", err)
		}
		if submission.Status() != SubmissionStatusDraft {
			t.Errorf("expected status to remain %s, got %s", SubmissionStatusDraft, submission.Status())
		}
		if submission.File() != nil {
			t.Errorf("expected file to remain nil, got %v", submission.File())
		}
		if !submission.SubmittedAt().IsZero() {
			t.Errorf("expected submittedAt to remain zero, got %v", submission.SubmittedAt())
		}
	})

	// 7. Сабмит по левому курсу
	t.Run("Сценарий 7: Сабмит по левому курсу", func(t *testing.T) {
		// Дано: опубликованное задание курса "course-arch-101".
		// Студент записан на ДРУГОЙ курс "course-history-202", но не записан на курс задания.
		_, _, assignment, _ := setupBaseEnvironment(t)

		otherStudent, err := NewStudent("student-other", "Сергей Чужой")
		if err != nil {
			t.Fatalf("failed to create other student: %v", err)
		}
		if err := otherStudent.Enroll(otherCourseID); err != nil {
			t.Fatalf("failed to enroll other student: %v", err)
		}

		submission, err := NewSubmission("sub-701", assignment.ID(), otherStudent.ID())
		if err != nil {
			t.Fatalf("failed to create submission: %v", err)
		}

		file, err := NewWorkFile("work.pdf", 250_000, "blob-701")
		if err != nil {
			t.Fatalf("failed to create work file: %v", err)
		}

		submitTime := fixedDeadline.Add(-1 * time.Hour)

		// Когда: студент не со своего курса пытается отправить работу по заданию
		err = submission.Submit(assignment, otherStudent, file, submitTime)

		// Тогда: отказ с ErrStudentNotEnrolled.
		// Состояние отправки не изменилось: статус Draft, файл nil, время нулевое (Правило 10)
		if !errors.Is(err, ErrStudentNotEnrolled) {
			t.Errorf("expected error ErrStudentNotEnrolled, got: %v", err)
		}
		if submission.Status() != SubmissionStatusDraft {
			t.Errorf("expected status to remain %s, got %s", SubmissionStatusDraft, submission.Status())
		}
		if submission.File() != nil {
			t.Errorf("expected file to remain nil, got %v", submission.File())
		}
		if !submission.SubmittedAt().IsZero() {
			t.Errorf("expected submittedAt to remain zero, got %v", submission.SubmittedAt())
		}
	})

	// 8. Сабмит с пустым файлом
	t.Run("Сценарий 8: Сабмит с пустым файлом", func(t *testing.T) {
		_, _, assignment, student := setupBaseEnvironment(t)
		submitTime := fixedDeadline.Add(-1 * time.Hour)

		// 8.1 Отправка nil-указателя на файл
		t.Run("8.1 nil-файл", func(t *testing.T) {
			// Дано: опубликованное задание, записанный студент, черновик отправки, файл nil
			submission, err := NewSubmission("sub-801a", assignment.ID(), student.ID())
			if err != nil {
				t.Fatalf("failed to create submission: %v", err)
			}

			// Когда: попытка сдать nil файл
			err = submission.Submit(assignment, student, nil, submitTime)

			// Тогда: отказ с ErrFileRequired, статус остаётся Draft, время нулевое
			if !errors.Is(err, ErrFileRequired) {
				t.Errorf("expected error ErrFileRequired, got: %v", err)
			}
			if submission.Status() != SubmissionStatusDraft {
				t.Errorf("expected status to remain %s, got %s", SubmissionStatusDraft, submission.Status())
			}
			if submission.File() != nil {
				t.Errorf("expected file to remain nil, got %v", submission.File())
			}
			if !submission.SubmittedAt().IsZero() {
				t.Errorf("expected submittedAt to remain zero, got %v", submission.SubmittedAt())
			}
		})

		// 8.2 Создание файла с размером 0 байт (пустое содержимое)
		t.Run("8.2 файл с нулевым размером (0 байт)", func(t *testing.T) {
			// Дано: имя файла "empty.pdf", размер 0 байт
			// Когда: попытка создать WorkFile с размером 0
			file, err := NewWorkFile("empty.pdf", 0, "blob-zero")

			// Тогда: отказ с ErrInvalidFileSize на этапе валидации Value Object, объект не создан
			if !errors.Is(err, ErrInvalidFileSize) {
				t.Errorf("expected error ErrInvalidFileSize, got: %v", err)
			}
			if file != nil {
				t.Errorf("expected file to be nil, got: %v", file)
			}
		})
	})

	// 9. Сабмит с файлом, нарушающим расширение
	t.Run("Сценарий 9: Сабмит с файлом, нарушающим расширение", func(t *testing.T) {
		// Дано: опубликованное задание с разрешёнными расширениями [".pdf", ".zip"].
		// Студент подготовил исполняемый файл с недопустимым расширением ".exe"
		_, _, assignment, student := setupBaseEnvironment(t)

		submission, err := NewSubmission("sub-901", assignment.ID(), student.ID())
		if err != nil {
			t.Fatalf("failed to create submission: %v", err)
		}

		file, err := NewWorkFile("malware.exe", 100_000, "blob-901")
		if err != nil {
			t.Fatalf("failed to create work file: %v", err)
		}

		submitTime := fixedDeadline.Add(-1 * time.Hour)

		// Когда: студент отправляет файл с недопустимым расширением
		err = submission.Submit(assignment, student, file, submitTime)

		// Тогда: отказ с ErrExtensionNotAllowed.
		// Состояние отправки не изменилось: статус Draft, файл nil, время нулевое (Правило 10)
		if !errors.Is(err, ErrExtensionNotAllowed) {
			t.Errorf("expected error ErrExtensionNotAllowed, got: %v", err)
		}
		if submission.Status() != SubmissionStatusDraft {
			t.Errorf("expected status to remain %s, got %s", SubmissionStatusDraft, submission.Status())
		}
		if submission.File() != nil {
			t.Errorf("expected file to remain nil, got %v", submission.File())
		}
		if !submission.SubmittedAt().IsZero() {
			t.Errorf("expected submittedAt to remain zero, got %v", submission.SubmittedAt())
		}
	})

	// 10. Сабмит с отрицательным сайзом
	t.Run("Сценарий 10: Сабмит с отрицательным сайзом", func(t *testing.T) {
		// Дано: параметры файла с отрицательным размером (-1024 байт)
		fileName := "corrupt.pdf"
		negativeSize := int64(-1024)
		storageID := "blob-neg"

		// Когда: попытка сконструировать WorkFile с отрицательным размером
		file, err := NewWorkFile(fileName, negativeSize, storageID)

		// Тогда: отказ с ErrInvalidFileSize, объект файла не создаётся (file == nil)
		if !errors.Is(err, ErrInvalidFileSize) {
			t.Errorf("expected error ErrInvalidFileSize, got: %v", err)
		}
		if file != nil {
			t.Errorf("expected file to be nil, got %v", file)
		}

		// Проверка: черновик отправки не может принять такой файл
		_, _, assignment, student := setupBaseEnvironment(t)
		submission, err := NewSubmission("sub-1001", assignment.ID(), student.ID())
		if err != nil {
			t.Fatalf("failed to create submission: %v", err)
		}

		// Попытка передать nil (так как файл не создался) приводит к ErrFileRequired
		err = submission.Submit(assignment, student, file, fixedDeadline.Add(-1*time.Hour))
		if !errors.Is(err, ErrFileRequired) {
			t.Errorf("expected error ErrFileRequired when file is nil, got: %v", err)
		}
		if submission.Status() != SubmissionStatusDraft {
			t.Errorf("expected status to remain %s, got %s", SubmissionStatusDraft, submission.Status())
		}
		if submission.File() != nil {
			t.Errorf("expected file to remain nil, got %v", submission.File())
		}
		if !submission.SubmittedAt().IsZero() {
			t.Errorf("expected submittedAt to remain zero, got %v", submission.SubmittedAt())
		}
	})

	// 11. Сабмит с сайзом сверх лимита
	t.Run("Сценарий 11: Сабмит с сайзом сверх лимита", func(t *testing.T) {
		// Дано: опубликованное задание с максимальным размером файла 5_000_000 байт (5 MB).
		// Файл работы имеет размер на 1 байт больше лимита: 5_000_001 байт.
		_, _, assignment, student := setupBaseEnvironment(t)

		submission, err := NewSubmission("sub-1101", assignment.ID(), student.ID())
		if err != nil {
			t.Fatalf("failed to create submission: %v", err)
		}

		oversizedBytes := assignment.MaxFileSize() + 1
		file, err := NewWorkFile("huge_archive.zip", oversizedBytes, "blob-1101")
		if err != nil {
			t.Fatalf("failed to create work file: %v", err)
		}

		submitTime := fixedDeadline.Add(-1 * time.Hour)

		// Когда: студент отправляет файл, размер которого превышает установленный лимит
		err = submission.Submit(assignment, student, file, submitTime)

		// Тогда: отказ с ErrFileSizeExceeded.
		// Состояние отправки не изменилось: статус Draft, файл nil, время нулевое (Правило 10)
		if !errors.Is(err, ErrFileSizeExceeded) {
			t.Errorf("expected error ErrFileSizeExceeded, got: %v", err)
		}
		if submission.Status() != SubmissionStatusDraft {
			t.Errorf("expected status to remain %s, got %s", SubmissionStatusDraft, submission.Status())
		}
		if submission.File() != nil {
			t.Errorf("expected file to remain nil, got %v", submission.File())
		}
		if !submission.SubmittedAt().IsZero() {
			t.Errorf("expected submittedAt to remain zero, got %v", submission.SubmittedAt())
		}
	})
}

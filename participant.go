package submission

import (
	"slices"
	"strings"
)

// Course представляет учебный курс.
type Course struct {
	id    string
	title string
}

// NewCourse создаёт новый курс. Идентификатор после создания неизменяем.
func NewCourse(id, title string) (*Course, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidID
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrInvalidTitle
	}
	return &Course{
		id:    id,
		title: title,
	}, nil
}

// ID возвращает неизменяемый идентификатор курса (Правило 1).
func (c *Course) ID() string {
	return c.id
}

// Title возвращает название курса.
func (c *Course) Title() string {
	return c.title
}

// ChangeTitle меняет название курса с проверкой на непустоту.
func (c *Course) ChangeTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrInvalidTitle
	}
	c.title = title
	return nil
}

// Student представляет студента, записанного на один или несколько курсов.
type Student struct {
	id        string
	name      string
	courseIDs map[string]struct{}
}

// NewStudent создаёт студента. Идентификатор неизменяем (Правило 1).
func NewStudent(id, name string) (*Student, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidID
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}
	return &Student{
		id:        id,
		name:      name,
		courseIDs: make(map[string]struct{}),
	}, nil
}

// ID возвращает неизменяемый идентификатор студента (Правило 1).
func (s *Student) ID() string {
	return s.id
}

// Name возвращает имя студента.
func (s *Student) Name() string {
	return s.name
}

// ChangeName меняет имя студента с валидацией.
func (s *Student) ChangeName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}
	s.name = name
	return nil
}

// Enroll записывает студента на курс.
func (s *Student) Enroll(courseID string) error {
	courseID = strings.TrimSpace(courseID)
	if courseID == "" {
		return ErrInvalidCourseID
	}
	if _, exists := s.courseIDs[courseID]; exists {
		return ErrCourseAlreadyAdded
	}
	s.courseIDs[courseID] = struct{}{}
	return nil
}

// Unenroll исключает студента из курса.
func (s *Student) Unenroll(courseID string) error {
	courseID = strings.TrimSpace(courseID)
	if courseID == "" {
		return ErrInvalidCourseID
	}
	if _, exists := s.courseIDs[courseID]; !exists {
		return ErrCourseNotFound
	}
	delete(s.courseIDs, courseID)
	return nil
}

// IsEnrolled проверяет, записан ли студент на курс.
func (s *Student) IsEnrolled(courseID string) bool {
	_, exists := s.courseIDs[courseID]
	return exists
}

// CourseIDs возвращает защищённую копию списка курсов студента.
func (s *Student) CourseIDs() []string {
	ids := make([]string, 0, len(s.courseIDs))
	for id := range s.courseIDs {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// Teacher представляет преподавателя, назначенного на один или несколько курсов.
type Teacher struct {
	id        string
	name      string
	courseIDs map[string]struct{}
}

// NewTeacher создаёт преподавателя. Идентификатор неизменяем (Правило 1).
func NewTeacher(id, name string) (*Teacher, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidID
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}
	return &Teacher{
		id:        id,
		name:      name,
		courseIDs: make(map[string]struct{}),
	}, nil
}

// ID возвращает неизменяемый идентификатор преподавателя (Правило 1).
func (t *Teacher) ID() string {
	return t.id
}

// Name возвращает имя преподавателя.
func (t *Teacher) Name() string {
	return t.name
}

// ChangeName меняет имя преподавателя с валидацией.
func (t *Teacher) ChangeName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}
	t.name = name
	return nil
}

// AssignCourse назначает преподавателя на курс.
func (t *Teacher) AssignCourse(courseID string) error {
	courseID = strings.TrimSpace(courseID)
	if courseID == "" {
		return ErrInvalidCourseID
	}
	if _, exists := t.courseIDs[courseID]; exists {
		return ErrCourseAlreadyAdded
	}
	t.courseIDs[courseID] = struct{}{}
	return nil
}

// RemoveCourse снимает преподавателя с курса.
func (t *Teacher) RemoveCourse(courseID string) error {
	courseID = strings.TrimSpace(courseID)
	if courseID == "" {
		return ErrInvalidCourseID
	}
	if _, exists := t.courseIDs[courseID]; !exists {
		return ErrCourseNotFound
	}
	delete(t.courseIDs, courseID)
	return nil
}

// IsAssignedTo проверяет, назначен ли преподаватель на указанный курс.
func (t *Teacher) IsAssignedTo(courseID string) bool {
	_, exists := t.courseIDs[courseID]
	return exists
}

// CourseIDs возвращает защищённую копию списка курсов преподавателя.
func (t *Teacher) CourseIDs() []string {
	ids := make([]string, 0, len(t.courseIDs))
	for id := range t.courseIDs {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

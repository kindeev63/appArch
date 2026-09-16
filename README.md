# Доменная модель сервиса сдачи учебных работ (Go)

Данный документ содержит согласованный контракт объектной модели и сопоставление с требованиями [case.md](../case.md).

---

## 1. Анализ исходного наброска и несоответствий с case.md

В первоначальном наброске были выявлены следующие архитектурные проблемы:
1. **Примесь RBAC/Admin в доменную модель**:
   - Введение роли `"admin"` и проверок `user.group == admin` внутри методов сущностей `Teacher`, `Course`, `Student`. В [case.md](../case.md) администратора нет — это контекст учебного взаимодействия, а не системы администрирования прав.
2. **Нарушение инкапсуляции и Правила 1 (неизменяемость ID)**:
   - Присутствовали мутаторы `set id`, `set courses`, `get set name`, что позволяло переопределять идентификаторы и связи постфактум.
3. **Ошибки сигнатур и Go-идиоматики в `Submission`**:
   - Метод `Submit(Assignment, Student, WorkFile, DateTimeOffset)` использовал value-receiver `(s Submission)`, из-за чего мутации терялись.
   - Метод не возвращал `error` (нарушение Правила 10).
   - Использовался тип `DateTimeOffset` (C#) вместо `time.Time`.
4. **Пропуски ключевых бизнес-правил**:
   - Отсутствовал метод замены работы (`ReplaceFile`) до дедлайна (Правила 6, 7).
   - Отсутствовала обратная связь (`Feedback`) и метод её добавления с проверкой назначения преподавателя на курс (Правила 8, 9).
   - `WorkFile` имел сеттер `name`, позволяющий подменить расширение в обход проверки задания.

---

## 2. Границы модели и схема классов

### Выбранные классы:
- **Подробно проработанные (Aggregate Root / ключевые сущности):**
  - `Submission` — агрегат сдачи работы: владеет статусом сдачи, принятым файлом, временем подтверждения, историей рецензий; отвечает за координацию дедлайна, студента, файла и преподавателя.
  - `Assignment` — задание курса: владеет статусом публикации, дедлайном, правилами валидации файлов (разрешённые расширения, лимит размера).
  - `Student` — участник учебного процесса: владеет списком курсов, на которые записан.
- **Вспомогательные сущности и Value Objects:**
  - `WorkFile` — неизменяемый Value Object с метаданными файла (имя, размер, storageID).
  - `Feedback` — неизменяемый Value Object рецензии преподавателя.
  - `Teacher` — преподаватель, назначенный на курсы.
  - `Course` — учебный курс.

### Схема контрактов

```text
class Course
- id: string
- title: string
+ ID(): string
+ Title(): string
+ ChangeTitle(title: string): error

class Student
- id: string
- name: string
- courseIDs: map[string]struct{}
+ ID(): string
+ Name(): string
+ ChangeName(name: string): error
+ Enroll(courseID: string): error
+ Unenroll(courseID: string): error
+ IsEnrolled(courseID: string): bool
+ CourseIDs(): []string

class Teacher
- id: string
- name: string
- courseIDs: map[string]struct{}
+ ID(): string
+ Name(): string
+ ChangeName(name: string): error
+ AssignCourse(courseID: string): error
+ RemoveCourse(courseID: string): error
+ IsAssignedTo(courseID: string): bool
+ CourseIDs(): []string

class WorkFile
- name: string
- size: int64
- storageID: string
+ Name(): string
+ Size(): int64
+ StorageID(): string
+ Extension(): string

class Assignment
- id: string
- courseID: string
- teacherID: string
- title: string
- deadline: time.Time
- status: AssignmentStatus (Draft | Published)
- allowedExtensions: []string
- maxFileSize: int64
+ ID(): string
+ CourseID(): string
+ TeacherID(): string
+ Title(): string
+ Deadline(): time.Time
+ Status(): AssignmentStatus
+ AllowedExtensions(): []string
+ MaxFileSize(): int64
+ Publish(): error
+ IsExtensionAllowed(ext: string): bool

class Feedback
- teacherID: string
- text: string
- createdAt: time.Time
+ TeacherID(): string
+ Text(): string
+ CreatedAt(): time.Time

class Submission
- id: string
- assignmentID: string
- studentID: string
- file: *WorkFile
- submittedAt: time.Time
- status: SubmissionStatus (Draft | Submitted | Reviewed)
- feedbackItems: []Feedback
+ ID(): string
+ AssignmentID(): string
+ StudentID(): string
+ File(): *WorkFile
+ SubmittedAt(): time.Time
+ Status(): SubmissionStatus
+ FeedbackItems(): []Feedback
+ Submit(assignment: *Assignment, student: *Student, file: *WorkFile, submittedAt: time.Time): error
+ ReplaceFile(assignment: *Assignment, student: *Student, file: *WorkFile, replacedAt: time.Time): error
+ AddFeedback(assignment: *Assignment, teacher: *Teacher, text: string, createdAt: time.Time): error
```

---

## 3. Rules (Таблица бизнес-правил)

| № | Правило из case.md | Тип правила | Публичная операция | Владелец проверки | Обеспечение невозможности обхода |
|---|---|---|---|---|---|
| 1 | Идентификаторы неизменяемы после создания | Локальное | Конструкторы `New*` | Все сущности | Поля не экспортированы, сеттеры отсутствуют |
| 2 | Сдача только по опубликованному заданию курса студента | Межобъектное | `Submission.Submit` | `Submission` | Проверяются `assignment.Status() == Published` и `student.IsEnrolled(courseID)` |
| 3 | Сдача принимается, если время не позже дедлайна | Межобъектное | `Submission.Submit`, `Submission.ReplaceFile` | `Submission` | Проверка `submittedAt <= assignment.Deadline()` |
| 4 | Файл один, непустой, разрешённое расширение, 0 < size <= max | Межобъектное / Value Object | `NewWorkFile`, `Submission.Submit`, `ReplaceFile` | `WorkFile`, `Submission` | Валидация инвариантов файла и задания до присвоения |
| 5 | Время подтверждения задаётся операцией отправки | Локальное | `Submission.Submit`, `Submission.ReplaceFile` | `Submission` | Нет публичного сеттера `SubmittedAt`, время фиксируется атомарно |
| 6 | До дедлайна студент может заменить текущую работу | Межобъектное | `Submission.ReplaceFile` | `Submission` | Проверка дедлайна, обновление активного файла и времени подтверждения |
| 7 | Замена только тем же студентом и для того же задания | Межобъектное | `Submission.ReplaceFile` | `Submission` | Проверка `student.ID() == s.studentID` и `assignment.ID() == s.assignmentID` |
| 8 | Обратную связь оставляет только преподаватель курса | Межобъектное | `Submission.AddFeedback` | `Submission` | Проверка `teacher.IsAssignedTo(assignment.CourseID())` |
| 9 | Текст обратной связи не может быть пустым | Локальное | `Submission.AddFeedback` | `Submission` | `strings.TrimSpace(text) != ""` |
| 10| Отказ при ошибке явный, без частичного изменения состояния | Системное | Все методы мутации | Все сущности | Валидация всех предусловий ДО любых изменений полей; возврат типизированных `error` |

---

## 4. Контракты отказов (Errors)

В качестве идиоматичного механизма Go используются типизированные sentinel errors (`errors.New`), позволяющие вызывающему коду использовать `errors.Is(err, target)`:
- `ErrInvalidID`, `ErrInvalidName`, `ErrInvalidTitle`, `ErrInvalidCourseID`
- `ErrAssignmentNotPublished`, `ErrAlreadyPublished`
- `ErrStudentNotEnrolled`, `ErrDeadlinePassed`
- `ErrFileRequired`, `ErrInvalidFileSize`, `ErrFileSizeExceeded`, `ErrExtensionNotAllowed`
- `ErrAlreadySubmitted`, `ErrSubmissionNotSubmitted`
- `ErrStudentMismatch`, `ErrAssignmentMismatch`
- `ErrTeacherNotAssigned`, `ErrEmptyFeedbackText`

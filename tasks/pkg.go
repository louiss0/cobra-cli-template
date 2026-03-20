package tasks

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/samber/lo"
)

const weeklyCleanupWindow = 7 * 24 * time.Hour

const (
	ListFilterAll        ListFilter = "all"
	ListFilterComplete   ListFilter = "complete"
	ListFilterIncomplete ListFilter = "incomplete"
)

var ErrTaskNotFound = errors.New("task not found")

type ListFilter string

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	createdAt   int64
	updatedAt   int64
}

type publicTask struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   string `json:"completed"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateTaskInput struct {
	Title       string
	Description string
	Completed   bool
}

type UpdateTaskInput struct {
	Title       *string
	Description *string
	Completed   OptionalBoolean
}

type OptionalBoolean struct {
	IsSet bool
	Value bool
}

type Store struct {
	dataDir   string
	now       func() time.Time
	newTaskID func() string
}

type taskDocument struct {
	LastCleanupAt int64       `json:"lastCleanupAt"`
	Tasks         taskBuckets `json:"tasks"`
}

type taskBuckets struct {
	Complete   []Task `json:"complete"`
	Incomplete []Task `json:"incomplete"`
}

type taskJSON struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

func OptionalBool(value bool) OptionalBoolean {
	return OptionalBoolean{
		IsSet: true,
		Value: value,
	}
}

func NewStore(dataDir string, now func() time.Time, newTaskID func() string) *Store {
	if now == nil {
		now = time.Now
	}

	if newTaskID == nil {
		newTaskID = NewTaskID
	}

	return &Store{
		dataDir:   dataDir,
		now:       now,
		newTaskID: newTaskID,
	}
}

func (task Task) CreatedAt() int64 {
	return task.createdAt
}

func (task Task) UpdatedAt() int64 {
	return task.updatedAt
}

func (task Task) MarshalJSON() ([]byte, error) {
	return json.Marshal(taskJSON{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.createdAt,
		UpdatedAt:   task.updatedAt,
	})
}

func (task *Task) UnmarshalJSON(content []byte) error {
	var taskValue taskJSON
	if err := json.Unmarshal(content, &taskValue); err != nil {
		return err
	}

	task.ID = taskValue.ID
	task.Title = taskValue.Title
	task.Description = taskValue.Description
	task.Completed = taskValue.Completed
	task.createdAt = taskValue.CreatedAt
	task.updatedAt = taskValue.UpdatedAt

	return nil
}

func PresentTask(task Task) any {
	return publicTask{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   completionLabel(task.Completed),
		CreatedAt:   formatUnixTimestamp(task.CreatedAt()),
		UpdatedAt:   formatUnixTimestamp(task.UpdatedAt()),
	}
}

func PresentTasks(list []Task) any {
	return lo.Map(list, func(task Task, _ int) publicTask {
		return PresentTask(task).(publicTask)
	})
}

func (store *Store) Create(username string, input CreateTaskInput) (Task, error) {
	document, err := store.load(username)
	if err != nil {
		return Task{}, err
	}

	if store.runWeeklyCleanup(&document) {
		if err := store.save(username, document); err != nil {
			return Task{}, err
		}
	}

	now := store.now().Unix()
	task := Task{
		ID:          store.newTaskID(),
		Title:       input.Title,
		Description: input.Description,
		Completed:   input.Completed,
		createdAt:   now,
		updatedAt:   now,
	}

	if task.Completed {
		document.Tasks.Complete = append(document.Tasks.Complete, task)
	} else {
		document.Tasks.Incomplete = append(document.Tasks.Incomplete, task)
	}

	if err := store.save(username, document); err != nil {
		return Task{}, err
	}

	return task, nil
}

func (store *Store) List(username string, filter ListFilter) ([]Task, error) {
	document, err := store.load(username)
	if err != nil {
		return nil, err
	}

	if store.runWeeklyCleanup(&document) {
		if err := store.save(username, document); err != nil {
			return nil, err
		}
	}

	switch filter {
	case ListFilterComplete:
		return append([]Task(nil), document.Tasks.Complete...), nil
	case ListFilterIncomplete:
		return append([]Task(nil), document.Tasks.Incomplete...), nil
	default:
		return lo.Concat(document.Tasks.Incomplete, document.Tasks.Complete), nil
	}
}

func (store *Store) Get(username string, id string) (Task, error) {
	document, err := store.load(username)
	if err != nil {
		return Task{}, err
	}

	if store.runWeeklyCleanup(&document) {
		if err := store.save(username, document); err != nil {
			return Task{}, err
		}
	}

	task, _, _, found := findTask(document, id)
	if !found {
		return Task{}, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	return task, nil
}

func (store *Store) Update(username string, id string, input UpdateTaskInput) (Task, error) {
	document, err := store.load(username)
	if err != nil {
		return Task{}, err
	}

	cleanupChanged := store.runWeeklyCleanup(&document)
	task, bucketName, index, found := findTask(document, id)
	if !found {
		if cleanupChanged {
			_ = store.save(username, document)
		}
		return Task{}, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	if input.Title != nil {
		task.Title = *input.Title
	}

	if input.Description != nil {
		task.Description = *input.Description
	}

	if input.Completed.IsSet {
		task.Completed = input.Completed.Value
	}

	task.updatedAt = store.now().Unix()
	document = removeTask(document, bucketName, index)

	if task.Completed {
		document.Tasks.Complete = append(document.Tasks.Complete, task)
	} else {
		document.Tasks.Incomplete = append(document.Tasks.Incomplete, task)
	}

	if err := store.save(username, document); err != nil {
		return Task{}, err
	}

	return task, nil
}

func (store *Store) Delete(username string, id string) error {
	document, err := store.load(username)
	if err != nil {
		return err
	}

	cleanupChanged := store.runWeeklyCleanup(&document)
	_, bucketName, index, found := findTask(document, id)
	if !found {
		if cleanupChanged {
			_ = store.save(username, document)
		}
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	document = removeTask(document, bucketName, index)

	return store.save(username, document)
}

func (store *Store) load(username string) (taskDocument, error) {
	if err := os.MkdirAll(store.tasksDir(), 0o755); err != nil {
		return taskDocument{}, fmt.Errorf("create task directory: %w", err)
	}

	path := store.taskFilePath(username)
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		document := newTaskDocument()
		if err := store.save(username, document); err != nil {
			return taskDocument{}, err
		}
		return document, nil
	}
	if err != nil {
		return taskDocument{}, fmt.Errorf("stat task file: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return taskDocument{}, fmt.Errorf("read task file: %w", err)
	}

	var document taskDocument
	if err := json.Unmarshal(content, &document); err != nil {
		return taskDocument{}, fmt.Errorf("decode task file: %w", err)
	}

	return document, nil
}

func (store *Store) save(username string, document taskDocument) error {
	content, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("encode task file: %w", err)
	}

	if err := os.WriteFile(store.taskFilePath(username), content, 0o644); err != nil {
		return fmt.Errorf("write task file: %w", err)
	}

	return nil
}

func (store *Store) runWeeklyCleanup(document *taskDocument) bool {
	now := store.now().Unix()

	if document.LastCleanupAt == 0 {
		document.LastCleanupAt = now
		return true
	}

	if time.Unix(now, 0).Sub(time.Unix(document.LastCleanupAt, 0)) < weeklyCleanupWindow {
		return false
	}

	if len(document.Tasks.Complete) == 0 {
		document.LastCleanupAt = now
		return true
	}

	document.Tasks.Complete = []Task{}
	document.LastCleanupAt = now
	return true
}

func findTask(document taskDocument, id string) (Task, string, int, bool) {
	if task, index, found := lo.FindIndexOf(document.Tasks.Incomplete, func(task Task) bool {
		return task.ID == id
	}); found {
		return task, "incomplete", index, true
	}

	if task, index, found := lo.FindIndexOf(document.Tasks.Complete, func(task Task) bool {
		return task.ID == id
	}); found {
		return task, "complete", index, true
	}

	return Task{}, "", -1, false
}

func removeTask(document taskDocument, bucketName string, index int) taskDocument {
	switch bucketName {
	case "complete":
		document.Tasks.Complete = lo.Reject(document.Tasks.Complete, func(_ Task, taskIndex int) bool {
			return taskIndex == index
		})
	default:
		document.Tasks.Incomplete = lo.Reject(document.Tasks.Incomplete, func(_ Task, taskIndex int) bool {
			return taskIndex == index
		})
	}

	return document
}

func newTaskDocument() taskDocument {
	return taskDocument{
		Tasks: taskBuckets{
			Complete:   []Task{},
			Incomplete: []Task{},
		},
	}
}

func completionLabel(completed bool) string {
	if completed {
		return "complete"
	}

	return "incomplete"
}

func formatUnixTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

func newRandomTaskID() string {
	buffer := make([]byte, 8)
	_, err := rand.Read(buffer)
	if err != nil {
		return fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(buffer)
}

func NewTaskID() string {
	return newRandomTaskID()
}

func (store *Store) taskFilePath(username string) string {
	return filepath.Join(store.tasksDir(), username+".json")
}

func (store *Store) tasksDir() string {
	return filepath.Join(store.dataDir, "tasks")
}

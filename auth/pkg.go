package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

var (
	ErrNotSignedIn  = errors.New("not signed in")
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
)

type Service struct {
	dataDir string
}

type usersDocument struct {
	Users []string `json:"users"`
}

type sessionDocument struct {
	Username string `json:"username"`
}

func NewService(dataDir string) *Service {
	return &Service{dataDir: dataDir}
}

func (service *Service) Register(username string) error {
	if err := validateUsername(username); err != nil {
		return err
	}

	users, err := service.loadUsers()
	if err != nil {
		return err
	}

	if slices.Contains(users.Users, username) {
		return fmt.Errorf("%w: %s", ErrUserExists, username)
	}

	users.Users = append(users.Users, username)

	if err := service.saveUsers(users); err != nil {
		return err
	}

	return service.ensureTaskFile(username)
}

func (service *Service) SignIn(username string) error {
	if err := validateUsername(username); err != nil {
		return err
	}

	users, err := service.loadUsers()
	if err != nil {
		return err
	}

	if !slices.Contains(users.Users, username) {
		return fmt.Errorf("%w: %s", ErrUserNotFound, username)
	}

	if err := service.ensureTaskFile(username); err != nil {
		return err
	}

	return service.writeJSON(service.sessionPath(), sessionDocument{Username: username})
}

func (service *Service) SignOut() error {
	err := os.Remove(service.sessionPath())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove session: %w", err)
	}

	return nil
}

func (service *Service) CurrentUser() (string, error) {
	session, err := service.loadSession()
	if err != nil {
		return "", err
	}

	if session.Username == "" {
		return "", ErrNotSignedIn
	}

	return session.Username, nil
}

func validateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("username %q must be lowercase alphanumeric with optional _ or -", username)
	}

	return nil
}

func (service *Service) loadUsers() (usersDocument, error) {
	if err := service.ensureBaseDirectories(); err != nil {
		return usersDocument{}, err
	}

	path := service.usersPath()
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return usersDocument{}, nil
	}
	if err != nil {
		return usersDocument{}, fmt.Errorf("stat users file: %w", err)
	}

	var users usersDocument
	if err := service.readJSON(path, &users); err != nil {
		return usersDocument{}, err
	}

	return users, nil
}

func (service *Service) saveUsers(users usersDocument) error {
	if err := service.ensureBaseDirectories(); err != nil {
		return err
	}

	return service.writeJSON(service.usersPath(), users)
}

func (service *Service) loadSession() (sessionDocument, error) {
	path := service.sessionPath()
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return sessionDocument{}, ErrNotSignedIn
	}
	if err != nil {
		return sessionDocument{}, fmt.Errorf("stat session file: %w", err)
	}

	var session sessionDocument
	if err := service.readJSON(path, &session); err != nil {
		return sessionDocument{}, err
	}

	return session, nil
}

func (service *Service) ensureTaskFile(username string) error {
	if err := service.ensureBaseDirectories(); err != nil {
		return err
	}

	path := filepath.Join(service.tasksDir(), username+".json")
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat task file: %w", err)
	}

	initialTaskDocument := map[string]any{
		"lastCleanupAt": int64(0),
		"tasks": map[string]any{
			"complete":   []any{},
			"incomplete": []any{},
		},
	}

	return service.writeJSON(path, initialTaskDocument)
}

func (service *Service) ensureBaseDirectories() error {
	if err := os.MkdirAll(service.tasksDir(), 0o755); err != nil {
		return fmt.Errorf("create data directories: %w", err)
	}

	return nil
}

func (service *Service) usersPath() string {
	return filepath.Join(service.dataDir, "users.json")
}

func (service *Service) sessionPath() string {
	return filepath.Join(service.dataDir, "session.json")
}

func (service *Service) tasksDir() string {
	return filepath.Join(service.dataDir, "tasks")
}

func (service *Service) readJSON(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}

	return nil
}

func (service *Service) writeJSON(path string, value any) error {
	content, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

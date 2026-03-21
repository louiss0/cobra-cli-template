package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	projectName := flag.String("project-name", "", "new project name")
	modulePath := flag.String("module-path", "", "new go module path")
	flag.Parse()

	currentModule, err := currentModulePath()
	if err != nil {
		fail(err.Error())
	}

	currentProjectName := filepath.Base(currentModule)
	if currentProjectName == "." || currentProjectName == string(filepath.Separator) {
		fail("could not infer current project name from go.mod module path")
	}

	if strings.TrimSpace(*projectName) == "" {
		inferredProjectName, err := currentDirectoryName()
		if err != nil {
			fail(err.Error())
		}

		*projectName = inferredProjectName
	}

	if strings.TrimSpace(*projectName) == "" {
		fail("project-name is empty")
	}

	if *projectName == currentProjectName && strings.TrimSpace(*modulePath) == "" {
		fail("project-name already matches current module name; nothing to rename")
	}

	if strings.TrimSpace(*modulePath) == "" {
		moduleParts := strings.Split(currentModule, "/")
		moduleParts[len(moduleParts)-1] = *projectName
		*modulePath = strings.Join(moduleParts, "/")
	}

	if *modulePath == currentModule && *projectName == currentProjectName {
		fail("module-path and project-name already match current values")
	}

	replacements := [][2]string{
		{currentModule, *modulePath},
		{currentProjectName, *projectName},
	}

	files, err := listTrackedFiles()
	if err != nil {
		fail(err.Error())
	}

	for _, path := range files {
		if err := rewriteFile(path, replacements); err != nil {
			fail(fmt.Sprintf("rewrite %s: %v", path, err))
		}
	}
}

func currentModulePath() (string, error) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
			if modulePath == "" {
				return "", fmt.Errorf("parse go.mod: module directive is empty")
			}

			return modulePath, nil
		}
	}

	return "", fmt.Errorf("parse go.mod: module directive not found")
}

func currentDirectoryName() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	projectName := filepath.Base(workingDirectory)
	if projectName == "." || projectName == string(filepath.Separator) {
		return "", fmt.Errorf("infer project name from working directory")
	}

	return projectName, nil
}

func listTrackedFiles() ([]string, error) {
	command := exec.Command("git", "ls-files")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list tracked files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		files = append(files, filepath.Clean(line))
	}

	return files, nil
}

func rewriteFile(path string, replacements [][2]string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if bytes.IndexByte(content, 0) >= 0 {
		return nil
	}

	updatedContent := string(content)
	for _, replacement := range replacements {
		updatedContent = strings.ReplaceAll(updatedContent, replacement[0], replacement[1])
	}

	if updatedContent == string(content) {
		return nil
	}

	if err := os.WriteFile(path, []byte(updatedContent), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func fail(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

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

	if strings.TrimSpace(*projectName) == "" {
		fail("project-name is required")
	}

	if strings.TrimSpace(*modulePath) == "" {
		*modulePath = fmt.Sprintf("github.com/louiss0/%s", *projectName)
	}

	replacements := [][2]string{
		{"github.com/louiss0/cobra-cli-template", *modulePath},
		{"cobra-cli-template", *projectName},
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

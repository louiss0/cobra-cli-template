package cmd

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/louiss0/cobra-cli-template/tasks"
	. "github.com/onsi/ginkgo/v2"
	tAssert "github.com/stretchr/testify/assert"
)

var _ = Describe("Task Output", func() {
	It("renders status cards at a fixed width", func() {
		assert := tAssert.New(GinkgoT())

		content := renderTaskWithStatus(tasks.Task{
			Title:       "Write docs",
			Description: "Document the task output styling",
			Completed:   false,
		})

		for _, line := range strings.Split(strings.TrimSuffix(content, "\n"), "\n") {
			assert.Equal(taskDisplayWidth, lipgloss.Width(line))
		}
	})

	It("expands a band for multiline content instead of forcing a height", func() {
		assert := tAssert.New(GinkgoT())

		content := renderTaskBand(
			"line one\nline two\nline three\nline four",
			descriptionBandTheme,
			lipgloss.NewStyle().Bold(true),
		)

		lines := strings.Split(content, "\n")

		assert.Greater(len(lines), 4)
		for _, line := range lines {
			assert.Equal(taskDisplayWidth, lipgloss.Width(line))
		}
	})

	It("renders id cards at a fixed width", func() {
		assert := tAssert.New(GinkgoT())

		content := renderTaskWithID(tasks.Task{
			ID:          "task-1",
			Title:       "Write docs",
			Description: "Document the task output styling",
		})

		for _, line := range strings.Split(strings.TrimSuffix(content, "\n"), "\n") {
			assert.Equal(taskDisplayWidth, lipgloss.Width(line))
		}
	})
})

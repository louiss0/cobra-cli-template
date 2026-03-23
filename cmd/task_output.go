package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var taskTitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")).
	Background(lipgloss.Color("12")).
	Bold(true).
	Padding(1, 3)

var taskStatusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("15")).
	Background(lipgloss.Color("1")).
	Bold(true).
	Padding(1, 2)

func writeStyledTaskOutput(cmd *cobra.Command, task tasks.Task) error {
	content := renderTaskWithStatus(task)
	if err := writeTaskOutput(cmd, content); err != nil {
		return err
	}

	return nil
}

func writeCreatedTaskOutput(cmd *cobra.Command, task tasks.Task) error {
	content := renderTaskWithID(task)
	if err := writeTaskOutput(cmd, content); err != nil {
		return err
	}

	return nil
}

func writeTaskOutput(cmd *cobra.Command, content string) error {
	_, err := fmt.Fprint(cmd.OutOrStdout(), content)
	if err != nil {
		return fmt.Errorf("write task output: %w", err)
	}

	return nil
}

func renderTaskWithStatus(task tasks.Task) string {
	status := "incomplete"
	if task.Completed {
		status = "complete"
	}

	return renderCenteredTask(task.Title, task.Description, "status", taskStatusStyle.Render(status))
}

func renderTaskWithID(task tasks.Task) string {
	return renderCenteredTask(task.Title, task.Description, "id", task.ID)
}

func renderCenteredTask(title string, description string, footerTitle string, footerValue string) string {
	titleText := taskTitleStyle.Render(title)
	descriptionText := pterm.DefaultParagraph.WithMaxWidth(60).Sprint(description)
	footerText := strings.ToUpper(footerTitle) + ": " + footerValue

	content := strings.Join([]string{
		titleText,
		"",
		descriptionText,
		"",
		footerText,
	}, "\n")

	return pterm.DefaultCenter.WithCenterEachLineSeparately().Sprint(content)
}

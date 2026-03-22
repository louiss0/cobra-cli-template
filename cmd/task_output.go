package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/spf13/cobra"
)

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
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	descriptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	statusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	status := "incomplete"
	if task.Completed {
		status = "complete"
		statusStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	}

	builder := strings.Builder{}
	builder.WriteString(titleStyle.Render(task.Title))
	builder.WriteString("\n")
	builder.WriteString(descriptionStyle.Render(task.Description))
	builder.WriteString("\n")
	builder.WriteString(statusStyle.Render(status))
	builder.WriteString("\n")

	return builder.String()
}

func renderTaskWithID(task tasks.Task) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	descriptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	idStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))

	builder := strings.Builder{}
	builder.WriteString(titleStyle.Render(task.Title))
	builder.WriteString("\n")
	builder.WriteString(descriptionStyle.Render(task.Description))
	builder.WriteString("\n")
	builder.WriteString(idStyle.Render(task.ID))
	builder.WriteString("\n")

	return builder.String()
}

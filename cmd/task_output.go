package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/spf13/cobra"
)

const taskDisplayWidth = 60

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
	statusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0"))
	if task.Completed {
		status = "complete"
		statusStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0"))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderTaskBand(task.Title, lipgloss.Color("210"), lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")), 5, 1, 8, lipgloss.Center),
		renderTaskBand(task.Description, lipgloss.Color("228"), lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")), 7, 2, 5, lipgloss.Left),
		renderTaskBand(status, lipgloss.Color("153"), statusStyle, 4, 1, 3, lipgloss.Left),
	) + "\n"
}

func renderTaskWithID(task tasks.Task) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderTaskBand(task.Title, lipgloss.Color("210"), lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")), 5, 1, 8, lipgloss.Center),
		renderTaskBand(task.Description, lipgloss.Color("228"), lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")), 7, 2, 5, lipgloss.Left),
		renderTaskBand(task.ID, lipgloss.Color("153"), lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")), 4, 1, 3, lipgloss.Left),
	) + "\n"
}

func renderTaskBand(
	value string,
	backgroundColor lipgloss.Color,
	textStyle lipgloss.Style,
	height int,
	paddingTop int,
	paddingX int,
	align lipgloss.Position,
) string {
	bandStyle := lipgloss.NewStyle().
		Width(taskDisplayWidth).
		Height(height).
		Padding(paddingTop, paddingX, 0, paddingX).
		Background(backgroundColor)

	contentStyle := textStyle.
		Width(taskDisplayWidth - (paddingX * 2)).
		Background(backgroundColor).
		Align(align)

	return bandStyle.Render(contentStyle.Render(value))
}

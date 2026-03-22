package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/louiss0/cobra-cli-template/tasks"
	"github.com/spf13/cobra"
)

const taskDisplayWidth = 35
const taskContentWidth = taskDisplayWidth - 4

type taskBandTheme struct {
	Background lipgloss.TerminalColor
	Foreground lipgloss.TerminalColor
}

var (
	titleBandTheme = taskBandTheme{
		Background: lipgloss.AdaptiveColor{Light: "#DCE7F4", Dark: "#223247"},
		Foreground: lipgloss.AdaptiveColor{Light: "#102A43", Dark: "#EAF2FF"},
	}
	descriptionBandTheme = taskBandTheme{
		Background: lipgloss.AdaptiveColor{Light: "#F2F5F7", Dark: "#273947"},
		Foreground: lipgloss.AdaptiveColor{Light: "#1F2933", Dark: "#F7FAFC"},
	}
	incompleteBandTheme = taskBandTheme{
		Background: lipgloss.AdaptiveColor{Light: "#FBE4D5", Dark: "#5C3520"},
		Foreground: lipgloss.AdaptiveColor{Light: "#5C2E12", Dark: "#FFF3E8"},
	}
	completeBandTheme = taskBandTheme{
		Background: lipgloss.AdaptiveColor{Light: "#D9F4E5", Dark: "#1F5134"},
		Foreground: lipgloss.AdaptiveColor{Light: "#0F3B21", Dark: "#E8FFF1"},
	}
	idBandTheme = taskBandTheme{
		Background: lipgloss.AdaptiveColor{Light: "#E4EBF5", Dark: "#31465E"},
		Foreground: lipgloss.AdaptiveColor{Light: "#102A43", Dark: "#F4F8FC"},
	}
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
	status := "incomplete"
	statusTheme := incompleteBandTheme
	if task.Completed {
		status = "complete"
		statusTheme = completeBandTheme
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderTaskBand(task.Title, titleBandTheme, lipgloss.NewStyle().Bold(true)),
		renderTaskBand(task.Description, descriptionBandTheme, lipgloss.NewStyle()),
		renderTaskBand(status, statusTheme, lipgloss.NewStyle().Bold(true)),
	) + "\n"
}

func renderTaskWithID(task tasks.Task) string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		renderTaskBand(task.Title, titleBandTheme, lipgloss.NewStyle().Bold(true)),
		renderTaskBand(task.Description, descriptionBandTheme, lipgloss.NewStyle()),
		renderTaskBand(task.ID, idBandTheme, lipgloss.NewStyle().Bold(true)),
	) + "\n"
}

func renderTaskBand(
	value string,
	theme taskBandTheme,
	textStyle lipgloss.Style,
) string {
	bandStyle := lipgloss.NewStyle().
		Width(taskDisplayWidth).
		Padding(1, 2).
		Background(theme.Background)

	contentStyle := textStyle.
		Foreground(theme.Foreground).
		Background(theme.Background).
		Width(taskContentWidth)

	return bandStyle.Render(contentStyle.Render(value))
}

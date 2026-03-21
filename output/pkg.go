package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/louiss0/g-tools/mode"
	"github.com/neilotoole/jsoncolor"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var modeOperator = mode.NewModeOperator()

func WriteModeAwareOutput(command *cobra.Command, message string, keyvals ...any) error {
	if modeOperator.IsProductionMode() {
		logger := log.New(command.OutOrStdout())
		logger.Info(message, keyvals...)
		return nil
	}

	chuncked := lo.Chunk(keyvals, 2)

	formattedPairs := lo.Map(chuncked, func(pair []any, _ int) string {
		return fmt.Sprintf("%s=%v", pair[0], pair[1])
	})

	finalString := strings.Join(append([]string{message}, formattedPairs...), " ")

	_, err := fmt.Fprint(command.OutOrStdout(), finalString)
	if err != nil {
		return fmt.Errorf("write development output: %w", err)
	}

	return nil
}

func WriteJSONOutput(command *cobra.Command, value any) error {
	if modeOperator.IsProductionMode() {
		encoder := jsoncolor.NewEncoder(command.OutOrStdout())
		encoder.SetIndent("", "  ")
		if jsoncolor.IsColorTerminal(command.OutOrStdout()) {
			encoder.SetColors(jsoncolor.DefaultColors())
		}

		if err := encoder.Encode(value); err != nil {
			return fmt.Errorf("encode output: %w", err)
		}

		return nil
	}

	content, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	_, err = fmt.Fprint(command.OutOrStdout(), string(content))
	if err != nil {
		return fmt.Errorf("write development json output: %w", err)
	}

	return nil
}

func WriteModeAwareError(stderr io.Writer, err error) error {
	if err == nil {
		return nil
	}

	logger := log.New(stderr)
	logger.Error("command failed", "error", err)
	return nil
}

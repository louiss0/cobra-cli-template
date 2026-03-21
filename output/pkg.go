package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/louiss0/g-tools/mode"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var modeOperator = mode.NewModeOperator()

func WriteModeAwareOutput(command *cobra.Command, message string, keyvals ...any) error {
	if modeOperator.IsProductionMode() {
		log.Info(message, keyvals...)
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
	content, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return WriteModeAwareOutput(command, string(content))
}

package cli

import "github.com/spf13/cobra"

func resolveOutputFormat(cmd *cobra.Command, resolveConfigApp ConfigResolver, jsonOut bool, outputFlag string) (string, error) {
	if jsonOut {
		return "json", nil
	}
	if outputFlag != "" && cmd.Flags().Changed(outputFlag) {
		value, err := cmd.Flags().GetString(outputFlag)
		if err != nil {
			return "", err
		}
		return value, nil
	}
	if resolveConfigApp == nil {
		return "table", nil
	}
	configApp, err := resolveConfigApp()
	if err != nil {
		return "", err
	}
	value, err := configApp.Get(cmd.Context(), "output.default")
	if err != nil {
		return "", err
	}
	if value == "" {
		return "table", nil
	}
	return value, nil
}

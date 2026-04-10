package cli

import (
	"errors"
	"fmt"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/spf13/cobra"
)

func NewConfigCommand(resolveConfigApp ConfigResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write local tick configuration",
	}

	get := &cobra.Command{
		Use:   "get <key>",
		Short: "Print a config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveConfigApp == nil {
				return errors.New("config command is unavailable")
			}
			configApp, err := resolveConfigApp()
			if err != nil {
				return err
			}
			value, err := configApp.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, value)
			return err
		},
	}

	set := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Persist a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveConfigApp == nil {
				return errors.New("config command is unavailable")
			}
			configApp, err := resolveConfigApp()
			if err != nil {
				return err
			}
			return configApp.Set(cmd.Context(), args[0], args[1])
		},
	}

	cmd.AddCommand(get, set)
	return cmd
}

type ConfigResolver func() (*app.ConfigApp, error)

package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/spf13/cobra"
)

func NewAuthCommand(resolveAuthApp AuthResolver, resolveAuthService AuthServiceResolver, resolveRegion RegionResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with TickTick",
	}

	login := &cobra.Command{
		Use:   "login",
		Short: "Start the TickTick OAuth login flow",
		Long:  "Start the TickTick OAuth login flow interactively. The CLI will prompt for region, client ID, and client secret, then try to capture localhost callbacks automatically and fall back to manual callback URL paste when needed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveAuthApp == nil {
				return errors.New("auth login is unavailable")
			}
			if !IsTerminal(streams) {
				return errors.New("auth login requires an interactive terminal")
			}
			authApp, err := resolveAuthApp()
			if err != nil {
				return err
			}

			cfg := app.LoginInput{}
			if authApp.ConfigStore != nil {
				loaded, err := authApp.ConfigStore.Load()
				if err == nil {
					cfg.Region = loaded.Region
					cfg.ClientID = loaded.ClientID
					cfg.ClientSecret = loaded.ClientSecret
				}
			}

			region, err := SelectRegion(streams, cfg.Region)
			if err != nil {
				return err
			}
			clientID, err := Prompt(streams, "Client ID", cfg.ClientID)
			if err != nil {
				return err
			}
			clientSecret, err := Prompt(streams, "Client Secret", cfg.ClientSecret)
			if err != nil {
				return err
			}

			if err := authApp.Login(cmd.Context(), app.LoginInput{
				Region:       region,
				ClientID:     clientID,
				ClientSecret: clientSecret,
			}); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Login successful")
			return err
		},
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveAuthService == nil {
				return errors.New("auth status is unavailable")
			}
			authService, err := resolveAuthService()
			if err != nil {
				return err
			}
			status, err := authService.Status(cmd.Context())
			if err != nil {
				return err
			}
			if status.Authenticated {
				if _, err = fmt.Fprintln(streams.Out, "authenticated"); err != nil {
					return err
				}
			} else {
				if _, err = fmt.Fprintln(streams.Out, "not authenticated"); err != nil {
					return err
				}
			}
			region := "ticktick"
			if resolveRegion != nil {
				region, err = resolveRegion()
				if err != nil {
					return err
				}
			}
			if _, err = fmt.Fprintf(streams.Out, "region: %s\n", region); err != nil {
				return err
			}
			if status.Authenticated && status.ExpiryKnown {
				if _, err = fmt.Fprintf(streams.Out, "expires_at: %s\n", time.Unix(status.ExpiresAtUnix, 0).UTC().Format(time.RFC3339)); err != nil {
					return err
				}
				if status.Expired {
					_, err = fmt.Fprintln(streams.Out, "expires_in: expired")
					return err
				}
				_, err = fmt.Fprintf(streams.Out, "expires_in: %ds\n", status.ExpiresInSeconds)
				return err
			}
			if status.Authenticated {
				_, err = fmt.Fprintln(streams.Out, "expires_at: unknown")
				return err
			}
			return nil
		},
	}

	logout := &cobra.Command{
		Use:   "logout",
		Short: "Delete stored TickTick credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveAuthService == nil {
				return errors.New("auth logout is unavailable")
			}
			authService, err := resolveAuthService()
			if err != nil {
				return err
			}
			if err := authService.Logout(cmd.Context()); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Logged out")
			return err
		},
	}

	cmd.AddCommand(login, status, logout)
	return cmd
}

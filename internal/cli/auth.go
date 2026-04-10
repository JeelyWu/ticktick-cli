package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/spf13/cobra"
)

func NewAuthCommand(resolveAuthApp AuthResolver, resolveAuthService AuthServiceResolver, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with TickTick",
	}

	var clientID string
	var clientSecret string
	var redirectURL string
	login := &cobra.Command{
		Use:   "login",
		Short: "Start the TickTick OAuth login flow",
		Long:  "Start the TickTick OAuth login flow. Prefer setting TICK_CLIENT_SECRET instead of passing secrets on the command line.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if resolveAuthApp == nil {
				return errors.New("auth login is unavailable")
			}
			authApp, err := resolveAuthApp()
			if err != nil {
				return err
			}
			loginSecret := clientSecret
			if loginSecret == "" {
				loginSecret = os.Getenv("TICK_CLIENT_SECRET")
			}
			if err := authApp.Login(cmd.Context(), app.LoginInput{
				ClientID:     clientID,
				ClientSecret: loginSecret,
				RedirectURL:  redirectURL,
			}); err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "Login successful")
			return err
		},
	}
	login.Flags().StringVar(&clientID, "client-id", "", "TickTick OAuth client ID")
	login.Flags().StringVar(&clientSecret, "client-secret", "", "TickTick OAuth client secret (defaults to TICK_CLIENT_SECRET)")
	login.Flags().StringVar(&redirectURL, "redirect-url", "", "TickTick OAuth redirect URL")
	_ = login.Flags().MarkDeprecated("client-secret", "prefer TICK_CLIENT_SECRET to avoid exposing secrets in shell history")
	_ = login.Flags().MarkHidden("client-secret")

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
				_, err = fmt.Fprintln(streams.Out, "authenticated")
				return err
			}
			_, err = fmt.Fprintln(streams.Out, "not authenticated")
			return err
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

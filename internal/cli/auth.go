package cli

import (
	"fmt"
	"os"

	"github.com/jeely/ticktick-cli/internal/app"
	"github.com/spf13/cobra"
)

func NewAuthCommand(authApp *app.AuthApp, streams Streams) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
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
			_, err := fmt.Fprintln(streams.Out, "Login successful")
			return err
		},
	}
	login.Flags().StringVar(&clientID, "client-id", "", "TickTick OAuth client ID")
	login.Flags().StringVar(&clientSecret, "client-secret", "", "TickTick OAuth client secret (defaults to TICK_CLIENT_SECRET)")
	login.Flags().StringVar(&redirectURL, "redirect-url", "", "TickTick OAuth redirect URL")

	status := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := authApp.Status(cmd.Context())
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
		Short: "Delete the stored TickTick token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := authApp.Logout(cmd.Context()); err != nil {
				return err
			}
			_, err := fmt.Fprintln(streams.Out, "Logged out")
			return err
		},
	}

	cmd.AddCommand(login, status, logout)
	return cmd
}

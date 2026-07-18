package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rie03p/appsflyer-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage the stored API token",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthStatusCmd(), newAuthLogoutCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save an API V2 token so you don't have to pass it on every call",
		Long: `Save an AppsFlyer API V2 token to the local config file.

An account admin can retrieve the token from the AppsFlyer dashboard
under Settings > API tokens. Once saved, all commands use it
automatically; --token and APPSFLYER_API_TOKEN still take precedence.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				var err error
				token, err = readToken(cmd)
				if err != nil {
					return err
				}
			}
			if token == "" {
				return fmt.Errorf("empty token")
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.Token = token
			path, err := config.Save(cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Token saved to %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "token to save (prompted for if omitted)")
	return cmd
}

// readToken hides input when stdin is a terminal so the token doesn't
// end up in scrollback; otherwise it reads a line (piped input, tests).
func readToken(cmd *cobra.Command) (string, error) {
	fmt.Fprint(cmd.ErrOrStderr(), "Paste your AppsFlyer API V2 token: ")
	if f, ok := cmd.InOrStdin().(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show which token would be used and where it comes from",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if t := os.Getenv("APPSFLYER_API_TOKEN"); t != "" {
				fmt.Fprintf(out, "Using token from APPSFLYER_API_TOKEN (%s)\n", mask(t))
				return nil
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.Token != "" {
				path, _ := config.Path()
				fmt.Fprintf(out, "Using token from %s (%s)\n", path, mask(cfg.Token))
				return nil
			}
			fmt.Fprintln(out, "Not logged in. Run: afcli auth login")
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Delete the stored token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.Token == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "No stored token.")
				return nil
			}
			cfg.Token = ""
			if _, err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Stored token deleted.")
			return nil
		},
	}
}

func mask(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:8] + "..."
}

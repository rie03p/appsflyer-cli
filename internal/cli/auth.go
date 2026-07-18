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
	var (
		token   string
		onelink bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save an API token so you don't have to pass it on every call",
		Long: `Save an AppsFlyer API token to the local config file.

By default this stores the API V2 (reporting) token; pass --onelink to
store the OneLink API token instead, which is a separate credential.
An account admin can retrieve both from the AppsFlyer dashboard. Once
saved, all commands use them automatically; the corresponding flags
and environment variables still take precedence.`,
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
			if onelink {
				cfg.OneLinkToken = token
			} else {
				cfg.Token = token
			}
			path, err := config.Save(cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Token saved to %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "token to save (prompted for if omitted)")
	cmd.Flags().BoolVar(&onelink, "onelink", false, "store the OneLink API token instead of the reporting token")
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
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			path, _ := config.Path()

			switch {
			case os.Getenv("APPSFLYER_API_TOKEN") != "":
				fmt.Fprintf(out, "Reporting token: APPSFLYER_API_TOKEN (%s)\n", mask(os.Getenv("APPSFLYER_API_TOKEN")))
			case cfg.Token != "":
				fmt.Fprintf(out, "Reporting token: %s (%s)\n", path, mask(cfg.Token))
			default:
				fmt.Fprintln(out, "Reporting token: not set. Run: afcli auth login")
			}

			switch {
			case os.Getenv("ONELINK_API_TOKEN") != "":
				fmt.Fprintf(out, "OneLink token:   ONELINK_API_TOKEN (%s)\n", mask(os.Getenv("ONELINK_API_TOKEN")))
			case cfg.OneLinkToken != "":
				fmt.Fprintf(out, "OneLink token:   %s (%s)\n", path, mask(cfg.OneLinkToken))
			default:
				fmt.Fprintln(out, "OneLink token:   not set. Run: afcli auth login --onelink")
			}
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Delete the stored tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.Token == "" && cfg.OneLinkToken == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "No stored tokens.")
				return nil
			}
			cfg.Token = ""
			cfg.OneLinkToken = ""
			if _, err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Stored tokens deleted.")
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

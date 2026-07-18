// Package cli implements the afcli command tree.
package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rie03p/appsflyer-cli/internal/appsflyer"
	"github.com/rie03p/appsflyer-cli/internal/config"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

type rootOptions struct {
	token   string
	baseURL string
	timeout time.Duration
	output  string
}

// NewRootCmd builds the afcli command tree.
func NewRootCmd() *cobra.Command {
	opts := &rootOptions{}

	cmd := &cobra.Command{
		Use:   "afcli",
		Short: "Command-line client for the AppsFlyer reporting APIs",
		Long: `afcli fetches AppsFlyer reports from the command line.

It covers the Pull API (raw and aggregate data) and the Master API.
Authentication uses an AppsFlyer API V2 token: run "afcli auth login"
once to store it, or pass --token / set APPSFLYER_API_TOKEN.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.token, "token", "", "AppsFlyer API V2 token (overrides $APPSFLYER_API_TOKEN and the token stored by auth login)")
	cmd.PersistentFlags().StringVar(&opts.baseURL, "base-url", appsflyer.DefaultBaseURL, "API base URL")
	cmd.PersistentFlags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "HTTP request timeout")
	cmd.PersistentFlags().StringVarP(&opts.output, "output", "o", "", "write the report to a file instead of stdout")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newRawCmd(opts))
	cmd.AddCommand(newAggCmd(opts))
	cmd.AddCommand(newMasterCmd(opts))
	cmd.AddCommand(newCohortCmd(opts))
	cmd.AddCommand(newSKANCmd(opts))
	cmd.AddCommand(newFreshnessCmd(opts))

	return cmd
}

func (o *rootOptions) client() (*appsflyer.Client, error) {
	token := o.token
	if token == "" {
		token = os.Getenv("APPSFLYER_API_TOKEN")
	}
	if token == "" {
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		token = cfg.Token
	}
	if token == "" {
		return nil, fmt.Errorf("no API token: run \"afcli auth login\", pass --token, or set APPSFLYER_API_TOKEN")
	}
	return appsflyer.New(token,
		appsflyer.WithBaseURL(o.baseURL),
		appsflyer.WithTimeout(o.timeout),
	), nil
}

func (o *rootOptions) write(cmd *cobra.Command, body io.ReadCloser) error {
	defer body.Close()
	out := cmd.OutOrStdout()
	if o.output != "" {
		f, err := os.Create(o.output)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}
	if _, err := io.Copy(out, body); err != nil {
		return err
	}
	if o.output != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", o.output)
	}
	return nil
}

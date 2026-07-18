// Package cli implements the afcli command tree.
package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rie03p/appsflyer-cli/internal/appsflyer"
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
Authentication uses an AppsFlyer API V2 token, taken from --token or
the APPSFLYER_API_TOKEN environment variable.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.token, "token", "", "AppsFlyer API V2 token (defaults to $APPSFLYER_API_TOKEN)")
	cmd.PersistentFlags().StringVar(&opts.baseURL, "base-url", appsflyer.DefaultBaseURL, "API base URL")
	cmd.PersistentFlags().DurationVar(&opts.timeout, "timeout", 5*time.Minute, "HTTP request timeout")
	cmd.PersistentFlags().StringVarP(&opts.output, "output", "o", "", "write the report to a file instead of stdout")

	cmd.AddCommand(newRawCmd(opts))
	cmd.AddCommand(newAggCmd(opts))
	cmd.AddCommand(newMasterCmd(opts))

	return cmd
}

func (o *rootOptions) client() (*appsflyer.Client, error) {
	token := o.token
	if token == "" {
		token = os.Getenv("APPSFLYER_API_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("no API token: pass --token or set APPSFLYER_API_TOKEN")
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

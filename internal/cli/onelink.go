package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/rie03p/appsflyer-cli/internal/appsflyer"
	"github.com/rie03p/appsflyer-cli/internal/config"
	"github.com/spf13/cobra"
)

type onelinkOptions struct {
	root  *rootOptions
	token string
}

func (o *onelinkOptions) client(timeout time.Duration) (*appsflyer.OneLinkClient, error) {
	token := o.token
	if token == "" {
		token = os.Getenv("ONELINK_API_TOKEN")
	}
	if token == "" {
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		token = cfg.OneLinkToken
	}
	if token == "" {
		return nil, fmt.Errorf("no OneLink API token: run \"afcli auth login --onelink\", pass --onelink-token, or set ONELINK_API_TOKEN")
	}
	return appsflyer.NewOneLink(token, appsflyer.WithOneLinkTimeout(timeout)), nil
}

func newOneLinkCmd(opts *rootOptions) *cobra.Command {
	olOpts := &onelinkOptions{root: opts}

	cmd := &cobra.Command{
		Use:   "onelink",
		Short: "Create and manage OneLink short links (API v2.0)",
		Long: `Create, fetch, update, and delete OneLink short links.

The OneLink API uses its own token (separate from the reporting API V2
token); an admin can retrieve it from the AppsFlyer dashboard. The
<onelink-id> argument is the template ID from the OneLink template
screen, e.g. "abc123" in "myapp.onelink.me/abc123/qwer9876".`,
	}
	cmd.PersistentFlags().StringVar(&olOpts.token, "onelink-token", "", "OneLink API token (defaults to $ONELINK_API_TOKEN, then the stored token)")

	cmd.AddCommand(newOneLinkCreateCmd(olOpts))
	cmd.AddCommand(newOneLinkGetCmd(olOpts))
	cmd.AddCommand(newOneLinkUpdateCmd(olOpts))
	cmd.AddCommand(newOneLinkDeleteCmd(olOpts))
	return cmd
}

func newOneLinkCreateCmd(opts *onelinkOptions) *cobra.Command {
	var (
		params    appsflyer.ShortlinkParams
		dataPairs []string
	)

	cmd := &cobra.Command{
		Use:   "create <onelink-id>",
		Short: "Create a short link",
		Example: `  afcli onelink create abc123 --param pid=email --param c=summer_sale
  afcli onelink create abc123 --shortlink-id promo2026 --ttl 90d \
    --param pid=sms --param deep_link_value=coupons`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			params.Data, err = parsePairs(dataPairs, "--param")
			if err != nil {
				return err
			}
			client, err := opts.client(opts.root.timeout)
			if err != nil {
				return err
			}
			body, err := client.CreateShortlink(cmd.Context(), args[0], params)
			if err != nil {
				return err
			}
			return opts.root.write(cmd, body)
		},
	}

	cmd.Flags().StringArrayVar(&dataPairs, "param", nil, "attribution parameter as key=value (pid is required); repeatable")
	cmd.Flags().StringVar(&params.ShortlinkID, "shortlink-id", "", "custom short link ID (random if omitted)")
	cmd.Flags().StringVar(&params.TTL, "ttl", "", "link time to live, e.g. 30m, 12h, 90d (default 31d, max 730d)")
	cmd.Flags().StringVar(&params.BrandDomain, "brand-domain", "", "branded domain (Branded Links feature required)")
	cmd.Flags().BoolVar(&params.RenewTTL, "renew-ttl", false, "extend the TTL on every click")
	return cmd
}

func newOneLinkGetCmd(opts *onelinkOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "get <onelink-id> <shortlink-id>",
		Short:   "Fetch a short link's parameters",
		Example: `  afcli onelink get abc123 qwer9876`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client(opts.root.timeout)
			if err != nil {
				return err
			}
			body, err := client.GetShortlink(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			return opts.root.write(cmd, body)
		},
	}
}

func newOneLinkUpdateCmd(opts *onelinkOptions) *cobra.Command {
	var (
		params    appsflyer.ShortlinkParams
		dataPairs []string
	)

	cmd := &cobra.Command{
		Use:     "update <onelink-id> <shortlink-id>",
		Short:   "Replace a short link's parameters",
		Example: `  afcli onelink update abc123 qwer9876 --param pid=email --param c=autumn_sale`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			params.Data, err = parsePairs(dataPairs, "--param")
			if err != nil {
				return err
			}
			client, err := opts.client(opts.root.timeout)
			if err != nil {
				return err
			}
			body, err := client.UpdateShortlink(cmd.Context(), args[0], args[1], params)
			if err != nil {
				return err
			}
			return opts.root.write(cmd, body)
		},
	}

	cmd.Flags().StringArrayVar(&dataPairs, "param", nil, "attribution parameter as key=value; repeatable (required)")
	cmd.Flags().StringVar(&params.TTL, "ttl", "", "link time to live, e.g. 30m, 12h, 90d")
	cmd.Flags().StringVar(&params.BrandDomain, "brand-domain", "", "branded domain")
	cmd.Flags().BoolVar(&params.RenewTTL, "renew-ttl", false, "extend the TTL on every click")
	return cmd
}

func newOneLinkDeleteCmd(opts *onelinkOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <onelink-id> <shortlink-id>",
		Short:   "Delete a short link",
		Example: `  afcli onelink delete abc123 qwer9876`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client(opts.root.timeout)
			if err != nil {
				return err
			}
			if err := client.DeleteShortlink(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Deleted.")
			return nil
		},
	}
}

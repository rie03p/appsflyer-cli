package cli

import (
	"fmt"
	"sort"

	"github.com/rie03p/appsflyer-cli/internal/appsflyer"
	"github.com/spf13/cobra"
)

func newRawCmd(opts *rootOptions) *cobra.Command {
	var (
		appID  string
		params appsflyer.RawDataParams
	)

	cmd := &cobra.Command{
		Use:   "raw <report>",
		Short: "Fetch a Pull API raw data report (CSV)",
		Example: `  afcli raw installs_report --app id123456789 --from 2026-07-01 --to 2026-07-07
  afcli raw in_app_events_report --app id123456789 --from 2026-07-01 --to 2026-07-07 \
    --event-name af_purchase --geo JP -o events.csv
  afcli raw list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}
			body, err := client.RawData(cmd.Context(), appID, args[0], params)
			if err != nil {
				return err
			}
			return opts.write(cmd, body)
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "app ID (required)")
	cmd.MarkFlagRequired("app")
	cmd.Flags().StringVar(&params.From, "from", "", `start date, "yyyy-mm-dd" or "yyyy-mm-dd hh:mm:ss" (required)`)
	cmd.MarkFlagRequired("from")
	cmd.Flags().StringVar(&params.To, "to", "", "end date (required)")
	cmd.MarkFlagRequired("to")
	cmd.Flags().StringVar(&params.MediaSource, "media-source", "", "filter by media source")
	cmd.Flags().StringVar(&params.Category, "category", "", "media source category (standard, facebook, twitter)")
	cmd.Flags().StringSliceVar(&params.EventName, "event-name", nil, "filter by in-app event names")
	cmd.Flags().StringVar(&params.Timezone, "timezone", "", "timezone for the data (default UTC)")
	cmd.Flags().StringVar(&params.Geo, "geo", "", "filter by country code")
	cmd.Flags().StringVar(&params.Currency, "currency", "", "revenue currency (preferred or USD)")
	cmd.Flags().IntVar(&params.MaximumRows, "max-rows", 0, "limit the number of rows")
	cmd.Flags().StringSliceVar(&params.AdditionalFields, "additional-fields", nil, "extra columns to include")

	cmd.AddCommand(newRawListCmd())
	return cmd
}

func newRawListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available raw data report names",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			groups := make([]string, 0, len(appsflyer.RawReports))
			for g := range appsflyer.RawReports {
				groups = append(groups, g)
			}
			sort.Strings(groups)
			for _, g := range groups {
				fmt.Fprintf(cmd.OutOrStdout(), "%s:\n", g)
				for _, r := range appsflyer.RawReports[g] {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", r)
				}
			}
		},
	}
}

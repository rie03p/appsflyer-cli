package cli

import (
	"fmt"
	"sort"

	"github.com/rie03p/appsflyer-cli/internal/appsflyer"
	"github.com/spf13/cobra"
)

func newAggCmd(opts *rootOptions) *cobra.Command {
	var (
		appID  string
		params appsflyer.AggregateParams
	)

	cmd := &cobra.Command{
		Use:   "agg <report>",
		Short: "Fetch a Pull API aggregate data report",
		Example: `  afcli agg partners_report --app id123456789 --from 2026-07-01 --to 2026-07-07
  afcli agg geo_by_date_report --app id123456789 --from 2026-07-01 --to 2026-07-07 --format json
  afcli agg list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}
			body, err := client.Aggregate(cmd.Context(), appID, args[0], params)
			if err != nil {
				return err
			}
			return opts.write(cmd, body)
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "app ID (required)")
	cmd.MarkFlagRequired("app")
	cmd.Flags().StringVar(&params.From, "from", "", `start date, "yyyy-mm-dd" (required)`)
	cmd.MarkFlagRequired("from")
	cmd.Flags().StringVar(&params.To, "to", "", "end date (required)")
	cmd.MarkFlagRequired("to")
	cmd.Flags().StringVar(&params.MediaSource, "media-source", "", "filter by media source")
	cmd.Flags().StringVar(&params.Category, "category", "", "media source category (standard, facebook, twitter)")
	cmd.Flags().StringVar(&params.AttributionTouchType, "attribution-touch-type", "", `set to "impression" for view-through attribution`)
	cmd.Flags().StringVar(&params.Currency, "currency", "", "revenue currency (preferred or USD)")
	cmd.Flags().StringVar(&params.Timezone, "timezone", "", "timezone for the data (default UTC)")
	cmd.Flags().BoolVar(&params.Reattr, "reattr", false, "fetch retargeting conversions instead of UA")
	cmd.Flags().StringVar(&params.Format, "format", "", "response format: csv (default) or json")

	cmd.AddCommand(newAggListCmd())
	return cmd
}

func newAggListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available aggregate report names",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			names := make([]string, 0, len(appsflyer.AggReports))
			for n := range appsflyer.AggReports {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				fmt.Fprintf(cmd.OutOrStdout(), "%-24s %s\n", n, appsflyer.AggReports[n])
			}
		},
	}
}

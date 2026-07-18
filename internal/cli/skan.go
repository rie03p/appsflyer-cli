package cli

import (
	"github.com/rie03p/appsflyer-cli/internal/appsflyer"
	"github.com/spf13/cobra"
)

func newSKANCmd(opts *rootOptions) *cobra.Command {
	var (
		appID  string
		params appsflyer.SKANParams
	)

	cmd := &cobra.Command{
		Use:   "skan",
		Short: "Fetch the SKAN aggregated performance report (iOS SKAdNetwork)",
		Example: `  afcli skan --app id123456789 --start-date 2026-07-01 --end-date 2026-07-07
  afcli skan --app id123456789 --start-date 2026-07-01 --end-date 2026-07-07 \
    --version v2 --view-type acquisition --modeled`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}
			body, err := client.SKAN(cmd.Context(), appID, params)
			if err != nil {
				return err
			}
			return opts.write(cmd, body)
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "app ID (required)")
	cmd.MarkFlagRequired("app")
	cmd.Flags().StringVar(&params.StartDate, "start-date", "", `install date range start, "yyyy-mm-dd" (required; range max 90 days)`)
	cmd.MarkFlagRequired("start-date")
	cmd.Flags().StringVar(&params.EndDate, "end-date", "", "install date range end (required)")
	cmd.MarkFlagRequired("end-date")
	cmd.Flags().StringVar(&params.Version, "version", "v3", "v3 for SKAN 4 postbacks, v2 for SKAN 3")
	cmd.Flags().StringVar(&params.DateType, "date-type", "", "install (default) or arrival")
	cmd.Flags().StringVar(&params.ViewType, "view-type", "", "unified (default), acquisition, or retargeting")
	cmd.Flags().BoolVar(&params.Modeled, "modeled", false, "include modeled conversion values (v2 only)")

	return cmd
}

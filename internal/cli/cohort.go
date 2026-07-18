package cli

import (
	"strings"

	"github.com/rie03p/appsflyer-cli/internal/appsflyer"
	"github.com/spf13/cobra"
)

func newCohortCmd(opts *rootOptions) *cobra.Command {
	var (
		appID   string
		params  appsflyer.CohortParams
		filters []string
	)

	cmd := &cobra.Command{
		Use:   "cohort",
		Short: "Fetch a Cohort API report (retention, LTV, ROAS by cohort day)",
		Example: `  afcli cohort --app id123456789 --from 2026-06-01 --to 2026-06-30 \
    --cohort-type user_acquisition --kpis users,roas --groupings pid,geo \
    --aggregation-type cumulative
  afcli cohort --app id123456789 --from 2026-06-01 --to 2026-06-30 \
    --cohort-type unified --kpis retention --groupings pid \
    --aggregation-type on_day --filter period=0,1,7,30 --format csv`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(filters) > 0 {
				params.Filters = make(map[string][]string, len(filters))
				for _, f := range filters {
					key, val, ok := strings.Cut(f, "=")
					if !ok || key == "" || val == "" {
						return errKeyValue("--filter", f)
					}
					params.Filters[key] = strings.Split(val, ",")
				}
			}
			client, err := opts.client()
			if err != nil {
				return err
			}
			body, err := client.Cohort(cmd.Context(), appID, params)
			if err != nil {
				return err
			}
			return opts.write(cmd, body)
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "app ID (required)")
	cmd.MarkFlagRequired("app")
	cmd.Flags().StringVar(&params.From, "from", "", `cohort start date, "yyyy-mm-dd" (required; up to 720 days back)`)
	cmd.MarkFlagRequired("from")
	cmd.Flags().StringVar(&params.To, "to", "", "cohort end date (required)")
	cmd.MarkFlagRequired("to")
	cmd.Flags().StringVar(&params.CohortType, "cohort-type", "", "user_acquisition, retargeting, or unified (required)")
	cmd.MarkFlagRequired("cohort-type")
	cmd.Flags().StringSliceVar(&params.KPIs, "kpis", nil, "KPIs: users, ecpi, cost, revenue, roas, roi, sessions, uninstalls, or an event name (required)")
	cmd.MarkFlagRequired("kpis")
	cmd.Flags().StringSliceVar(&params.Groupings, "groupings", nil, "1-7 dimensions, e.g. pid,geo,date (required)")
	cmd.MarkFlagRequired("groupings")
	cmd.Flags().StringVar(&params.AggregationType, "aggregation-type", "", "cumulative or on_day (required)")
	cmd.MarkFlagRequired("aggregation-type")
	cmd.Flags().IntVar(&params.MinCohortSize, "min-cohort-size", 0, "hide cohorts smaller than this")
	cmd.Flags().StringVar(&params.Granularity, "granularity", "", "hour or day")
	cmd.Flags().BoolVar(&params.PartialData, "partial-data", false, "include cohorts whose measurement period is incomplete")
	cmd.Flags().StringArrayVar(&filters, "filter", nil, "dimension filter as key=v1,v2 (e.g. pid=facebook or period=0,7,30); repeatable")
	cmd.Flags().BoolVar(&params.PreferredCurrency, "preferred-currency", false, "use the app's preferred currency instead of USD")
	cmd.Flags().BoolVar(&params.PreferredTimezone, "preferred-timezone", false, "use the app's timezone instead of UTC")
	cmd.Flags().BoolVar(&params.PerUser, "per-user", false, "divide KPI values by the number of users")
	cmd.Flags().StringVar(&params.Format, "format", "", "response format: json (default) or csv")

	return cmd
}

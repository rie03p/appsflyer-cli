package cli

import (
	"fmt"
	"strings"

	"github.com/rie03p/appsflyer-cli/internal/appsflyer"
	"github.com/spf13/cobra"
)

func newMasterCmd(opts *rootOptions) *cobra.Command {
	var (
		appID      string
		params     appsflyer.MasterParams
		filters    []string
		calculated []string
	)

	cmd := &cobra.Command{
		Use:   "master",
		Short: "Fetch a Master API report (cross-app aggregate KPIs)",
		Example: `  afcli master --app id123456789 --from 2026-07-01 --to 2026-07-07 \
    --groupings pid,geo --kpis installs,clicks,impressions
  afcli master --app all --from 2026-07-01 --to 2026-07-07 \
    --groupings app_id,pid --kpis installs --filter pid=facebook --format json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			params.Filters, err = parsePairs(filters, "--filter")
			if err != nil {
				return err
			}
			params.CalculatedKPIs, err = parsePairs(calculated, "--calculated-kpi")
			if err != nil {
				return err
			}
			for key := range params.Filters {
				if !contains(appsflyer.MasterFilterKeys, key) {
					return fmt.Errorf("unknown filter %q (valid: %s)", key, strings.Join(appsflyer.MasterFilterKeys, ", "))
				}
			}
			client, err := opts.client()
			if err != nil {
				return err
			}
			body, err := client.Master(cmd.Context(), appID, params)
			if err != nil {
				return err
			}
			return opts.write(cmd, body)
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", `app ID, comma-separated list, or "all" (required)`)
	cmd.MarkFlagRequired("app")
	cmd.Flags().StringVar(&params.From, "from", "", `start date, "yyyy-mm-dd" (required; range max 31 days)`)
	cmd.MarkFlagRequired("from")
	cmd.Flags().StringVar(&params.To, "to", "", "end date (required)")
	cmd.MarkFlagRequired("to")
	cmd.Flags().StringSliceVar(&params.Groupings, "groupings", nil, "grouping dimensions, e.g. app_id,pid,geo (required)")
	cmd.MarkFlagRequired("groupings")
	cmd.Flags().StringSliceVar(&params.KPIs, "kpis", nil, "KPIs to include, e.g. installs,clicks,impressions (required)")
	cmd.MarkFlagRequired("kpis")
	cmd.Flags().StringArrayVar(&filters, "filter", nil, "dimension filter as key=value (pid, c, af_prt, af_channel, af_siteid, geo); repeatable")
	cmd.Flags().StringArrayVar(&calculated, "calculated-kpi", nil, "custom KPI as name=formula, e.g. ctr=clicks/impressions; repeatable")
	cmd.Flags().StringVar(&params.Currency, "currency", "", "revenue currency (preferred or USD)")
	cmd.Flags().StringVar(&params.Timezone, "timezone", "", "timezone for the data")
	cmd.Flags().StringVar(&params.Format, "format", "", "response format: csv (default) or json")

	return cmd
}

func parsePairs(pairs []string, flagName string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		key, val, ok := strings.Cut(p, "=")
		if !ok || key == "" || val == "" {
			return nil, errKeyValue(flagName, p)
		}
		m[key] = val
	}
	return m, nil
}

func errKeyValue(flagName, got string) error {
	return fmt.Errorf("%s expects key=value, got %q", flagName, got)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

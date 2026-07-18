package cli

import (
	"github.com/spf13/cobra"
)

func newFreshnessCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "freshness",
		Short:   "Show when Master API aggregated data was last updated",
		Example: `  afcli freshness`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}
			body, err := client.Freshness(cmd.Context())
			if err != nil {
				return err
			}
			return opts.write(cmd, body)
		},
	}
}

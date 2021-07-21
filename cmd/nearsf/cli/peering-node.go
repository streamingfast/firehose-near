package cli

import "github.com/spf13/cobra"

func init() {
	registerNode("peering", func(cmd *cobra.Command) error {
		return nil
	}, NodeManagerAPIAddr)
}

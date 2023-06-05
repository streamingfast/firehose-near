package main

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var numberRegex = regexp.MustCompile(`(\d{10})`)
var errStopWalk = fmt.Errorf("stop walk")

func newToolsBackfillCmd(logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{Use: "backfill", Short: "Various tools for updating merged block files"}
	cmd.PersistentFlags().StringP("range", "r", "", "Block range to use for the check")

	cmd.AddCommand(newToolsBackfillFixEncodingCmd(logger))
	cmd.AddCommand(newToolsBackfillPrevHeightCmd(logger))
	cmd.AddCommand(newToolsbackfillPrevHeightCheckCmd(logger))

	return cmd
}

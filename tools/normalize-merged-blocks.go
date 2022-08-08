package tools

import (
	sftools "github.com/streamingfast/sf-tools"
)

func init() {
	Cmd.AddCommand(NormalizeMergedBlocksCmd)
}

var NormalizeMergedBlocksCmd = sftools.GetMergedBlocksNormalizer(zlog, tracer, nil)

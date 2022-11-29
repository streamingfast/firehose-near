package tools

import (
	sftools "github.com/streamingfast/sf-tools"
)

func init() {
	prometheusExporterCmd := sftools.GetFirehosePrometheusExporterCmd(zlog, tracer, transformsSetter)
	prometheusExporterCmd.Flags().String("receipt-account-filters", "", "comma-separated accounts to use as filter/index. If it contains a colon (:), it will be interpreted as <prefix>:<suffix> (each of which can be empty, ex. hello: or :world)")
	Cmd.AddCommand(prometheusExporterCmd)
}

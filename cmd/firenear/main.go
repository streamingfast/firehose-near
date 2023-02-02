package main

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/streamingfast/firehose-near/cmd/firenear/cli"
)

// Version value, injected via go build `ldflags` at build time
var version = "dev"

func init() {
	cli.RootCmd.Version = versionString(version)
}

func main() {
	cli.Main()
}

func versionString(version string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("we should have been able to retrieve info from 'runtime/debug#ReadBuildInfo'")
	}

	commit := findSetting("vcs.revision", info.Settings)
	date := findSetting("vcs.time", info.Settings)

	var labels []string
	if len(commit) >= 7 {
		labels = append(labels, fmt.Sprintf("Commit %s", commit[0:7]))
	}

	if date != "" {
		labels = append(labels, fmt.Sprintf("Built %s", date))
	}

	if len(labels) == 0 {
		return version
	}

	return fmt.Sprintf("%s (%s)", version, strings.Join(labels, ", "))
}

func findSetting(key string, settings []debug.BuildSetting) (value string) {
	for _, setting := range settings {
		if setting.Key == key {
			return setting.Value
		}
	}

	return ""
}

// Copyright 2021 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/launcher"
	"go.uber.org/zap"
)

var StartCmd = &cobra.Command{Use: "start", Short: "Starts `firenear` services all at once", RunE: nearStartE, Args: cobra.ArbitraryArgs}

func init() {
	RootCmd.AddCommand(StartCmd)
}

func nearStartE(cmd *cobra.Command, args []string) (err error) {
	cmd.SilenceUsage = true

	dataDir := viper.GetString("global-data-dir")
	zlog.Debug("firenear binary started", zap.String("data_dir", dataDir))

	configFile := viper.GetString("global-config-file")
	zlog.Info(fmt.Sprintf("Starting StreamingFast on NEAR with config file '%s'", configFile))

	err = Start(configFile, dataDir, args)
	if err != nil {
		return fmt.Errorf("unable to launch: %w", err)
	}

	// If an error occurred, saying Goodbye is not greate
	zlog.Info(fmt.Sprintf("Goodbye"))
	return
}

func Start(configFile string, dataDir string, args []string) (err error) {
	dataDirAbs, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("unable to setup directory structure: %w", err)
	}

	err = makeDirs([]string{dataDirAbs})
	if err != nil {
		return err
	}

	modules := &launcher.Runtime{
		AbsDataDir: dataDirAbs,
	}

	atmCacheEnabled := viper.GetBool("common-blocks-cache-enabled")
	if atmCacheEnabled {
		bstream.GetBlockPayloadSetter = bstream.ATMCachedPayloadSetter

		cacheDir := MustReplaceDataDir(modules.AbsDataDir, viper.GetString("common-blocks-cache-dir"))
		storeUrl := MustReplaceDataDir(modules.AbsDataDir, viper.GetString("common-merged-blocks-store-url"))
		maxRecentEntryBytes := viper.GetInt("common-blocks-cache-max-recent-entry-bytes")
		maxEntryByAgeBytes := viper.GetInt("common-blocks-cache-max-entry-by-age-bytes")
		bstream.InitCache(storeUrl, cacheDir, maxRecentEntryBytes, maxEntryByAgeBytes)
	}

	bstream.GetProtocolFirstStreamableBlock = uint64(viper.GetInt("common-first-streamable-block"))

	err = bstream.ValidateRegistry()
	if err != nil {
		return fmt.Errorf("protocol specific hooks not configured correctly: %w", err)
	}

	launch := launcher.NewLauncher(zlog, modules)
	zlog.Debug("launcher created")

	runByDefault := func(app string) bool {
		if app == "archive-node" {
			return false
		}
		return true
	}

	apps := launcher.ParseAppsFromArgs(args, runByDefault)
	if len(args) == 0 {
		apps = launcher.ParseAppsFromArgs(launcher.Config["start"].Args, runByDefault)
	}
	zlog.Info(fmt.Sprintf("Launching applications: %s", strings.Join(apps, ",")))
	if err = launch.Launch(apps); err != nil {
		return err
	}

	signalHandler := derr.SetupSignalHandler(viper.GetDuration("common-system-shutdown-signal-delay"))
	select {
	case <-signalHandler:
		zlog.Info(fmt.Sprintf("Received termination signal, quitting"))
		go launch.Close()
	case appID := <-launch.Terminating():
		if launch.Err() == nil {
			zlog.Info(fmt.Sprintf("Application %s triggered a clean shutdown, quitting", appID))
		} else {
			zlog.Info(fmt.Sprintf("Application %s shutdown unexpectedly, quitting", appID))
			err = launch.Err()
		}
	}

	launch.WaitForTermination()

	return
}

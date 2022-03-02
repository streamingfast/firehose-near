package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/transform"
	dauthAuthenticator "github.com/streamingfast/dauth/authenticator"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/dmetering"
	"github.com/streamingfast/dmetrics"
	"github.com/streamingfast/dstore"
	firehoseApp "github.com/streamingfast/firehose/app/firehose"
	"github.com/streamingfast/logging"
	sftransform "github.com/streamingfast/sf-near/transform"
	"go.uber.org/zap"
)

var metricset = dmetrics.NewSet()
var headBlockNumMetric = metricset.NewHeadBlockNumber("firehose")
var headTimeDriftmetric = metricset.NewHeadTimeDrift("firehose")

func init() {
	appLogger := zap.NewNop()
	logging.Register("github.com/streamingfast/sf-near/firehose", &appLogger)

	launcher.RegisterApp(&launcher.AppDef{
		ID:          "firehose",
		Title:       "Block Firehose",
		Description: "Provides on-demand filtered blocks, depends on common-blocks-store-url and common-blockstream-addr",
		MetricsID:   "merged-filter",
		Logger:      launcher.NewLoggingDef("github.com/streamingfast/sf-near/firehose.*", nil),
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().String("firehose-grpc-listen-addr", FirehoseGRPCServingAddr, "Address on which the firehose will listen")
			cmd.Flags().StringSlice("firehose-blocks-store-urls", nil, "If non-empty, overrides common-blocks-store-url with a list of blocks stores")
			cmd.Flags().Duration("firehose-real-time-tolerance", 1*time.Minute, "firehose will became alive if now - block time is smaller then tolerance")

			// irreversible indices
			cmd.Flags().String("firehose-irreversible-blocks-index-url", "", "If non-empty, will use this URL as a store to read irreversibility data on blocks and optimize replay")
			cmd.Flags().IntSlice("firehose-irreversible-blocks-index-bundle-sizes", []int{100000, 10000, 1000, 100}, "list of sizes for irreversible block indices")
			// block indices
			cmd.Flags().String("firehose-block-index-url", "", "If non-empty, will use this URL as a store to load index data used by some transforms")
			cmd.Flags().IntSlice("firehose-block-index-sizes", []int{100000, 10000, 1000, 100}, "list of sizes for block indices")

			return nil
		},

		FactoryFunc: func(runtime *launcher.Runtime) (launcher.App, error) {
			sfDataDir := runtime.AbsDataDir
			tracker := runtime.Tracker.Clone()
			blockstreamAddr := viper.GetString("common-blockstream-addr")
			if blockstreamAddr != "" {
				tracker.AddGetter(bstream.BlockStreamLIBTarget, bstream.StreamLIBBlockRefGetter(blockstreamAddr))
			}

			// FIXME: That should be a shared dependencies across `dfuse for EOSIO`
			authenticator, err := dauthAuthenticator.New(viper.GetString("common-auth-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dauth: %w", err)
			}

			// FIXME: That should be a shared dependencies across `dfuse for EOSIO`, it will avoid the need to call `dmetering.SetDefaultMeter`
			metering, err := dmetering.New(viper.GetString("common-metering-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dmetering: %w", err)
			}
			dmetering.SetDefaultMeter(metering)

			firehoseBlocksStoreURLs := viper.GetStringSlice("firehose-blocks-store-urls")
			if len(firehoseBlocksStoreURLs) == 0 {
				firehoseBlocksStoreURLs = []string{viper.GetString("common-blocks-store-url")}
			} else if len(firehoseBlocksStoreURLs) == 1 && strings.Contains(firehoseBlocksStoreURLs[0], ",") {
				if viper.GetBool("common-atm-cache-enabled") {
					panic("cannot use ATM cache with firehose multi blocks store URLs")
				}
				// Providing multiple elements from config doesn't work with `viper.GetStringSlice`, so let's also handle the case where a single element has separator
				firehoseBlocksStoreURLs = strings.Split(firehoseBlocksStoreURLs[0], ",")
			}

			for i, url := range firehoseBlocksStoreURLs {
				firehoseBlocksStoreURLs[i] = mustReplaceDataDir(sfDataDir, url)
			}

			shutdownSignalDelay := viper.GetDuration("common-system-shutdown-signal-delay")
			grcpShutdownGracePeriod := time.Duration(0)
			if shutdownSignalDelay.Seconds() > 5 {
				grcpShutdownGracePeriod = shutdownSignalDelay - (5 * time.Second)
			}

			indexStoreUrl := viper.GetString("firehose-block-index-url")
			var indexStore dstore.Store
			if indexStoreUrl != "" {
				s, err := dstore.NewStore(indexStoreUrl, "", "", false)
				if err != nil {
					return nil, fmt.Errorf("couldn't create indexStore: %w", err)
				}
				indexStore = s
			}

			var possibleIndexSizes []uint64
			for _, size := range viper.GetIntSlice("firehose-block-index-sizes") {
				if size < 0 {
					return nil, fmt.Errorf("invalid negative size for firehose-block-index-sizes: %d", size)
				}
				possibleIndexSizes = append(possibleIndexSizes, uint64(size))
			}

			registry := transform.NewRegistry()
			registry.Register(sftransform.BasicReceiptFilterFactory(indexStore, possibleIndexSizes))

			var possibleIrreversibleIndexSizes []uint64
			for _, size := range viper.GetIntSlice("firehose-irreversible-blocks-index-bundle-sizes") {
				if size < 0 {
					return nil, fmt.Errorf("invalid negative size for firehose-irreversible-blocks-index-bundle-sizes: %d", size)
				}
				possibleIrreversibleIndexSizes = append(possibleIrreversibleIndexSizes, uint64(size))
			}

			return firehoseApp.New(appLogger, &firehoseApp.Config{
				BlockStoreURLs:                  firehoseBlocksStoreURLs,
				BlockStreamAddr:                 blockstreamAddr,
				GRPCListenAddr:                  viper.GetString("firehose-grpc-listen-addr"),
				GRPCShutdownGracePeriod:         grcpShutdownGracePeriod,
				IrreversibleBlocksIndexStoreURL: viper.GetString("firehose-irreversible-blocks-index-url"),
				IrreversibleBlocksBundleSizes:   possibleIrreversibleIndexSizes,
				RealtimeTolerance:               viper.GetDuration("firehose-real-time-tolerance"),
			}, &firehoseApp.Modules{
				Authenticator: authenticator,
				//				BlockTrimmer:              blockstreamv2.BlockTrimmerFunc(trimBlock),
				HeadTimeDriftMetric:   headTimeDriftmetric,
				HeadBlockNumberMetric: headBlockNumMetric,
				Tracker:               tracker,
			}), nil
		},
	})
}

func passthroughPreprocessBlock(blk *bstream.Block) (interface{}, error) {
	return nil, nil
}

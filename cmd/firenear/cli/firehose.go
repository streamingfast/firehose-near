package cli

import (
	"fmt"
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

	launcher.RegisterApp(zlog, &launcher.AppDef{
		ID:          "firehose",
		Title:       "Block Firehose",
		Description: "Provides on-demand filtered blocks, depends on common-blocks-store-url and common-blockstream-addr",
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().String("firehose-grpc-listen-addr", FirehoseGRPCServingAddr, "Address on which the firehose will listen")

			cmd.Flags().String("firehose-block-index-url", "", "If non-empty, will use this URL as a store to load index data used by some transforms")
			cmd.Flags().IntSlice("firehose-block-index-sizes", []int{100000, 10000, 1000, 100}, "list of sizes for block indices")

			return nil
		},

		FactoryFunc: func(runtime *launcher.Runtime) (launcher.App, error) {
			sfDataDir := runtime.AbsDataDir
			blockstreamAddr := viper.GetString("common-blockstream-addr")

			authenticator, err := dauthAuthenticator.New(viper.GetString("common-auth-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dauth: %w", err)
			}

			metering, err := dmetering.New(viper.GetString("common-metering-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dmetering: %w", err)
			}
			dmetering.SetDefaultMeter(metering)

			mergedBlocksStoreURL := MustReplaceDataDir(sfDataDir, viper.GetString("common-blocks-store-url"))
			oneBlocksStoreURL := MustReplaceDataDir(sfDataDir, viper.GetString("common-oneblock-store-url"))
			forkedBlocksStoreURL := MustReplaceDataDir(sfDataDir, viper.GetString("common-forkedblocks-store-url"))

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

			return firehoseApp.New(appLogger, &firehoseApp.Config{
				MergedBlocksStoreURL:    mergedBlocksStoreURL,
				OneBlocksStoreURL:       oneBlocksStoreURL,
				ForkedBlocksStoreURL:    forkedBlocksStoreURL,
				BlockStreamAddr:         blockstreamAddr,
				GRPCListenAddr:          viper.GetString("firehose-grpc-listen-addr"),
				GRPCShutdownGracePeriod: time.Second,
			}, &firehoseApp.Modules{
				Authenticator:         authenticator,
				HeadTimeDriftMetric:   headTimeDriftmetric,
				HeadBlockNumberMetric: headBlockNumMetric,
				TransformRegistry:     registry,
			}), nil
		},
	})
}

func passthroughPreprocessBlock(blk *bstream.Block) (interface{}, error) {
	return nil, nil
}

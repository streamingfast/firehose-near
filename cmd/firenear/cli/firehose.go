package cli

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream/transform"
	dauthAuthenticator "github.com/streamingfast/dauth/authenticator"
	_ "github.com/streamingfast/dauth/authenticator/gcp"
	discoveryservice "github.com/streamingfast/dgrpc/server/discovery-service"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/dmetering"
	"github.com/streamingfast/dmetrics"
	"github.com/streamingfast/dstore"
	sftransform "github.com/streamingfast/firehose-near/transform"
	firehoseApp "github.com/streamingfast/firehose/app/firehose"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/substreams/client"
	substreamsService "github.com/streamingfast/substreams/service"
)

var metricset = dmetrics.NewSet()
var headBlockNumMetric = metricset.NewHeadBlockNumber("firehose")
var headTimeDriftmetric = metricset.NewHeadTimeDrift("firehose")

func init() {
	appLogger, _ := logging.PackageLogger("firehose", "github.com/streamingfast/firehose-near/firehose")

	launcher.RegisterApp(zlog, &launcher.AppDef{
		ID:          "firehose",
		Title:       "Block Firehose",
		Description: "Provides on-demand filtered blocks, depends on common-merged-blocks-store-url and common-live-blocks-addr",
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().String("firehose-grpc-listen-addr", FirehoseGRPCServingAddr, "Address on which the Firehose will listen")
			cmd.Flags().String("firehose-discovery-service-url", "", "URL to configure the gRPC discovery service") //traffic-director://xds?vpc_network=vpc-global&use_xds_reds=true

			cmd.Flags().Bool("substreams-enabled", false, "Whether to enable substreams")
			cmd.Flags().Bool("substreams-partial-mode-enabled", false, "Whether to enable partial stores generation support on this instance (usually for internal deployments only)")
			cmd.Flags().Bool("substreams-request-stats-enabled", false, "Enables stats per request, like block rate. Should only be enabled in debugging instance not in production")
			cmd.Flags().String("substreams-state-store-url", "{sf-data-dir}/localdata", "where substreams state data are stored")
			cmd.Flags().Uint64("substreams-cache-save-interval", uint64(1_000), "Interval in blocks at which to save store snapshots and output caches")
			cmd.Flags().Int("substreams-parallel-subrequest-limit", 4, "Number of parallel subrequests substream can make to synchronize its stores")
			cmd.Flags().String("substreams-client-endpoint", FirehoseGRPCServingAddr, "Firehose endpoint for substreams client.")
			cmd.Flags().String("substreams-client-jwt", "", "JWT for substreams client authentication")
			cmd.Flags().Bool("substreams-client-insecure", false, "Substreams client in insecure mode")
			cmd.Flags().Bool("substreams-client-plaintext", true, "Substreams client in plaintext mode")
			cmd.Flags().Uint64("substreams-sub-request-parallel-jobs", 5, "Substreams subrequest parallel jobs for the scheduler")
			cmd.Flags().Uint64("substreams-sub-request-block-range-size", 1000, "Substreams subrequest block range size value for the scheduler")

			return nil
		},

		FactoryFunc: func(runtime *launcher.Runtime) (launcher.App, error) {
			blockstreamAddr := viper.GetString("common-live-blocks-addr")

			authenticator, err := dauthAuthenticator.New(viper.GetString("common-auth-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dauth: %w", err)
			}

			metering, err := dmetering.New(viper.GetString("common-metering-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dmetering: %w", err)
			}
			dmetering.SetDefaultMeter(metering)

			mergedBlocksStoreURL, oneBlocksStoreURL, forkedBlocksStoreURL, err := getCommonStoresURLs(runtime.AbsDataDir)
			if err != nil {
				return nil, err
			}

			indexStore, possibleIndexSizes, err := GetIndexStore(runtime.AbsDataDir)
			if err != nil {
				return nil, fmt.Errorf("unable to initialize indexes: %w", err)
			}

			sfDataDir := runtime.AbsDataDir
			var registerServiceExt firehoseApp.RegisterServiceExtensionFunc

			rawServiceDiscoveryURL := viper.GetString("firehose-discovery-service-url")
			var serviceDiscoveryURL *url.URL
			if rawServiceDiscoveryURL != "" {
				serviceDiscoveryURL, err = url.Parse(rawServiceDiscoveryURL)
				if err != nil {
					return nil, fmt.Errorf("unable to parse discovery service url: %w", err)
				}
				err = discoveryservice.Bootstrap(serviceDiscoveryURL)
				if err != nil {
					return nil, fmt.Errorf("unable to bootstrap discovery service: %w", err)
				}
			}

			if viper.GetBool("substreams-enabled") {
				stateStore, err := dstore.NewStore(MustReplaceDataDir(sfDataDir, viper.GetString("substreams-state-store-url")), "zst", "zstd", true)
				if err != nil {
					return nil, fmt.Errorf("setting up state store for data: %w", err)
				}

				opts := []substreamsService.Option{
					substreamsService.WithCacheSaveInterval(viper.GetUint64("substreams-cache-save-interval")),
				}

				if viper.GetBool("substreams-request-stats-enabled") {
					opts = append(opts, substreamsService.WithRequestStats())
				}

				if viper.GetBool("substreams-partial-mode-enabled") {
					opts = append(opts, substreamsService.WithPartialMode())
				}

				substreamsClientConfig := client.NewSubstreamsClientConfig(
					viper.GetString("substreams-client-endpoint"),
					os.ExpandEnv(viper.GetString("substreams-client-jwt")),
					viper.GetBool("substreams-client-insecure"),
					viper.GetBool("substreams-client-plaintext"),
				)

				sss, err := substreamsService.New(
					stateStore,
					"sf.near.type.v1.Block",
					viper.GetUint64("substreams-sub-request-parallel-jobs"),
					viper.GetUint64("substreams-sub-request-block-range-size"),
					substreamsClientConfig,
					opts...,
				)

				if err != nil {
					return nil, fmt.Errorf("creating substreams service: %w", err)
				}

				registerServiceExt = sss.Register
			}

			registry := transform.NewRegistry()
			registry.Register(sftransform.HeaderOnlyTransformFactory)
			registry.Register(sftransform.BasicReceiptFilterFactory(indexStore, possibleIndexSizes))

			return firehoseApp.New(appLogger, &firehoseApp.Config{
				MergedBlocksStoreURL:    mergedBlocksStoreURL,
				OneBlocksStoreURL:       oneBlocksStoreURL,
				ForkedBlocksStoreURL:    forkedBlocksStoreURL,
				BlockStreamAddr:         blockstreamAddr,
				GRPCListenAddr:          viper.GetString("firehose-grpc-listen-addr"),
				ServiceDiscoveryURL:     serviceDiscoveryURL,
				GRPCShutdownGracePeriod: time.Second,
			}, &firehoseApp.Modules{
				Authenticator:            authenticator,
				HeadTimeDriftMetric:      headTimeDriftmetric,
				HeadBlockNumberMetric:    headBlockNumMetric,
				TransformRegistry:        registry,
				RegisterServiceExtension: registerServiceExt,
			}), nil
		},
	})
}

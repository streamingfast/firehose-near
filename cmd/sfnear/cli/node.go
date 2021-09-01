package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream/blockstream"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/logging"
	nodeManager "github.com/streamingfast/node-manager"
	nodeManagerApp "github.com/streamingfast/node-manager/app/node_manager2"
	"github.com/streamingfast/node-manager/metrics"
	"github.com/streamingfast/node-manager/operator"
	pbbstream "github.com/streamingfast/pbgo/dfuse/bstream/v1"
	pbheadinfo "github.com/streamingfast/pbgo/dfuse/headinfo/v1"
	"github.com/streamingfast/sf-near/nodemanager"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func registerCommonNodeFlags(cmd *cobra.Command, flagPrefix string, managerAPIAddr string) {
	cmd.Flags().String(flagPrefix+"path", "neard", "Command that will be launched by the node manager")
	cmd.Flags().String(flagPrefix+"data-dir", "{sf-data-dir}/{node-role}/data", "Directory for node data ({node-role} is either mindreader, peering or dev-miner)")
	cmd.Flags().String(flagPrefix+"config-file", "./{node-role}/config.json", "Node configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {sf-data-dir}/{node-role}/data folder")
	cmd.Flags().String(flagPrefix+"genesis-file", "./{node-role}/genesis.json", "Node configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {sf-data-dir}/{node-role}/data folder")
	cmd.Flags().String(flagPrefix+"node-key-file", "./{node-role}/node_key.json", "Node key configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {sf-data-dir}/{node-role}/data folder")
	cmd.Flags().Bool(flagPrefix+"debug-deep-mind", false, "[DEV] Prints deep mind instrumentation logs to standard output, should be use for debugging purposes only")
	cmd.Flags().Bool(flagPrefix+"log-to-zap", true, "Enable all node logs to transit into node's logger directly, when false, prints node logs directly to stdout")
	cmd.Flags().String(flagPrefix+"manager-api-addr", managerAPIAddr, "Near node manager API address")
	cmd.Flags().Duration(flagPrefix+"readiness-max-latency", 30*time.Second, "Determine the maximum head block latency at which the instance will be determined healthy. Some chains have more regular block production than others.")
	cmd.Flags().String(flagPrefix+"node-extra-arguments", "", "Extra arguments to be passed when executing superviser binary")

	// FIXME: Right now our near-dm-indexer doesn't support it, we should plan on adding it!
	// cmd.Flags().String(flagPrefix+"node-boot-nodes", "", "Set the node's boot nodes to bootstrap network from")
}

func registerNode(kind string, extraFlagRegistration func(cmd *cobra.Command) error, managerAPIaddr string) {
	if kind != "mindreader" && kind != "peering" {
		panic(fmt.Errorf("invalid kind value, must be either 'mindreader' or 'peering', got %q", kind))
	}

	app := fmt.Sprintf("%s-node", kind)
	flagPrefix := fmt.Sprintf("%s-", app)
	appLogger := zap.NewNop()
	nodeLogger := zap.NewNop()

	logging.Register(fmt.Sprintf("github.com/streamingfast/sf-near/%s", app), &appLogger)
	logging.Register(fmt.Sprintf("github.com/streamingfast/sf-near/%s/node", app), &nodeLogger)

	launcher.RegisterApp(&launcher.AppDef{
		ID:          app,
		Title:       fmt.Sprintf("NEAR Node (%s)", kind),
		Description: fmt.Sprintf("NEAR %s node with built-in operational manager", kind),
		MetricsID:   app,
		Logger: launcher.NewLoggingDef(
			fmt.Sprintf("github.com/streamingfast/sf-near/%s.*", app),
			[]zapcore.Level{zap.WarnLevel, zap.WarnLevel, zap.InfoLevel, zap.DebugLevel},
		),
		RegisterFlags: func(cmd *cobra.Command) error {
			registerCommonNodeFlags(cmd, flagPrefix, managerAPIaddr)
			extraFlagRegistration(cmd)
			return nil
		},
		InitFunc: func(runtime *launcher.Runtime) error {
			return nil
		},
		FactoryFunc: nodeFactoryFunc(flagPrefix, kind, &appLogger, &nodeLogger),
	})

}

func nodeFactoryFunc(flagPrefix, kind string, appLogger, nodeLogger **zap.Logger) func(*launcher.Runtime) (launcher.App, error) {
	return func(runtime *launcher.Runtime) (launcher.App, error) {
		sfDataDir := runtime.AbsDataDir

		nodePath := viper.GetString(flagPrefix + "path")
		nodeDataDir := replaceNodeRole(kind, mustReplaceDataDir(sfDataDir, viper.GetString(flagPrefix+"data-dir")))
		configFile := replaceNodeRole(kind, viper.GetString(flagPrefix+"config-file"))
		genesisFile := replaceNodeRole(kind, viper.GetString(flagPrefix+"genesis-file"))
		nodeKeyFile := replaceNodeRole(kind, viper.GetString(flagPrefix+"node-key-file"))
		readinessMaxLatency := viper.GetDuration(flagPrefix + "readiness-max-latency")
		debugDeepMind := viper.GetBool(flagPrefix + "debug-deep-mind")
		logToZap := viper.GetBool(flagPrefix + "log-to-zap")
		shutdownDelay := viper.GetDuration("common-system-shutdown-signal-delay") // we reuse this global value
		httpAddr := viper.GetString(flagPrefix + "manager-api-addr")

		nodeArguments, err := buildNodeArguments(
			nodeDataDir,
			flagPrefix,
			kind,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot build node bootstrap arguments")
		}
		extraArgs := getExtraArguments(flagPrefix)
		if len(extraArgs) > 0 {
			nodeArguments = append(nodeArguments, extraArgs...)
		}

		metricsAndReadinessManager := buildMetricsAndReadinessManager(flagPrefix, readinessMaxLatency)

		superviser := nodemanager.NewSuperviser(
			nodePath,
			nodeArguments,
			nodeDataDir,
			metricsAndReadinessManager.UpdateHeadBlock,
			debugDeepMind,
			logToZap,
			*appLogger,
			*nodeLogger,
		)

		bootstrapper := &bootstrapper{
			configFile:  configFile,
			genesisFile: genesisFile,
			nodeKeyFile: nodeKeyFile,
			nodeDataDir: nodeDataDir,
		}

		chainOperator, err := operator.New(
			*appLogger,
			superviser,
			metricsAndReadinessManager,
			&operator.Options{
				ShutdownDelay:              shutdownDelay,
				EnableSupervisorMonitoring: true,
				Bootstrapper:               bootstrapper,
			})
		if err != nil {
			return nil, fmt.Errorf("unable to create chain operator: %w", err)
		}

		if kind != "mindreader" {
			return nodeManagerApp.New(&nodeManagerApp.Config{
				HTTPAddr: httpAddr,
			}, &nodeManagerApp.Modules{
				Operator:                   chainOperator,
				MetricsAndReadinessManager: metricsAndReadinessManager,
			}, *appLogger), nil
		}

		blockStreamServer := blockstream.NewUnmanagedServer(blockstream.ServerOptionWithLogger(*appLogger))
		oneBlockStoreURL := mustReplaceDataDir(sfDataDir, viper.GetString("common-oneblock-store-url"))
		mergedBlockStoreURL := mustReplaceDataDir(sfDataDir, viper.GetString("common-blocks-store-url"))
		workingDir := mustReplaceDataDir(sfDataDir, viper.GetString("mindreader-node-working-dir"))
		gprcListenAdrr := viper.GetString("mindreader-node-grpc-listen-addr")
		mergeAndStoreDirectly := viper.GetBool("mindreader-node-merge-and-store-directly")
		mergeThresholdBlockAge := viper.GetDuration("mindreader-node-merge-threshold-block-age")
		batchStartBlockNum := viper.GetUint64("mindreader-node-start-block-num")
		batchStopBlockNum := viper.GetUint64("mindreader-node-stop-block-num")
		waitTimeForUploadOnShutdown := viper.GetDuration("mindreader-node-wait-upload-complete-on-shutdown")
		oneBlockFileSuffix := viper.GetString("mindreader-node-oneblock-suffix")
		blocksChanCapacity := viper.GetInt("mindreader-node-blocks-chan-capacity")

		tracker := runtime.Tracker.Clone()

		mindreaderPlugin, err := getMindreaderLogPlugin(
			blockStreamServer,
			oneBlockStoreURL,
			mergedBlockStoreURL,
			mergeAndStoreDirectly,
			mergeThresholdBlockAge,
			workingDir,
			batchStartBlockNum,
			batchStopBlockNum,
			blocksChanCapacity,
			false,
			waitTimeForUploadOnShutdown,
			oneBlockFileSuffix,
			chainOperator.Shutdown,
			metricsAndReadinessManager,
			tracker,
			*appLogger,
		)
		if err != nil {
			return nil, fmt.Errorf("new mindreader plugin: %w", err)
		}

		superviser.RegisterLogPlugin(mindreaderPlugin)

		return nodeManagerApp.New(&nodeManagerApp.Config{
			HTTPAddr: httpAddr,
			GRPCAddr: gprcListenAdrr,
		}, &nodeManagerApp.Modules{
			Operator:                   chainOperator,
			MindreaderPlugin:           mindreaderPlugin,
			MetricsAndReadinessManager: metricsAndReadinessManager,
			RegisterGRPCService: func(server *grpc.Server) error {
				pbheadinfo.RegisterHeadInfoServer(server, blockStreamServer)
				pbbstream.RegisterBlockStreamServer(server, blockStreamServer)

				return nil
			},
		}, *appLogger), nil
	}
}

type bootstrapper struct {
	configFile  string
	genesisFile string
	nodeKeyFile string
	nodeDataDir string
}

func (b *bootstrapper) Bootstrap() error {
	configFileInDataDir := filepath.Join(b.nodeDataDir, "config.json")
	genesisFileInDataDir := filepath.Join(b.nodeDataDir, "genesis.json")
	nodeKeyFileInDataDir := filepath.Join(b.nodeDataDir, "node_key.json")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := os.MkdirAll(b.nodeDataDir, os.ModePerm); err != nil {
		return fmt.Errorf("create all dirs of %q: %w", b.nodeDataDir, err)
	}

	if err := copyFile(ctx, b.configFile, configFileInDataDir); err != nil {
		return fmt.Errorf("unable to copy config file %q to %q: %w", b.configFile, configFileInDataDir, err)
	}

	if err := copyFile(ctx, b.genesisFile, genesisFileInDataDir); err != nil {
		return fmt.Errorf("unable to copy genesis file %q to %q: %w", b.genesisFile, genesisFileInDataDir, err)
	}

	if err := copyFile(ctx, b.nodeKeyFile, nodeKeyFileInDataDir); err != nil {
		return fmt.Errorf("unable to copy node key file %q to %q: %w", b.nodeKeyFile, nodeKeyFileInDataDir, err)
	}

	return nil
}

type nodeArgsByRole map[string]string

func buildNodeArguments(nodeDataDir, flagPrefix, nodeRole string) ([]string, error) {
	typeRoles := nodeArgsByRole{
		"peering":    "--home={node-data-dir} run",
		"mindreader": "--home={node-data-dir} run",
	}

	roleArgs, ok := typeRoles[nodeRole]
	if !ok {
		return nil, fmt.Errorf("invalid node role: %s", nodeRole)
	}
	args := strings.Fields(strings.Replace(roleArgs, "{node-data-dir}", nodeDataDir, -1))

	bootNodes := viper.GetString(flagPrefix + "node-boot-nodes")
	if bootNodes != "" {
		args = append(args, "--boot-nodes", viper.GetString(flagPrefix+"node-boot-nodes"))
	}

	return args, nil
}

func buildMetricsAndReadinessManager(name string, maxLatency time.Duration) *nodeManager.MetricsAndReadinessManager {
	headBlockTimeDrift := metrics.NewHeadBlockTimeDrift(name)
	headBlockNumber := metrics.NewHeadBlockNumber(name)

	metricsAndReadinessManager := nodeManager.NewMetricsAndReadinessManager(
		headBlockTimeDrift,
		headBlockNumber,
		maxLatency,
	)
	return metricsAndReadinessManager
}

func getExtraArguments(prefix string) (out []string) {
	extraArguments := viper.GetString(prefix + "node-extra-arguments")
	if extraArguments != "" {
		out = append(out, strings.Split(extraArguments, " ")...)
	}

	return
}

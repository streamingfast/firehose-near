package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/dfuse-io/logging"

	"github.com/dfuse-io/bstream"
	"github.com/streamingfast/dgrpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/near-sf/nodemanager"
	nodeManager "github.com/streamingfast/node-manager"
	nodeManagerApp "github.com/streamingfast/node-manager/app/node_manager"
	nodeMindReaderApp "github.com/streamingfast/node-manager/app/node_mindreader"
	"github.com/streamingfast/node-manager/metrics"
	"github.com/streamingfast/node-manager/operator"
	"go.uber.org/zap"
)

func registerNode(kind string, extraFlagRegistration func(cmd *cobra.Command) error, managerAPIaddr string) {
	if kind != "mindreader" && kind != "peering" {
		panic(fmt.Errorf("invalid kind value, must be either 'mindreader' or 'peering', got %q", kind))
	}

	app := fmt.Sprintf("%s-node", kind)
	flagPrefix := fmt.Sprintf("%s-", app)
	appLogger := zap.NewNop()
	nodeLogger := zap.NewNop()

	logging.Register(fmt.Sprintf("github.com/streamingfast/near-sf/%s", app), &appLogger)
	logging.Register(fmt.Sprintf("github.com/streamingfast/near-sf/%s/node", app), &nodeLogger)

	launcher.RegisterApp(&launcher.AppDef{
		ID:          app,
		Title:       fmt.Sprintf("Near Node (%s)", kind),
		Description: fmt.Sprintf("Near %s node with built-in operational manager", kind),
		MetricsID:   app,
		Logger: launcher.NewLoggingDef(
			fmt.Sprintf("github.com/dfuse-io/dfuse-solana/%s.*", app),
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
		dfuseDataDir := runtime.AbsDataDir

		nodePath := viper.GetString(flagPrefix + "path")
		nodeDataDir := replaceNodeRole(kind, mustReplaceDataDir(dfuseDataDir, viper.GetString(flagPrefix+"data-dir")))
		configFile := replaceNodeRole(kind, viper.GetString(flagPrefix+"config-file"))
		genesisFile := replaceNodeRole(kind, viper.GetString(flagPrefix+"genesis-file"))
		nodeKeyFile := replaceNodeRole(kind, viper.GetString(flagPrefix+"node-key-file"))
		readinessMaxLatency := viper.GetDuration(flagPrefix + "readiness-max-latency")
		debugDeepMind := viper.GetBool(flagPrefix + "debug-deep-mind")
		logToZap := viper.GetBool(flagPrefix + "log-to-zap")
		shutdownDelay := viper.GetDuration("common-system-shutdown-signal-delay") // we reuse this global value
		managerAPIAddress := viper.GetString(flagPrefix + "manager-api-addr")

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
				ManagerAPIAddress: managerAPIAddress,
			}, &nodeManagerApp.Modules{
				Operator:                   chainOperator,
				MetricsAndReadinessManager: metricsAndReadinessManager,
			}, *appLogger), nil
		}

		oneBlockStoreURL := mustReplaceDataDir(dfuseDataDir, viper.GetString("common-oneblock-store-url"))
		mergedBlockStoreURL := mustReplaceDataDir(dfuseDataDir, viper.GetString("common-blocks-store-url"))
		workingDir := mustReplaceDataDir(dfuseDataDir, viper.GetString("mindreader-node-working-dir"))
		mergeAndStoreDirectly := viper.GetBool("mindreader-node-merge-and-store-directly")
		mergeThresholdBlockAge := viper.GetDuration("mindreader-node-merge-threshold-block-age")
		batchStartBlockNum := viper.GetUint64("mindreader-node-start-block-num")
		batchStopBlockNum := viper.GetUint64("mindreader-node-stop-block-num")
		failOnNonContiguousBlock := false //FIXME ?
		waitTimeForUploadOnShutdown := viper.GetDuration("mindreader-node-wait-upload-complete-on-shutdown")
		oneBlockFileSuffix := viper.GetString("mindreader-node-oneblock-suffix")
		blocksChanCapacity := viper.GetInt("mindreader-node-blocks-chan-capacity")
		gs := dgrpc.NewServer(dgrpc.WithLogger(*appLogger))

		mindreaderPlugin, err := getMindreaderLogPlugin(
			oneBlockStoreURL,
			mergedBlockStoreURL,
			workingDir,
			mergeAndStoreDirectly,
			mergeThresholdBlockAge,
			batchStartBlockNum,
			batchStopBlockNum,
			failOnNonContiguousBlock,
			waitTimeForUploadOnShutdown,
			oneBlockFileSuffix,
			blocksChanCapacity,
			chainOperator.Shutdown,
			metricsAndReadinessManager,
			bstream.NewTracker(25),
			gs,
			*appLogger,
		)
		if err != nil {
			return nil, err
		}

		superviser.RegisterLogPlugin(mindreaderPlugin)
		return nodeMindReaderApp.New(&nodeMindReaderApp.Config{
			ManagerAPIAddress: managerAPIAddress,
		}, &nodeMindReaderApp.Modules{
			Operator:                   chainOperator,
			MetricsAndReadinessManager: metricsAndReadinessManager,
			GrpcServer:                 gs,
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

func registerCommonNodeFlags(cmd *cobra.Command, flagPrefix string, managerAPIAddr string) {
	cmd.Flags().String(flagPrefix+"path", "neard", "command that will be launched by the node manager")
	cmd.Flags().String(flagPrefix+"data-dir", "{dfuse-data-dir}/{node-role}/data", "Directory for node data ({node-role} is either mindreader, peering or dev-miner)")
	cmd.Flags().String(flagPrefix+"config-file", "./{node-role}/config.json", "Node configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {dfuse-data-dir}/{node-role}/data folder")
	cmd.Flags().String(flagPrefix+"genesis-file", "./{node-role}/genesis.json", "Node configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {dfuse-data-dir}/{node-role}/data folder")
	cmd.Flags().String(flagPrefix+"node-key-file", "./{node-role}/node_key.json", "Node key configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {dfuse-data-dir}/{node-role}/data folder")
	cmd.Flags().Bool(flagPrefix+"debug-deep-mind", false, "[DEV] Prints deep mind instrumentation logs to standard output, should be use for debugging purposes only")
	cmd.Flags().Bool(flagPrefix+"log-to-zap", true, "Enable all node logs to transit into node's logger directly, when false, prints node logs directly to stdout")
	cmd.Flags().String(flagPrefix+"manager-api-addr", managerAPIAddr, "Near node manager API address")
	cmd.Flags().Duration(flagPrefix+"readiness-max-latency", 30*time.Second, "Determine the maximum head block latency at which the instance will be determined healthy. Some chains have more regular block production than others.")
	cmd.Flags().String(flagPrefix+"node-boot-nodes", "", "Set the node's boot nodes to bootstrap network from")
	cmd.Flags().String(flagPrefix+"node-extra-arguments", "", "Extra arguments to be passed when executing superviser binary")
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
	extraArguments := viper.GetString(prefix + "-node-extra-arguments")
	if extraArguments != "" {
		for _, arg := range strings.Split(extraArguments, " ") {
			out = append(out, arg)
		}
	}
	return
}

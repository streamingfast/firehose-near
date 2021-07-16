package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dfuse-io/bstream"
	"github.com/dfuse-io/dgrpc"
	"github.com/dfuse-io/dlauncher/launcher"
	nodeManager "github.com/dfuse-io/node-manager"
	nodeManagerApp "github.com/dfuse-io/node-manager/app/node_manager"
	nodeMindReaderApp "github.com/dfuse-io/node-manager/app/node_mindreader"
	"github.com/dfuse-io/node-manager/metrics"
	"github.com/dfuse-io/node-manager/operator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/near-sf/nodemanager"
	"go.uber.org/zap"
)

func nodeFactoryFunc(isMindreader bool, appLogger, nodeLogger **zap.Logger) func(*launcher.Runtime) (launcher.App, error) {
	return func(runtime *launcher.Runtime) (launcher.App, error) {
		dfuseDataDir := runtime.AbsDataDir

		prefix := "node-"
		nodeRole := viper.GetString("node-role")
		if isMindreader {
			prefix = "mindreader-node-"
			nodeRole = "mindreader"
		}

		nodePath := viper.GetString(prefix + "path")
		nodeDataDir := replaceNodeRole(nodeRole, mustReplaceDataDir(dfuseDataDir, viper.GetString(prefix+"data-dir")))
		configFile := replaceNodeRole(nodeRole, viper.GetString(prefix+"config-file"))
		genesisFile := replaceNodeRole(nodeRole, viper.GetString(prefix+"genesis-file"))
		nodeKeyFile := replaceNodeRole(nodeRole, viper.GetString(prefix+"node-key-file"))
		readinessMaxLatency := viper.GetDuration(prefix + "readiness-max-latency")
		debugDeepMind := viper.GetBool(prefix + "debug-deep-mind")
		logToZap := viper.GetBool(prefix + "log-to-zap")
		shutdownDelay := viper.GetDuration("common-system-shutdown-signal-delay") // we reuse this global value
		managerAPIAddress := viper.GetString(prefix + "manager-api-addr")

		nodeArguments, err := buildNodeArguments(
			nodeDataDir,
			viper.GetString(prefix+"arguments"),
			nodeRole,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot build node bootstrap arguments")
		}

		metricsAndReadinessManager := buildMetricsAndReadinessManager(prefix, readinessMaxLatency)

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

		if !isMindreader {
			return nodeManagerApp.New(&nodeManagerApp.Config{
				ManagerAPIAddress: managerAPIAddress,
			}, &nodeManagerApp.Modules{
				Operator:                   chainOperator,
				MetricsAndReadinessManager: metricsAndReadinessManager,
			}, *appLogger), nil
		} else {
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

	if err := os.MkdirAll(b.nodeDataDir, os.ModePerm); err != nil {
		return fmt.Errorf("create all dirs of %q: %w", b.nodeDataDir, err)
	}

	if err := copyFile(b.configFile, configFileInDataDir); err != nil {
		return fmt.Errorf("unable to copy config file %q to %q: %w", b.configFile, configFileInDataDir, err)
	}

	if err := copyFile(b.genesisFile, genesisFileInDataDir); err != nil {
		return fmt.Errorf("unable to copy genesis file %q to %q: %w", b.genesisFile, genesisFileInDataDir, err)
	}

	if err := copyFile(b.nodeKeyFile, nodeKeyFileInDataDir); err != nil {
		return fmt.Errorf("unable to copy node key file %q to %q: %w", b.nodeKeyFile, nodeKeyFileInDataDir, err)
	}

	return nil
}

func registerCommonNodeFlags(cmd *cobra.Command, isMindreader bool) {
	prefix := "node-"
	managerAPIAddr := NodeManagerAPIAddr
	if isMindreader {
		prefix = "mindreader-node-"
		managerAPIAddr = MindreaderNodeManagerAPIAddr
	}

	cmd.Flags().String(prefix+"path", "neard", "command that will be launched by the node manager")
	cmd.Flags().String(prefix+"data-dir", "{dfuse-data-dir}/{node-role}/data", "Directory for node data ({node-role} is either mindreader, peering or dev-miner)")
	cmd.Flags().String(prefix+"config-file", "./{node-role}/config.json", "Node configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {dfuse-data-dir}/{node-role}/data folder")
	cmd.Flags().String(prefix+"genesis-file", "./{node-role}/genesis.json", "Node configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {dfuse-data-dir}/{node-role}/data folder")
	cmd.Flags().String(prefix+"node-key-file", "./{node-role}/node_key.json", "Node key configuration file where ({node-role} is either mindreader, peering or dev-miner), the file is copied inside the {dfuse-data-dir}/{node-role}/data folder")
	cmd.Flags().Bool(prefix+"debug-deep-mind", false, "[DEV] Prints deep mind instrumentation logs to standard output, should be use for debugging purposes only")
	cmd.Flags().Bool(prefix+"log-to-zap", true, "Enable all node logs to transit into node's logger directly, when false, prints node logs directly to stdout")
	cmd.Flags().String(prefix+"arguments", "", "If not empty, overrides the list of default node arguments (computed from node type and role). Start with '+' to append to default args instead of replacing. You can use the {public-ip} token, that will be matched against space-separated hostname:IP pairs in PUBLIC_IPS env var, taking hostname from HOSTNAME env var.")
	cmd.Flags().String(prefix+"manager-api-addr", managerAPIAddr, "Near node manager API address")
	cmd.Flags().Duration(prefix+"readiness-max-latency", 30*time.Second, "Determine the maximum head block latency at which the instance will be determined healthy. Some chains have more regular block production than others.")

}

type nodeArgsByRole map[string]string

func buildNodeArguments(nodeDataDir, providedArgs, nodeRole string) ([]string, error) {
	typeRoles := nodeArgsByRole{
		"peering":    "--home={node-data-dir} run",
		"mindreader": "--home={node-data-dir} run",
	}

	args, ok := typeRoles[nodeRole]
	if !ok {
		return nil, fmt.Errorf("invalid node role: %s", nodeRole)
	}

	if providedArgs != "" {
		if strings.HasPrefix(providedArgs, "+") {
			args += " " + strings.TrimLeft(providedArgs, "+")
		} else {
			args = providedArgs // discard info provided by node type / role
		}
	}

	args = strings.Replace(args, "{node-data-dir}", nodeDataDir, -1)

	return strings.Fields(args), nil
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

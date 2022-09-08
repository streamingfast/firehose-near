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
	"time"

	"github.com/spf13/cobra"
	"github.com/streamingfast/bstream/blockstream"
	"github.com/streamingfast/firehose-near/codec"
	"github.com/streamingfast/logging"
	nodeManager "github.com/streamingfast/node-manager"
	"github.com/streamingfast/node-manager/mindreader"
	"go.uber.org/zap"
)

func init() {
	registerNode("reader", registerReaderNodeFlags, ReaderNodeManagerAPIAddr)
}

func registerReaderNodeFlags(cmd *cobra.Command) error {
	cmd.Flags().String("reader-node-grpc-listen-addr", ReaderGRPCAddr, "gRPC listening address to use for serving real-time blocks")
	cmd.Flags().Bool("reader-node-merge-and-store-directly", false, "[BATCH] When enabled, do not write oneblock files, sidestep the merger and write the merged 100-blocks logs directly to --common-merged-blocks-store-url")
	cmd.Flags().Bool("reader-node-discard-after-stop-num", false, "Ignore remaining blocks being processed after stop num (only useful if we discard the reader data after reprocessing a chunk of blocks)")
	cmd.Flags().String("reader-node-working-dir", "{sf-data-dir}/reader/work", "Path where reader will stores its files")
	cmd.Flags().Uint("reader-node-start-block-num", 0, "Blocks that were produced with smaller block number then the given block num are skipped")
	cmd.Flags().Uint("reader-node-stop-block-num", 0, "Shutdown reader when we the following 'stop-block-num' has been reached, inclusively.")
	cmd.Flags().Int("reader-node-blocks-chan-capacity", 100, "Capacity of the channel holding blocks read by the reader. Process will shutdown superviser/geth if the channel gets over 90% of that capacity to prevent horrible consequences. Raise this number when processing tiny blocks very quickly")
	cmd.Flags().String("reader-node-one-block-suffix", "default", "Unique identifier for that reader, so that it can produce 'oneblock files' in the same store as another instance without competing for writes.")
	cmd.Flags().Duration("reader-node-wait-upload-complete-on-shutdown", 30*time.Second, "When the reader is shutting down, it will wait up to that amount of time for the archiver to finish uploading the blocks before leaving anyway")
	cmd.Flags().String("reader-node-merge-threshold-block-age", "24h", "When processing blocks with a blocktime older than this threshold, they will be automatically merged (you can also use \"always\" or \"never\")")

	return nil
}

func getReaderLogPlugin(
	blockStreamServer *blockstream.Server,
	oneBlockStoreURL string,
	mergedBlockStoreURL string,
	mergeThresholdBlockAge string,
	workingDir string,
	batchStartBlockNum uint64,
	batchStopBlockNum uint64,
	blocksChanCapacity int,
	waitTimeForUploadOnShutdown time.Duration,
	oneBlockFileSuffix string,
	operatorShutdownFunc func(error),
	metricsAndReadinessManager *nodeManager.MetricsAndReadinessManager,
	appLogger *zap.Logger,
	appTracer logging.Tracer,
) (*mindreader.MindReaderPlugin, error) {

	consoleReaderFactory := func(lines chan string) (mindreader.ConsolerReader, error) {
		return codec.NewConsoleReader(lines, NodeRPCAddr)
	}

	return mindreader.NewMindReaderPlugin(
		oneBlockStoreURL,
		workingDir,
		consoleReaderFactory,
		batchStartBlockNum,
		batchStopBlockNum,
		blocksChanCapacity,
		metricsAndReadinessManager.UpdateHeadBlock,
		func(error) {
			operatorShutdownFunc(nil)
		},
		oneBlockFileSuffix,
		blockStreamServer,
		appLogger,
		appTracer,
	)
}

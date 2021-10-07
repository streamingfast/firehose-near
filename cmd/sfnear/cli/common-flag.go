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
	"github.com/spf13/cobra"
	"github.com/streamingfast/dlauncher/launcher"
)

func init() {
	launcher.RegisterCommonFlags = func(cmd *cobra.Command) error {
		//Common stores configuration flags
		cmd.Flags().String("common-blocks-store-url", MergedBlocksStoreURL, "[COMMON] Store URL (with prefix) where to read/write. Used by: relayer, fluxdb, trxdb-loader, blockmeta, search-indexer, search-live, search-forkresolver")
		cmd.Flags().String("common-oneblock-store-url", OneBlockStoreURL, "[COMMON] Store URL (with prefix) to read/write one-block files. Used by: mindreader, merger")
		cmd.Flags().String("common-blockstream-addr", RelayerServingAddr, "[COMMON] gRPC endpoint to get real-time blocks. Used by: fluxdb, trxdb-loader, blockmeta, search-indexer, search-live (relayer uses its own --relayer-blockstream-addr)")
		//cmd.Flags().StringSlice("common-trxstream-addresses", []string{MindreaderGRPCAddr}, "[COMMON] gRPC endpoint to get transaction stream. Used by: dweb3, trx-state-tracker")
		//
		//// Network config
		//cmd.Flags().Uint32("common-chain-id", DefaultChainID, "[COMMON] ETH chain ID (from EIP-155) as returned from JSON-RPC 'eth_chainId' call Used by: dgraphql")
		//cmd.Flags().Uint32("common-network-id", DefaultNetworkID, "[COMMON] ETH network ID as returned from JSON-RPC 'net_version' call. Used by: miner-geth-node, mindreader-geth-node, mindreader-openeth-node, peering-geth-node, peering-openeth-node")
		//cmd.Flags().String("common-dfuse-network-id", DefaultDfuseNetworkID, "[COMMON] dfuse network ID, used for some billing functions by dgraphql")
		//

		// Authentication, metering and rate limiter plugins
		cmd.Flags().String("common-auth-plugin", "null://", "[COMMON] Auth plugin URI, see dfuse-io/dauth repository")
		cmd.Flags().String("common-metering-plugin", "null://", "[COMMON] Metering plugin URI, see dfuse-io/dmetering repository")
		// cmd.Flags().String("common-ratelimiter-plugin", "null://", "[COMMON] Rate Limiter plugin URI, see dfuse-io/dauth repository")

		//
		////// Database connection strings
		//cmd.Flags().String("common-trxdb-dsn", TrxdbDSN, "[COMMON] kvdb connection string to trxdb database. Used by: trxdb-loader, dgraphql")
		//
		////// RPC access
		//cmd.Flags().String("common-rpc-endpoint", NodeRPCPort, "[COMMON] RPC endpoint to use to perform Ethereum JSON-RPC. Used by: dweb3, tokenmeta")

		// System Behavior
		cmd.Flags().Duration("common-system-shutdown-signal-delay", 0, "[COMMON] Add a delay between receiving SIGTERM signal and shutting down apps. Apps will respond negatively to /healthz during this period")

		//// Service addresses
		//cmd.Flags().String("common-search-addr", RouterServingAddr, "[COMMON] gRPC endpoint to reach the Search Router. Used by: abicodec, dgraphql")
		//cmd.Flags().String("common-blockmeta-addr", BlockmetaServingAddr, "[COMMON] gRPC endpoint to reach the Blockmeta. Used by: search-indexer, search-router, search-live, dgraphql")
		//cmd.Flags().String("common-evm-executor-addr", EVMExecutorGRPCServingAddr, "[COMMON] gRPC endpoint to reach the EVM executor. Used by: trx-state-tracker. Feature disabled if empty")

		//// Search flags
		//// Register common search flags once for all the services
		//cmd.Flags().String("search-common-mesh-store-addr", "", "[COMMON] Address of the backing etcd cluster for mesh service discovery.")
		//cmd.Flags().String("search-common-mesh-dsn", DmeshDSN, "[COMMON] Dmesh DSN, supports local & etcd")
		//cmd.Flags().String("search-common-mesh-service-version", DmeshServiceVersion, "[COMMON] Dmesh service version (v1)")
		//cmd.Flags().Duration("search-common-mesh-publish-interval", 0*time.Second, "[COMMON] How often does search archive poll dmesh")
		//cmd.Flags().String("search-common-indices-store-url", IndicesStoreURL, "[COMMON] Indices path to read or write index shards Used by: search-indexer, search-archiver.")
		//cmd.Flags().String("search-common-indexed-terms", ethSearch.DefaultIndexedTerms, "[COMMON] Comma separated list of terms available for indexing. These include: calltype, callindex, signer, nonce, from, to, value, method, balancechange, storagechange, input.<index>, address, topic.<index>, data.<index>, where <index> for data & input is a number between 0 to 15 and 0 to 3 for topic.")

		return nil
	}
}

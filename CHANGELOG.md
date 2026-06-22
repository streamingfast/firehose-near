# Change log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). See [MAINTAINERS.md](./MAINTAINERS.md)
for instructions to keep up to date.

## [2.5.1]

* Bumped to [firehose-core v1.14.6](https://github.com/streamingfast/firehose-core/releases/tag/v1.14.6).

## [2.5.0]

* Bumped to [firehose-core v1.11.3](https://github.com/streamingfast/firehose-core/releases/tag/v1.12.7).

* Bumped to [substreams v1.17.11](https://github.com/streamingfast/substreams/releases/tag/v1.17.11).

## [2.4.0]

* Bumped to [firehose-core v1.12.1](https://github.com/streamingfast/firehose-core/releases/tag/v1.12.1).
* Added protobuf definitions for nearcore 2.10.0-rc.1

## [2.3.0]

* Bumped to [firehose-core v1.11.3](https://github.com/streamingfast/firehose-core/releases/tag/v1.11.3).

### Firehose Core Changes (copied over, see diff v1.10.2 => v1.11.3)

* Improved panic message when reader node encounter a block whose finality is bigger than the block itself to include `lib_num`, `block_num`, `distance`, and `max_distance` for easier debugging.

* Updated `firehose-networks` dependency to `v0.2.2` (latest).

* Fixed `common-one-block-store-url` flag not expanding environment variables in all apps.

#### Substreams v1.16.6

* Updated Wasmtime runtime from v30.0.0 to v36.0.0, bringing performance improvements, inlining support, Component Model async implementation, and enhanced security features.
* Added WASM bindgen shims support for Wasmtime runtime to handle WASM modules with WASM bindgen imports (when Substreams Module binary is defined as type `wasm/rust-v1+wasm-bindgen-shims`).
* Added support for foundational-store (in wasmtime and wazero).
* Added foundational-store grpc client to substreams engine.
* Fixed module caching to properly handle modules with different runtime extensions.
* 'paymentgateway' metering plugin renamed to `tgm`,  now supports the `indexer-api-key` parameter.

##### Tier1 thread / memory leak

* Fix thread leak on filereader.

* If `--advertise-chain-name` is sey, `substreams-tier1` app will now infer default `--substreams-tier1-block-type` value by using chain's name and extracting chain's block type Protobuf package id, which will fix some cases where `substreams-tier1` waits for 100 blocks before starting up.

##### Authentication changes

People using their own authentication layer will need to consider these changes before upgrading!

* Renamed config headers that come from authentication layer:
  - `x-sf-user-id` renamed to `x-user-id` (from dauth module)
  - `x-sf-api-key-id` renamed to `x-api-key-id` (from dauth module)
  - `x-sf-meta` renamed to `x-meta` (from dauth module)
  - `x-sf-substreams-parallel-jobs` renamed to `x-substreams-parallel-workers`
* Allow decreasing `x-substreams-parallel-workers` through an HTTP headers (auth layer determines higher bound)
* Detect value for the 'stage layer parallel executor max count' based on the `x-plan-tier` header (removed `x-sf-substreams-stage-layer-parallel-executor-max-count` handling)

##### New authentication plugin

* Added `tgm://auth.thegraph.market?indexer-api-key=<API_KEY>&reissue-jwt-max-age-secs=600` plugin that allows an indexer to use The Graph Market as the authentication source.
  An API key with special "indexer" feature is needed to allow repeated calls to the API without rate limiting (for Key-based authentication and reissuance of "untrusted long-lived JWTs").

##### Session (stream + workers management)

* Concurrent streams and workers limits are now handled under the new session plugin, available under `common-session-plugin` argument.

* The following flags were removed, now handled by that session plugin
  - `substreams-tier1-global-worker-pool-address`
  - `substreams-tier1-global-request-pool-address`
  - `substreams-tier1-global-worker-pool-keep-alive-delay`
  - `substreams-tier1-global-request-pool-keep-alive-delay`
  - `substreams-tier1-default-max-request-per-use`
  - `substreams-tier1-default-minimal-request-life-time-second`

* To use thegraph.market as a session plugin, use:
  `--common-session-plugin=tgm://session.thegraph.market:443?indexer-api-key={your-api-key}` (requires specific indexer API key)
  see https://github.com/streamingfast/tgm-gateway/tree/develop/session for details on the various flags

* To use simple local session management, use:
  `--common-session-plugin=local://?max_sessions=30&max_sessions_per_user=3&max_workers_per_user=10&max_workers_per_session=10`
  see https://github.com/streamingfast/dsession/tree/main/local for details on those flags

* Note: The 'max_sessions' parameter from the `common-session-plugin` is now also used to limit the number of firehose streams.

* If you were using a custom GRPC implementation for `--substreams-tier1-global-worker-pool-address` and `--substreams-tier1-global-request-pool-address` (ex: localhost:9010),
  simply use this value for the session plugin: `--common-session-plugin=tgm://localhost:9010?plaintext=true`, it is compatible.

##### Stability

* Fix a slow memory leak around metering plugin on tier2
* Add a maximum execution time for a full tier2 segment. By default, this is 60 minutes. It will fail with `rpc error: code = DeadlineExceeded desc = request active for too long`.
  It can be configured from the --substreams-tier2-segment-execution-timeout flag
* Fix `subscription channel at max capacity` error: when the LIVE channel is full (ex: slow module execution or slow client reader), the request will be continued from merged files instead of failing, and gracefully recover if performance is restored.
* Improve log message for 'request active for a long time', adding stats.

#### CLI

* Improved how `firenear tools --output=protojson` and `firenear tools --output=json` renders `pbbstream.Block` type now printing the underlying chain's specific block.

## [2.2.2]

* Re-release of 2.2.1, with latest firehose-core

## [2.2.1]

* Re-release of 2.2.0 but now with Docker images built with `Ubuntu 24.04`.

## [2.2.0]

### Substreams

* Fix another `cannot resolve 'old cursor' from files in passthrough mode -- not implemented` bug when receiving a request in production-mode with a cursor that is below the "linear handoff" block

* Rust modules will now be executed with `wasmtime` by default instead of `wazero`.
  - Prevents the whole server from stalling in certain memory-intensive operations in wazero.
  - Speed improvement: cuts the execution time in half in some circumstances.
  - Wazero is still used for modules with `wbindgen` and modules compiled with `tinygo`.
  - Set env var `SUBSTREAMS_WASM_RUNTIME=wazero` to revert to previous behavior.

* Implement "QuickSave" feature to save the state of "live running" substreams stores when shutting down, and then resume processing from that point if the cursor matches.
  - Added flag `substreams-tier1-quicksave-store` to enable quicksave when non-empty
    (requires `--common-system-shutdown-signal-delay` to be set to a long enough value to save the in-flight stores)

- The `substreams-tier1` app now has two new configuration flags named respectively `substreams-tier1-active-requests-soft-limit` and `substreams-tier1-active-requests-hard-limit`
  helping better load balance active requests across a pool of `tier1` instances.

  The `substreams-tier1-active-requests-soft-limit` limits the number of client active requests that a tier1 accepts before starting
  to be report itself as 'unready' within the health check endpoint. A limit of 0 or less means no limit.

  This is useful to load balance active requests more easily across a pool of tier1 instance. When the instance reaches the soft
  limit, it will start to be unready from the load balancer standpoint. The load balancer in return will remove it from the list
  of available instances, and new connections will be routed to remaining clients, spreading the load.

      The `substreams-tier1-active-requests-hard-limit` limits the number of client active requests that a tier1 accepts before

  rejecting incoming gRPC requests with 'Unavailable' code and setting itself as unready. A limit of 0 or less means no limit.

  This is useful to prevent the tier1 from being overwhelmed by too many requests, most client auto-reconnects on 'Unavailable' code
  so they should end up on another tier1 instance, assuming you have proper auto-scaling of the number of instances available.

- The `substreams-tier1` app now exposes a new Prometheus metric `substreams_tier1_rejected_request_counter` that tracks rejected
  requests. The counter is labelled by the gRPC/ConnectRPC returned code (`ok` and `canceled` are not considered rejected requests).

- The `substreams-tier2` app now exposes a new Prometheus metric `substreams_tier2_rejected_request_counter` that tracks rejected
  requests. The counter is labelled by the gRPC/ConnectRPC returned code (`ok` and `canceled` are not considered rejected requests).

- Properly accept and compress responses with `gzip` for browser HTTP clients using ConnectWeb with `Accept-Encoding` header
- Allow setting subscription channel max capacity via `SOURCE_CHAN_SIZE` env var (default: 100)

- Fix an issue preventing proper detection of gzip compression when multiple headers are set (ex: python grpc client)
- Add support for zstd compression on server
- Fix an issue preventing some tier2 requests on last-stage from correctly generating stores. This could lead to some missing "backfilling" jobs and slower time to first block on reconnection.
- Fix a thread leak on cursor resolution resulting in bad counter for active connections
> **Note** All caches for stores using the updatePolicy `set_sum` (added in substreams v1.7.0) and modules that depend on them will need to be deleted, since they may contain bad data.

- Fix bad data in stores using `set_sum` policy: squashing of store segments incorrectly "summed" some values that should have been "set" if the last event for a key on this segment was a "sum"
- Fix small bug making some requests in development-mode slow to start (when starting close to the module initialBlock with a store that doesn't start on a boundary)
- Fixed an(other) issue where multiple stores running on the same stage with different initialBlocks will fail to proress (and hang)
- Fix "cannot resolve 'old cursor' from files in passthrough mode" error on some requests with an old cursor
- Fix handling of 'special case' substreams module with only "params" as its input: should not skip this execution (used in graph-node for head tracking)
  -> empty files in module cache with hash `d3b1920483180cbcd2fd10abcabbee431146f4c8` should be deleted for consistency
- Fix bug where some invalid cursors may be sent (with 'LIB' being above the block being sent) and add safeguard/loggin if the bug appears again
- Fix panic in the whole tier2 process when stores go above the size limit while being read from "kvops" cached changes

#### Capacity Management

* Integrated the `GlobalRequestPool` service in the `Tier1App` to manage global requests pooling.
* Integrated the `GlobalWorkerPool` service in the `Tier1App` to manage global worker pooling.

* Added flag `substreams-tier1-global-worker-pool-address`, the address of the global worker pool to use for the substreams tier1. (disabled if empty)
* Added flag `substreams-tier1-global-worker-pool-keep-alive-delay` delay between two keep alive call to the global worker pool. Default is 25s")
* Added flag `substreams-tier1-global-request-pool-keep-alive-delay` delay between two keep alive call to the global worker pool for request. Default is 25s
* Added flag `substreams-tier1-default-max-request-per-user` default max request per user, this will be use of the global worker pool is not reachable. Default is 5
* Added flag `substreams-tier1-default-minimal-request-life-time-second` default minimal request life time, this will be use of the global worker pool is not reachable. . Default is 180

* Limit parallel execution of a stage's layer: Previously, the engine was executing modules in a stage's layer all in parallel.
  We now change that behavior, development mode will from now on execute every sequentially and when in production mode will
  limit parallelism to 2 (hard-coded) for now.
  The auth plugin can control that value dynamically by providing a trusted header `X-Sf-Substreams-Stage-Layer-Parallel-Executor-Max-Count`.

#### Performance

* Add shared cache for tier1 execution near HEAD, to prevent multiple tier1 instances from reprocessing the same module on the same block when it comes in (ex: foundational modules)
* Improved fetching of state caches on tier1 requests to speed up "time to first data"

* Fixed a regression since "v1.7.3" where the SkipEmptyOutput instruction was ignored in substreams mappers

### Tools

* make 'compare-blocks' command support one-blocks stores as well as merged-blocks

* The `firenear tools print one-block` is now able to print from a file directly.

- Improved logging of requests beginning/end
- Improved `noop` mode (now sends less data)

- fix panic when using an index that allows `skip_empty_output`

- Fixed `substreams-tier2` not setting itself ready correctly on startup since `v1.7.0`.

- Added support for `--output=bytes` mode which prints the chain's specific Protobuf block as bytes, the encoding for the bytes string printed is determined by `--bytes-encoding`, uses `hex` by default.

- Added back `-o` as shortand for `--output` in `firenear tools ...` sub-commands.

- Add back `grpc.health.v1.Health` service to `firehose` and `substreams-tier1` services (regression in 1.7.0)
- Give precedence to the tracing header `X-Cloud-Trace-Context` over `Traceparent` to prevent user systems' trace IDs from leaking passed a GCP load-balancer

- Reader Node Manager HTTP API now accepts `POST http://localhost:10011/v1/restart<?sync=true>` to restart the underlying reader node binary sub-process. This is a alias for `/v1/reload`.

- Enhanced `firenear tools print merged-blocks` with various small quality of life improvements:
  - Now accepts a block range instead of a single start block.
  - Passing a single block as the block range will print this single block alone.
  - Block range is now optional, defaulting to run until there is no more files to read.
  - It's possible to pass a merged blocks file directly, with or without an optional range.

### Firehose

> [!IMPORTANT]
> This release will reject firehose connections from clients that don't support GZIP or ZSTD compression. Use `--firehose-enforce-compression=false` to keep previous behavior, then check the logs for `incoming Substreams Blocks request` logs with the value `compressed: false` to track users who are not using compressed HTTP connections.

> [!IMPORTANT]
> This release removes the old `sf.firehose.v1` protocol (replaced by `sf.firehose.v2` in 2022, this should not affect any reasonably recent client).

- Add support for ConnectWeb firehose requests.
- Always use gzip compression on firehose requests for clients that support it (instead of always answering with the same compression as the request).

> [!NOTE]
> This release will reject substreams connections from clients that don't support GZIP compression. Use `--substreams-tier1-enforce-compression=false` to keep previous behavior, then check the logs for `incoming Substreams Blocks request` logs with the value `compressed: false` to track users who are not using compressed HTTP connections.

- Substreams: add `--substreams-tier1-enforce-compression` to reject connections from clients that do not support GZIP compression
- Substreams performance: reduced the number of `mallocs` (patching some third-party libraries)
- Substreams performance: removed heavy tracing (that wasn't exposed to the client)
- Fixed `reader-node-line-buffer-size` flag that was not being respected in `reader-node-stdin` app
- Well-known chains: change genesis block for near-mainnet from 9820214 to 9820210
- BlockPoller library: reworked logic to support more flexible balancing strategy

- `firehose-grpc-listen-addr` and `substreams-tier1-grpc-listen-addr` flags now accepts comma-separated addresses (allows listening as plaintext and snakeoil-ssl at the same time or on specific ip addresses)
- removed old `RegisterServiceExtension` implementation (not used anywhere anymore)
- rpc-poller lib: fix fetching the first block on an endpoint (was not following the cursor, failing unnecessarily on non-archive nodes)

- Bump `substreams` and `dmetering` to latest version adding the `outputModuleHash` to metering sender.

- [Operator] Node Manager HTTP `/v1/resume` call now accepts `extra-env=<key>=<value>&extra-env=<keyN>=<valueN>` enabling to override environment variables for the next restart **only**. Use `curl -XPOST "http://localhost:10011/v1/resume?sync=true&extra-env=NODE_DEBUG=true"` (change `localhost:10011` accordingly to your setup).

  > This is **not** persistent upon restart!

- [Metering] Revert undesired Firehose metric `Endpoint` changes, the correct new value used is `sf.firehose.v2.Firehose/Blocks` (had been mistakenly set to `sf.firehose.v2.Firehose/Block` between version v1.6.1 and v1.6.4 inclusively).

- fix: reader-node-stdin not shutting down after receiving an EOF

## [2.1.0]

* Bump protobuf model to be compatible with v2.5.0-rc.2

## [2.0.0]

* Update to use `firehose-core`

## [1.1.14](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.14)

* Fixed `reader block stats` to print properly time of importing block.

## [1.1.13](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.13)

### Substreams

* Performance: prevent reprocessing jobs when there is only a mapper in production mode and everything is already cached
* Performance: prevent "UpdateStats" from running too often and stalling other operations when running with a high parallel jobs count
* Performance: fixed bug in scheduler ramp-up function sometimes waiting before raising the number of workers

## [1.1.12](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.12)

* 1.1.12 replaces the deleted release 1.1.11, which should not be used

### Operators

> [!IMPORTANT]
> We have had reports of older versions of this software creating corrupted merged-blocks-files (with duplicate or extra out-of-bound blocks)
> This release adds additional validation of merged-blocks to prevent serving duplicate blocks from the firehose or substreams service.
> This may cause service outage if you have produced those blocks or downloaded them from another party who was affected by this bug.

1. Find the affected files by running the following command (can be run multiple times in parallel, over smaller ranges)

```
tools check merged-blocks-batch <merged-blocks-store> <start> <stop>
```

2. If you see any affected range, produce fixed merged-blocks files with the following command, on each range:

```
tools fix-bloated-merged-blocks <merged-blocks-store> <output-store> <start>:<stop>
```

3. Copy the merged-blocks files created in output-store over to the your merged-blocks-store, replacing the corrupted files.

### Added

* Added `tools check merged-blocks-batch` to simplify checking blocks continuity in batched mode, optionally writing results to a store
* Added the command `tools fix-bloated-merged-blocks` to try to fix merged-blocks that contain duplicates and blocks outside of their range.
* Command `tools print one-block and merged-blocks` now supports a new `--output-format` `jsonl` format. Bytes data can now printed as hex or base58 string instead of base64 string.
* Added retry loop for merger when walking one block files. Some use-cases where the bundle reader was sending files too fast and the merger was not waiting to accumulate enough files to start bundling merged files
* Firehose logs now include auth information (userID, keyID, realIP) along with blocks + egress bytes sent.

### Fixed

* Bumped `bstream`: the `filesource` will now refuse to read blocks from a merged-files if they are not ordered or if there are any duplicate.
* The command `tools download-from-firehose` will now fail if it is being served blocks "out of order", to prevent any corrupted merged-blocks from being created.
* The command `tools print merged-blocks` did not print the whole merged-blocks file, the arguments were confusing: now it will parse <start_block> as a uint64.
* The command `tools unmerge-blocks` did not cover the whole given range, now fixed

### Removed

* **Breaking** The `reader-node-log-to-zap` flag has been removed. This was a source of confusion for operators reporting Firehose on <Chain> bugs because the node's logs where merged within normal Firehose on <Chain> logs and it was not super obvious.

  Now, logs from the node will be printed to `stdout` unformatted exactly like presented by the chain. Filtering of such logs must now be delegated to the node's implementation and how it deals depends on the node's binary. Refer to it to determine how you can tweak the logging verbosity emitted by the node.
* Flag substreams-rpc-endpoints removed, this was present by mistake and unused actually.
* Flag substreams-rpc-cache-store-url removed, this was present by mistake and unused actually.
* Flag substreams-rpc-cache-chunk-size removed, this was present by mistake and unused actually.

## 1.1.11

* This version has been deleted because it contains a critical bug: it produced blocks with invalid version number

## [1.1.10](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.10)

* bump firehose-core to `v0.1.12`

## [1.1.9](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.9)

### Changed

* bump firehose-core to `v0.1.10` with regression fix for start block in reversible segment


## [1.1.8](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.8)

### Changed

* bump firehose-core to `v0.1.10` with new metrics `substreams_active_requests` and `substreams_counter`


## [1.1.7](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.7)

* Bumped firehose-core to `0.1.9`
* Go version requirement bumped to `1.21`

## [1.1.6](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.6)

> [!IMPORTANT]
> The Substreams service exposed from this version will send progress messages that cannot be decoded by substreams clients prior to v1.1.12.
> Streaming of the actual data will not be affected. Clients will need to be upgraded to properly decode the new progress messages.

### Changed

* Bumped firehose-core to `0.1.8`
* Bumped substreams to `v1.1.12` to support the new progress message format. Progression now relates to **stages** instead of modules. You can get stage information using the `substreams info` command starting at version `v1.1.12`.
* Migrated to firehose-core
* change block reader-node block encoding from hex to base64

### Removed

*  Removed --substreams-tier1-request-stats and --substreams-tier1-request-stats (substreams request-stats are now always sent to clients)

### Fixed

* More tolerant retry/timeouts on filesource (prevent "Context Deadline Exceeded")

## [1.1.5-rc1](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.5-rc1)

This release candidate is a hotfix for an issue introduced at block v1.1.3 and affecting `production-mode` where the stream will hang and some `map_outputs` will not be produced over some specific ranges of the chains.

## [1.1.4](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.4)

This release bumps substreams to v1.1.10 and firehose-core to v0.1.4

### Fixes

* Fixed: jobs would hang when flags `--substreams-state-bundle-size` and `--substreams-tier1-subrequests-size` had different values. The latter flag has been completely **removed**, subrequests will be bound to the state bundle size.

### Added

* Added support for *continuous authentication* via the grpc auth plugin (allowing cutoff triggered by the auth system).


## [1.1.3](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.3)

This release bumps substreams to v1.1.9 and firehose-core to v0.1.3

### Highlights

#### Substreams Scheduler Improvements for Parallel Processing

The `substreams` scheduler has been improved to reduce the number of required jobs for parallel processing. This affects `backprocessing` (preparing the states of modules up to a "start-block") and `forward processing` (preparing the states and the outputs to speed up streaming in production-mode).

Jobs on `tier2` workers are now divided in "stages", each stage generating the partial states for all the modules that have the same dependencies. A `substreams` that has a single store won't be affected, but one that has 3 top-level stores, which used to run 3 jobs for every segment now only runs a single job per segment to get all the states ready.


#### Substreams State Store Selection

The `substreams` server now accepts `X-Sf-Substreams-Cache-Tag` header to select which Substreams state store URL should be used by the request. When performing a Substreams request, the servers will optionally pick the state store based on the header. This enable consumers to stay on the same cache version when the operators needs to bump the data version (reasons for this could be a bug in Substreams software that caused some cached data to be corrupted on invalid).

To benefit from this, operators that have a version currently in their state store URL should move the version part from `--substreams-state-store-url` to the new flag `--substreams-state-store-default-tag`. For example if today you have in your config:

```yaml
start:
  ...
  flags:
    substreams-state-store-url: /<some>/<path>/v3
```

You should convert to:

```yaml
start:
  ...
  flags:
    substreams-state-store-url: /<some>/<path>
    substreams-state-store-default-tag: v3
```

### Operators Upgrade

The app `substreams-tier1` and `substreams-tier2` should be upgraded concurrently. Some calls will fail while versions are misaligned.

### Backend Changes

* Authentication plugin `trust` can now specify an exclusive list of `allowed` headers (all lowercase), ex: `trust://?allowed=x-sf-user-id,x-sf-api-key-id,x-real-ip,x-sf-substreams-cache-tag`

* The `tier2` app no longer uses the `common-auth-plugin`, `trust` will always be used, so that `tier1` can pass down its headers (ex: `X-Sf-Substreams-Cache-Tag`).

* Fixed some loggers to not render a shortname (so appearing as `<n/a>` in the log).

### CLI Changes

* Added `firenear tools check forks <forked-blocks-store-url> [--min-depth=<depth>]` that reads forked blocks you have and prints resolved longest forks you have seen. The command works for any chain, here a sample output:

    ```log
    ...

    Fork Depth 3
    #45236230 [ea33194e0a9bb1d8 <= 164aa1b9c8a02af0 (on chain)]
    #45236231 [f7d2dc3fbdd0699c <= ea33194e0a9bb1d8]
        #45236232 [ed588cca9b1db391 <= f7d2dc3fbdd0699c]

    Fork Depth 2
    #45236023 [b6b1c68c30b61166 <= 60083a796a079409 (on chain)]
    #45236024 [6d64aec1aece4a43 <= b6b1c68c30b61166]

    ...
    ```

* The `firenear tools` commands and sub-commands have better rendering `--help` by hidden not needed global flags with long description.

## [1.1.2](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.2)

#### Backend Changes

### Highlights

This release brings various renames to fully align with all Firehose <Chain> out there. The repository is now using `firehose-core` which should make easier to follow up with latest Firehose/Substreams feature(s).

This brings in a few breaking changes to align the flags across all chains.

### Breaking Changes

* Removed support for `archive-node` app, if you were using this, please use a standard NEAR Archive node to do the same job.

* Flag `common-block-index-sizes` has been renamed to `common-index-block-sizes`.

* String variable `{sf-data-dir}` which interpolates at runtime to Firehose data directory is now `{data-dir}`. If any of your parameter value has `{sf-data-dir}` in its value, change it to `{data-dir}`.

  > **Note** This is an important change, forgetting to change it will change expected locations of data leading to errors or wrong data.

* The default value for `config-file` changed from `sf.yaml` to `firehose.yaml`. If you didn't had this flag defined and wish to keep the old default, define `config-file: sf.yaml`.

* The default value for `data-dir` changed from `sf-data` to `firehose-data`. If you didn't had this flag defined before, you should either move `sf-data` to `firehose-data` or define `data-dir: sf-data`.

  > **Note** This is an important change, forgetting to change it will change expected locations of data leading to errors or wrong data.

* The flag `verbose` has been renamed to `log-verbosity`.

* The default value for `common-blocks-cache-dir` changed from `{sf-data-dir}/blocks-cache` to `file://{data-dir}/storage/blocks-cache`. If you didn't had this flag defined and you had `common-blocks-cache-enabled: true`, you should define `common-blocks-cache-dir: {data-dir}/blocks-cache`.

* The default value for `common-live-blocks-addr` changed from `:15011` to `:10014`. If you didn't had this flag defined and wish to keep the old default, define `common-live-blocks-addr: 15011` and ensure you also modify `relayer-grpc-listen-addr: :15011` (see next entry for details).

* The default value for `relayer-grpc-listen-addr` changed from `:15011` to `:10014`. If you didn't had this flag defined and wish to keep the old default, define `relayer-grpc-listen-addr: 15011` and ensure you also modify `common-live-blocks-addr: :15011` (see previous entry for details).

* The default value for `relayer-source` changed from `:15010` to `:10010`. If you didn't had this flag defined and wish to keep the old default, define `relayer-source: 15010` and ensure you also modify `reader-node-grpc-listen-addr: :15010` (see next entry for details).

* The default value for `reader-node-grpc-listen-addr` changed from `:15010` to `:10010`. If you didn't had this flag defined and wish to keep the old default, define `reader-node-grpc-listen-addr: :15010` and ensure you also modify `relayer-source: :15010` (see previous entry for details).

* The default value for `reader-node-manager-api-addr` changed from `:15009` to `:10011`. If you didn't had this flag defined and wish to keep the old default, define `reader-node-manager-api-addr: :15010`.

* The `reader-node-arguments` is not populated anymore with default `--home={node-data-dir} <extra-args> run` which means you must now specify those manually. The variables `{data-dir}`, `{node-data-dir}` and `{hostname}` are interpolated respectively to Firehose absolute `data-dir` value, to Firehose absolute `reader-node-data-dir` value and to current hostname. To upgrade, if you had no `reader-node-arguments` defined, you must now define `reader-node-arguments: --home="{node-data-dir}" run`, if you had a `+` in your `reader-node-arguments: +--some-flag`, you must now define it like `reader-node-arguments: --home="{node-data-dir}" --some-flag run`.

  > **Note** This is an important change, forgetting to change it will change expected locations of data leading to errors or wrong data.

* The `reader-node-boot-nodes` flag has been removed entirely, if you have boot nodes to specify, specify them in `reader-node-arguments` using `--boot-nodes=...` instead.

* Removed unused flags `reader-node-merge-and-store-directly`, `reader-node-merge-threshold-block-age` and `reader-node-wait-upload-complete-on-shutdown`.

* The flag `receipt-index-builder-index-size` has been renamed to `index-builder-index-size`.

* The flag `receipt-index-builder-start-block` has been renamed to `index-builder-start-block`.

* The flag `receipt-index-builder-stop-block` has been renamed to `index-builder-stop-block`.

* The default value for `firehose-grpc-listen-addr` changed from `:15042` to `:10015`. If you didn't had this flag defined and wish to keep the old default, define `firehose-grpc-listen-addr: :15010`.

* The default value for `merger-grpc-listen-addr` changed from `:15012` to `:10012`. If you didn't had this flag defined and wish to keep the old default, define `merger-grpc-listen-addr: :15012`.
* Update firehose-core to v0.1.1:
  - added missing `--substreams-tier2-request-stats` request debugging flag
  - added missing firehose rate limiting options flags, `--firehose-rate-limit-bucket-size` and `--firehose-rate-limit-bucket-fill-rate` to manage concurrent connection attempts to Firehose.

## [1.1.1](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.1)

#### Backend Changes

* Fixed Substreams accepted block which was not working properly.

## [1.1.0](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.0)

### Highlights

This release brings various renames to fully align with all Firehose <Chain> out there. The repository is now using `firehose-core` which should make easier to follow up with latest Firehose/Substreams feature(s).

This brings in a few breaking changes to align the flags across all chains.

### Breaking Changes

* Removed support for `archive-node` app, if you were using this, please use a standard NEAR Archive node to do the same job.

* Flag `common-block-index-sizes` has been renamed to `common-index-block-sizes`.

* String variable `{sf-data-dir}` which interpolates at runtime to Firehose data directory is now `{data-dir}`. If any of your parameter value has `{sf-data-dir}` in its value, change it to `{data-dir}`.

  > **Note** This is an important change, forgetting to change it will change expected locations of data leading to errors or wrong data.

* The default value for `config-file` changed from `sf.yaml` to `firehose.yaml`. If you didn't had this flag defined and wish to keep the old default, define `config-file: sf.yaml`.

* The default value for `data-dir` changed from `sf-data` to `firehose-data`. If you didn't had this flag defined before, you should either move `sf-data` to `firehose-data` or define `data-dir: sf-data`.

  > **Note** This is an important change, forgetting to change it will change expected locations of data leading to errors or wrong data.

* The flag `verbose` has been renamed to `log-verbosity`.

* The default value for `common-blocks-cache-dir` changed from `{sf-data-dir}/blocks-cache` to `file://{data-dir}/storage/blocks-cache`. If you didn't had this flag defined and you had `common-blocks-cache-enabled: true`, you should define `common-blocks-cache-dir: {data-dir}/blocks-cache`.

* The default value for `common-live-blocks-addr` changed from `:15011` to `:10014`. If you didn't had this flag defined and wish to keep the old default, define `common-live-blocks-addr: 15011` and ensure you also modify `relayer-grpc-listen-addr: :15011` (see next entry for details).

* The default value for `relayer-grpc-listen-addr` changed from `:15011` to `:10014`. If you didn't had this flag defined and wish to keep the old default, define `relayer-grpc-listen-addr: 15011` and ensure you also modify `common-live-blocks-addr: :15011` (see previous entry for details).

* The default value for `relayer-source` changed from `:15010` to `:10010`. If you didn't had this flag defined and wish to keep the old default, define `relayer-source: 15010` and ensure you also modify `reader-node-grpc-listen-addr: :15010` (see next entry for details).

* The default value for `reader-node-grpc-listen-addr` changed from `:15010` to `:10010`. If you didn't had this flag defined and wish to keep the old default, define `reader-node-grpc-listen-addr: :15010` and ensure you also modify `relayer-source: :15010` (see previous entry for details).

* The default value for `reader-node-manager-api-addr` changed from `:15009` to `:10011`. If you didn't had this flag defined and wish to keep the old default, define `reader-node-manager-api-addr: :15010`.

* The `reader-node-arguments` is not populated anymore with default `--home={node-data-dir} <extra-args> run` which means you must now specify those manually. The variables `{data-dir}`, `{node-data-dir}` and `{hostname}` are interpolated respectively to Firehose absolute `data-dir` value, to Firehose absolute `reader-node-data-dir` value and to current hostname. To upgrade, if you had no `reader-node-arguments` defined, you must now define `reader-node-arguments: --home="{node-data-dir}" run`, if you had a `+` in your `reader-node-arguments: +--some-flag`, you must now define it like `reader-node-arguments: --home="{node-data-dir}" --some-flag run`.

  > **Note** This is an important change, forgetting to change it will change expected locations of data leading to errors or wrong data.

* The `reader-node-boot-nodes` flag has been removed entirely, if you have boot nodes to specify, specify them in `reader-node-arguments` using `--boot-nodes=...` instead.

* Removed unused flags `reader-node-merge-and-store-directly`, `reader-node-merge-threshold-block-age` and `reader-node-wait-upload-complete-on-shutdown`.

* The flag `receipt-index-builder-index-size` has been renamed to `index-builder-index-size`.

* The flag `receipt-index-builder-start-block` has been renamed to `index-builder-start-block`.

* The flag `receipt-index-builder-stop-block` has been renamed to `index-builder-stop-block`.

* The default value for `firehose-grpc-listen-addr` changed from `:15042` to `:10015`. If you didn't had this flag defined and wish to keep the old default, define `firehose-grpc-listen-addr: :15010`.

* The default value for `merger-grpc-listen-addr` changed from `:15012` to `:10012`. If you didn't had this flag defined and wish to keep the old default, define `merger-grpc-listen-addr: :15012`.

## [1.0.6](https://github.com/streamingfast/firehose-near/releases/tag/v1.0.6)

### Highlights

Before this release, the merger would create incorrect merged-blocks bundles if the chain skipped too many block (i.e. skipping a full bundle).

While in some cases, it would not cause any issue while reading over these blocks, it WOULD result in skipped blocks from firehose if the requested start-block was within a problematic range, or even fail to start if given a cursor within that range.

Here are the known affected block ranges showing these problems on NEAR Testnet:
* 102435000-102449000
* 102457500-102458500

### Actions required

* Upgrade your merger deployment to **v1.0.6**
* Run this command to detect ranges with invalid merged-blocks:
`firenear tools check merged-blocks /path/to/merged/blocks   -r 0:123879100 -e |tee /check-all-blocks` (the upper boundary should be adjusted to cover the chain up to the HEAD before you upgraded the merger.
* You will see this kind of output:
```
(...)
❌ invalid block 102448970 in segment 0102448800
❌ invalid block 102458195 in segment 0102457700
❌ invalid block 102458195 in segment 0102457800
(...)
```

* Regroup the affected segments into larger ranges, adding 200 blocks 'before' and 1000 blocks 'after'
* For each of these large ranges (START:END), run the following commands (with **v1.0.6**):

```
# get the one-blocks-files locally
firenear tools unmerge /path/to/your/merged/blocks /tmp/one-blocks START END

# bundle the files into new merged-blocks
firenear start merger --config-file= --common-first-streamable-block=START  --common-merged-blocks-store-url=/tmp/re-merged-blocks --common-one-block-store-url=/tmp/one-blocks --merger-stop-block=END

# copy the new merged files over
cp /tmp/re-merged-blocks/* /path/to/your/merged/blocks/

# clean up your tmp folders
rm /tmp/one-blocks/* /tmp/re-merged-blocks/*
```

## [1.0.5](https://github.com/streamingfast/firehose-near/releases/tag/v1.0.5)

### Changed

* Add s5cmd download to Docker bundle image

## [1.0.4](https://github.com/streamingfast/firehose-near/releases/tag/v1.0.4)

### Changed

* All workflows now use ubuntu-20.04

## [1.0.3](https://github.com/streamingfast/firehose-near/releases/tag/v1.0.3)

### Changed

* More fixes to GitHub workflows

## [1.0.2](https://github.com/streamingfast/firehose-near/releases/tag/v1.0.2)

### Changed

* Update GitHub workflow to use ubuntu-20.04

## [1.0.1](https://github.com/streamingfast/firehose-near/releases/tag/v1.0.1)

### Changed

* **Breaking** Flag `receipt-index-builder-lookup-index-sizes` has been replaced by `common-block-index-sizes`.
    * Migration path is to replace any flag or configuration value named `receipt-index-builder-lookup-index-sizes` by `common-block-index-sizes`.

* **Breaking** Flag `receipt-index-builder-index-store-url` has been replaced by `common-index-store-url`.
    * Migration path is to replace any flag or configuration value named `receipt-index-builder-index-store-url` by `common-index-store-url`.

* **Breaking** Flag `firehose-block-index-url` has been replaced by `common-index-store-url`.
    * Migration path is to replace any flag or configuration value named `firehose-block-index-url` by `common-index-store-url`.

### Added

* Added `firenear tools generate-node-key` command to easily generate a new `node_key.json` file.

## [1.0.0](https://github.com/streamingfast/firehose-near/releases/tag/v1.0.0)

### Added

* Added support for "requester pays" buckets on Google Storage in url, ex: `gs://my-bucket/path?project=my-project-id`

* Added support for Substreams version `v0.2.0` please refer to [release page](https://github.com/streamingfast/substreams/releases/tag/v0.2.0) for further info about Substreams changes.

## [0.3.0](https://github.com/streamingfast/firehose-near/releases/tag/v0.3.0)

* Actual release, no changelog was maintained back then

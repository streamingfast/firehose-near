# Change log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). See [MAINTAINERS.md](./MAINTAINERS.md)
for instructions to keep up to date.

## [1.1.2](https://github.com/streamingfast/firehose-near/releases/tag/v1.1.2)

#### Backend Changes

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

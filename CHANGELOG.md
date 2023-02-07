# Change log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). See [MAINTAINERS.md](./MAINTAINERS.md)
for instructions to keep up to date.

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

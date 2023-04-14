# Change log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). See [MAINTAINERS.md](./MAINTAINERS.md)
for instructions to keep up to date.

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

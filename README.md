# Near on StreamingFast
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/firehose-near)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Usage

## Release

1. Define the version information that we are about to release:

    ```bash
    version=1.0.1 # Use correct version
    ```

    > **Note** Those instructions uses [sd](https://github.com/chmln/sd#installation), `brew install sd` (or see [sd](https://github.com/chmln/sd#installation))

1. Prepare the release by updating the [CHANGELOG.md](./CHANGELOG.md) file, change `## Unreleased` to become `## [1.0.1](https://github.com/streamingfast/firehose-near/releases/tag/v1.0.1)`:

    ```bash
    sd "## Unreleased" "## [$version](https://github.com/streamingfast/firehose-near/releases/tag/v$version)" CHANGELOG.md
    ```

1. Update [substreams.yaml](./substreams/substreams.yaml) `version: v1.0.0` to `version: v1.0.1`:

    ```bash
    sd "version: v.*" "version: v$version" substreams/substreams.yaml
    ```

1. Commit to prepare release:

    ```bash
    git add CHANGELOG.md substreams/substreams.yaml
    git commit -m "Preparing for release v$version"
    ```

1. Run the [./bin/release.sh](./bin/release.sh) Bash script to perform a new release. It will ask you questions as well as driving all the required commands, performing the necessary operation automatically. The Bash script publishes a GitHub release by default, so you can check first that everything is all right.

    ```bash
    ./bin/release.sh v$version
    ```

## Contributing

**Issues and PR in this repo related strictly to the NEAR on StreamingFast.**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.

This codebase uses unit tests extensively, please write and run tests.

## License

[Apache 2.0](LICENSE)

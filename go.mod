module github.com/streamingfast/firehose-near

go 1.16

require (
	github.com/RoaringBitmap/roaring v0.9.4
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/golang/protobuf v1.5.2
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.8.1
	github.com/streamingfast/bstream v0.0.2-0.20221117104246-5660c4ba5e8c
	github.com/streamingfast/cli v0.0.4-0.20220630165922-bc58c6666fc8
	github.com/streamingfast/dauth v0.0.0-20221027185237-b209f25fa3ff
	github.com/streamingfast/dbin v0.9.1-0.20220513054835-1abebbb944ad
	github.com/streamingfast/derr v0.0.0-20221125175206-82e01d420d45
	github.com/streamingfast/dgrpc v0.0.0-20220909121013-162e9305bbfc
	github.com/streamingfast/dlauncher v0.0.0-20220909121534-7a9aa91dbb32
	github.com/streamingfast/dmetering v0.0.0-20220307162406-37261b4b3de9
	github.com/streamingfast/dmetrics v0.0.0-20221107142404-e88fe183f07d
	github.com/streamingfast/dstore v0.1.1-0.20221021155138-4baa2d406146
	github.com/streamingfast/firehose v0.1.1-0.20221101130227-3a0b1980aa0b
	github.com/streamingfast/firehose-near/types v0.0.0-20220906143314-cd1a739fc58f
	github.com/streamingfast/index-builder v0.0.0-20220810183227-6de1a5d962c7
	github.com/streamingfast/logging v0.0.0-20220813175024-b4fbb0e893df
	github.com/streamingfast/merger v0.0.3-0.20221123202507-445dfd357868
	github.com/streamingfast/near-go v0.0.0-20220302163233-b638f5b48a2d
	github.com/streamingfast/node-manager v0.0.2-0.20221115101723-d9823ffd7ad5
	github.com/streamingfast/pbgo v0.0.6-0.20221014191646-3a05d7bc30c8
	github.com/streamingfast/relayer v0.0.2-0.20220909122435-e67fbc964fd9
	github.com/streamingfast/sf-tools v0.0.0-20221129150749-9a8582eb4a04
	github.com/streamingfast/sf-tracing v0.0.0-20221104190152-7f721cb9b60c // indirect
	github.com/streamingfast/snapshotter v0.0.0-20220901201120-f4b8e3920987
	github.com/stretchr/testify v1.8.0
	github.com/tidwall/gjson v1.9.3
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.0
)

replace github.com/ShinyTrinkets/overseer => github.com/dfuse-io/overseer v0.2.1-0.20210326144022-ee491780e3ef

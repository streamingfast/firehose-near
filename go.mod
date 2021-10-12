module github.com/streamingfast/sf-near

go 1.16

require (
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/abourget/llerrgroup v0.2.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/streamingfast/bstream v0.0.2-0.20211006133909-71ab5cca098d
	github.com/streamingfast/cli v0.0.3-0.20210811201236-5c00ec55462d
	github.com/streamingfast/dauth v0.0.0-20210811181149-e8fd545948cc
	github.com/streamingfast/dbin v0.0.0-20210809205249-73d5eca35dc5
	github.com/streamingfast/derr v0.0.0-20210811180100-9138d738bcec
	github.com/streamingfast/dgrpc v0.0.0-20210901144702-c57c3701768b
	github.com/streamingfast/dlauncher v0.0.0-20211012125617-3178fefea1d5
	github.com/streamingfast/dmetering v0.0.0-20210811181351-eef120cfb817
	github.com/streamingfast/dmetrics v0.0.0-20210811180524-8494aeb34447
	github.com/streamingfast/dstore v0.1.1-0.20210811180812-4db13e99cc22
	github.com/streamingfast/firehose v0.1.1-0.20210901164748-403e4d029276
	github.com/streamingfast/logging v0.0.0-20210908162127-bdc5856d5341
	github.com/streamingfast/merger v0.0.3-0.20210913201227-b6049773cace
	github.com/streamingfast/node-manager v0.0.2-0.20210913201120-d5940f6b4634
	github.com/streamingfast/pbgo v0.0.6-0.20210820205306-ba5335146052
	github.com/streamingfast/relayer v0.0.2-0.20211006134256-f3773fc7f46d
	github.com/streamingfast/sf-tools v0.0.0-20210823043548-13a30de7c1b1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.1
	google.golang.org/grpc v1.39.1
	google.golang.org/protobuf v1.27.1
)

replace github.com/ShinyTrinkets/overseer => github.com/dfuse-io/overseer v0.2.1-0.20210326144022-ee491780e3ef

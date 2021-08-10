module github.com/streamingfast/near-sf

go 1.16

require (
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/dfuse-io/bstream v0.0.2-0.20210810055526-1c2b722f0cf6
	github.com/dfuse-io/dauth v0.0.0-20200601190857-60bc6a4b4665 // indirect
	github.com/dfuse-io/dgrpc v0.0.0-20210424033943-10e04dd5b19c
	github.com/dfuse-io/dmetering v0.0.0-20210112023524-c3ddadbc0d6a // indirect
	github.com/dfuse-io/dmetrics v0.0.0-20200508170817-3b8cb01fee68
	github.com/dfuse-io/dstore v0.1.1-0.20210507180120-88a95674809f
	github.com/dfuse-io/logging v0.0.0-20210109005628-b97a57253f70
	github.com/dfuse-io/pbgo v0.0.6-0.20210429181308-d54fc7723ad3
	github.com/golang/protobuf v1.5.2
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/streamingfast/dauth v0.0.0-20210809192433-4c758fd333ac
	github.com/streamingfast/dbin v0.0.0-20210809205249-73d5eca35dc5
	github.com/streamingfast/derr v0.0.0-20210810022442-32249850a4fb
	github.com/streamingfast/dlauncher v0.0.0-20210809203103-fe27d137acc7
	github.com/streamingfast/dmetering v0.0.0-20210809193048-81d008c90843
	github.com/streamingfast/firehose v0.1.1-0.20210809193802-776cf9f9942e
	github.com/streamingfast/merger v0.0.3-0.20210810060447-9dbf5a0ab9e9
	github.com/streamingfast/node-manager v0.0.2-0.20210810060535-08c7a3f7fa5a
	github.com/streamingfast/relayer v0.0.2-0.20210809195208-c686bf91e083
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/ShinyTrinkets/overseer => github.com/dfuse-io/overseer v0.2.1-0.20210326144022-ee491780e3ef

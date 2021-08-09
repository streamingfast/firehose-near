module github.com/streamingfast/near-sf

go 1.16

require (
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/dfuse-io/bstream v0.0.2-0.20210218160250-ce6144227e87
	github.com/dfuse-io/dauth v0.0.0-20210330175213-9154c2cf75be
	github.com/dfuse-io/dbin v0.0.0-20200406215642-ec7f22e794eb
	github.com/dfuse-io/derr v0.0.0-20201001203637-4dc9d8014152
	github.com/dfuse-io/dgrpc v0.0.0-20210424033943-10e04dd5b19c
	github.com/dfuse-io/dlauncher v0.0.0-20210401132540-cc35cfce1757
	github.com/dfuse-io/dmetering v0.0.0-20210208201314-50d9add52005
	github.com/dfuse-io/dmetrics v0.0.0-20200508170817-3b8cb01fee68
	github.com/dfuse-io/dstore v0.1.1-0.20210507180120-88a95674809f
	github.com/dfuse-io/firehose v0.1.1-0.20210424034522-6e5c05a23323
	github.com/dfuse-io/logging v0.0.0-20210109005628-b97a57253f70
	github.com/dfuse-io/node-manager v0.0.2-0.20210723133121-26b375a773ec
	github.com/dfuse-io/pbgo v0.0.6-0.20210429181308-d54fc7723ad3
	github.com/dfuse-io/relayer v0.0.2-0.20210525190602-eeec872ddedf
	github.com/golang/protobuf v1.5.2
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/streamingfast/merger v0.0.3-0.20210809165038-14f85d21b69b
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/ShinyTrinkets/overseer => github.com/dfuse-io/overseer v0.2.1-0.20210326144022-ee491780e3ef

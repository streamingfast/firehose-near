module github.com/streamingfast/near-sf

go 1.16

require (
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/dfuse-io/bstream v0.0.2-0.20210218160250-ce6144227e87
	github.com/dfuse-io/dbin v0.0.0-20200406215642-ec7f22e794eb
	github.com/dfuse-io/derr v0.0.0-20201001203637-4dc9d8014152
	github.com/dfuse-io/dgrpc v0.0.0-20210128133958-db1ca95920e4
	github.com/dfuse-io/dlauncher v0.0.0-20210401132540-cc35cfce1757
	github.com/dfuse-io/logging v0.0.0-20210109005628-b97a57253f70
	github.com/dfuse-io/merger v0.0.3-0.20210226144304-7e370a347999
	github.com/dfuse-io/node-manager v0.0.2-0.20210716050720-2c168713e51c
	github.com/dfuse-io/pbgo v0.0.6-0.20210429181308-d54fc7723ad3
	github.com/golang/protobuf v1.5.2
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/ShinyTrinkets/overseer => github.com/dfuse-io/overseer v0.2.1-0.20210326144022-ee491780e3ef

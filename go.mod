module github.com/streamingfast/near-sf

go 1.16

require (
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/dfuse-io/bstream v0.0.2-0.20210218160250-ce6144227e87
	github.com/dfuse-io/dauth v0.0.0-20210330175213-9154c2cf75be // indirect
	github.com/dfuse-io/derr v0.0.0-20201001203637-4dc9d8014152
	github.com/dfuse-io/dgrpc v0.0.0-20210128133958-db1ca95920e4
	github.com/dfuse-io/dlauncher v0.0.0-20210401132540-cc35cfce1757
	github.com/dfuse-io/logging v0.0.0-20210109005628-b97a57253f70
	github.com/dfuse-io/node-manager v0.0.2-0.20210510211158-85801370a2bf
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/tidwall/gjson v1.5.0
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.38.0 // indirect
)


replace github.com/ShinyTrinkets/overseer => github.com/dfuse-io/overseer v0.2.1-0.20210326144022-ee491780e3ef
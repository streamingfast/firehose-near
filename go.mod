module github.com/streamingfast/near-sf

go 1.16

require (
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/abourget/llerrgroup v0.2.0 // indirect
	github.com/dfuse-io/bstream v0.0.2-0.20210810200727-06a4c4aaeb99
	github.com/dfuse-io/logging v0.0.0-20210518215502-2d920b2ad1f2
	github.com/dfuse-io/pbgo v0.0.6-0.20210810190312-2053a3b38779
	github.com/golang/protobuf v1.5.2
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/streamingfast/dauth v0.0.0-20210809192433-4c758fd333ac
	github.com/streamingfast/dbin v0.0.0-20210809205249-73d5eca35dc5
	github.com/streamingfast/derr v0.0.0-20210810022442-32249850a4fb
	github.com/streamingfast/dgrpc v0.0.0-20210810185305-905172f728e8
	github.com/streamingfast/dhammer v0.0.0-20210810184929-89abe4f2b612 // indirect
	github.com/streamingfast/dlauncher v0.0.0-20210810201623-b4bad04a0782
	github.com/streamingfast/dmesh v0.0.0-20210810205752-f210f374556e // indirect
	github.com/streamingfast/dmetering v0.0.0-20210809193048-81d008c90843
	github.com/streamingfast/dmetrics v0.0.0-20210810205551-6071d7bae2cd // indirect
	github.com/streamingfast/dstore v0.1.1-0.20210810110932-928f221474e4
	github.com/streamingfast/firehose v0.1.1-0.20210810201729-f4f65f7bc597
	github.com/streamingfast/merger v0.0.3-0.20210810201721-8308c7731ce1
	github.com/streamingfast/node-manager v0.0.2-0.20210810201828-5033a297edfa
	github.com/streamingfast/relayer v0.0.2-0.20210810201213-52e46787d413
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/ShinyTrinkets/overseer => github.com/dfuse-io/overseer v0.2.1-0.20210326144022-ee491780e3ef

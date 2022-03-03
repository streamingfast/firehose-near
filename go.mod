module github.com/streamingfast/sf-near

go 1.16

require (
	github.com/RoaringBitmap/roaring v0.9.4
	github.com/ShinyTrinkets/overseer v0.3.0
	github.com/abourget/llerrgroup v0.2.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mr-tron/base58 v1.2.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/streamingfast/bstream v0.0.2-0.20220303215418-35f97f725c45
	github.com/streamingfast/cli v0.0.3-0.20210811201236-5c00ec55462d
	github.com/streamingfast/dauth v0.0.0-20210812020920-1c83ba29add1
	github.com/streamingfast/dbin v0.0.0-20210809205249-73d5eca35dc5
	github.com/streamingfast/derr v0.0.0-20210811180100-9138d738bcec
	github.com/streamingfast/dgrpc v0.0.0-20220301153539-536adf71b594
	github.com/streamingfast/dlauncher v0.0.0-20211210162313-cf4aa5fc4878
	github.com/streamingfast/dmetering v0.0.0-20220301165106-a642bb6a21bd
	github.com/streamingfast/dmetrics v0.0.0-20210811180524-8494aeb34447
	github.com/streamingfast/dstore v0.1.1-0.20220203133825-30eb2f9c5cd3
	github.com/streamingfast/firehose v0.1.1-0.20220303153047-161af394ee22
	github.com/streamingfast/logging v0.0.0-20220222131651-12c3943aac2e
	github.com/streamingfast/merger v0.0.3-0.20220301162603-c0129b6f1ad4
	github.com/streamingfast/near-go v0.0.0-20220302163233-b638f5b48a2d
	github.com/streamingfast/node-manager v0.0.2-0.20211029201743-0b82ab7f9de4
	github.com/streamingfast/pbgo v0.0.6-0.20220228185940-1bbaafec7d8a
	github.com/streamingfast/relayer v0.0.2-0.20220301162545-2db510359d2a
	github.com/streamingfast/sf-tools v0.0.0-20211222184149-d5ac7ff965f7
	github.com/streamingfast/snapshotter v0.0.0-20210906180247-1ec27a37764f
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.9.3
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/ShinyTrinkets/overseer => github.com/dfuse-io/overseer v0.2.1-0.20210326144022-ee491780e3ef

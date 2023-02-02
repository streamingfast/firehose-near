package cli

const (
	MetricsListenAddr         string = ":9102"
	ReaderGRPCAddr            string = ":15010"
	NodeManagerAPIAddr        string = ":15041"
	ReaderNodeManagerAPIAddr  string = ":15009"
	ArchiveNodeManagerAPIAddr string = ":15014"
	MergerServingAddr         string = ":15012"
	RelayerServingAddr        string = ":15011"
	FirehoseGRPCServingAddr   string = ":15042"

	BlocksCacheDirectory string = "{sf-data-dir}/blocks-cache"
	FirstStreamableBlock int    = 3

	MergedBlocksStoreURL string = "file://{sf-data-dir}/storage/merged-blocks"
	OneBlockStoreURL     string = "file://{sf-data-dir}/storage/one-blocks"
	ForkedBlocksStoreURL string = "file://{sf-data-dir}/storage/forked-blocks"
	IndexStoreURL        string = "file://{sf-data-dir}/storage/index"

	NodeRPCPort string = "3030"
	NodeRPCAddr string = "http://localhost:" + NodeRPCPort

	CommonAutoMaxProcsFlag              string = "common-auto-max-procs"
	CommonAutoMemLimitFlag              string = "common-auto-mem-limit-percent"
	CommonSystemShutdownSignalDelayFlag string = "common-system-shutdown-signal-delay"
)

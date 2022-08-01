package tools

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/sf-near/codec"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
	sftools "github.com/streamingfast/sf-tools"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func init() {
	Cmd.AddCommand(DownloadFromFirehoseCmd)
	DownloadFromFirehoseCmd.Flags().StringP("api-token-env-var", "a", "FIREHOSE_API_TOKEN", "Look for a JWT in this environment variable to authenticate against endpoint")
	DownloadFromFirehoseCmd.Flags().BoolP("plaintext", "p", false, "Use plaintext connection to firehose")
	DownloadFromFirehoseCmd.Flags().BoolP("insecure", "k", false, "Skip SSL certificate validation when connecting to firehose")
}

var DownloadFromFirehoseCmd = &cobra.Command{
	Use:     "download-from-firehose",
	Short:   "download blocks from firehose and save them to merged-blocks",
	Args:    cobra.ExactArgs(4),
	RunE:    downloadFromFirehoseE,
	Example: "sfnear tools download-from-firehose api.streamingfast.io 1000 2000 ./outputdir",
}

func downloadFromFirehoseE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	endpoint := args[0]
	start, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing start block num: %w", err)
	}
	stop, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing stop block num: %w", err)
	}
	destFolder := args[3]

	apiTokenEnvVar := mustGetString(cmd, "api-token-env-var")
	apiToken := os.Getenv(apiTokenEnvVar)

	var fixerFunc func(*bstream.Block) (*bstream.Block, error)
	plaintext := mustGetBool(cmd, "plaintext")
	insecure := mustGetBool(cmd, "insecure")

	return sftools.DownloadFirehoseBlocks(
		ctx,
		endpoint,
		apiToken,
		insecure,
		plaintext,
		start,
		stop,
		destFolder,
		decodeAnyPB,
		fixerFunc,
		zlog,
	)
}

func decodeAnyPB(in *anypb.Any) (*bstream.Block, error) {
	block := &pbcodec.Block{}
	if err := anypb.UnmarshalTo(in, block, proto.UnmarshalOptions{}); err != nil {
		return nil, fmt.Errorf("unmarshal anypb: %w", err)
	}

	return codec.BlockFromProto(block)
}

package types

import (
	"github.com/streamingfast/bstream"
	pbnear "github.com/streamingfast/firehose-near/types/pb/sf/near/type/v1"
)

func BlockFromProto(b *pbnear.Block) (*bstream.Block, error) {
	return b.ToBstreamBlock()
}

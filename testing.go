package firenear

import (
	"github.com/streamingfast/bstream"
	pbnear "github.com/streamingfast/firehose-near/pb/sf/near/type/v1"
	"google.golang.org/protobuf/proto"
)

func TestingInitBstream() {
	// Should be aligned with firecore.Chain as defined in `cmd/firenear/main.go``
	bstream.InitGeneric("NEA", 1, func() proto.Message {
		return new(pbnear.Block)
	})
}

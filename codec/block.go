package codec

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/dfuse/bstream/v1"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
)

func BlockFromProto(b *pbcodec.BlockWrapper) (*bstream.Block, error) {
	content, err := proto.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to binary form: %s", err)
	}

	return &bstream.Block{
		Id:             b.ID(),
		Number:         b.Number(),
		PreviousId:     b.PreviousID(),
		Timestamp:      b.Time(),
		LibNum:         b.LIBNum(),
		PayloadKind:    pbbstream.Protocol_NEAR,
		PayloadVersion: 1,
		PayloadBuffer:  content,
	}, nil
}

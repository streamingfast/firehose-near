package codec

import (
	"fmt"

	"github.com/dfuse-io/bstream"
	pbbstream "github.com/dfuse-io/pbgo/dfuse/bstream/v1"
	"github.com/golang/protobuf/proto"
	pbcodec "github.com/streamingfast/near-sf/pb/sf/near/codec/v1"
)

func BlockFromProto(b *pbcodec.Block) (*bstream.Block, error) {
	//blockTime, err := b.Time()
	//if err != nil {
	//	return nil, err
	//}

	content, err := proto.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to binary form: %s", err)
	}

	return &bstream.Block{
		//Id:             b.ID(),
		Number: b.Number,
		//PreviousId:     b.PreviousID(),
		//Timestamp:      blockTime,
		//LibNum:         b.LIBNum(),
		PayloadKind:    pbbstream.Protocol_ETH,
		PayloadVersion: b.Ver,
		PayloadBuffer:  content,
	}, nil
}

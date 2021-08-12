package codec

import (
	"fmt"

	"github.com/dfuse-io/bstream"
	"github.com/golang/protobuf/proto"
	pbbstream "github.com/streamingfast/pbgo/dfuse/bstream/v1"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
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

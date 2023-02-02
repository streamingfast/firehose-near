package pbnear

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mr-tron/base58"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	"google.golang.org/protobuf/proto"
)

func (x *Block) ID() string {
	return x.Header.Hash.AsString()
}

func (x *Block) Num() uint64 {
	return x.Header.Height
}

func (x *Block) LIBNum() uint64 {
	return x.Header.LastFinalBlockHeight
}

func (x *Block) PreviousID() string {
	return x.Header.PrevHash.AsString()
}

func (x *Block) Time() time.Time {
	return time.Unix(0, int64(x.Header.TimestampNanosec)).UTC()
}

func (b *Block) ToBstreamBlock() (*bstream.Block, error) {
	content, err := proto.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to binary form: %s", err)
	}

	block := &bstream.Block{
		Id:             b.ID(),
		Number:         b.Num(),
		PreviousId:     b.PreviousID(),
		Timestamp:      b.Time(),
		LibNum:         b.LIBNum(),
		PayloadKind:    pbbstream.Protocol_NEAR,
		PayloadVersion: 1,
	}

	return bstream.GetBlockPayloadSetter(block, content)
}

func (x *CryptoHash) AsString() string {
	return hex.EncodeToString(x.Bytes)
}

func (x *CryptoHash) AsBase58String() string {
	return base58.Encode(x.Bytes)
}

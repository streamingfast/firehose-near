package pbcodec

import (
	"encoding/hex"
	"time"

	"github.com/streamingfast/bstream"
)

func (x *Block) ID() string {
	return x.Header.Hash.AsString()
}

func (x *Block) Number() uint64 {
	return x.Header.Height
}

func (x *Block) LIBNum() uint64 {
	if x.Number() == bstream.GetProtocolFirstStreamableBlock {
		return bstream.GetProtocolGenesisBlock
	}

	if x.Number() <= 25+bstream.GetProtocolFirstStreamableBlock {
		return bstream.GetProtocolFirstStreamableBlock
	}

	return x.Number() - 25
}

func (x *Block) PreviousID() string {
	return x.Header.PrevHash.AsString()
}

func (x *Block) Time() time.Time {
	return time.Unix(0, int64(x.Header.TimestampNanosec)).UTC()
}

func (x *CryptoHash) AsString() string {
	return hex.EncodeToString(x.Bytes)
}

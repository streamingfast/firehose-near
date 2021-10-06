package pbcodec

import (
	"encoding/hex"
	"time"

	"github.com/streamingfast/bstream"
)

func (x *BlockWrapper) ID() string {
	return x.Block.Header.Hash.AsString()
}

func (x *BlockWrapper) Number() uint64 {
	return x.Block.Header.Height
}

func (x *BlockWrapper) LIBNum() uint64 {
	if x.Number() == bstream.GetProtocolFirstStreamableBlock {
		return bstream.GetProtocolGenesisBlock
	}

	if x.Number() <= 25+bstream.GetProtocolFirstStreamableBlock {
		return bstream.GetProtocolFirstStreamableBlock
	}

	return x.Number() - 25
}

func (x *BlockWrapper) PreviousID() string {
	return x.Block.Header.PrevHash.AsString()
}

func (x *BlockWrapper) Time() time.Time {
	return time.Unix(0, int64(x.Block.Header.TimestampNanosec)).UTC()
}

func (x *CryptoHash) AsString() string {
	return hex.EncodeToString(x.Bytes)
}

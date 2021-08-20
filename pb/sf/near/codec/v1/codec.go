package pbcodec

import (
	"encoding/hex"
	"time"
)

func (x *BlockWrapper) ID() string {
	return x.Block.Header.Hash.AsString()
}

func (x *BlockWrapper) Number() uint64 {
	return x.Block.Header.Height
}

func (x *BlockWrapper) LIBNum() uint64 {
	// FIXME: What is the correct way to get was is the last irreversible num of a given block
	return x.Number() - 25
}

func (x *BlockWrapper) PreviousID() string {
	return x.Block.Header.PrevHash.AsString()
}

func (x *BlockWrapper) Time() time.Time {
	return time.Unix(0, int64(x.Block.Header.TimestampNanosec))
}

func (x *CryptoHash) AsString() string {
	return hex.EncodeToString(x.Bytes)
}

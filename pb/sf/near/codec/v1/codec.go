package pbcodec

import (
	"encoding/hex"
	"time"

	"github.com/mr-tron/base58"
)

func (x *BlockWrapper) ID() string {
	return x.Block.Header.Hash.AsString()
}

func (x *BlockWrapper) Number() uint64 {
	return x.Block.Header.Height
}

func (x *BlockWrapper) LIBNum() uint64 {
	return x.Block.Header.LastFinalBlockHeight
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

func (x *CryptoHash) AsBase58String() string {
	return base58.Encode(x.Bytes)
}

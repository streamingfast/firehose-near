package transform

import (
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
)

type blockIndexer interface {
	Add(keys []string, blockNum uint64)
}

type NearBlockIndexer struct {
	BlockIndexer blockIndexer
}

func NewNearBlockIndexer(indexStore dstore.Store, indexSize uint64, startBlock uint64) *NearBlockIndexer {
	bi := transform.NewBlockIndexer(indexStore, indexSize, ReceiptAddressIndexShortName, transform.WithDefinedStartBlock(startBlock))
	return &NearBlockIndexer{
		BlockIndexer: bi,
	}
}

func (i *NearBlockIndexer) ProcessBlock(blk *pbcodec.Block) {
	keyMap := make(map[string]bool)
	for _, shard := range blk.Shards {
		for _, outcome := range shard.ReceiptExecutionOutcomes {
			if outcome.Receipt.GetAction() != nil {
				keyMap[outcome.Receipt.ReceiverId] = true
			}
		}
	}
	var keys []string
	for key := range keyMap {
		keys = append(keys, key)
	}

	i.BlockIndexer.Add(keys, blk.Num())
	return
}

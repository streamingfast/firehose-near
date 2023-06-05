package transform

import (
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	firecore "github.com/streamingfast/firehose-core"
	pbnear "github.com/streamingfast/firehose-near/pb/sf/near/type/v1"
)

var _ firecore.BlockIndexer[*pbnear.Block] = (*NearBlockIndexer)(nil)

type blockIndexer interface {
	Add(keys []string, blockNum uint64)
}

type NearBlockIndexer struct {
	BlockIndexer blockIndexer
}

func NewNearBlockIndexer(indexStore dstore.Store, indexSize uint64) (firecore.BlockIndexer[*pbnear.Block], error) {
	bi := transform.NewBlockIndexer(indexStore, indexSize, ReceiptAddressIndexShortName)

	return &NearBlockIndexer{
		BlockIndexer: bi,
	}, nil
}

func (i *NearBlockIndexer) ProcessBlock(blk *pbnear.Block) error {
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
	return nil
}

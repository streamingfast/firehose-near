package codec

import (
	"fmt"
	"strings"

	"github.com/streamingfast/bstream"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
)

type FilteringPreprocessor struct {
	Filter *BlockFilter
}

func (f *FilteringPreprocessor) PreprocessBlock(blk *bstream.Block) (interface{}, error) {
	return nil, f.Filter.TransformInPlace(blk)
}

type BlockFilter struct {
	IncludeReceivers map[string]bool
	ExcludeReceivers map[string]bool
}

//receiver-ids:bozo.near|matt.near
func NewBlockFilter(includeExpression, excludeExpression string) (*BlockFilter, error) {
	filterType, includes, err := splitExpression(includeExpression)
	if err != nil {
		return nil, fmt.Errorf("parsing includes: %w", err)
	}

	if filterType != "receiver-ids" {
		return nil, fmt.Errorf("invalid include filter type, supported types are: receiver-ids")
	}

	filterType, excludes, err := splitExpression(excludeExpression)
	if err != nil {
		return nil, fmt.Errorf("parsing excludes: %w", err)
	}

	if filterType != "receiver-ids" {
		return nil, fmt.Errorf("invalid exclude filter type, supported types are: receiver-ids")
	}

	return &BlockFilter{
		IncludeReceivers: includes,
		ExcludeReceivers: excludes,
	}, nil
}

func splitExpression(expression string) (filterType string, values map[string]bool, err error) {
	values = make(map[string]bool)

	parts := strings.Split(expression, ":")
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("bad expression format")
	}

	filterType = parts[0]
	splitValues := strings.Split(parts[1], "|")
	for _, v := range splitValues {
		values[v] = true
	}
	return
}

func (f *BlockFilter) TransformInPlace(blk *bstream.Block) error {
	block := blk.ToNative().(*pbcodec.Block)

	var filteredShards []*pbcodec.IndexerShard
	for _, shard := range block.Shards {
		//Filter execution transaction
		if shard.Chunk == nil {
			continue
		}

		var filteredTransaction []*pbcodec.IndexerTransactionWithOutcome
		if shard.Chunk.Transactions != nil {
			for _, transaction := range shard.Chunk.Transactions {
				if _, found := f.ExcludeReceivers[transaction.Transaction.ReceiverId]; found {
					continue
				}
				if _, found := f.IncludeReceivers[transaction.Transaction.ReceiverId]; found {
					filteredTransaction = append(filteredTransaction, transaction)
				}
			}
		}
		shard.Chunk.Transactions = filteredTransaction

		//Filter execution receipt
		var filteredOutcomes []*pbcodec.IndexerExecutionOutcomeWithReceipt
		if shard.ReceiptExecutionOutcomes != nil {
			for _, executionOutcome := range shard.ReceiptExecutionOutcomes {
				if _, found := f.ExcludeReceivers[executionOutcome.Receipt.ReceiverId]; found {
					continue
				}
				if _, found := f.IncludeReceivers[executionOutcome.Receipt.ReceiverId]; found {
					filteredOutcomes = append(filteredOutcomes, executionOutcome)
				}
			}
		}
		shard.ReceiptExecutionOutcomes = filteredOutcomes

		if len(filteredOutcomes) > 0 || len(filteredTransaction) > 0 {
			filteredShards = append(filteredShards, shard)
		}
	}
	block.Shards = filteredShards

	return nil
}

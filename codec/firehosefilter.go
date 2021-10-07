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
	filterTYpe, includes, err := splitExpression(includeExpression)
	if err != nil {
		return nil, fmt.Errorf("parsing includes: %w", err)
	}

	if filterTYpe != "receiver-ids" {
		return nil, fmt.Errorf("invalid include filter type, supported types are: receiver-ids")
	}

	filterTYpe, excludes, err := splitExpression(excludeExpression)
	if err != nil {
		return nil, fmt.Errorf("parsing excludes: %w", err)
	}

	if filterTYpe != "receiver-ids" {
		return nil, fmt.Errorf("invalid exclude filter type, supported types are: receiver-ids")
	}

	return &BlockFilter{
		IncludeReceivers: includes,
		ExcludeReceivers: excludes,
	}, nil
}

func splitExpression(expression string) (filterType string, values map[string]bool, err error) {
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
		var filteredTransaction []*pbcodec.IndexerTransactionWithOutcome
		for _, transaction := range shard.Chunk.Transactions {
			if _, found := f.ExcludeReceivers[transaction.Transaction.ReceiverId]; found {
				continue
			}
			if _, found := f.ExcludeReceivers[transaction.Transaction.ReceiverId]; found {
				filteredTransaction = append(filteredTransaction, transaction)
			}
		}

		//Filter execution receipt
		var filteredOutcomes []*pbcodec.IndexerExecutionOutcomeWithReceipt
		for _, executionOutcome := range shard.ReceiptExecutionOutcomes {
			if _, found := f.ExcludeReceivers[executionOutcome.Receipt.ReceiverId]; found {
				continue
			}
			if _, found := f.ExcludeReceivers[executionOutcome.Receipt.ReceiverId]; found {
				filteredOutcomes = append(filteredOutcomes, executionOutcome)
			}
		}

		if len(filteredOutcomes) > 0 || len(filteredTransaction) > 0 {
			shard.ReceiptExecutionOutcomes = filteredOutcomes
			filteredShards = append(filteredShards, shard)
		}
	}
	block.Shards = filteredShards

	return nil
}

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

func (f *BlockFilter) included(receiverID string) bool {
	if len(f.IncludeReceivers) == 0 {
		return true
	}
	_, found := f.IncludeReceivers[receiverID]
	return found
}
func (f *BlockFilter) excluded(receiverID string) bool {
	if len(f.ExcludeReceivers) == 0 {
		return false
	}
	_, found := f.ExcludeReceivers[receiverID]
	return found
}

func (f *BlockFilter) TransformInPlace(blk *bstream.Block) error {
	block := blk.ToNative().(*pbcodec.Block)

	if len(f.IncludeReceivers) == 0 && len(f.ExcludeReceivers) == 0 {
		return nil
	}

	var filteredShards []*pbcodec.IndexerShard
	for _, shard := range block.Shards {

		//Filter execution transaction
		var filteredTransaction []*pbcodec.IndexerTransactionWithOutcome
		for _, transaction := range shard.Chunk.Transactions {
			if f.excluded(transaction.Transaction.ReceiverId) {
				continue
			}
			if f.included(transaction.Transaction.ReceiverId) {
				filteredTransaction = append(filteredTransaction, transaction)
			}
		}

		//Filter execution receipt
		var filteredOutcomes []*pbcodec.IndexerExecutionOutcomeWithReceipt
		for _, executionOutcome := range shard.ReceiptExecutionOutcomes {
			if f.excluded(executionOutcome.Receipt.ReceiverId) {
				continue
			}
			if f.included(executionOutcome.Receipt.ReceiverId) {
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

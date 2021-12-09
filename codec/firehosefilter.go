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
	blockfilter := &BlockFilter{
		IncludeReceivers: make(map[string]bool),
		ExcludeReceivers: make(map[string]bool),
	}

	if includeExpression != "" {
		filterType, includes, err := splitExpression(includeExpression)
		if err != nil {
			return nil, fmt.Errorf("parsing includes: %w", err)
		}

		if filterType != "receiver-ids" {
			return nil, fmt.Errorf("invalid include filter type, supported types are: receiver-ids")
		}
		blockfilter.IncludeReceivers = includes
	}

	if excludeExpression != "" {
		filterType, excludes, err := splitExpression(excludeExpression)
		if err != nil {
			return nil, fmt.Errorf("parsing excludes: %w", err)
		}

		if filterType != "receiver-ids" {
			return nil, fmt.Errorf("invalid exclude filter type, supported types are: receiver-ids")
		}
		blockfilter.ExcludeReceivers = excludes
	}

	return blockfilter, nil
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
		if shard.Chunk == nil {
			continue
		}

		var filteredTransaction []*pbcodec.IndexerTransactionWithOutcome
		if shard.Chunk.Transactions != nil {
			for _, transaction := range shard.Chunk.Transactions {
				if f.excluded(transaction.Transaction.ReceiverId) {
					continue
				}
				if f.included(transaction.Transaction.ReceiverId) {
					filteredTransaction = append(filteredTransaction, transaction)
				}
			}
		}
		shard.Chunk.Transactions = filteredTransaction

		//Filter execution receipt
		var filteredOutcomes []*pbcodec.IndexerExecutionOutcomeWithReceipt
		if shard.ReceiptExecutionOutcomes != nil {
			for _, executionOutcome := range shard.ReceiptExecutionOutcomes {
				if f.excluded(executionOutcome.Receipt.ReceiverId) {
					continue
				}
				if f.included(executionOutcome.Receipt.ReceiverId) {
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

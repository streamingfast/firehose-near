package transform

import (
	"fmt"

	"github.com/streamingfast/dstore"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
	pbtransform "github.com/streamingfast/sf-near/pb/sf/near/transform/v1"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/transform"
	"google.golang.org/protobuf/proto"
)

var ReceiptFilterMessageName = proto.MessageName(&pbtransform.BasicReceiptFilter{})

func BasicReceiptFilterFactory(indexStore dstore.Store, possibleIndexSizes []uint64) *transform.Factory {
	return &transform.Factory{
		Obj: &pbtransform.BasicReceiptFilter{},
		NewFunc: func(message *anypb.Any) (transform.Transform, error) {
			mname := message.MessageName()
			if mname != ReceiptFilterMessageName {
				return nil, fmt.Errorf("expected type url %q, received %q", ReceiptFilterMessageName, message.TypeUrl)
			}

			filter := &pbtransform.BasicReceiptFilter{}
			err := proto.Unmarshal(message.Value, filter)
			if err != nil {
				return nil, fmt.Errorf("unexpected unmarshall error: %w", err)
			}

			if len(filter.Accounts) == 0 {
				return nil, fmt.Errorf("a basic account filter requires at least one account")
			}

			accountMap := make(map[string]bool)
			for _, acc := range filter.Accounts {
				accountMap[acc] = true
			}
			f := &BasicReceiptFilter{
				Accounts:           accountMap,
				possibleIndexSizes: possibleIndexSizes,
				indexStore:         indexStore,
			}
			return f, nil
		},
	}
}

type BasicReceiptFilter struct {
	Accounts map[string]bool

	indexStore         dstore.Store
	possibleIndexSizes []uint64
}

func (p *BasicReceiptFilter) String() string {
	return fmt.Sprintf("%v", p.Accounts)
}

func (p *BasicReceiptFilter) Transform(readOnlyBlk *bstream.Block, in transform.Input) (transform.Output, error) {
	nearBlock := readOnlyBlk.ToProtocol().(*pbcodec.Block)
	var outShards []*pbcodec.IndexerShard
	for _, shard := range nearBlock.Shards {
		var outcomes []*pbcodec.IndexerExecutionOutcomeWithReceipt
		for _, outcome := range shard.ReceiptExecutionOutcomes {
			if outcome.Receipt.GetAction() != nil && p.Accounts[outcome.Receipt.ReceiverId] {
				outcomes = append(outcomes, outcome)
			}
		}
		if len(outcomes) != 0 {
			shard.ReceiptExecutionOutcomes = outcomes
			outShards = append(outShards, shard)
		}
	}
	nearBlock.Shards = outShards
	return nearBlock, nil
}

func (p *BasicReceiptFilter) GetIndexProvider() bstream.BlockIndexProvider {
	if p.indexStore == nil {
		return nil
	}

	if len(p.Accounts) == 0 {
		return nil
	}

	return NewNearBlockIndexProvider(
		p.indexStore,
		p.possibleIndexSizes,
		p.Accounts,
	)
}

package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/cli/sflags"
	pbtransform "github.com/streamingfast/firehose-near/pb/sf/near/transform/v1"
	pbnear "github.com/streamingfast/firehose-near/pb/sf/near/type/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

func printBlock(blk *bstream.Block, alsoPrintTransactions bool, out io.Writer) error {
	block := blk.ToProtocol().(*pbnear.Block)

	transactionCount := 0
	for _, shard := range block.Shards {
		if shard.Chunk != nil {
			transactionCount += len(shard.Chunk.Transactions)
		}
	}

	if _, err := fmt.Fprintf(out, "Block #%d (%s) (prev: %s): %d transactions\n",
		block.Num(),
		block.ID(),
		block.PreviousID()[0:7],
		transactionCount,
	); err != nil {
		return err
	}

	if alsoPrintTransactions {
		for _, shard := range block.Shards {
			if shard.Chunk != nil {
				if _, err := fmt.Fprintf(out, "- Shard %d\n", shard.ShardId); err != nil {
					return err
				}

				for _, trx := range shard.Chunk.Transactions {
					if _, err := fmt.Fprintf(out, "  - Transaction %s\n", trx.Transaction.Hash.AsBase58String()); err != nil {
						return err
					}
				}

				for _, receipt := range shard.Chunk.Receipts {
					if _, err := fmt.Fprintf(out, "  - Receipt %s\n", receipt.ReceiptId.AsBase58String()); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func parseReceiptAccountFilters(cmd *cobra.Command, logger *zap.Logger) ([]*anypb.Any, error) {
	in := sflags.MustGetString(cmd, "receipt-account-filters")
	if in == "" {
		return nil, nil
	}

	var pairs []*pbtransform.PrefixSuffixPair
	var accounts []string

	for _, unit := range strings.Split(in, ",") {
		if parts := strings.Split(unit, ":"); len(parts) == 2 {
			pairs = append(pairs, &pbtransform.PrefixSuffixPair{
				Prefix: parts[0],
				Suffix: parts[1],
			})
			continue
		}
		accounts = append(accounts, unit)
	}

	filters := &pbtransform.BasicReceiptFilter{
		Accounts:             accounts,
		PrefixAndSuffixPairs: pairs,
	}

	any, err := anypb.New(filters)
	if err != nil {
		return nil, err
	}
	return []*anypb.Any{any}, nil
}

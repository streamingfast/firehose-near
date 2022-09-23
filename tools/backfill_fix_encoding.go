package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/streamingfast/dbin"
	"github.com/streamingfast/dstore"
	pbnear "github.com/streamingfast/firehose-near/types/pb/sf/near/type/v1"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	sftools "github.com/streamingfast/sf-tools"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var backfillFixEncodingCmd = &cobra.Command{
	Use:   "fix-encoding {input-store-url} {output-store-url}",
	Short: "update merge block files to fix encoding issues caused by 1.29.0 migration",
	Args:  cobra.ExactArgs(2),
	RunE:  backfillFixEncodingE,
}

func init() {
	BackfillCmd.AddCommand(backfillFixEncodingCmd)
}

func backfillFixEncodingE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	blockRange, err := sftools.Flags.GetBlockRange("range")
	if err != nil {
		return fmt.Errorf("error parsing block range: %w", err)
	}

	fileBlockSize := 100
	walkPrefix := sftools.WalkBlockPrefix(blockRange, uint32(fileBlockSize))

	inputBlocksStore, err := dstore.NewDBinStore(args[0])
	if err != nil {
		return fmt.Errorf("error opening input store %s: %w", args[0], err)
	}

	var baseNum32 uint32

	err = inputBlocksStore.Walk(ctx, walkPrefix, func(filename string) (err error) {
		match := numberRegex.FindStringSubmatch(filename)
		if match == nil {
			zlog.Debug("file does not match pattern", zap.String("filename", filename))
			return nil
		}

		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		if baseNum+uint64(fileBlockSize)-1 < blockRange.Start {
			zlog.Debug("file is before block range start", zap.String("filename", filename), zap.Uint64("block_range_start", blockRange.Start))
			return nil
		}
		baseNum32 = uint32(baseNum)

		try := 0
		var obj io.ReadCloser
		for {
			try += 1
			obj, err = inputBlocksStore.OpenObject(ctx, filename)
			if err != nil {
				if try > 10 {
					return fmt.Errorf("error reading file %s from input store %s: %w", filename, args[0], err)
				}
				time.Sleep(time.Duration(try) * time.Second)
				continue
			}
			break
		}

		binReader := dbin.NewReader(obj)
		contentType, version, err := binReader.ReadHeader()
		if err != nil {
			return fmt.Errorf("error reading file header for file %s: %w", filename, err)
		}
		defer binReader.Close()

		buffer := bytes.NewBuffer(nil)
		binWriter := dbin.NewWriter(buffer)
		defer binWriter.Close()

		err = binWriter.WriteHeader(contentType, int(version))
		if err != nil {
			return fmt.Errorf("error writing block header: %w", err)
		}

		for {
			line, err := binReader.ReadMessage()
			if err == io.EOF {
				zlog.Debug("eof", zap.String("filename", filename))
				break
			}

			if err != nil {
				return fmt.Errorf("reading block: %w", err)
			}

			if len(line) == 0 {
				zlog.Debug("empty line", zap.String("filename", filename))
				break
			}

			// decode block data
			bstreamBlock := new(pbbstream.Block)
			err = proto.Unmarshal(line, bstreamBlock)
			if err != nil {
				return fmt.Errorf("unmarshaling block proto: %w", err)
			}

			// check range
			if bstreamBlock.Number < blockRange.Start || bstreamBlock.Number > blockRange.Stop {
				zlog.Debug("block is outside block range, skipping post-processing")
				if err := binWriter.WriteMessage(line); err != nil {
					return fmt.Errorf("error writing block: %w", err)
				}

				continue
			}

			blockBytes := bstreamBlock.GetPayloadBuffer()

			block := new(pbnear.Block)
			err = proto.Unmarshal(blockBytes, block)
			if err != nil {
				return fmt.Errorf("unmarshaling block proto: %w", err)
			}

			for _, shard := range block.Shards {
				if shard.Chunk == nil {
					continue
				}

				for _, trx := range shard.Chunk.Transactions {
					if v := trx.Outcome.ExecutionOutcome.Outcome.GetSuccessValue(); v != nil {
						data, err := base64.StdEncoding.DecodeString(string(v.Value))
						if err != nil {
							return fmt.Errorf("unable to base64 executionOutcome.SuccessValue.value for receipt %q decode: %w", hex.EncodeToString(trx.Transaction.Hash.Bytes), err)
						}

						v.Value = data
					}

					for _, action := range trx.Transaction.Actions {

						if v := action.GetFunctionCall(); v != nil {
							data, err := base64.StdEncoding.DecodeString(string(v.Args))
							if err != nil {
								return fmt.Errorf("unable to base64 functionCall.args for receipt %q in block %d decode: %w", hex.EncodeToString(trx.Transaction.Hash.Bytes), block.Num(), err)
							}

							v.Args = data
						} else if v := action.GetDeployContract(); v != nil {
							data, err := base64.StdEncoding.DecodeString(string(v.Code))
							if err != nil {
								return fmt.Errorf("unable to base64 deployContractCall.code for receipt %q decode: %w", hex.EncodeToString(trx.Transaction.Hash.Bytes), err)
							}

							v.Code = data
						}
					}
				}
			}

			// encode block data
			backFilledBlockBytes, err := proto.Marshal(block)
			if err != nil {
				return fmt.Errorf("marshaling block proto: %w", err)
			}

			bstreamBlock.PayloadBuffer = backFilledBlockBytes

			bstreamBlockBytes, err := proto.Marshal(bstreamBlock)
			if err != nil {
				return fmt.Errorf("marshaling bstream block: %w", err)
			}

			err = binWriter.WriteMessage(bstreamBlockBytes)
			if err != nil {
				return fmt.Errorf("error writing block: %w", err)
			}
		}

		err = obj.Close()
		if err != nil {
			return fmt.Errorf("error closing object %s: %w", filename, err)
		}

		outputBlocksStore, err := dstore.NewDBinStore(args[1])
		if err != nil {
			return fmt.Errorf("error opening output store %s: %w", args[0], err)
		}

		outputBlocksStore.SetOverwrite(true)

		try = 0
		for {
			try += 1
			err = outputBlocksStore.WriteObject(ctx, filename, buffer)
			if err != nil {
				if try > 10 {
					return fmt.Errorf("error writing file %s to output store %s: %w", filename, args[1], err)
				}
				time.Sleep(time.Duration(try) * time.Second)
				continue
			}
			zlog.Debug("saved output file", zap.String("store", outputBlocksStore.BaseURL().String()), zap.String("filename", filename))
			break
		}

		// check range upper bound
		if !blockRange.Unbounded() {
			roundedEndBlock := sftools.RoundToBundleEndBlock(baseNum32, uint32(fileBlockSize))
			if roundedEndBlock >= uint32(blockRange.Stop-1) {
				return errStopWalk
			}
		}

		zlog.Info("updated blocks", zap.String("file", filename))
		return nil
	})

	if err != nil && err != errStopWalk {
		return err
	}

	return nil
}

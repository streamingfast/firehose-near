package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/streamingfast/dbin"
	"github.com/streamingfast/dstore"
	firecore "github.com/streamingfast/firehose-core"
	pbnear "github.com/streamingfast/firehose-near/pb/sf/near/type/v1"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	sftools "github.com/streamingfast/sf-tools"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func newToolsBackfillPrevHeightCmd(logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "prev-height {input-store-url} {output-store-url}",
		Short: "update merge block files to set previous height data",
		Args:  cobra.ExactArgs(2),
		RunE:  createBackfillPrevHeightE(logger),
	}
}

func newToolsbackfillPrevHeightCheckCmd(logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "prev-height-check {input-store-url}",
		Short: "Check merge block files to validate previous height data",
		Args:  cobra.ExactArgs(1),
		RunE:  createBackfillPrevHeightCheckE(logger),
	}
}

func createBackfillPrevHeightE(logger *zap.Logger) firecore.CommandExecutor {
	return func(cmd *cobra.Command, args []string) error {
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
		heightMap := make(map[string]uint64)

		err = inputBlocksStore.Walk(ctx, walkPrefix, func(filename string) (err error) {
			match := numberRegex.FindStringSubmatch(filename)
			if match == nil {
				logger.Debug("file does not match pattern", zap.String("filename", filename))
				return nil
			}

			baseNum, _ := strconv.ParseUint(match[1], 10, 32)
			if baseNum+uint64(fileBlockSize)-1 < blockRange.Start {
				logger.Debug("file is before block range start", zap.String("filename", filename), zap.Uint64("block_range_start", blockRange.Start))
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

			firstSeenBlock := true
			for {
				line, err := binReader.ReadMessage()
				if err == io.EOF {
					logger.Debug("eof", zap.String("filename", filename))
					break
				}

				if err != nil {
					return fmt.Errorf("reading block: %w", err)
				}

				if len(line) == 0 {
					logger.Debug("empty line", zap.String("filename", filename))
					break
				}

				// decode block data
				bstreamBlock := new(pbbstream.Block)
				err = proto.Unmarshal(line, bstreamBlock)
				if err != nil {
					return fmt.Errorf("unmarshaling block proto: %w", err)
				}

				blockBytes := bstreamBlock.GetPayloadBuffer()

				block := new(pbnear.Block)
				err = proto.Unmarshal(blockBytes, block)
				if err != nil {
					return fmt.Errorf("unmarshaling block proto: %w", err)
				}

				// save current id/height
				heightMap[block.ID()] = block.Num()

				prevHeight, ok := heightMap[block.PreviousID()]
				if !ok {
					if firstSeenBlock {
						firstSeenBlock = false
						logger.Debug("skipping first block update. no prev_height data yet", zap.Uint64("block", block.Num()))
					} else {
						return fmt.Errorf("could not find previous height for block id %s", block.ID())
					}
				} else {
					// update current block prev_height
					block.Header.PrevHeight = prevHeight
					logger.Debug("updated prev_height", zap.Uint64("block", block.Num()), zap.Uint64("prev_height", prevHeight))
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
				logger.Debug("saved output file", zap.String("store", outputBlocksStore.BaseURL().String()), zap.String("filename", filename))
				break
			}

			// check range upper bound
			if !blockRange.Unbounded() {
				roundedEndBlock := sftools.RoundToBundleEndBlock(baseNum32, uint32(fileBlockSize))
				if roundedEndBlock >= uint32(blockRange.Stop-1) {
					return errStopWalk
				}
			}

			logger.Info("updated blocks", zap.String("file", filename))
			return nil
		})

		if err != nil && err != errStopWalk {
			return err
		}

		return nil
	}
}

func createBackfillPrevHeightCheckE(logger *zap.Logger) firecore.CommandExecutor {
	return func(cmd *cobra.Command, args []string) error {
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
		firstBlockSeen := true

		err = inputBlocksStore.Walk(ctx, walkPrefix, func(filename string) (err error) {
			match := numberRegex.FindStringSubmatch(filename)
			if match == nil {
				return nil
			}

			baseNum, _ := strconv.ParseUint(match[1], 10, 32)
			if baseNum+uint64(fileBlockSize)-1 < blockRange.Start {
				return nil
			}
			baseNum32 = uint32(baseNum)

			logger.Debug("checking file", zap.String("filename", filename))

			obj, err := inputBlocksStore.OpenObject(ctx, filename)
			if err != nil {
				return fmt.Errorf("error reading file %s from input store %s: %w", filename, args[0], err)
			}

			binReader := dbin.NewReader(obj)
			_, _, err = binReader.ReadHeader()
			if err != nil {
				if err == io.EOF {
					logger.Info("eof reading file header", zap.String("filename", filename))
					return nil
				}
				return fmt.Errorf("error reading file header for file %s: %w", filename, err)
			}
			defer binReader.Close()

			for {
				line, err := binReader.ReadMessage()
				if err == io.EOF {
					break
				}

				if err != nil {
					return fmt.Errorf("reading block: %w", err)
				}

				if len(line) == 0 {
					break
				}

				// decode block data
				bstreamBlock := new(pbbstream.Block)
				err = proto.Unmarshal(line, bstreamBlock)
				if err != nil {
					return fmt.Errorf("unmarshaling block proto: %w", err)
				}

				blockBytes := bstreamBlock.GetPayloadBuffer()

				block := new(pbnear.Block)
				err = proto.Unmarshal(blockBytes, block)
				if err != nil {
					return fmt.Errorf("unmarshaling block proto: %w", err)
				}

				prevHeight := block.Header.PrevHeight
				if prevHeight == 0 {
					if firstBlockSeen {
						logger.Debug("first block, skipping check", zap.Uint64("block", block.Num()))
						firstBlockSeen = false
						continue
					}
					return fmt.Errorf("previous height not set for block number %d in file %s", block.Num(), filename)
				}
			}

			err = obj.Close()
			if err != nil {
				return fmt.Errorf("error closing object %s: %w", filename, err)
			}

			// check range upper bound
			if !blockRange.Unbounded() {
				roundedEndBlock := sftools.RoundToBundleEndBlock(baseNum32, uint32(fileBlockSize))
				if roundedEndBlock >= uint32(blockRange.Stop-1) {
					return errStopWalk
				}
			}

			logger.Info("checked file", zap.String("file", filename))
			return nil
		})

		logger.Debug("file walk ended", zap.Error(err))

		if err != nil && err != errStopWalk {
			return err
		}

		return nil
	}
}

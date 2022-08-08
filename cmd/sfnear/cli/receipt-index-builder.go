package cli

import (
	"context"
	"fmt"

	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	bstransform "github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/dstore"
	indexerApp "github.com/streamingfast/index-builder/app/index-builder"
	"github.com/streamingfast/sf-near/transform"
)

func init() {
	launcher.RegisterApp(&launcher.AppDef{
		ID:          "receipt-index-builder",
		Title:       "Receipt Index Builder",
		Description: "Produces a receipt index for a given set of blocks",
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().Uint64("receipt-index-builder-index-size", 10000, "size of receipt index bundles that will be created")
			cmd.Flags().IntSlice("receipt-index-builder-lookup-index-sizes", []int{1000000, 100000, 10000, 1000}, "index bundle sizes that we will look for on start to find first unindexed block")
			cmd.Flags().String("receipt-index-builder-index-store-url", "", "url of the index store")
			cmd.Flags().Uint64("receipt-index-builder-start-block", 0, "block number to start indexing")
			cmd.Flags().Uint64("receipt-index-builder-stop-block", 0, "block number to stop indexing")
			return nil
		},
		InitFunc: func(runtime *launcher.Runtime) error {
			return nil
		},
		FactoryFunc: func(runtime *launcher.Runtime) (launcher.App, error) {
			sfDataDir := runtime.AbsDataDir

			indexStoreURL := MustReplaceDataDir(sfDataDir, viper.GetString("receipt-index-builder-index-store-url"))
			blockStoreURL := MustReplaceDataDir(sfDataDir, viper.GetString("common-blocks-store-url"))

			indexStore, err := dstore.NewStore(indexStoreURL, "", "", false)
			if err != nil {
				return nil, err
			}

			var lookupIdxSizes []uint64
			lookupIndexSizes := viper.GetIntSlice("receipt-index-builder-lookup-index-sizes")
			for _, size := range lookupIndexSizes {
				if size < 0 {
					return nil, fmt.Errorf("invalid negative size for bundle-sizes: %d", size)
				}
				lookupIdxSizes = append(lookupIdxSizes, uint64(size))
			}

			startBlockResolver := func(ctx context.Context) (uint64, error) {
				select {
				case <-ctx.Done():
					return 0, ctx.Err()
				default:
				}

				startBlockNum := bstransform.FindNextUnindexed(
					ctx,
					viper.GetUint64("receipt-index-builder-start-block"),
					lookupIdxSizes,
					transform.ReceiptAddressIndexShortName,
					indexStore,
				)

				return startBlockNum, nil
			}
			stopBlockNum := viper.GetUint64("receipt-index-builder-stop-block")

			receiptIndexer := transform.NewNearBlockIndexer(indexStore, viper.GetUint64("receipt-index-builder-index-size"))
			handler := bstream.HandlerFunc(func(blk *bstream.Block, obj interface{}) error {
				receiptIndexer.ProcessBlock(blk.ToNative().(*pbcodec.Block))
				return nil
			})

			app := indexerApp.New(&indexerApp.Config{
				BlockHandler:       handler,
				StartBlockResolver: startBlockResolver,
				EndBlock:           stopBlockNum,
				BlockStorePath:     blockStoreURL,
			})

			return app, nil
		},
	})
}

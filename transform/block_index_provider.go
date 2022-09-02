package transform

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	pbtransform "github.com/streamingfast/firehose-near/types/pb/sf/near/transform/v1"
)

const ReceiptAddressIndexShortName = "rcptaddr"

func NewNearBlockIndexProvider(
	store dstore.Store,
	possibleIndexSizes []uint64,
	addresses map[string]bool,
	prefixSuffixPairs []*pbtransform.PrefixSuffixPair,
) *transform.GenericBlockIndexProvider {
	return transform.NewGenericBlockIndexProvider(
		store,
		ReceiptAddressIndexShortName,
		possibleIndexSizes,
		getFilterFunc(addresses, prefixSuffixPairs),
	)
}

func getFilterFunc(accounts map[string]bool, prefixSuffixPairs []*pbtransform.PrefixSuffixPair) func(transform.BitmapGetter) []uint64 {
	return func(bitmaps transform.BitmapGetter) (matchingBlocks []uint64) {
		out := roaring64.NewBitmap()
		for a := range accounts {
			if bm := bitmaps.Get(a); bm != nil {
				out.Or(bm)
			}
		}

		for _, pair := range prefixSuffixPairs {
			if bm := bitmaps.GetByPrefixAndSuffix(pair.Prefix, pair.Suffix); bm != nil {
				out.Or(bm)
			}
		}

		return nilIfEmpty(out.ToArray())
	}
}

// nilIfEmpty is a convenience method which returns nil if the provided slice is empty
func nilIfEmpty(in []uint64) []uint64 {
	if len(in) == 0 {
		return nil
	}
	return in
}

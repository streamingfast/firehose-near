package transform

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
)

const ReceiptAddressIndexShortName = "rcptaddr"

func NewNearBlockIndexProvider(
	store dstore.Store,
	possibleIndexSizes []uint64,
	addresses map[string]bool,
) *transform.GenericBlockIndexProvider {
	return transform.NewGenericBlockIndexProvider(
		store,
		ReceiptAddressIndexShortName,
		possibleIndexSizes,
		getFilterFunc(addresses),
	)
}

func getFilterFunc(accounts map[string]bool) func(transform.BitmapGetter) []uint64 {
	return func(getBitmap transform.BitmapGetter) (matchingBlocks []uint64) {
		out := roaring64.NewBitmap()
		for a := range accounts {
			if bm := getBitmap(a); bm != nil {
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

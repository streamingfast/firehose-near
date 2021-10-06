package codec

import (
	"github.com/streamingfast/bstream"
)

type FilteringPreprocessor struct {
	Filter *BlockFilter
}

func (f *FilteringPreprocessor) PreprocessBlock(blk *bstream.Block) (interface{}, error) {
	return nil, f.Filter.TransformInPlace(blk)
}

type BlockFilter struct {
	IncludeExpression string
	ExcludeExpression string
}

func NewBlockFilter(includeExpression, excludeExpression string) (*BlockFilter, error) {

	return &BlockFilter{
		IncludeExpression: includeExpression,
		ExcludeExpression: excludeExpression,
	}, nil
}

func (f *BlockFilter) TransformInPlace(blk *bstream.Block) error {
	//block := blk.ToNative().(*pbcodec.Block)

	// FIXME: Re-add when proto is changed to know about filtering
	// if filterExprContains(block.FilteringIncludeFilterExpr, include.code) {
	// 	include = includeNOOP
	// }
	// if filterExprContains(block.FilteringExcludeFilterExpr, exclude.code) {
	// 	exclude = excludeNOOP
	// }
	// if include.IsNoop() && exclude.IsNoop() {
	// 	return nil
	// }

	//transformInPlaceV2(block, include, exclude)

	//block.GetChunks()[0].ValidatorProposals[0].

	return nil
}

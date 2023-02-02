package transform

import (
	"testing"

	"github.com/streamingfast/bstream/transform"
	pbtransform "github.com/streamingfast/firehose-near/types/pb/sf/near/transform/v1"
	pbnear "github.com/streamingfast/firehose-near/types/pb/sf/near/type/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func headerOnlyTransform(t *testing.T) *anypb.Any {
	transform := &pbtransform.HeaderOnly{}
	a, err := anypb.New(transform)
	require.NoError(t, err)
	return a
}

func TestHeaderOnly_Transform(t *testing.T) {
	transformReg := transform.NewRegistry()
	transformReg.Register(HeaderOnlyTransformFactory)

	transforms := []*anypb.Any{headerOnlyTransform(t)}

	preprocFunc, x, _, err := transformReg.BuildFromTransforms(transforms)
	require.NoError(t, err)
	require.Nil(t, x)

	block := &pbnear.Block{
		Header: &pbnear.BlockHeader{
			Height:     160,
			PrevHeight: 158,
			Hash:       &pbnear.CryptoHash{Bytes: []byte{0x00, 0xa0}},
			PrevHash:   &pbnear.CryptoHash{Bytes: []byte{0x00, 0x9e}},
		},
		Shards: []*pbnear.IndexerShard{
			{
				ShardId: 1,
			},
		},
		Author: "somehone",
		ChunkHeaders: []*pbnear.ChunkHeader{
			{ChunkHash: []byte{0x01}},
		},
		StateChanges: []*pbnear.StateChangeWithCause{
			{
				Value: &pbnear.StateChangeValue{
					Value: &pbnear.StateChangeValue_AccessKeyUpdate_{},
				},
			},
		},
	}
	blk, err := block.ToBstreamBlock()
	require.NoError(t, err)

	output, err := preprocFunc(blk)
	require.NoError(t, err)

	assertProtoEqual(t, &pbnear.Block{
		Header: &pbnear.BlockHeader{
			Height:     160,
			PrevHeight: 158,
			Hash:       &pbnear.CryptoHash{Bytes: []byte{0x00, 0xa0}},
			PrevHash:   &pbnear.CryptoHash{Bytes: []byte{0x00, 0x9e}},
		},
	}, output.(*pbnear.Block))
}

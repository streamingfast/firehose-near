package codec

import (
	"encoding/hex"
	"testing"

	"github.com/golang/protobuf/proto"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockDecoder(t *testing.T) {
	tests := []struct {
		name        string
		hex         string
		expected    *pbcodec.BlockWrapper
		expectedErr error
	}{
		{
			"standard",
			"0a56125408032a220a20529523882d12fb39302181f2b7b97b0134b83cfd3078656e1cd4645dbcd7030232220a2081b08f80a7c38ef04e4678c51d10312fde4077674319ada99ad79736dd80c1ad78c0e8e19bea88f4ce16",
			&pbcodec.BlockWrapper{
				Block: &pbcodec.Block{
					Author: "",
					Header: &pbcodec.BlockHeader{
						Height:           3,
						Hash:             hash(t, "529523882d12fb39302181f2b7b97b0134b83cfd3078656e1cd4645dbcd70302"),
						PrevHash:         hash(t, "81b08f80a7c38ef04e4678c51d10312fde4077674319ada99ad79736dd80c1ad"),
						TimestampNanosec: 1629687641986856000,
					},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(test.hex)
			require.NoError(t, err)

			actual := new(pbcodec.BlockWrapper)
			err = proto.Unmarshal(bytes, actual)

			if test.expectedErr == nil {
				require.NoError(t, err)
				assertProtoEqual(t, test.expected, actual)
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func hash(t *testing.T, in string) *pbcodec.CryptoHash {
	t.Helper()

	out, err := hex.DecodeString(in)
	require.NoError(t, err, "invalid hex %q", in)

	return &pbcodec.CryptoHash{
		Bytes: out,
	}
}

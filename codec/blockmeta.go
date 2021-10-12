package codec

import (
	"container/heap"
	"context"
	"fmt"
	"time"

	"github.com/streamingfast/near-go/rpc"
)

type blockMeta struct {
	id        string
	number    uint64
	blockTime time.Time
}

type blockMetaGetter interface {
	getBlockMeta(id string) (*blockMeta, error)
}

type RPCBlockMetaGetter struct {
	client *rpc.Client
}

func NewRPCBlockMetaGetter(endpointURL string) *RPCBlockMetaGetter {
	return &RPCBlockMetaGetter{client: rpc.NewClient(endpointURL)}
}

func (g *RPCBlockMetaGetter) getBlockMeta(id string) (*blockMeta, error) {
	res, err := g.client.GetBlock(context.Background(), id)
	if err != nil {
		return nil, err
	}

	return &blockMeta{
		id:        res.Header.Hash,
		number:    uint64(res.Header.Height),
		blockTime: time.Unix(res.Header.Timestamp, 0),
	}, nil
}

type blockMetaHeap struct {
	metas  []*blockMeta
	getter blockMetaGetter
}

func newBlockMetaHeap(getter blockMetaGetter) *blockMetaHeap {
	h := &blockMetaHeap{
		metas:  []*blockMeta{},
		getter: getter,
	}
	return h
}

func (h *blockMetaHeap) get(id string) *blockMeta {
	for _, bm := range h.metas {
		if bm.id == id {
			return bm
		}
	}

	bm, err := h.getter.getBlockMeta(id)
	if err != nil {
		panic(fmt.Errorf("getting block for id: %s, %w", id, err))
	}

	if bm == nil {
		panic(fmt.Errorf("block getter return nil block for id: %s", id))
	}

	heap.Push(h, bm)
	return bm
}

func (h *blockMetaHeap) purge(upToID string) {
	if bm := h.get(upToID); bm == nil {
		return
	}
	for {
		bm := heap.Pop(h).(*blockMeta)
		if bm.id == upToID {
			break
		}
	}
}

func (h *blockMetaHeap) Len() int {
	return len(h.metas)
}

func (h *blockMetaHeap) Less(i, j int) bool {
	return h.metas[i].blockTime.Before(h.metas[j].blockTime)
}

func (h *blockMetaHeap) Swap(i, j int) {
	if len(h.metas) == 0 {
		return
	}

	h.metas[i], h.metas[j] = h.metas[j], h.metas[i]
}

func (h *blockMetaHeap) Push(x interface{}) {
	bm := x.(*blockMeta)
	h.metas = append(h.metas, bm)
}

func (h *blockMetaHeap) Pop() interface{} {
	old := h.metas
	if len(old) == 0 {
		return nil
	}
	n := len(old)
	bm := old[n-1]
	h.metas = old[0 : n-1]
	return bm
}

package substrate

import (
	"strconv"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

type Client struct {
	// w       *wallet.Key
	api     *gsrpc.SubstrateAPI
	vu      modules.VU
	metrics ethMetrics
}

func (c *Client) Exports() modules.Exports {
	return modules.Exports{}
}

// GetBlockHashLatests returns the current block number.
func (c *Client) GetBlockHashLatest() (types.Hash, error) {
	return c.api.RPC.Chain.GetBlockHashLatest()
}

// PollBlocks polls for new blocks and emits a "block" metric.
func (c *Client) subscribeNewHeads() {
	sub, err := c.api.RPC.Chain.SubscribeNewHeads()
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		head := <-sub.Chan()
		bh, err := c.api.RPC.Chain.GetBlockHash(uint64(head.Number))
		if err != nil {
			panic(err)
		}

		block, err := c.api.RPC.Chain.GetBlock(bh)
		if err != nil {
			panic(err)
		}

		if c.vu.Context() != nil {
			metrics.PushIfNotDone(c.vu.Context(), c.vu.State().Samples, metrics.ConnectedSamples{
				Samples: []metrics.Sample{
					{
						Metric: c.metrics.Block,
						Tags: metrics.NewSampleTags(map[string]string{
							"extrinsics": strconv.Itoa(len(block.Block.Extrinsics)),
						}),
						Value: float64(head.Number),
						Time:  time.Now(),
					},
					// {
					// 	Metric: c.metrics.TPS,
					// 	// Tags: metrics.NewSampleTags(map[string]string{}),
					// 	Value: tps,
					// 	Time:  time.Now(),
					// },
				},
			})
		}
	}

}

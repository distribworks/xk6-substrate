package substrate

import (
	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"go.k6.io/k6/js/modules"
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

// xk6 build --with github.com/distribworks/xk6-substrate
package substrate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

const (
	privateKey = ""
)

type ethMetrics struct {
	RequestDuration *metrics.Metric
	TimeToMine      *metrics.Metric
	Block           *metrics.Metric
	TPS             *metrics.Metric
}

func init() {
	modules.Register("k6/x/substrate", &Root{})
}

// Eth is the root module
type Root struct{}

// NewModuleInstance implements the modules.Module interface returning a new instance for each VU.
func (*Root) NewModuleInstance(vu modules.VU) modules.Instance {
	m, err := registerMetrics(vu)
	if err != nil {
		common.Throw(vu.Runtime(), err)
	}

	return &ModuleInstance{
		vu: vu,
		m:  m,
	}
}

type ModuleInstance struct {
	vu modules.VU
	m  ethMetrics
}

// Exports implements the modules.Instance interface and returns the exported types for the JS module.
func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{Named: map[string]interface{}{
		"Client": mi.NewClient,
	}}
}

func (mi *ModuleInstance) NewClient(call goja.ConstructorCall) *goja.Object {
	rt := mi.vu.Runtime()

	var optionsArg map[string]interface{}
	err := rt.ExportTo(call.Arguments[0], &optionsArg)
	if err != nil {
		common.Throw(rt, errors.New("unable to parse options object"))
	}

	opts, err := newOptionsFrom(optionsArg)
	if err != nil {
		common.Throw(rt, fmt.Errorf("invalid options; reason: %w", err))
	}

	if opts.URL == "" {
		opts.URL = "http://localhost:8545"
	}

	if opts.PrivateKey == "" {
		opts.PrivateKey = privateKey
	}

	api, err := gsrpc.NewSubstrateAPI("wss://poc3-rpc.polkadot.io")
	if err != nil {
		common.Throw(rt, fmt.Errorf("invalid options; reason: %w", err))
	}

	client := &Client{
		vu:      mi.vu,
		metrics: mi.m,
		api:     api,
		// w:       wa,
	}

	// go client.pollForBlocks()

	return rt.ToValue(client).ToObject(rt)
}

func registerMetrics(vu modules.VU) (ethMetrics, error) {
	var err error
	registry := vu.InitEnv().Registry
	m := ethMetrics{}

	m.RequestDuration, err = registry.NewMetric("ethereum_req_duration", metrics.Trend, metrics.Time)
	if err != nil {
		return m, err
	}
	m.TimeToMine, err = registry.NewMetric("ethereum_time_to_mine", metrics.Trend, metrics.Time)
	if err != nil {
		return m, err
	}
	m.Block, err = registry.NewMetric("ethereum_block", metrics.Counter, metrics.Default)
	if err != nil {
		return m, err
	}
	m.TPS, err = registry.NewMetric("ethereum_tps", metrics.Gauge, metrics.Default)
	if err != nil {
		return m, err
	}

	return m, nil
}

func (c *Client) reportMetricsFromStats(call string, t time.Duration) {
	now := time.Now()
	tags := metrics.NewSampleTags(map[string]string{"call": call})
	ctx := c.vu.Context()
	metrics.PushIfNotDone(ctx, c.vu.State().Samples, metrics.ConnectedSamples{
		Samples: []metrics.Sample{
			{
				Metric: c.metrics.RequestDuration,
				Tags:   tags,
				Value:  float64(t / time.Millisecond),
				Time:   now,
			},
		},
	})
}

// options defines configuration options for the client.
type options struct {
	URL        string `json:"url,omitempty"`
	Mnemonic   string `json:"mnemonic,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
}

// newOptionsFrom validates and instantiates an options struct from its map representation
// as obtained by calling a Goja's Runtime.ExportTo.
func newOptionsFrom(argument map[string]interface{}) (*options, error) {
	jsonStr, err := json.Marshal(argument)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize options to JSON %w", err)
	}

	// Instantiate a JSON decoder which will error on unknown
	// fields. As a result, if the input map contains an unknown
	// option, this function will produce an error.
	decoder := json.NewDecoder(bytes.NewReader(jsonStr))
	decoder.DisallowUnknownFields()

	var opts options
	err = decoder.Decode(&opts)
	if err != nil {
		return nil, fmt.Errorf("unable to decode options %w", err)
	}

	return &opts, nil
}

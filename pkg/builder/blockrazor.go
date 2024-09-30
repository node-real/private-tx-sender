package builder

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	jsoniter "github.com/json-iterator/go"

	"github.com/node-real/private-tx-sender/pkg/rpc"
)

const BlockrazorMethod = "eth_sendBundle"

func newBlockrazor(cfg Config) Builder {
	return &blockrazor{
		key: cfg.Key,
		builder: &builder{
			url:   cfg.URL,
			brand: cfg.Brand,
		},
	}
}

type blockrazor struct {
	key string
	*builder
}

func (b *blockrazor) SendBundle(ctx context.Context, args *types.SendBundleArgs, bundleLifeNumber uint64) error {
	req, err := newBlockrazorRequest(args, bundleLifeNumber)
	if err != nil {
		log.Error("failed to create blockrazor jsonrpc request", "err", err)
		return err
	}

	opt := rpc.WithHeader(map[string]string{
		"Authorization": b.key,
	})

	err = SendBundleCall(ctx, b.url, req, opt)
	if err != nil {
		log.Error("failed to send blockrazor bundle", "err", err)
		return err
	}

	return nil
}

func (b *blockrazor) GetBrand() string {
	return string(b.brand)
}

type blockrazorBody struct {
	Txs               []hexutil.Bytes `json:"txs"`
	MaxBlockNumber    uint64          `json:"maxBlockNumber"`
	MinTimestamp      uint64          `json:"minTimestamp,omitempty"`
	MaxTimestamp      uint64          `json:"maxTimestamp,omitempty"`
	RevertingTxHashes []common.Hash   `json:"revertingTxHashes,omitempty"`
}

func newBlockrazorRequest(args *types.SendBundleArgs, _ uint64) (*rpc.JsonrpcRequest, error) {
	body := blockrazorBody{
		Txs:               args.Txs,
		MaxBlockNumber:    args.MaxBlockNumber,
		RevertingTxHashes: args.RevertingTxHashes,
	}

	if args.MaxTimestamp != nil {
		body.MaxTimestamp = *args.MaxTimestamp
	}

	if args.MinTimestamp != nil {
		body.MinTimestamp = *args.MinTimestamp
	}

	bodybyte, err := jsoniter.Marshal(body)
	if err != nil {
		log.Error("failed to marshal blockrazor body", "err", err)
		return nil, err
	}

	param := bodybyte
	params := []rpc.Param{param}

	return &rpc.JsonrpcRequest{
		ID:      1,
		Version: "2.0",
		Method:  BlockrazorMethod,
		Params:  params,
	}, nil
}

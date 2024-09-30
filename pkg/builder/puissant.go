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

const PuissantMethod = "eth_sendPuissant"

func newPuissant(cfg Config) Builder {
	return &puissant{
		builder: &builder{
			url:   cfg.URL,
			brand: cfg.Brand,
		},
	}
}

type puissant struct {
	*builder
}

func (b *puissant) SendBundle(ctx context.Context, args *types.SendBundleArgs, _ uint64) error {
	req, err := newPuissantRequest(args)
	if err != nil {
		log.Error("failed to create puissant jsonrpc request", "err", err)
		return err
	}

	err = SendBundleCall(ctx, b.url, req)
	if err != nil {
		log.Error("failed to send puissant bundle", "err", err)
		return err
	}

	return nil
}

func (b *puissant) GetBrand() string {
	return string(b.brand)
}

type puissantBundleBody struct {
	Txs             []hexutil.Bytes `json:"txs"`
	MaxTimestamp    uint64          `json:"maxTimestamp"`
	AcceptReverting []common.Hash   `json:"acceptReverting"`
}

func newPuissantRequest(args *types.SendBundleArgs) (*rpc.JsonrpcRequest, error) {
	body := &puissantBundleBody{
		Txs:             args.Txs,
		AcceptReverting: args.RevertingTxHashes,
	}

	if args.MaxTimestamp != nil {
		body.MaxTimestamp = *args.MaxTimestamp
	}

	bodybytes, err := jsoniter.Marshal(body)
	if err != nil {
		log.Error("failed to marshal puissant body", "err", err)
		return nil, err
	}

	param := bodybytes
	params := []rpc.Param{param}

	return &rpc.JsonrpcRequest{
		ID:      1,
		Version: "2.0",
		Method:  PuissantMethod,
		Params:  params,
	}, nil
}

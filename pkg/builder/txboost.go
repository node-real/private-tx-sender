package builder

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	jsoniter "github.com/json-iterator/go"

	"github.com/node-real/private-tx-sender/pkg/rpc"
)

const TxboostMethod = "eth_sendBundle"

func newTxboost(cfg Config) Builder {
	return &txboost{
		key: cfg.Key,
		builder: &builder{
			url:   cfg.URL,
			brand: cfg.Brand,
		},
	}
}

type txboost struct {
	key string
	*builder
}

func (b *txboost) SendBundle(ctx context.Context, args *types.SendBundleArgs, bundleLifeNumber uint64) error {
	req, err := newTxboostRequest(args, bundleLifeNumber)
	if err != nil {
		log.Error("failed to create txboost jsonrpc request", "err", err)
		return err
	}

	opt := rpc.WithHeader(map[string]string{
		"Authorization": b.key,
	})

	err = SendBundleCall(ctx, b.url, req, opt)
	if err != nil {
		log.Error("failed to send txboost bundle", "err", err)
		return err
	}

	return nil
}

func (b *txboost) GetBrand() string {
	return string(b.brand)
}

type txboostBody struct {
	Txs               []hexutil.Bytes `json:"txs"`
	BlockNumber       string          `json:"blockNumber"`
	MinTimestamp      uint64          `json:"minTimestamp,omitempty"`
	MaxTimestamp      uint64          `json:"maxTimestamp,omitempty"`
	RevertingTxHashes []common.Hash   `json:"revertingTxHashes,omitempty"`
	IgnoreBlockNumber bool            `json:"ignoreBlockNumber"`
}

func newTxboostRequest(args *types.SendBundleArgs, bundleLifeNumber uint64) (*rpc.JsonrpcRequest, error) {
	maxBlockNumber := args.MaxBlockNumber
	nextBlockNumber := maxBlockNumber - bundleLifeNumber + 1
	blockNumber := hexutil.EncodeBig(big.NewInt(int64(nextBlockNumber)))

	body := &txboostBody{
		Txs:               args.Txs,
		BlockNumber:       blockNumber,
		RevertingTxHashes: args.RevertingTxHashes,
		IgnoreBlockNumber: true,
	}

	if args.MaxTimestamp != nil {
		body.MaxTimestamp = *args.MaxTimestamp
	}

	if args.MinTimestamp != nil {
		body.MinTimestamp = *args.MinTimestamp
	}

	bodybyte, err := jsoniter.Marshal(body)
	if err != nil {
		log.Error("failed to marshal puissant body", "err", err)
		return nil, err
	}

	param := bodybyte
	params := []rpc.Param{param}

	return &rpc.JsonrpcRequest{
		ID:      1,
		Version: "2.0",
		Method:  TxboostMethod,
		Params:  params,
	}, nil
}

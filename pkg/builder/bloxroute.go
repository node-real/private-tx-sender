package builder

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/node-real/private-tx-sender/pkg/rpc"
)

const (
	BloxrouteBlockchainNetwork = "BSC-Mainnet"
	BloxrouteMethod            = "blxr_submit_bundle"
)

func newBloxroute(cfg Config) Builder {
	return &bloxroute{
		key: cfg.Key,
		builder: &builder{
			url:   cfg.URL,
			brand: cfg.Brand,
		},
	}
}

type bloxroute struct {
	key string
	*builder
}

// SendBundle sends a bundle to bloxroute TODO customize bundler for paying to bloxroute builder
func (b *bloxroute) SendBundle(ctx context.Context, args *types.SendBundleArgs, bundleLifeNumber uint64) error {
	req, err := newBloxrouteRequest(args, bundleLifeNumber)
	if err != nil {
		log.Error("failed to create bloxroute jsonrpc request", "err", err)
		return err
	}

	opt := rpc.WithHeader(map[string]string{
		"Authorization": b.key,
	})

	err = SendBundleCall(ctx, b.url, req, opt)
	if err != nil {
		log.Error("failed to send bloxroute bundle", "err", err)
		return err
	}

	return nil
}

func (b *bloxroute) GetBrand() string {
	return string(b.brand)
}

type bloxrouteBundleBody struct {
	Transaction       []hexutil.Bytes `json:"transaction"`
	BlockchainNetwork string          `json:"blockchain_network"`
	BlockNumber       string          `json:"block_number"`
	MaxTimestamp      uint64          `json:"max_timestamp"`
	RevertingHashes   []common.Hash   `json:"reverting_hashes"`
}

func newBloxrouteRequest(args *types.SendBundleArgs, bundleLifeNumber uint64) (*rpcRequest, error) {
	maxBlockNumber := args.MaxBlockNumber
	nextBlockNumber := maxBlockNumber - bundleLifeNumber + 1
	blockNumber := hexutil.EncodeBig(big.NewInt(int64(nextBlockNumber)))

	body := bloxrouteBundleBody{
		Transaction:       args.Txs,
		BlockchainNetwork: BloxrouteBlockchainNetwork,
		BlockNumber:       blockNumber,
		RevertingHashes:   args.RevertingTxHashes,
	}

	if args.MaxTimestamp != nil {
		body.MaxTimestamp = *args.MaxTimestamp
	}

	return &rpcRequest{
		ID:     1,
		Method: BloxrouteMethod,
		Params: body,
	}, nil
}

type rpcRequest struct {
	ID     uint64              `json:"id"`
	Method string              `json:"method"`
	Params bloxrouteBundleBody `json:"params"`
}

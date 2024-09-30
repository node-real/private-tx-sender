package builder

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

func newNodeReal(cfg Config) Builder {
	client, err := ethclient.Dial(cfg.URL)
	if err != nil {
		log.Crit("failed to dial ethclient", "url", cfg.URL, "err", err)
		return nil
	}

	return &nodeReal{
		builder: &builder{
			url:   cfg.URL,
			brand: cfg.Brand,
		},
		ethclient: client,
	}
}

type nodeReal struct {
	*builder
	ethclient *ethclient.Client
}

func (b *nodeReal) SendBundle(ctx context.Context, args *types.SendBundleArgs, _ uint64) error {
	bundlehash, err := b.ethclient.SendBundle(ctx, *args)
	if err != nil {
		log.Error("failed to send bundle", "url", b.url, "err", err)
		return err
	}

	log.Info("send bundle success", "url", b.url, "bundle_hash", bundlehash)
	return nil
}

func (b *nodeReal) GetBrand() string {
	return string(b.brand)
}

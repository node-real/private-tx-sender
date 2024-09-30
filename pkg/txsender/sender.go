package txsender

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/hashicorp/go-multierror"
	"github.com/tredeske/u/ustrings"

	"github.com/node-real/private-tx-sender/pkg/builder"
	"github.com/node-real/private-tx-sender/pkg/rpc"
)

type PrivateTxSender interface {
	SendRawTransaction(ctx context.Context, input hexutil.Bytes, revertible bool) error
}

type Duration time.Duration

func (d *Duration) MarshalText() ([]byte, error) {
	return ustrings.UnsafeStringToBytes(time.Duration(*d).String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	dd, err := time.ParseDuration(ustrings.UnsafeBytesToString(text))
	*d = Duration(dd)
	return err
}

type Config struct {
	ChainURL         string
	BlockInterval    Duration
	BundleLifeNumber uint64
}

type privateTxSender struct {
	cfg            Config
	bundleLifeTime time.Duration
	client         *ethclient.Client
	latestHeader   atomic.Pointer[types.Header]
	builders       []builder.Builder
}

func NewPrivateTxSender(ctx context.Context, cfg Config, builders []builder.Builder) PrivateTxSender {
	client, err := ethclient.DialOptions(ctx, cfg.ChainURL, ethrpc.WithHTTPClient(rpc.HTTPClient))
	if err != nil {
		log.Error("failed to dial chain", "err", err)
		return nil
	}

	s := &privateTxSender{
		cfg:            cfg,
		bundleLifeTime: time.Duration(cfg.BundleLifeNumber) * time.Duration(cfg.BlockInterval),
		client:         client,
		builders:       builders,
	}

	s.storeHeader()

	go s.refresh(ctx)

	return s
}

func (s *privateTxSender) refresh(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.storeHeader()
		}
	}
}

func (s *privateTxSender) storeHeader() {
	header, err := s.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Error("failed to get latest header", "err", err)
		return
	}

	s.latestHeader.Store(header)
}

func (s *privateTxSender) SendRawTransaction(_ context.Context, input hexutil.Bytes, revertible bool) error {
	latestHeader := s.latestHeader.Load()
	minTimestamp := uint64(time.Unix(int64(latestHeader.Time), 0).Add(time.Duration(s.cfg.BlockInterval)).Unix())
	maxTimestamp := uint64(time.Unix(int64(latestHeader.Time), 0).Add(s.bundleLifeTime).Unix())
	sendBundlerArgs := &types.SendBundleArgs{
		Txs:            []hexutil.Bytes{input},
		MaxBlockNumber: latestHeader.Number.Uint64() + s.cfg.BundleLifeNumber,
		MinTimestamp:   &minTimestamp,
		MaxTimestamp:   &maxTimestamp,
	}

	if revertible {
		tx := new(types.Transaction)
		if err := tx.UnmarshalBinary(input); err != nil {
			log.Error("failed to unmarshal tx", "err", err)
			return err
		}

		sendBundlerArgs.RevertingTxHashes = []common.Hash{tx.Hash()}
	}

	sendTasks := make([]func() (common.Hash, error), len(s.builders))

	for idx, builder := range s.builders {
		builder := builder

		sendTasks[idx] = func() (common.Hash, error) {
			err := builder.SendBundle(context.Background(), sendBundlerArgs, s.cfg.BundleLifeNumber)
			if err != nil {
				log.Error("send bundle to builder failed", "builder", builder.GetBrand(), "err", err.Error())
			} else {
				log.Info("send bundle to builder success", "builder", builder.GetBrand())
			}

			return common.Hash{}, err
		}
	}

	if _, err := RunForOnlyOneSucceed(sendTasks...); err != nil {
		return err
	}

	return nil
}

// RunForOnlyOneSucceed returns in two conditions:
// 1. one task succeed
// 2. all tasks failed
func RunForOnlyOneSucceed[T any](tasks ...func() (T, error)) (t T, err error) {
	respChan := make(chan batchResp[T], len(tasks))
	for _, task := range tasks {
		task := task

		go func() {
			var resp batchResp[T]
			resp.t, resp.err = task()
			respChan <- resp
		}()
	}

	respCounter := 0
	for resp := range respChan {
		respCounter++
		if resp.err == nil {
			t = resp.t
			return t, nil
		}

		err = multierror.Append(err, resp.err)

		if respCounter == len(tasks) {
			break
		}
	}
	return
}

type batchResp[T any] struct {
	t   T
	err error
}

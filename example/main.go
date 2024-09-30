package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"math/big"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/node-real/private-tx-sender/pkg/builder"
	"github.com/node-real/private-tx-sender/pkg/txsender"
)

var (
	configPath = flag.String("config", "./config.toml", "Give a config file path")
	privatekey = flag.String("privatekey", "", "Give a private key")
)

type Config struct {
	Sender   txsender.Config
	Builders []builder.Config
}

func main() {
	flag.Parse()

	cfg := LoadConfig(*configPath)

	ctx, _ := context.WithCancel(context.Background())

	client, err := ethclient.Dial(cfg.Sender.ChainURL)
	if err != nil {
		panic("failed to dial chain")
	}

	builders := make([]builder.Builder, 0)
	for _, v := range cfg.Builders {
		builders = append(builders, builder.New(v))
	}

	txSender := txsender.NewPrivateTxSender(ctx, cfg.Sender, builders)

	txHash, txByte := generateTx(client)

	err = txSender.SendRawTransaction(ctx, txByte, true)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Duration(cfg.Sender.BundleLifeNumber) * time.Duration(cfg.Sender.BlockInterval))
	receipt, err := client.TransactionReceipt(ctx, txHash)
	if err != nil {
		panic(err)
	}

	println("txHash:", txHash.Hex(), "status:", receipt.Status)
}

func generateTx(client *ethclient.Client) (common.Hash, hexutil.Bytes) {
	key, err := crypto.HexToECDSA(*privatekey)
	if err != nil {
		panic("failed to load private key")
	}

	pubKey := key.Public()
	pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		panic("public key is not *ecdsa.PublicKey")
	}

	addr := crypto.PubkeyToAddress(*pubKeyECDSA)

	nonce, err := client.NonceAt(context.Background(), addr, nil)
	if err != nil {
		panic(err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &addr,
		Value:    big.NewInt(0),
		Gas:      21000,
		GasPrice: big.NewInt(1e9),
		Data:     nil,
	})

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), key)
	if err != nil {
		panic(err)
	}

	rawTx, err := signedTx.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return signedTx.Hash(), rawTx
}

func LoadConfig(path string) Config {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		panic(err)
	}

	return cfg
}

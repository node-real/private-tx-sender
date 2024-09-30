package builder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"

	"github.com/node-real/private-tx-sender/pkg/rpc"
)

type Brand string

const (
	Nodereal   Brand = "nodereal"
	Puissant   Brand = "puissant"
	Txboost    Brand = "txboost"
	Bloxroute  Brand = "bloxroute"
	Blockrazor Brand = "blockrazor"
)

type Config struct {
	Brand Brand
	URL   string
	Key   string // api key for authentication
}

func New(cfg Config) Builder {
	switch cfg.Brand {
	case Nodereal:
		return newNodeReal(cfg)
	case Puissant:
		return newPuissant(cfg)
	case Txboost:
		return newTxboost(cfg)
	case Bloxroute:
		return newBloxroute(cfg)
	case Blockrazor:
		return newBlockrazor(cfg)
	default:
		log.Crit("invalid builder brand", "brand", cfg.Brand)
	}

	return nil
}

type Builder interface {
	SendBundle(ctx context.Context, args *types.SendBundleArgs, bundleLifeNumber uint64) error
	GetBrand() string
}

type builder struct {
	brand Brand
	url   string
}

func SendBundleCall(ctx context.Context, url string, req interface{}, options ...rpc.CallOption) error {
	opt := &rpc.CallOptions{Header: map[string]string{"Content-Type": gin.MIMEJSON}}
	opt.ApplyOptions(options...)

	reqByte, err := jsoniter.Marshal(req)
	if err != nil {
		log.Error("failed to marshal jsonrpc request body", "url", url, "err", err)
		return err
	}

	httpReq, httpErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqByte))
	if httpErr != nil {
		log.Error("failed to create jsonrpc http request", "url", url, "err", httpErr)
		return err
	}

	for k, v := range opt.Header {
		httpReq.Header.Set(k, v)
	}

	httpResp, err := rpc.HTTPClient.Do(httpReq)
	if err != nil {
		ErrorCounter.WithLabelValues(url).Inc()

		log.Error("failed to send jsonrpc http request", "url", url, "err", err)
		return err
	}
	defer httpResp.Body.Close()

	if !rpc.HTTPCode(httpResp.StatusCode).Success() {
		ErrorCounter.WithLabelValues(url).Inc()

		log.Error("failed to send jsonrpc http request", "url", url, "code", httpResp.StatusCode)
		return fmt.Errorf("failed to send jsonrpc http request, code: %d", httpResp.StatusCode)
	}

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		log.Error("failed to read response of jsonrpc call", "url", url, "err", err)
		return err
	}

	resp := rpc.JsonrpcResponse{}
	err = jsoniter.Unmarshal(body, &resp)
	if err != nil {
		log.Error("failed to unmarshal response of jsonrpc call", "url", url, "err", err)
		return err
	}

	if resp.Error != nil {
		ErrorCounter.WithLabelValues(url).Inc()

		jrError := rpc.JsonrpcError{}
		err = jsoniter.Unmarshal(*resp.Error, &jrError)
		if err != nil {
			log.Error("failed to unmarshal resp.Error", "url", url, "err", err)
			return err
		}

		if jrError.Code == rpc.InternalErrorCode {
			log.Error(" response internal error", "url", url)
			return errors.New(" response internal error")
		}

		log.Error(" response error", "code", jrError.Code, "message", jrError.Message)
		return errors.New(jrError.Message)
	}

	log.Info("send bundle success", "url", url, "bundle_hash", resp.Result)
	return nil
}

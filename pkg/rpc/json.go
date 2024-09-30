package rpc

import (
	"encoding/json"
	"errors"
)

var (
	InternalErrorCode = -32603
)

type Param json.RawMessage
type Params []Param

// MarshalJSON returns m as the JSON encoding of m.
func (m Param) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *Param) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

type JsonrpcRequest struct {
	ID      uint64 `json:"id"`
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  Params `json:"params"`
}

// JsonrpcResponse keeps Result and Error as unparsed JSON
// It is meant to be used to deserialize JSONPRC responses from downstream components
// while Response is meant to be used to craft our own responses to clients.
type JsonrpcResponse struct {
	ID      uint64           `json:"id"`
	JSONRPC string           `json:"jsonrpc"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *json.RawMessage `json:"error,omitempty"`
}

type JsonrpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

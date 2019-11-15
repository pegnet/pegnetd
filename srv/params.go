// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package srv

import (
	"time"

	jrpc "github.com/AdamSLevy/jsonrpc2/v13"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/node/pegnet"
)

type Params interface {
	IsValid() error
	ValidChainID() *factom.Bytes32
	HasIncludePending() bool
}

type ParamsGetGlobalRichList struct {
	Count int `json:"count,omitempty"`
}

func (p ParamsGetGlobalRichList) HasIncludePending() bool { return false }
func (p ParamsGetGlobalRichList) IsValid() error {
	if p.Count < 0 {
		return jrpc.ErrorInvalidParams("count must be >= 0")
	}
	return nil
}
func (p ParamsGetGlobalRichList) ValidChainID() *factom.Bytes32 {
	return nil
}

type ParamsGetRichList struct {
	Asset string `json:"asset,omitempty"`
	Count int    `json:"count,omitempty"`
}

func (p ParamsGetRichList) HasIncludePending() bool { return false }
func (p ParamsGetRichList) IsValid() error {
	ticker := fat2.StringToTicker(p.Asset)
	if ticker == fat2.PTickerInvalid {
		return jrpc.ErrorInvalidParams("invalid asset")
	}
	if p.Count < 0 {
		return jrpc.ErrorInvalidParams("count must be >= 0")
	}
	return nil
}
func (p ParamsGetRichList) ValidChainID() *factom.Bytes32 {
	return nil
}

// ParamsToken scopes a request down to a single FAT token using either the
// ChainID or both the TokenID and the IssuerChainID.
type ParamsToken struct {
	ChainID *factom.Bytes32 `json:"chainid,omitempty"`
}

func (p ParamsToken) IsValid() error {
	if p.ChainID != nil {
		return nil
	}
	return jrpc.ErrorInvalidParams(`required: "chainid"`)
}

func (p ParamsToken) HasIncludePending() bool { return false }

func (p ParamsToken) ValidChainID() *factom.Bytes32 {
	if p.ChainID != nil {
		return p.ChainID
	}
	return p.ChainID
}

// ParamsGetTransaction is used to query for a single particular transaction
// with the given Entry Hash.
type ParamsGetTransaction struct {
	ParamsToken
	Hash    *factom.Bytes32 `json:"entryhash"`
	TxIndex uint64          `json:"txindex,omitempty"`
}

func (p ParamsGetTransaction) IsValid() error {
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if p.Hash == nil {
		return jrpc.ErrorInvalidParams(`required: "entryhash"`)
	}
	return nil
}

type ParamsGetPegnetRates struct {
	Height *uint32 `json:"height,omitempty"`
}

func (ParamsGetPegnetRates) HasIncludePending() bool { return false }

func (p ParamsGetPegnetRates) IsValid() error {
	if p.Height == nil {
		return jrpc.ErrorInvalidParams(`required: "height"`)
	}
	return nil
}
func (ParamsGetPegnetRates) ValidChainID() *factom.Bytes32 {
	return nil
}

type ParamsGetPegnetTransactionStatus struct {
	Hash *factom.Bytes32 `json:"entryhash,omitempty"`
}

func (p ParamsGetPegnetTransactionStatus) HasIncludePending() bool { return false }
func (p ParamsGetPegnetTransactionStatus) IsValid() error {
	if p.Hash == nil {
		return jrpc.ErrorInvalidParams(`required: "entryhash"`)
	}
	return nil
}
func (p ParamsGetPegnetTransactionStatus) ValidChainID() *factom.Bytes32 {
	return nil
}

// ParamsGetPegnetTransaction are the parameters for retrieving transactions from
// the history system.
// You need to specify exactly one of either `hash`, `address`, or `height`.
// `offset` is the value from a previous query's `nextoffset`.
// `desc` returns transactions in newest->oldest order
type ParamsGetPegnetTransaction struct {
	Hash       string `json:"entryhash,omitempty"`
	Address    string `json:"address,omitempty"`
	Height     int    `json:"height,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	Desc       bool   `json:"desc,omitempty"`
	Transfer   bool   `json:"transfer,omitempty"`
	Conversion bool   `json:"conversion,omitempty"`
	Coinbase   bool   `json:"coinbase,omitempty"`
	Burn       bool   `json:"burn,omitempty"`
	Asset      string `json:"asset,omitempty"`

	// TxID is in the format #-[Entryhash], where '#' == tx index
	TxID string `json:"txid,omitempty"`
	// Used by the server to store the entryhash in the txid
	txEntryHash string
}

func (p ParamsGetPegnetTransaction) HasIncludePending() bool { return false }
func (p ParamsGetPegnetTransaction) IsValid() error {
	if p.Offset < 0 {
		return jrpc.ErrorInvalidParams(`offset must be >= 0`)
	}
	// check that only one is set
	var count int
	if p.Hash != "" {
		count++
	}
	if p.Address != "" {
		count++
	}
	if p.TxID != "" {
		count++
	}
	if p.Height > 0 {
		count++
	}
	if count != 1 {
		if count == 0 {
			return jrpc.ErrorInvalidParams(`need to specify either "entryhash" or "address", "txid", or "height"`)
		}
		return jrpc.ErrorInvalidParams(`cannot specify more than one of "entryhash", "address", "txid", or "height"`)
	}

	if p.Asset != "" {
		ticker := fat2.StringToTicker(p.Asset)
		if ticker == fat2.PTickerInvalid {
			return jrpc.ErrorInvalidParams("invalid asset filter")
		}
	}

	// error check input
	if p.Address != "" {
		if _, err := factom.NewFAAddress(p.Address); err != nil {
			return jrpc.ErrorInvalidParams("address: " + err.Error())
		}
	}
	if p.Hash != "" {
		hash := new(factom.Bytes32)
		if err := hash.UnmarshalText([]byte(p.Hash)); err != nil {
			return jrpc.ErrorInvalidParams("entryhash: " + err.Error())
		}
	}
	if p.TxID != "" {
		_, _, err := pegnet.SplitTxID(p.TxID)
		if err != nil {
			return jrpc.ErrorInvalidParams("txid: " + err.Error())
		}
	}

	return nil
}
func (p ParamsGetPegnetTransaction) ValidChainID() *factom.Bytes32 {
	return nil
}

type ParamsGetPegnetBalances struct {
	Address *factom.FAAddress `json:"address,omitempty"`
}

func (p ParamsGetPegnetBalances) HasIncludePending() bool { return false }

func (p ParamsGetPegnetBalances) IsValid() error {
	if p.Address == nil {
		return jrpc.ErrorInvalidParams(`required: "address"`)
	}
	return nil
}
func (p ParamsGetPegnetBalances) ValidChainID() *factom.Bytes32 {
	return nil
}

type ParamsSendTransaction struct {
	ParamsToken
	ExtIDs  []factom.Bytes `json:"extids,omitempty"`
	Content factom.Bytes   `json:"content,omitempty"`
	Raw     factom.Bytes   `json:"raw,omitempty"`
	DryRun  bool           `json:"dryrun,omitempty"`
	entry   factom.Entry
}

func (p *ParamsSendTransaction) IsValid() error {
	if p.Raw != nil {
		if p.ExtIDs != nil || p.Content != nil || p.ParamsToken != (ParamsToken{}) {
			return jrpc.ErrorInvalidParams(
				`"raw cannot be used with "content" or "extids"`)
		}
		if err := p.entry.UnmarshalBinary(p.Raw); err != nil {
			return jrpc.ErrorInvalidParams(err)
		}
		p.entry.Timestamp = time.Now()
		p.ChainID = p.entry.ChainID
		return nil
	}
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if len(p.Content) == 0 || len(p.ExtIDs) == 0 {
		return jrpc.ErrorInvalidParams(`required: "raw" or "content" and "extids"`)
	}
	p.entry = factom.Entry{
		ExtIDs:    p.ExtIDs,
		Content:   p.Content,
		Timestamp: time.Now(),
		ChainID:   p.ChainID,
	}
	if _, err := p.entry.Cost(); err != nil {
		return jrpc.ErrorInvalidParams(err)
	}

	return nil
}

func (p ParamsSendTransaction) Entry() factom.Entry {
	return p.entry
}

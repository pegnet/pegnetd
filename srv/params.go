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

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/factom"
)

type Params interface {
	IsValid() error
	ValidChainID() *factom.Bytes32
	HasIncludePending() bool
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
	return jrpc.InvalidParams(`required: "chainid"`)
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
		return jrpc.InvalidParams(`required: "entryhash"`)
	}
	return nil
}

type ParamsGetPegnetBalances struct {
	Address *factom.FAAddress `json:"address,omitempty"`
}

func (p ParamsGetPegnetBalances) HasIncludePending() bool { return false }

func (p ParamsGetPegnetBalances) IsValid() error {
	if p.Address == nil {
		return jrpc.InvalidParams(`required: "address"`)
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
			return jrpc.InvalidParams(
				`"raw cannot be used with "content" or "extids"`)
		}
		if err := p.entry.UnmarshalBinary(p.Raw); err != nil {
			return jrpc.InvalidParams(err)
		}
		p.entry.Timestamp = time.Now()
		p.ChainID = p.entry.ChainID
		return nil
	}
	if err := p.ParamsToken.IsValid(); err != nil {
		return err
	}
	if len(p.Content) == 0 || len(p.ExtIDs) == 0 {
		return jrpc.InvalidParams(`required: "raw" or "content" and "extids"`)
	}
	p.entry = factom.Entry{
		ExtIDs:    p.ExtIDs,
		Content:   p.Content,
		Timestamp: time.Now(),
		ChainID:   p.ChainID,
	}
	if _, err := p.entry.Cost(false); err != nil {
		return jrpc.InvalidParams(err)
	}

	return nil
}

func (p ParamsSendTransaction) Entry() factom.Entry {
	return p.entry
}

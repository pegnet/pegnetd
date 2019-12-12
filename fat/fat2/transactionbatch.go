package fat2

import (
	"encoding/json"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat103"
	"github.com/Factom-Asset-Tokens/factom/jsonlen"
)

// TransactionBatch represents a fat2 entry, which can be a list of one or more
// transactions to be executed in order
type TransactionBatch struct {
	Version      uint          `json:"version"`
	Transactions []Transaction `json:"transactions"`

	Metadata json.RawMessage `json:"metadata,omitempty"`
	Entry    factom.Entry    `json:"-"`
}

// NewTransactionBatch returns a TransactionBatch initialized with the given
// entry.
// The height is for validation purposes. For certain heights, only
// certain rcd types are valid. If the height -1 is passed in, then
// all rcd types are valid. This is helpful when making a tx.
func NewTransactionBatch(entry factom.Entry, height int32) (*TransactionBatch, error) {
	t := TransactionBatch{Entry: entry}
	if err := t.UnmarshalJSON(entry.Content); err != nil {
		return nil, err
	}

	if err := t.Validate(height); err != nil {
		return nil, err
	}

	return &t, nil
}

type transactionBatch TransactionBatch

// UnmarshalJSON unmarshals the bytes of JSON into a TransactionBatch
// ensuring that there are no duplicate JSON keys.
func (t *TransactionBatch) UnmarshalJSON(data []byte) error {
	data = jsonlen.Compact(data)
	tRaw := struct {
		Version      json.RawMessage `json:"version"`
		Transactions json.RawMessage `json:"transactions"`
		Metadata     json.RawMessage `json:"metadata,omitempty"`
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf(
			"%T: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Version, &t.Version); err != nil {
		return fmt.Errorf("%T.Version: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Transactions, &t.Transactions); err != nil {
		return fmt.Errorf("%T.Transactions: %v", t, err)
	}
	t.Metadata = tRaw.Metadata

	expectedJSONLen := len(`{"version":,"transactions":}`) +
		len(tRaw.Version) + len(tRaw.Transactions)
	if expectedJSONLen != len(data) {
		return fmt.Errorf("%T: unexpected JSON length", t)
	}
	return nil
}

// MarshalJSON marshals the TransactionBatch content field as JSON, but will
// raise an error if the batch fails the checks in ValidData()
func (t TransactionBatch) MarshalJSON() ([]byte, error) {
	if err := t.ValidData(); err != nil {
		return nil, err
	}
	return json.Marshal(transactionBatch(t))
}

func (t TransactionBatch) String() string {
	data, err := t.MarshalJSON()
	if err != nil {
		return err.Error()
	}
	return string(data)
}

//// UnmarshalEntry unmarshals the Entry content as a TransactionBatch
//func (t *TransactionBatch) UnmarshalEntry() error {
//	return t.Entry.UnmarshalEntry(t)
//}
//
//// MarshalEntry marshals the TransactionBatch into the entry content
//func (t *TransactionBatch) MarshalEntry() error {
//	return t.Entry.MarshalEntry(t)
//}

func (t TransactionBatch) Sign(signingSet ...factom.RCDSigner) (factom.Entry, error) {
	e := t.Entry
	content, err := json.Marshal(t)
	if err != nil {
		return e, err
	}
	e.Content = content
	signed := fat103.Sign(e, signingSet...)
	t.Entry = signed
	return t.Entry, nil
}

// Validate performs all validation checks and returns nil if it is a valid
// batch. This function assumes the struct's entry field is populated.
// Validate requires a height for rcd signature validation.
// Not all rcd types are valid for all heights
func (t *TransactionBatch) Validate(height int32) error {
	err := t.ValidData()
	if err != nil {
		return err
	}
	if err = t.ValidExtIDs(height); err != nil {
		return err
	}
	return nil
}

// ValidData validates all Transaction data included in the batch and returns
// nil if it is valid. This function assumes that the entry content (or an
// independent JSON object) has been unmarshaled.
func (t *TransactionBatch) ValidData() error {
	if t.Version != 1 {
		return fmt.Errorf("invalid version")
	}
	if len(t.Transactions) == 0 {
		return fmt.Errorf("at least one output required")
	}
	uniqueInputs := make(map[factom.FAAddress]struct{})
	for i, tx := range t.Transactions {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction at index %d: %v", i, err)
		}
		uniqueInputs[tx.Input.Address] = struct{}{}
	}
	// There can only be one input address for the batch. This is to make
	// potential sharding of the pegnet easier (from Paul Snow and Clay Douglass).
	if len(uniqueInputs) != 1 {
		return fmt.Errorf("only one input address allowed")
	}
	return nil
}

// ValidExtIDs validates the structure of the external IDs of the entry to make
// sure that it has the correct number of RCD/signature pairs. If no errors are
// found, it will then validate the content of the RCD/signature pair. This
// function assumes that the entry content has been unmarshaled and that
// ValidData returns nil.
func (t TransactionBatch) ValidExtIDs(height int32) error {
	// Count unique inputs to know how many signatures are needed on the entry

	uniqueInputs := make(map[factom.Bytes32]struct{})
	for _, tx := range t.Transactions {
		uniqueInputs[factom.Bytes32(tx.Input.Address)] = struct{}{}
	}

	flag := factom.R_RCD1
	if height > int32(Fat2RCDEActivation) {
		flag = flag | factom.R_RCDe
	}
	// < 0 means accept all rcd types
	if height < 0 {
		flag = flag | factom.R_ALL
	}

	if err := fat103.Validate(t.Entry, uniqueInputs, flag); err != nil {
		return err
	}

	return nil
}

// HasConversions returns true if this batch contains at least one transaction
// with a conversion input/output pair. This function assumes that
// TransactionBatch.Valid() returns nil
func (t *TransactionBatch) HasConversions() bool {
	for _, tx := range t.Transactions {
		if tx.IsConversion() {
			return true
		}
	}
	return false
}

// HasPEGRequest returns if the tx batch has a conversion request into PEG
func (t *TransactionBatch) HasPEGRequest() bool {
	for _, tx := range t.Transactions {
		if tx.IsPEGRequest() {
			return true
		}
	}
	return false
}

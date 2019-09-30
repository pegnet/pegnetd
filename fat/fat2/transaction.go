package fat2

import (
	"encoding/json"
	"fmt"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/jsonlen"
)

// TypedAddressAmountTuple represents a tuple of a Factoid address sending
// or receiving an Amount of funds of a type that is inferred based on
// outside context
type AddressAmountTuple struct {
	Address factom.FAAddress `json:"address"`
	Amount  uint64           `json:"amount"`
}

// UnmarshalJSON unmarshals the bytes of JSON into a AddressAmountTuple
// ensuring that there are no duplicate JSON keys.
func (t *AddressAmountTuple) UnmarshalJSON(data []byte) error {
	data = jsonlen.Compact(data)
	tRaw := struct {
		Address json.RawMessage `json:"address"`
		Amount  json.RawMessage `json:"amount"`
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Address, &t.Address); err != nil {
		return fmt.Errorf("%T.Address: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Amount, &t.Amount); err != nil {
		return fmt.Errorf("%T.Amount: %v", t, err)
	}

	expectedJSONLen := len(`{"address":,"amount":}`) +
		len(tRaw.Address) + len(tRaw.Amount)
	if expectedJSONLen != len(data) {
		return fmt.Errorf("%T: unexpected JSON length", t)
	}
	return nil
}

// TypedAddressAmountTuple represents a 3-tuple of a Factoid address sending
// or receiving an Amount of funds of a given Type
type TypedAddressAmountTuple struct {
	Address factom.FAAddress `json:"address"`
	Amount  uint64           `json:"amount"`
	Type    PTicker          `json:"type"`
}

// UnmarshalJSON unmarshals the bytes of JSON into a TypedAddressAmountTuple
// ensuring that there are no duplicate JSON keys.
func (t *TypedAddressAmountTuple) UnmarshalJSON(data []byte) error {
	data = jsonlen.Compact(data)
	tRaw := struct {
		Address json.RawMessage `json:"address"`
		Amount  json.RawMessage `json:"amount"`
		Type    json.RawMessage `json:"type"`
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Address, &t.Address); err != nil {
		return fmt.Errorf("%T.Address: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Amount, &t.Amount); err != nil {
		return fmt.Errorf("%T.Amount: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Type, &t.Type); err != nil {
		return fmt.Errorf("%T.Type: %v", t, err)
	}

	expectedJSONLen := len(`{"address":,"amount":,"type":}`) +
		len(tRaw.Address) + len(tRaw.Amount) + len(tRaw.Type)
	if expectedJSONLen != len(data) {
		return fmt.Errorf("%T: unexpected JSON length", t)
	}
	return nil
}

// Transaction represents a fat2 transaction, which can be a value transfer
// or a conversion depending on present fields
type Transaction struct {
	Input      TypedAddressAmountTuple `json:"input"`
	Transfers  []AddressAmountTuple    `json:"transfers,omitempty"`
	Conversion PTicker                 `json:"conversion,omitempty"`
	Metadata   interface{}             `json:"metadata,omitempty"`
}

// UnmarshalJSON unmarshals the bytes of JSON into a Transaction
// ensuring that there are no duplicate JSON keys.
func (t *Transaction) UnmarshalJSON(data []byte) error {
	data = jsonlen.Compact(data)
	tRaw := struct {
		Input      json.RawMessage `json:"input"`
		Transfers  json.RawMessage `json:"transfers,omitempty"`
		Conversion json.RawMessage `json:"conversion,omitempty"`
		Metadata   json.RawMessage `json:"metadata,omitempty"`
	}{}
	if err := json.Unmarshal(data, &tRaw); err != nil {
		return fmt.Errorf("%T: %v", t, err)
	}
	if err := json.Unmarshal(tRaw.Input, &t.Input); err != nil {
		return fmt.Errorf("%T.Input: %v", t, err)
	}
	if 0 < len(tRaw.Transfers) {
		if err := json.Unmarshal(tRaw.Transfers, &t.Transfers); err != nil {
			return fmt.Errorf("%T.Transfers: %v", t, err)
		}
	}
	if 0 < len(tRaw.Conversion) {
		if err := json.Unmarshal(tRaw.Conversion, &t.Conversion); err != nil {
			return fmt.Errorf("%T.Conversion: %v", t, err)
		}
	}
	t.Metadata = tRaw.Metadata

	var expectedJSONLen int
	if tRaw.Metadata != nil {
		expectedJSONLen += len(`,"metadata":`) + len(tRaw.Metadata)
	}
	if t.IsConversion() {
		expectedJSONLen += len(`{"input":,"conversion":}`) +
			len(tRaw.Input) + len(tRaw.Conversion)
	} else {
		expectedJSONLen += len(`{"input":,"transfers":}`) +
			len(tRaw.Input) + len(tRaw.Transfers)
	}
	if expectedJSONLen != len(data) {
		return fmt.Errorf("%T: unexpected JSON length", t)
	}
	return nil
}

// Validate performs all validation checks and returns nil if t is a valid
// Transaction
func (t *Transaction) Validate() error {
	if t.Input.Address == fat.Coinbase() {
		return fmt.Errorf("invalid input: %v reserved as burn address", t.Input.Address)
	} else if t.Input == (TypedAddressAmountTuple{}) { // TODO: is there a better way to check for zero value struct?
		return fmt.Errorf("invalid input: empty")
	}
	if len(t.Transfers) == 0 && t.Conversion == PTickerInvalid {
		return fmt.Errorf("at least one transfer or exactly one conversion type required")
	} else if 0 < len(t.Transfers) && PTickerInvalid < t.Conversion {
		return fmt.Errorf("transfers and conversion type must be mutually exclusive")
	}

	remainingInputAmount := t.Input.Amount
	for _, transfer := range t.Transfers {
		if remainingInputAmount < transfer.Amount {
			return fmt.Errorf("insufficient input to cover outputs")
		}
		remainingInputAmount -= transfer.Amount
	}
	return nil
}

// IsConversion returns true if this transaction has zero transfers and a
// valid conversion PTicker
func (t *Transaction) IsConversion() bool {
	return len(t.Transfers) == 0 && PTickerInvalid < t.Conversion && t.Conversion < PTickerMax
}

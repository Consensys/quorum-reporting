package types

import (
	"errors"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ParsedTransaction struct {
	Sig            string                 `json:"txSig"`
	Func4Bytes     hexutil.Bytes          `json:"func4Bytes"`
	ParsedData     map[string]interface{} `json:"parsedData"`
	ParsedEvents   []*ParsedEvent         `json:"parsedEvents"`
	RawTransaction *Transaction           `json:"rawTransaction"`
}

func (ptx *ParsedTransaction) ParseTransaction(rawABI string) error {
	parsedABI, _ := abi.JSON(strings.NewReader(rawABI))
	if ptx.RawTransaction == nil {
		return errors.New("transaction is nil or invalid")
	}
	log.Printf("Parse transaction %v.\n", ptx.RawTransaction.Hash.Hex())
	// set defaults
	var data []byte
	if len(ptx.RawTransaction.PrivateData) > 0 {
		data = ptx.RawTransaction.PrivateData
	} else {
		data = ptx.RawTransaction.Data
	}
	ptx.ParsedData = map[string]interface{}{}
	// parse transaction data
	if ptx.RawTransaction.To != (common.Address{0}) {
		ptx.Func4Bytes = data[:4]
		// check against all abi methods
		for _, method := range parsedABI.Methods {
			if string(method.ID()) == string(ptx.Func4Bytes) {
				ptx.Sig = method.Sig()
				method.Inputs.UnpackIntoMap(ptx.ParsedData, data[4:])
				break
			}
		}
	} else {
		// contract deployment transaction
		ptx.Sig = "constructor" + parsedABI.Constructor.Sig()
		if len(data) > 32*parsedABI.Constructor.Inputs.LengthNonIndexed() {
			parsedABI.Constructor.Inputs.UnpackIntoMap(ptx.ParsedData, data[(len(data)-32*parsedABI.Constructor.Inputs.LengthNonIndexed()):])
			// TODO: Parsing inputs for complex data type in constructor is not supported unless the exact contract bin is provided.
		}
	}
	return nil
}

type ParsedEvent struct {
	Sig        string                 `json:"eventSig"`
	ParsedData map[string]interface{} `json:"parsedData"`
	RawEvent   *Event                 `json:"rawEvent"`
}

func (pe *ParsedEvent) ParseEvent(rawABI string) error {
	parsedABI, _ := abi.JSON(strings.NewReader(rawABI))
	if pe.RawEvent == nil || pe.RawEvent.Topics == nil {
		return errors.New("event is nil or invalid")
	}
	log.Printf("Parse event %v.\n", pe.RawEvent.Topics[0].Hex())
	if eventABI, err := parsedABI.EventByID(pe.RawEvent.Topics[0]); err == nil {
		pe.Sig = eventABI.String()
		pe.ParsedData = map[string]interface{}{}
		eventABI.Inputs.UnpackIntoMap(pe.ParsedData, pe.RawEvent.Data)
	}
	return nil
}

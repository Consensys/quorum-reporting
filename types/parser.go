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
		dataHex := hexutil.Encode(data)
		if index := strings.Index(dataHex, "a165627a7a72305820"); index > 0 {
			// search for pattern a165627a7a72305820 for solidity < 0.5.10
			// <bytecode> + "a165627a7a72305820" + <256 bits whisperHash> + "0029"
			index = (index - 2 + 18 + 64 + 4) / 2 // remove 0x, find hex position 18+64+4 after
			parsedABI.Constructor.Inputs.UnpackIntoMap(ptx.ParsedData, data[index:])
		} else if index := strings.LastIndex(dataHex, "64736f6c6343"); index > 0 {
			// search for pattern 64736f6c6343 for solidity >= 0.5.10,
			// <bytecode> + "a265627a7a72305820" + <256 bits whisperHash> + "64736f6c6343" + compiler_version(e.g. 000608) + "0033"
			index = (index - 2 + 12 + 6 + 4) / 2 // remove 0x, find hex position 12+6+4 after
			parsedABI.Constructor.Inputs.UnpackIntoMap(ptx.ParsedData, data[index:])
		} else {
			ptx.ParsedData["error"] = "unable to parse params"
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

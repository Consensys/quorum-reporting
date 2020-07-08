package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"quorumengineering/quorum-report/log"
)

type ParsedTransaction struct {
	Sig            string                 `json:"txSig"`
	Func4Bytes     hexutil.Bytes          `json:"func4Bytes"`
	ParsedData     map[string]interface{} `json:"parsedData"`
	ParsedEvents   []*ParsedEvent         `json:"parsedEvents"`
	RawTransaction *Transaction           `json:"rawTransaction"`
}

func (ptx *ParsedTransaction) ParseTransaction(rawABI string) error {
	if ptx.RawTransaction == nil {
		return errors.New("transaction is nil or invalid")
	}

	var structure ABIStructure
	if err := json.Unmarshal([]byte(rawABI), &structure); err != nil {
		log.Error("Could not unmarshal ABI", "abi", rawABI)
		return errors.New("could not unmarshal ABI")
	}
	internalAbi := structure.To()

	log.Debug("Parse transaction", "tx", ptx.RawTransaction.Hash.Hex())

	// set defaults
	var data []byte
	if len(ptx.RawTransaction.PrivateData) > 0 {
		data = ptx.RawTransaction.PrivateData
	} else {
		data = ptx.RawTransaction.Data
	}
	ptx.ParsedData = map[string]interface{}{}
	// parse transaction data
	if ptx.RawTransaction.To != (common.Address{}) {
		ptx.Func4Bytes = data[:4]
		// check against all abi methods
		for _, method := range internalAbi.Functions {
			if method.Signature() == hex.EncodeToString(ptx.Func4Bytes) {
				ptx.Sig = method.String()
				ptx.ParsedData = method.Parse(data[4:])
			}
		}
		return nil
	}

	// contract deployment transaction
	ptx.Sig = "constructor" + internalAbi.Constructor.String()
	dataHex := hexutil.Encode(data)
	if index := strings.Index(dataHex, "a165627a7a72305820"); index > 0 {
		// search for pattern a165627a7a72305820 for solidity < 0.5.10
		// <bytecode> + "a165627a7a72305820" + <256 bits whisperHash> + "0029"
		index = (index - 2 + 18 + 64 + 4) / 2 // remove 0x, find hex position 18+64+4 after
		ptx.ParsedData = internalAbi.Constructor.Parse(data[index:])
	} else if index := strings.LastIndex(dataHex, "64736f6c6343"); index > 0 {
		// search for pattern 64736f6c6343 for solidity >= 0.5.10,
		// <bytecode> + "a265627a7a72305820" + <256 bits whisperHash> + "64736f6c6343" + compiler_version(e.g. 000608) + "0033"
		index = (index - 2 + 12 + 6 + 4) / 2 // remove 0x, find hex position 12+6+4 after
		ptx.ParsedData = internalAbi.Constructor.Parse(data[index:])
	} else {
		ptx.ParsedData["error"] = "unable to parse params"
	}
	return nil
}

type ParsedEvent struct {
	Sig        string                 `json:"eventSig"`
	ParsedData map[string]interface{} `json:"parsedData"`
	RawEvent   *Event                 `json:"rawEvent"`
}

func (pe *ParsedEvent) ParseEvent(rawABI string) error {
	if pe.RawEvent == nil || len(pe.RawEvent.Topics) == 0 {
		return errors.New("event is nil or invalid")
	}

	var structure ABIStructure
	if err := json.Unmarshal([]byte(rawABI), &structure); err != nil {
		log.Error("Could not unmarshal ABI", "abi", rawABI)
		return errors.New("could not unmarshal ABI")
	}
	internalAbi := structure.To()

	log.Debug("Parse event", "event", pe.RawEvent.Topics[0].Hex())
	for _, ev := range internalAbi.Events {
		if ev.Signature() == pe.RawEvent.Topics[0].String() {
			pe.Sig = "event " + ev.String()
			pe.ParsedData = ev.Parse(pe.RawEvent.Data)
		}
	}
	return nil
}

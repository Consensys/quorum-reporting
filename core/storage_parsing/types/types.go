package types

import (
	"encoding/json"
	"errors"
	"strconv"
)

type SolidityStorageEntries []SolidityStorageEntry

type SolidityStorageDocument struct {
	Storage SolidityStorageEntries       `json:"storage"`
	Types   map[string]SolidityTypeEntry `json:"types"`
}

type SolidityStorageEntry struct {
	Label  string `json:"label"`
	Offset uint64 `json:"offset"`
	Slot   uint64 `json:"slot"`
	Type   string `json:"type"`
}

func (sse SolidityStorageEntry) UnmarshalJSON(b []byte) error {
	var simple map[string]string
	err := json.Unmarshal(b, &simple)
	if err != nil {
		return err
	}
	sse.Label = simple["label"]
	sse.Type = simple["type"]

	if simple["offset"] == "" {
		return errors.New("offset not set")
	}
	offsetAsUint64, err := strconv.ParseUint(simple["offset"], 10, 0)
	if err != nil {
		return err
	}
	sse.Offset = offsetAsUint64

	if simple["slot"] == "" {
		return errors.New("slot not set")
	}

	slotAsUint64, err := strconv.ParseUint(simple["slot"], 10, 0)
	if err != nil {
		return err
	}
	sse.Slot = slotAsUint64

	return nil
}

type SolidityTypeEntry struct {
	Encoding      string                 `json:"encoding"`
	Key           string                 `json:"key"`
	Label         string                 `json:"label"`
	NumberOfBytes string                 `json:"numberOfBytes"`
	Value         string                 `json:"value"`
	Base          string                 `json:"base"`
	Members       SolidityStorageEntries `json:"members"`
}

func (sse SolidityStorageEntries) Len() int {
	return len(sse)
}

func (sse SolidityStorageEntries) Less(i, j int) bool {
	return (sse[i].Slot < sse[j].Slot) || (sse[i].Offset < sse[j].Offset)
}

func (sse SolidityStorageEntries) Swap(i, j int) {
	sse[i], sse[j] = sse[j], sse[i]
}

type StorageItem struct {
	VarName  string      `json:"name"`
	VarIndex uint64      `json:"index"`
	VarType  string      `json:"type"`
	Value    interface{} `json:"value,omitempty"`
	// for map only
	Values map[string]interface{} `json:"values,omitempty"`
}

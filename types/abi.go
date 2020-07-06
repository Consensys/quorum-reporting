package types

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/sha3"
	"strings"
)

// JSON formatted type

type ABIStructure []ABIStructureEntry

type ABIStructureEntry struct {
	Type            string                 `json:"type"` //TODO: should default to "function"
	Name            string                 `json:"name"` //TODO: if type is "constructor", then name should be blank
	Inputs          []ABIStructureArgument `json:"inputs"`
	Outputs         []ABIStructureArgument `json:"outputs"`
	StateMutability string                 `json:"stateMutability"`
	Anonymous       bool                   `json:"anonymous"`
}

type ABIStructureArgument struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Components []ABIStructureArgument `json:"components"`
	Indexed    bool                   `json:"indexed"`
}

func (abi ABIStructure) To() *ContractABI {
	contractAbi := new(ContractABI)

	for _, entry := range abi {
		switch entry.Type {
		case "constructor", "receive", "fallback":
			//Nothing to do as they have no impact for the ABI
			//Only here to show the have been intentionally left out
		case "", "function":
			var inputs []ContractABIFunctionArgument
			for _, input := range entry.Inputs {
				inputs = append(inputs, input.ToFunctionArgument())
			}
			var outputs []ContractABIFunctionArgument
			for _, output := range entry.Outputs {
				outputs = append(outputs, output.ToFunctionArgument())
			}

			stateMutability := entry.StateMutability
			if stateMutability == "" {
				stateMutability = "nonpayable"
			}

			functionDefinition := ContractABIFunction{
				Type:            "function",
				Name:            entry.Name,
				Inputs:          inputs,
				Outputs:         outputs,
				StateMutability: stateMutability,
			}
			contractAbi.Functions = append(contractAbi.Functions, functionDefinition)
		case "event":
			var inputs []ContractABIEventArgument
			for _, input := range entry.Inputs {
				inputs = append(inputs, input.ToEventArgument())
			}
			eventDefinition := ContractABIEvent{
				Type:      "event",
				Name:      entry.Name,
				Inputs:    inputs,
				Anonymous: entry.Anonymous,
			}
			contractAbi.Events = append(contractAbi.Events, eventDefinition)
		}
	}

	return contractAbi
}

func (entry ABIStructureArgument) ToFunctionArgument() ContractABIFunctionArgument {
	var components []ContractABIFunctionArgument
	for _, component := range entry.Components {
		components = append(components, component.ToFunctionArgument())
	}

	return ContractABIFunctionArgument{
		Name:       entry.Name,
		Type:       entry.Type,
		Components: components,
	}
}

func (entry ABIStructureArgument) ToEventArgument() ContractABIEventArgument {
	var components []ContractABIEventArgument
	for _, component := range entry.Components {
		components = append(components, component.ToEventArgument())
	}

	return ContractABIEventArgument{
		Name:       entry.Name,
		Type:       entry.Type,
		Components: components,
		Indexed:    entry.Indexed,
	}
}

// Internal representation

type ContractABI struct {
	Functions []ContractABIFunction
	Events    []ContractABIEvent
}

type ContractABIFunction struct {
	Type            string
	Name            string
	Inputs          []ContractABIFunctionArgument
	Outputs         []ContractABIFunctionArgument
	StateMutability string
}

func (function ContractABIFunction) String() string {
	var inputSigs []string
	for _, input := range function.Inputs {
		inputSigs = append(inputSigs, input.String())
	}
	return fmt.Sprintf("%s(%s)", function.Name, strings.Join(inputSigs, ","))
}

func (function ContractABIFunction) Signature() string {
	definition := function.String()

	d := sha3.NewLegacyKeccak256()
	d.Write([]byte(definition))
	d.Sum(nil)

	hsh := d.Sum(nil)[:4]
	return hex.EncodeToString(hsh)
}

type ContractABIFunctionArgument struct {
	Name       string
	Type       string
	Components []ContractABIFunctionArgument
}

func (arg ContractABIFunctionArgument) String() string {
	if arg.Type == "tuple" || arg.Type == "tuple[]" {
		arraySuffix := ""
		if arg.Type == "tuple[]" {
			arraySuffix = "[]"
		}
		var inputSigs []string
		for _, input := range arg.Components {
			inputSigs = append(inputSigs, input.String())
		}
		return fmt.Sprintf("(%s)%s", strings.Join(inputSigs, ","), arraySuffix)
	}
	return arg.Type
}

type ContractABIEvent struct {
	Type      string
	Name      string
	Inputs    []ContractABIEventArgument
	Anonymous bool
}

func (event ContractABIEvent) String() string {
	var inputSigs []string
	for _, input := range event.Inputs {
		inputSigs = append(inputSigs, input.String())
	}
	return fmt.Sprintf("%s(%s)", event.Name, strings.Join(inputSigs, ","))
}

func (event ContractABIEvent) Signature() string {
	definition := event.String()

	d := sha3.NewLegacyKeccak256()
	d.Write([]byte(definition))
	d.Sum(nil)

	hsh := d.Sum(nil)
	return hex.EncodeToString(hsh)
}

func (event ContractABIEvent) ParseData(data []byte) map[string]interface{} {

}

type ContractABIEventArgument struct {
	Name       string
	Type       string
	Components []ContractABIEventArgument
	Indexed    bool
}

func (arg ContractABIEventArgument) String() string {
	if arg.Type == "tuple" || arg.Type == "tuple[]" {
		arraySuffix := ""
		if arg.Type == "tuple[]" {
			arraySuffix = "[]"
		}
		var inputSigs []string
		for _, input := range arg.Components {
			inputSigs = append(inputSigs, input.String())
		}
		return fmt.Sprintf("(%s)%s", strings.Join(inputSigs, ","), arraySuffix)
	}
	return arg.Type
}

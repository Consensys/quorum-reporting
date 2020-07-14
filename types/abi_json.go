package types

import "encoding/json"

// Defined according to https://solidity.readthedocs.io/en/develop/abi-spec.html#json

type ABIStructure []ABIStructureEntry

func NewABIStructureFromJSON(abi string) (ABIStructure, error) {
	var structure ABIStructure
	err := json.Unmarshal([]byte(abi), &structure)
	return structure, err
}

func (abi ABIStructure) ToInternalABI() *ContractABI {
	contractAbi := new(ContractABI)

	for _, entry := range abi {
		switch entry.Type {
		case "receive", "fallback":
		//Nothing to do as they have no impact for the ABI
		//Only here to show the have been intentionally left out
		case "constructor":
			contractAbi.Constructor = entry.AsConstructor()
		case "", "function":
			contractAbi.Functions = append(contractAbi.Functions, entry.AsFunction())
		case "event":
			contractAbi.Events = append(contractAbi.Events, entry.AsEvent())
		}
	}

	return contractAbi
}

type ABIStructureEntry struct {
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Inputs    []ABIStructureArgument `json:"inputs"`
	Outputs   []ABIStructureArgument `json:"outputs"`
	Anonymous bool                   `json:"anonymous"`
}

func (entry ABIStructureEntry) AsConstructor() ContractABIFunction {
	return ContractABIFunction{"constructor", "", entry.AsFunction().Inputs, nil}
}

func (entry ABIStructureEntry) AsFunction() ContractABIFunction {
	var inputs []ContractABIArgument
	for _, input := range entry.Inputs {
		inputs = append(inputs, input.AsArgument())
	}
	var outputs []ContractABIArgument
	for _, output := range entry.Outputs {
		outputs = append(outputs, output.AsArgument())
	}
	return ContractABIFunction{"function", entry.Name, inputs, outputs}
}

func (entry ABIStructureEntry) AsEvent() ContractABIEvent {
	var inputs []ContractABIEventArgument
	for _, input := range entry.Inputs {
		inputs = append(inputs, input.AsIndexedArgument())
	}
	return ContractABIEvent{"event", entry.Name, inputs, entry.Anonymous}
}

type ABIStructureArgument struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Components []ABIStructureArgument `json:"components"`
	Indexed    bool                   `json:"indexed"`
}

func (arg ABIStructureArgument) AsArgument() ContractABIArgument {
	var components []ContractABIArgument
	for _, component := range arg.Components {
		components = append(components, component.AsArgument())
	}
	return ContractABIArgument{arg.Name, arg.Type, components}
}

func (arg ABIStructureArgument) AsIndexedArgument() ContractABIEventArgument {
	return ContractABIEventArgument{arg.AsArgument(), arg.Indexed}
}

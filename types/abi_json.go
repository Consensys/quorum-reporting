package types

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
		case "receive", "fallback":
		//Nothing to do as they have no impact for the ABI
		//Only here to show the have been intentionally left out
		case "constructor":
			var inputs []ContractABIArgument
			for _, input := range entry.Inputs {
				inputs = append(inputs, input.ToFunctionArgument())
			}
			contractAbi.Constructor = ContractABIFunction{
				Type:    "constructor",
				Name:    "",
				Inputs:  inputs,
				Outputs: nil,
			}
		case "", "function":
			var inputs []ContractABIArgument
			for _, input := range entry.Inputs {
				inputs = append(inputs, input.ToFunctionArgument())
			}
			var outputs []ContractABIArgument
			for _, output := range entry.Outputs {
				outputs = append(outputs, output.ToFunctionArgument())
			}

			stateMutability := entry.StateMutability
			if stateMutability == "" {
				stateMutability = "nonpayable"
			}

			functionDefinition := ContractABIFunction{
				Type:    "function",
				Name:    entry.Name,
				Inputs:  inputs,
				Outputs: outputs,
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

func (entry ABIStructureArgument) ToFunctionArgument() ContractABIArgument {
	var components []ContractABIArgument
	for _, component := range entry.Components {
		components = append(components, component.ToFunctionArgument())
	}

	return ContractABIArgument{
		Name:       entry.Name,
		Type:       entry.Type,
		Components: components,
	}
}

func (entry ABIStructureArgument) ToEventArgument() ContractABIEventArgument {
	var components []ContractABIArgument
	for _, component := range entry.Components {
		components = append(components, component.ToFunctionArgument())
	}

	return ContractABIEventArgument{
		ContractABIArgument: ContractABIArgument{
			Name:       entry.Name,
			Type:       entry.Type,
			Components: components,
		},
		Indexed: entry.Indexed,
	}
}

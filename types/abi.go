package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type ContractABI struct {
	Constructor ContractABIFunction
	Functions   []ContractABIFunction
	Events      []ContractABIEvent
}

type ContractABIFunction struct {
	Type    string
	Name    string
	Inputs  []ContractABIArgument
	Outputs []ContractABIArgument
}

func (function ContractABIFunction) String() string {
	var inputSigs []string
	for _, input := range function.Inputs {
		inputSigs = append(inputSigs, input.String())
	}
	return fmt.Sprintf("%s(%s)", function.Name, strings.Join(inputSigs, ","))
}

func (function ContractABIFunction) StringNoName() string {
	var inputSigs []string
	for _, input := range function.Inputs {
		inputSigs = append(inputSigs, input.StringNoName())
	}
	return fmt.Sprintf("%s(%s)", function.Name, strings.Join(inputSigs, ","))
}

func (function ContractABIFunction) Signature() string {
	return hex.EncodeToString(hash(function.StringNoName())[:4])
}

func (function ContractABIFunction) Parse(data []byte) (map[string]interface{}, error) {
	return ParseAllData(function.Inputs, data)
}

type ContractABIArgument struct {
	Name       string
	Type       string
	Components []ContractABIArgument
}

func (arg ContractABIArgument) String() string {
	if strings.HasPrefix(arg.Type, "tuple") {
		var componentSigs []string
		for _, input := range arg.Components {
			componentSigs = append(componentSigs, input.String())
		}
		return fmt.Sprintf("(%s)%s %s", strings.Join(componentSigs, ","), arg.Type[5:], arg.Name)
	}
	return fmt.Sprintf("%s %s", arg.Type, arg.Name)
}

func (arg ContractABIArgument) StringNoName() string {
	if strings.HasPrefix(arg.Type, "tuple") {
		var componentSigs []string
		for _, input := range arg.Components {
			componentSigs = append(componentSigs, input.StringNoName())
		}
		return fmt.Sprintf("(%s)%s", strings.Join(componentSigs, ","), arg.Type[5:])
	}
	return arg.Type
}

/*
+-----------------+--------+---------+
|      Type       | Static | Dynamic |
+-----------------+--------+---------+
| uint<x>         | ✔      |         |
| int<x>          | ✔      |         |
| address         | ✔      |         |
| bool            | ✔      |         |
| bytes<x>        | ✔      |         |
| bytes           |        | ✔       |
| string          |        | ✔       |
| T[]             |        | ✔       |
| T<static>[m]    | ✔      |         |
| T<dynamic>[m]   |        | ✔       |
| Tuple<static>   | ✔      |         |
| Tuple<dynamic>  |        | ✔       |
+-----------------+--------+---------+

Note: a tuple is dynamic if at least one element is dynamic

Note: fixed sized arrays are not explicitly handled because they only
		depend on their type to determine if they're static or not
*/
func (arg ContractABIArgument) IsDynamic() bool {
	if strings.HasSuffix(arg.Type, "[]") {
		return true
	}

	if arg.Type == "bytes" || arg.Type == "string" {
		return true
	}
	// the case for fixed size array (dynamic array would be handled above)
	if arg.Type == "bytes[" || arg.Type == "string[" {
		return true
	}

	if strings.HasPrefix(arg.Type, "uint") {
		return false
	}
	if strings.HasPrefix(arg.Type, "int") {
		return false
	}
	if strings.HasPrefix(arg.Type, "bytes") {
		return false
	}
	if strings.HasPrefix(arg.Type, "address") {
		return false
	}
	if strings.HasPrefix(arg.Type, "bool") {
		return false
	}

	if strings.HasPrefix(arg.Type, "tuple") {
		for _, comp := range arg.Components {
			if comp.IsDynamic() {
				return true
			}
		}
	}
	return false
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

func (event ContractABIEvent) StringNoName() string {
	var inputSigs []string
	for _, input := range event.Inputs {
		inputSigs = append(inputSigs, input.StringNoName())
	}
	return fmt.Sprintf("%s(%s)", event.Name, strings.Join(inputSigs, ","))
}

func (event ContractABIEvent) Signature() string {
	return hex.EncodeToString(hash(event.StringNoName()))
}

func (event ContractABIEvent) Parse(data []byte) (map[string]interface{}, error) {
	var args []ContractABIArgument
	for _, arg := range event.Inputs {
		if !arg.Indexed {
			args = append(args, ContractABIArgument{arg.Name, arg.Type, arg.Components})
		}
	}
	return ParseAllData(args, data)
}

type ContractABIEventArgument struct {
	ContractABIArgument
	Indexed bool
}

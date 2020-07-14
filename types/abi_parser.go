package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

type HeadItem struct {
	StartOffset uint64
	EndOffset   uint64
	Arg         ContractABIArgument
}

func ParseInt(bytes []byte) *big.Int {
	//2s complement, so negative is Most Significant Bit is set
	isPositive := bytes[0] < 128

	if isPositive {
		return ParseUint(bytes)
	}

	// negative, so invert all the bits, add 1 and flip the sign
	for i := 0; i < len(bytes); i++ {
		bytes[i] = ^bytes[i]
	}

	i := ParseUint(bytes)
	i.Add(i, new(big.Int).SetUint64(1))
	i.Neg(i)

	return i
}

func ParseUint(bytes []byte) *big.Int {
	return new(big.Int).SetBytes(bytes)
}

func hash(input string) []byte {
	d := sha3.NewLegacyKeccak256()
	d.Write([]byte(input))
	return d.Sum(nil)
}

/*
Rules for parsing can be found at: https://solidity.readthedocs.io/en/develop/abi-spec.html#argument-encoding

ParseAllData parses a set of elements from the ABI, assuming they fill the provided data

1) Check the "head" of the next element:
2) Check if the element is statically typed
2a) if statically typed, the "head" is the element itself, so it is parsed immediately, and the current pointer for where we are within the data
	array is updated. Note there is no "tail" for a static element.
2b) if it is dynamically typed, make a note of its starting position (data[pointer:pointer+32] interpreted as a uint) and update the last dynamic
	element with its ending position (which is the same value))
3) Once all the elements have been parsed, start processing the dynamic element tails, which may recursively call ParseData
4) Concatenate all the results into a map against their variable names and return

The "head" and "tail" concept of the encoding can be visualised as follows:

If we have elements: X1, X2, X3, X4, then the encoding of that will be:
Head(X1) + Head(X2) + Head(X3) + Head(X4) + Tail(X1) + Tail(X2) + Tail(X3) + Tail(X4)
- if Xn is static, then Tail(Xn) is omitted, and the encoding of the element is in Head(Xn)
- if Xn is dynamic then:
    - Head(Xn) is 32 bytes, representing a uint256 of the starting position of Tail(Xn) within the data array
    - Tail(Xn) is the actual encoding  of the element

Because we only get the starting position of dynamic elements, we have to infer the end position by taking the start
position of the next dynamic element, leading to 2b) that we update the last dynamic element with the end point (as we
now know this value)
*/
func ParseAllData(inputs []ContractABIArgument, data []byte) (map[string]interface{}, error) {
	currentOffset := uint64(0)
	tailOffsets := make([]*HeadItem, 0)
	allResults := make(map[string]interface{})

	//handle all the heads, then handle all the tails
	for _, input := range inputs {
		if input.IsDynamic() {
			elementStartOffsetBytes := data[currentOffset : currentOffset+32]
			elementStartOffset := ParseUint(elementStartOffsetBytes).Uint64()

			if len(tailOffsets) == 0 {
				tailOffsets = append(tailOffsets, &HeadItem{
					StartOffset: elementStartOffset,
					Arg:         input,
				})
			} else {
				maxIndex := len(tailOffsets) - 1
				tailOffsets[maxIndex].EndOffset = elementStartOffset
				tailOffsets = append(tailOffsets, &HeadItem{
					StartOffset: elementStartOffset,
					Arg:         input,
				})
			}
			//parse it later
			currentOffset += 32
			continue
		}

		result, newOffset, err := ParseStaticType(input, data, currentOffset)
		if err != nil {
			return nil, err
		}

		allResults[input.Name] = result
		currentOffset = newOffset
	}

	//set the end of the last dynamic element, if there are any
	if len(tailOffsets) != 0 {
		maxIndex := len(tailOffsets) - 1
		tailOffsets[maxIndex].EndOffset = uint64(len(data))
	}

	for _, headItem := range tailOffsets {
		dynamicTypeData := data[headItem.StartOffset:headItem.EndOffset]

		var err error
		allResults[headItem.Arg.Name], err = ParseDynamicType(headItem.Arg, dynamicTypeData)
		if err != nil {
			return nil, err
		}
	}

	return allResults, nil
}

//TODO: add bounds check with errors on data array

func ParseDynamicType(arg ContractABIArgument, data []byte) (interface{}, error) {
	//A dynamically sized array of either a static or dynamic type
	//Extract the array size from the first 32 bytes, and then treat it
	//as a fixed size array of the extracted size
	//This means we can cycle it through the parser again, stripping it of the first
	//32 bytes, where it will get picked up by the fixed-sized array parser
	if strings.HasSuffix(arg.Type, "[]") {
		//this is a dynamic array

		//read the number of elements from the first 32 bytes
		numberOfElements := ParseUint(data[:32]).Uint64()
		//cut of the ending [] and add [numberOfElements]
		typeName := fmt.Sprintf("%s[%d]", arg.Type[:len(arg.Type)-2], numberOfElements)

		///construct a new type that turns T[] -> T[numberOfElements]
		fixedSizeType := ContractABIArgument{
			Name:       arg.Name,
			Type:       typeName,
			Components: arg.Components,
		}

		arrayData := data[32:] // remove the size from the start of the data, as we've parsed that already
		return ParseDynamicType(fixedSizeType, arrayData)
	}

	//A fixed size array of a dynamic type
	//This is the same as listing all the elements individually and parsing, so do that
	//NOTE: this may actually be a static type is the "real" type was a dynamic array
	//		this doesn't really matter, since it will get interpreted correctly when
	//		doing a top-level parse again
	if strings.Contains(arg.Type, "[") && strings.Contains(arg.Type, "]") {
		start := strings.LastIndex(arg.Type, "[")
		end := strings.LastIndex(arg.Type, "]")

		numberOfElements, err := strconv.ParseUint(arg.Type[start+1:end], 10, 0)
		if err != nil {
			return nil, errors.New("error parsing fixed sized, dynamically typed array size: " + err.Error())
		}

		//ABI spec says we can treat a fixed size array as a tuple with same number of elements
		var arrayAsTuple []ContractABIArgument
		for i := uint64(0); i < numberOfElements; i++ {
			repeatedElement := ContractABIArgument{
				Name:       strconv.FormatUint(i, 10),
				Type:       arg.Type[:start],
				Components: arg.Components,
			}
			arrayAsTuple = append(arrayAsTuple, repeatedElement)
		}

		result, err := ParseAllData(arrayAsTuple, data)
		if err != nil {
			return nil, err
		}

		// make sure the results get sorted in order
		keys := make([]uint64, 0)
		for key := range result {
			asUint, _ := strconv.ParseUint(key, 10, 0)
			keys = append(keys, asUint)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		var resultAsList []interface{}
		for _, k := range keys {
			resultAsList = append(resultAsList, result[strconv.FormatUint(k, 10)])
		}
		return resultAsList, nil
	}

	//a bytes array is prefixed with its length,
	//the data is right-padded to the next multiple of 32, but we can just ignore this extra data
	//each byte is treated individually in the result
	if arg.Type == "bytes" {
		numberOfBytes := ParseUint(data[:32]).Uint64()
		allBytes := data[33 : numberOfBytes+33]
		var converted []string
		for _, b := range allBytes {
			converted = append(converted, "0x"+hex.EncodeToString([]byte{b}))
		}
		return converted, nil
	}

	//string parsing is the same as bytes, but just interpreting
	//the result as a string instead of individual bytes
	if arg.Type == "string" {
		numberOfBytes := ParseUint(data[:32]).Uint64()
		return string(data[32 : numberOfBytes+32]), nil
	}

	//a dynamic tuple may contain a mix of dynamic and static elements
	//but contains at least one dynamic element
	//parse it as though its components were indivudally listed, since the
	//data array is only made up of this tuple
	if arg.Type == "tuple" {
		return ParseAllData(arg.Components, data)
	}

	return nil, errors.New("unknown type: " + arg.Type)
}

//TODO: add bounds check with errors on data array

//ParseStaticType will attempt to parse all the possible static types defined by the
//ABI encoding spec. It returns the result of parsing, as well as the next offset from which to parse
//the next element - i.e. the starting offset + how many bytes it read to parse this element
func ParseStaticType(arg ContractABIArgument, data []byte, startingPosition uint64) (interface{}, uint64, error) {

	//a fixed size array of a static type
	//treat it as though it is X number of individually defined elements
	if strings.Contains(arg.Type, "[") && strings.Contains(arg.Type, "]") {
		start := strings.LastIndex(arg.Type, "[")
		end := strings.LastIndex(arg.Type, "]")

		numberOfElements, err := strconv.ParseUint(arg.Type[start+1:end], 10, 0)
		if err != nil {
			return nil, 0, errors.New("error parsing static array size: " + err.Error())
		}

		results := make([]interface{}, 0)
		nextOffset := startingPosition
		for i := uint64(0); i < numberOfElements; i++ {
			nextResult, updatedOffset, err := ParseStaticType(ContractABIArgument{Type: arg.Type[:start], Components: arg.Components}, data, nextOffset)
			if err != nil {
				return nil, 0, err
			}
			results = append(results, nextResult)
			nextOffset = updatedOffset
		}
		return results, nextOffset, nil
	}

	//a set of bytes, from bytes1 upto bytes32
	//the number of bytes is extracted from the type as to truncate the
	//left-padded zeroes, returning the hex encoded value of the bytes
	if strings.HasPrefix(arg.Type, "bytes") {
		numberOfBytes, _ := strconv.ParseUint(arg.Type[5:], 10, 0)
		nextChunk := data[startingPosition : startingPosition+32]
		b := nextChunk[:numberOfBytes]
		return "0x" + hex.EncodeToString(b), startingPosition + 32, nil
	}

	//a bool value, left-padded to 32 bytes
	if arg.Type == "bool" {
		nextChunk := data[startingPosition : startingPosition+32]
		b := nextChunk[31]
		val := b != 0
		return val, startingPosition + 32, nil
	}

	// a fixed 32 byte int. Handled int8 upto int256
	if strings.HasPrefix(arg.Type, "int") {
		nextChunk := data[startingPosition : startingPosition+32]
		val := ParseInt(nextChunk)
		return val, startingPosition + 32, nil
	}

	// a fixed 32 byte uint. Handled uint8 upto uint256
	if strings.HasPrefix(arg.Type, "uint") {
		nextChunk := data[startingPosition : startingPosition+32]
		val := ParseUint(nextChunk)
		return val, startingPosition + 32, nil
	}

	// a fixed 20 byte address, with leading 0s to pad it to 32 bytes
	if strings.HasPrefix(arg.Type, "address") {
		nextChunk := data[startingPosition : startingPosition+32]
		addressBytes := nextChunk[12:32]
		return "0x" + hex.EncodeToString(addressBytes), startingPosition + 32, nil
	}

	// this is a static tuple, we can treat this as though the elements were
	// individually named (instead of being grouped in the tuple), parsing one at a time inline
	if strings.HasPrefix(arg.Type, "tuple") {
		results := make([]interface{}, 0)
		nextOffset := startingPosition
		for _, comp := range arg.Components {
			nextResult, updatedOffset, err := ParseStaticType(comp, data, nextOffset)
			if err != nil {
				return nil, 0, err
			}
			results = append(results, nextResult)
			nextOffset = updatedOffset
		}
		return results, nextOffset, nil
	}

	return nil, 0, errors.New("unknown type: " + arg.Type)
}

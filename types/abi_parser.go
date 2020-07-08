package types

import (
	"encoding/hex"
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
	d.Sum(nil)

	return d.Sum(nil)
}

func ParseAllData(inputs []ContractABIArgument, data []byte) map[string]interface{} {

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

		result, newOffset := ParseStaticType(input, data, currentOffset)

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

		allResults[headItem.Arg.Name] = ParseDynamicType(headItem.Arg, dynamicTypeData)
	}

	return allResults
}

func ParseDynamicType(arg ContractABIArgument, data []byte) interface{} {
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

	//fixed size array
	if strings.Contains(arg.Type, "[") && strings.Contains(arg.Type, "]") {
		start := strings.LastIndex(arg.Type, "[")
		end := strings.LastIndex(arg.Type, "]")

		//TODO: handle error
		numberOfElements, _ := strconv.ParseUint(arg.Type[start+1:end], 10, 0)

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

		result := ParseAllData(arrayAsTuple, data)

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
		return resultAsList
	}

	if arg.Type == "bytes" {
		numberOfBytes := ParseUint(data[:32]).Uint64()
		allBytes := data[33 : numberOfBytes+33]
		var converted []string
		for _, b := range allBytes {
			converted = append(converted, "0x"+hex.EncodeToString([]byte{b}))
		}
		return converted
	}

	if arg.Type == "string" {
		numberOfBytes := ParseUint(data[:32]).Uint64()
		return string(data[32 : numberOfBytes+32])
	}

	if arg.Type == "tuple" {
		return ParseAllData(arg.Components, data)
	}

	//implicit error case
	panic("can't get here")
	return nil
}

func ParseStaticType(arg ContractABIArgument, data []byte, startingPosition uint64) (interface{}, uint64) {

	if strings.Contains(arg.Type, "[") && strings.Contains(arg.Type, "]") {
		start := strings.LastIndex(arg.Type, "[")
		end := strings.LastIndex(arg.Type, "]")

		//TODO: handle error
		numberOfElements, _ := strconv.ParseUint(arg.Type[start+1:end], 10, 0)

		results := make([]interface{}, 0)
		nextOffset := startingPosition
		for i := uint64(0); i < numberOfElements; i++ {
			nextResult, updatedOffset := ParseStaticType(ContractABIArgument{Type: arg.Type[:start], Components: arg.Components}, data, nextOffset)
			results = append(results, nextResult)
			nextOffset = updatedOffset
		}
		return results, nextOffset
	}

	if strings.HasPrefix(arg.Type, "bytes") {
		numberOfBytes, _ := strconv.ParseUint(arg.Type[5:], 10, 0)
		nextChunk := data[startingPosition : startingPosition+32]
		b := nextChunk[:numberOfBytes]
		return "0x" + hex.EncodeToString(b), startingPosition + 32
	}

	if arg.Type == "bool" {
		nextChunk := data[startingPosition : startingPosition+32]
		b := nextChunk[31]
		val := b != 0
		return val, startingPosition + 32
	}

	if strings.HasPrefix(arg.Type, "int") || strings.HasPrefix(arg.Type, "uint") {
		nextChunk := data[startingPosition : startingPosition+32]
		val := ParseInt(nextChunk)
		return val, startingPosition + 32
	}

	if strings.HasPrefix(arg.Type, "address") {
		nextChunk := data[startingPosition : startingPosition+32]
		addressBytes := nextChunk[12:32]
		return "0x" + hex.EncodeToString(addressBytes), startingPosition + 32
	}

	if strings.HasPrefix(arg.Type, "tuple") {
		results := make([]interface{}, 0)
		nextOffset := startingPosition
		for _, comp := range arg.Components {
			nextResult, updatedOffset := ParseStaticType(comp, data, nextOffset)
			results = append(results, nextResult)
			nextOffset = updatedOffset
		}
		return results, nextOffset
	}

	//TODO: make explicit error case
	panic("can't get here")
	return nil, 0
}

pragma solidity ^0.6.0;
pragma experimental ABIEncoderV2;

contract ABIParsingContract {

    struct Custom {
        string first;
        bytes32 second;
        bool third;
    }

    bytes moreThan31;
    uint256[10] uintArr;
    bool[6] boolArr;

    uint256[] uintArrDyn;
    bool[] boolArrDyn;

    Custom[5] customArr;
    Custom[] customArr2;
    string[7] firstStringArr;
    string[] secondStringArr;

    event BytesFixedSize(bytes32 first, bytes1 second, byte third);
    event BoolFixed(bool first, bool second);
    event IntFixed(int256 first, int16 second, int64 third, int fourth);
    event AddressFixed(address first, address second);
    event UintFixed(uint256 first, uint16 second, uint64 third, uint fourth);
    event StringFixed(string first, string second);
    event BytesFixed(bytes first, bytes second);

    event ArrayFixedSize(uint256[10] first, bool[6] second);
    event ArrayDynamicSize(uint256[] first, bool[] second);
    event TupleDynamic(Custom first, Custom second);

    event Mixed(Custom first, bytes32 second, int16 third, Custom fourth, string fifth);
    event StructArray(Custom[5] first, Custom[] second);

    constructor() public {
        emit BytesFixedSize(0x1234567890123456789012345678901234567890123456789012345678901234, 0x74, 0x12);
        emit BoolFixed(true, false);
        emit IntFixed(-98765432109876543210, 12345, -43857439857398534, 12345678901234567890);
        emit AddressFixed(0x1932c48b2bF8102Ba33B4A6B545C32236e342f34, 0x9d13C6D3aFE1721BEef56B55D303B09E021E27ab);
        emit UintFixed(98765432109876543210, 12345, 43857439857398534, 12345678901234567890);
        emit StringFixed("small", "some really large string that will go over the thirty-two byte limit for a single variable");

        for (uint8 i = 0; i < 100; i++) { moreThan31.push(byte(i)); }
        emit BytesFixed(moreThan31, moreThan31);

        for (uint8 i = 0; i < 10; i++) { uintArr[i] = i*100; }
        for (uint8 i = 0; i < 6; i++) { boolArr[i] = (i%2) == 0; }
        emit ArrayFixedSize(uintArr, boolArr);

        for (uint8 i = 0; i < 10; i++) { uintArrDyn.push(i*100); }
        for (uint8 i = 0; i < 20; i++) { boolArrDyn.push((i%2) == 0); }
        emit ArrayDynamicSize(uintArrDyn, boolArrDyn);

        emit TupleDynamic(Custom("first", 0x1234567890123456789012345678901234567890123456789012345678901234, true), Custom("second", 0x4013487654674538507684738547847680974039786439857345674358096798, false));

        emit Mixed(
            Custom("qwerty", 0x1234567890123456789012345678901234567890123456789012345678901234, true),
            0x4386578974647808650460543048357430403897631067453064043476584731,
            3455,
            Custom("asdfghjkl", 0x4013487654674538507684738547847680974039786439857345674358096798, false),
            "string fifth"
        );

        for (uint8 i = 3; i < 5; i++) { customArr[i] = Custom("qwerty", 0x1234567890123456789012345678901234567890123456789012345678901234, true); }
        for (uint8 i = 3; i < 5; i++) { customArr2.push(Custom("qwerty", 0x1234567890123456789012345678901234567890123456789012345678901234, true)); }
        delete customArr2[0];
        emit StructArray(customArr, customArr2);

    }
}
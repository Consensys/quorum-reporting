pragma solidity ^0.6.0;

contract SimpleStorage {

  //Value types

  uint a;
  uint8 b;
  uint8 c;
  int d;
  bool e;

  address f;
  SimpleStorage g;

  byte h1;
  bytes1 h2;
  bytes2 h3;
  bytes31 h4;
  bytes32 h5;

  enum ActionChoices { GoLeft, GoRight, GoStraight, SitStill}
  ActionChoices choice;

  //Arrays
  bytes i1;
  string i2;
  string i5;
  byte[] h6;
  byte[10] h7;
  address[] i3;
  SimpleStorage[] i4;

  struct Funder {
      string addr;
      uint amount;
  }
  Funder funder1;
  Funder[2] fundersFixed;
  Funder[] fundersDyn;

  constructor() public {
    a = 42;
    b = 6;
    c = 9;
    d = -42;
    e = true;

    f = 0xdCad3a6d3569DF655070DEd06cb7A1b2Ccd1D3AF;
    g = SimpleStorage(f);

    h1 = 0x01;
    h2 = 0x00;
    h3 = 0x1000;
    h4 = 0x10000000000000000000000000000000000000000000000000000000000000;
    h5 = 0x1000000000000000000000000000000000000000000000000000000000000000;

    h6.push(0x01);
    h7[0] = 0x01;
    h7[2] = 0x01;
    h7[3] = 0x01;
    h7[4] = 0x01;
    h7[9] = 0x01;

    choice = ActionChoices.GoLeft;

    i1.push(0x01);
    i1.push(0x02);
    i1.push(0x03);
    i1.push(0x04);
    i1.push(0x05);
    i2 = "mystring";
    i5 = "my really long string that is definitely longer than the 32 byte limit";

    funder1 = Funder("some addr", 56);
    fundersFixed[0] = Funder("some addr fixed 1", 85);
    fundersFixed[1] = Funder("some addr fixed 2", 6565);
    fundersDyn.push(Funder("some addr fixed 3", 76309));
    fundersDyn.push(Funder("some addr fixed 4", 5876));
    fundersDyn.push(Funder("some addr fixed 5", 4875443));

  }
}

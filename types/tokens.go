package types

type ERC721Token struct {
	Contract  Address `json:"contract"`
	Holder    Address `json:"holder"`
	Token     string  `json:"token"`
	HeldFrom  uint64  `json:"heldFrom"`
	HeldUntil *uint64 `json:"heldUntil"`
}

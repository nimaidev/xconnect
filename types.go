package main

type NPDU struct {
	Version     byte
	ControlFlag byte
}

type BLVC struct {
	Type     byte
	Function byte
	Length   uint16
}

type BACnetPacket struct {
	BVLCType      byte
	BVLCFunc      byte
	BVLCLength    uint16
	NPDUVersion   byte
	NPDUControl   byte
	DNET          uint16
	DADR          []byte
	SNET          uint16
	SADR          []byte
	HopCount      byte
	APDU          []byte
	PDUType       byte
	ServiceChoice byte
}

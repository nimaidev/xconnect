package main

import (
	"encoding/binary"
	"fmt"
	"log"
)

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

func ParseAPDUPackets(buf []byte) (*BACnetPacket, error) {

	// BVLC header (first 4 bytes)
	//
	pkt := BACnetPacket{
		BVLCType:   buf[0], // Should be 0x81 for BACnet/IP
		BVLCFunc:   buf[1], // BVLC function (0x0A = Original-Unicast-NPDU)
		BVLCLength: binary.BigEndian.Uint16(buf[2:4]),
	}
	log.Printf("Found BlVC Type:  %d | BLVC Func: %d \n", pkt.BVLCType, pkt.BVLCFunc)

	// Skip BVLC header → Get NPDU+APDU
	payload := buf[4:pkt.BVLCLength]

	if len(payload) < 2 {
		return nil, fmt.Errorf("packet too small to contain NPDU + APDU")
	}

	// First byte of NPDU is always version (0x01)
	npduVersion := payload[0]

	if npduVersion != 0x01 {
		return nil, fmt.Errorf("invalid Version %d", npduVersion)
	}
	pkt.NPDUVersion = payload[0]
	pkt.NPDUControl = payload[1]

	index := 2 // [1: NPDU Version. 2: NPDU Control]
	//--- Handle Destination [DNET/DADR] ---
	if pkt.NPDUControl&0x20 != 0 { // if Control is 0x20
		if len(payload) < index+3 {
			return nil, fmt.Errorf("invalid DNET/DADR length")
		}
		// Read DNET (Destination Network) - 2 bytes in big-endian format
		pkt.DNET = binary.BigEndian.Uint16(payload[index : index+2])

		// Read DLEN 1 byte
		dlen := int(payload[index+2])
		// Mov next 3 bytes
		index += 3

		if len(payload) < index+int(dlen) {
			return nil, fmt.Errorf("invalid DADR length")
		}

		// Read DADR (Destination Address)
		pkt.DADR = payload[index : index+dlen]
		index += dlen // Move past DADR byte
	}

	// --- Handle SNET/SADR ---
	if pkt.NPDUControl&0x08 != 0 { // SNET/SADR present
		if len(payload) < index+3 {
			return nil, fmt.Errorf("invalid SNET/SADR length")
		}
		pkt.SNET = binary.BigEndian.Uint16(payload[index : index+2])
		slen := int(payload[index+2])
		index += 3
		if len(payload) < index+slen {
			return nil, fmt.Errorf("invalid SADR length")
		}
		pkt.SADR = payload[index : index+slen]
		index += slen
	}
	// --- Hop Count ---
	if pkt.NPDUControl&0x20 != 0 { // Hop count present
		if len(payload) < index+1 {
			return nil, fmt.Errorf("missing hop count")
		}
		pkt.HopCount = payload[index]
		index++
	}
	// --- APDU ---
	if len(payload) <= index {
		return nil, fmt.Errorf("missing APDU")
	}
	pkt.APDU = payload[index:]

	// Extract PDU Type
	pkt.PDUType = (pkt.APDU[0] & 0xF0) >> 4

	// If Confirmed Request, extract Service Choice
	if pkt.PDUType == 0 { // Confirmed-REQ
		if len(pkt.APDU) >= 4 {
			pkt.ServiceChoice = pkt.APDU[3]
		}
	} else if pkt.PDUType == 1 { // Unconfirmed-REQ
		if len(pkt.APDU) >= 2 {
			pkt.ServiceChoice = pkt.APDU[1]
		}
	}

	return &pkt, nil
}

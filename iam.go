package main

import (
	"bytes"
	"encoding/binary"
)

func PrepareIAmResponse(deviceId int, vendorId int) []byte {
	var apduData []byte

	// APDU Type and Service Choice
	apduData = append(apduData, 0x10) // Unconfirmed Request
	apduData = append(apduData, 0x00) // I-Am service

	// Object Identifier (Device Object)
	objectID := encodeObjectID(8, uint32(deviceId)) // 8 = Device object type
	objectIDBytes := encodeBACnetObjectID(objectID)
	objectIDTag := encodeBACnetApplicationTag(APP_TAG_BACNET_OBJECT_ID, objectIDBytes)
	apduData = append(apduData, objectIDTag...)

	// Maximum APDU Length Accepted (Unsigned Integer)
	maxAPDUBytes := encodeUnsignedInt(1476)
	maxAPDUTag := encodeBACnetApplicationTag(APP_TAG_UNSIGNED_INT, maxAPDUBytes)
	apduData = append(apduData, maxAPDUTag...)

	// Segmentation Supported (Enumerated - 0=segmentation not supported)
	segmentationBytes := []byte{0x00}
	segmentationTag := encodeBACnetApplicationTag(APP_TAG_ENUMERATED, segmentationBytes)
	apduData = append(apduData, segmentationTag...)

	// Vendor ID (Unsigned Integer)
	vendorIDBytes := encodeUnsignedInt(uint32(vendorId))
	vendorIDTag := encodeBACnetApplicationTag(APP_TAG_UNSIGNED_INT, vendorIDBytes)
	apduData = append(apduData, vendorIDTag...)

	// NPDU Header
	npdu := NPDU{
		Version:     0x01, // BACnet version
		ControlFlag: 0x00, // No destination, no source
	}

	npduData := []byte{npdu.Version, npdu.ControlFlag}

	// Calculate total length: BVLC header (4) + NPDU (2) + APDU
	totalLength := 4 + len(npduData) + len(apduData)

	// BVLC Header
	bvlc := BLVC{
		Type:     0x81, // BACnet/IP
		Function: 0x0A, // Original-Unicast-NPDU
		Length:   uint16(totalLength),
	}

	// Build final packet
	var packet []byte
	buf := new(bytes.Buffer)

	// Write BVLC header
	binary.Write(buf, binary.BigEndian, bvlc.Type)
	binary.Write(buf, binary.BigEndian, bvlc.Function)
	binary.Write(buf, binary.BigEndian, bvlc.Length)
	packet = append(packet, buf.Bytes()...)

	// Write NPDU
	packet = append(packet, npduData...)

	// Write APDU
	packet = append(packet, apduData...)

	return packet
}

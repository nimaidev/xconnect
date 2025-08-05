package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const BACnetPort = 47808

func main() {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: BACnetPort}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	fmt.Println("Listening for BACnet packets on UDP 47808...")

	buf := make([]byte, 1500)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Read error:", err)
			continue
		}

		fmt.Printf("\nPacket from %s (%d bytes):\n", clientAddr, n)
		//remove the zero'd buffer
		packet, err := ParseAPDUPackets(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Packet Type: %#x | Request Type: %#x \n", packet.ServiceChoice, packet.PDUType)
		if packet.PDUType == 0x01 && packet.ServiceChoice == 0x08 {
			log.Println("Got Who-Is Request")
		}
		// validate the packet
		// BACnet length field is 2 bytes (big-endian) at buf[2:4]
		// length := binary.BigEndian.Uint16(buf[2:4])
		// log.Println("Size of the request: ", length, len(buf))
		// buf = buf[:length]
		log.Println(buf[:n])

		// // Get the last packet to see the request type
		// reqType := buf[length-1]
		// log.Println(reqType)

		// if reqType == 0x08 { //Got a who is Request
		// 	log.Println("Got an Who Is Broadcast")
		// 	//Send I-AM Response
		// 	packet := prepareIAmResponse(1223, 12)
		// 	//write the packet to the UDP
		// 	log.Println("Packets", packet)
		// 	log.Println("Sending I-am Response")
		// 	conn.WriteToUDP(packet, clientAddr)
		// } else {
		// 	log.Println("Interesting")
		// }
	}
}

type BLVC struct {
	Type     byte
	Function byte
	Length   uint16
}

type NPDU struct {
	Version     byte
	ControlFlag byte
}

// BACnet Application Tag constants
const (
	APP_TAG_NULL             = 0x00
	APP_TAG_BOOLEAN          = 0x10
	APP_TAG_UNSIGNED_INT     = 0x20
	APP_TAG_SIGNED_INT       = 0x30
	APP_TAG_REAL             = 0x40
	APP_TAG_DOUBLE           = 0x50
	APP_TAG_OCTET_STRING     = 0x60
	APP_TAG_CHARACTER_STRING = 0x70
	APP_TAG_BIT_STRING       = 0x80
	APP_TAG_ENUMERATED       = 0x90
	APP_TAG_DATE             = 0xA0
	APP_TAG_TIME             = 0xB0
	APP_TAG_BACNET_OBJECT_ID = 0xC0
)

// Helper to encode BACnet Object Identifier (ObjectType=8 for Device)
func encodeObjectID(objectType, instanceNumber uint32) uint32 {
	return (objectType << 22) | (instanceNumber & 0x3FFFFF)
}

// Encode BACnet Application Tag with data
func encodeBACnetApplicationTag(tagNumber byte, data []byte) []byte {
	var result []byte
	dataLen := len(data)

	if dataLen < 5 {
		// Length fits in 3 bits of tag byte
		tagByte := tagNumber | byte(dataLen)
		result = append(result, tagByte)
	} else {
		// Extended length encoding
		tagByte := tagNumber | 0x05 // 5 indicates extended length
		result = append(result, tagByte)
		if dataLen < 254 {
			result = append(result, byte(dataLen))
		} else {
			result = append(result, 254) // Extended length indicator
			result = append(result, byte(dataLen>>8), byte(dataLen&0xFF))
		}
	}

	result = append(result, data...)
	return result
}

// Encode unsigned integer
func encodeUnsignedInt(value uint32) []byte {
	if value < 256 {
		return []byte{byte(value)}
	} else if value < 65536 {
		return []byte{byte(value >> 8), byte(value & 0xFF)}
	} else if value < 16777216 {
		return []byte{byte(value >> 16), byte(value >> 8), byte(value & 0xFF)}
	} else {
		return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value & 0xFF)}
	}
}

// Encode BACnet Object Identifier
func encodeBACnetObjectID(objectID uint32) []byte {
	return []byte{
		byte(objectID >> 24),
		byte(objectID >> 16),
		byte(objectID >> 8),
		byte(objectID & 0xFF),
	}
}

func prepareIAmResponse(deviceId int, vendorId int) []byte {
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

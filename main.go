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
		fmt.Printf("Raw packet: %v\n", buf[:n])

		if n < 4 {
			log.Println("Packet too small")
			continue
		}

		if buf[0] == 0x81 {
			fmt.Println("BACnet/IP protocol detected")
		}

		// BACnet length field is 2 bytes (big-endian) at buf[2:4]
		length := binary.BigEndian.Uint16(buf[2:4])
		fmt.Printf("BACnet packet length: %d, UDP payload: %d\n", length, n)

		if int(length) > n {
			log.Println("Invalid packet: BACnet length exceeds UDP payload")
			continue
		}

		// Parse NPDU to find APDU start
		if length < 6 {
			log.Println("Packet too small for BVLC + NPDU")
			continue
		}

		npduStart := 4
		npduVersion := buf[npduStart]
		npduControl := buf[npduStart+1]

		fmt.Printf("NPDU Version: 0x%02X, Control: 0x%02X\n", npduVersion, npduControl)

		apduStart := npduStart + 2 // Base NPDU size

		// Check for destination network info
		if (npduControl & 0x20) != 0 {
			if apduStart+2 >= int(length) {
				log.Println("Invalid NPDU: missing DNET")
				continue
			}
			dnet := binary.BigEndian.Uint16(buf[apduStart : apduStart+2])
			fmt.Printf("DNET: 0x%04X\n", dnet)
			apduStart += 2
			if dnet != 0xFFFF {
				if apduStart >= int(length) {
					log.Println("Invalid NPDU: missing DLEN")
					continue
				}
				dlen := buf[apduStart]
				fmt.Printf("DLEN: %d\n", dlen)
				apduStart += 1 + int(dlen)
			} else {
				apduStart += 1 // Skip DLEN for broadcast
			}
		}

		// Check for source network info
		if (npduControl & 0x08) != 0 {
			if apduStart+2 >= int(length) {
				log.Println("Invalid NPDU: missing SNET")
				continue
			}
			snet := binary.BigEndian.Uint16(buf[apduStart : apduStart+2])
			fmt.Printf("SNET: 0x%04X\n", snet)
			apduStart += 2
			if apduStart >= int(length) {
				log.Println("Invalid NPDU: missing SLEN")
				continue
			}
			slen := buf[apduStart]
			fmt.Printf("SLEN: %d\n", slen)
			apduStart += 1 + int(slen)
		}

		// Skip hop count if present
		if (npduControl & 0x20) != 0 {
			if apduStart >= int(length) {
				log.Println("Invalid NPDU: missing hop count")
				continue
			}
			hopCount := buf[apduStart]
			fmt.Printf("Hop Count: %d\n", hopCount)
			apduStart++
		}

		if apduStart >= int(length) {
			log.Println("No APDU data")
			continue
		}

		fmt.Printf("APDU starts at byte %d\n", apduStart)
		fmt.Printf("APDU data: %v\n", buf[apduStart:length])

		// Check for Who-Is request
		apduType := buf[apduStart]
		fmt.Printf("APDU Type: 0x%02X\n", apduType)

		if apduType == 0x10 && apduStart+1 < int(length) {
			serviceChoice := buf[apduStart+1]
			fmt.Printf("Service Choice: 0x%02X\n", serviceChoice)

			if serviceChoice == 0x08 {
				log.Println("Got a Who-Is broadcast - sending I-Am response")
				packet := prepareIAmResponse(1223, 12)
				logPacketDetails(packet)

				// Send as broadcast to the BACnet broadcast address
				broadcastAddr := &net.UDPAddr{
					IP:   net.IPv4bcast, // 255.255.255.255
					Port: BACnetPort,
				}

				_, err := conn.WriteToUDP(packet, broadcastAddr)
				if err != nil {
					log.Printf("Error sending broadcast response: %v\n", err)
				} else {
					log.Println("I-Am broadcast response sent successfully")
				}
			}
		}
	}
}

func logPacketDetails(packet []byte) {
	fmt.Printf("Generated I-Am packet (%d bytes): %v\n", len(packet), packet)

	if len(packet) < 4 {
		log.Println("Packet too small")
		return
	}

	// Parse and log each section
	fmt.Printf("BVLC: Type=0x%02X, Function=0x%02X, Length=%d\n",
		packet[0], packet[1], binary.BigEndian.Uint16(packet[2:4]))

	if len(packet) > 6 {
		fmt.Printf("NPDU: Version=0x%02X, Control=0x%02X\n", packet[4], packet[5])

		if len(packet) > 8 {
			fmt.Printf("APDU: PDU_Type=0x%02X, Service=0x%02X\n", packet[6], packet[7])
			fmt.Printf("APDU Data: %v\n", packet[8:])
		}
	}
}

// Helper to encode BACnet Object Identifier (ObjectType=8 for Device)
func encodeObjectID(objectType, instanceNumber uint32) uint32 {
	return (objectType << 22) | (instanceNumber & 0x3FFFFF)
}

func prepareIAmResponse(deviceId int, vendorId int) []byte {
	log.Printf("Preparing I-Am response for Device ID: %d, Vendor ID: %d", deviceId, vendorId)

	var packet bytes.Buffer

	// BVLC Header - Use BROADCAST for I-Am responses
	packet.WriteByte(0x81) // Type: BACnet/IP
	packet.WriteByte(0x0B) // Function: Original-Broadcast-NPDU (I-Am should be broadcast)
	packet.WriteByte(0x00) // Length high byte (placeholder)
	packet.WriteByte(0x00) // Length low byte (placeholder)

	// NPDU Header
	packet.WriteByte(0x01) // Version 1
	packet.WriteByte(0x20) // Control flags: destination network present (for broadcast)

	// NPDU Destination (for broadcast)
	packet.WriteByte(0xFF) // DNET high byte (0xFFFF = global broadcast)
	packet.WriteByte(0xFF) // DNET low byte
	packet.WriteByte(0x00) // DLEN (0 = broadcast to all)
	packet.WriteByte(0xFF) // Hop count

	// APDU Header
	packet.WriteByte(0x10) // PDU Type: Unconfirmed Request
	packet.WriteByte(0x00) // Service Choice: I-Am

	// Parameter 1: Object Identifier (Device object) - Context tag 0
	objectID := encodeObjectID(8, uint32(deviceId))
	log.Printf("Encoded Object ID: 0x%08X", objectID)

	// BACnet Object ID encoding: tag + 4 bytes of data
	packet.WriteByte(0x0C) // Context tag 0, 4 bytes following
	packet.WriteByte(byte(objectID >> 24))
	packet.WriteByte(byte(objectID >> 16))
	packet.WriteByte(byte(objectID >> 8))
	packet.WriteByte(byte(objectID))

	// Parameter 2: Maximum APDU Length Accepted - Context tag 1
	maxAPDU := uint16(1476)
	log.Printf("Max APDU Length: %d", maxAPDU)

	// BACnet Unsigned encoding: tag + 2 bytes of data
	packet.WriteByte(0x19) // Context tag 1, 2 bytes following
	packet.WriteByte(byte(maxAPDU >> 8))
	packet.WriteByte(byte(maxAPDU))

	// Parameter 3: Segmentation Support - Context tag 2
	segmentation := byte(0x03) // SEGMENTATION_NOT_SUPPORTED
	log.Printf("Segmentation: 0x%02X", segmentation)

	// BACnet Enumerated encoding: tag + 1 byte of data
	packet.WriteByte(0x22) // Context tag 2, 1 byte following
	packet.WriteByte(segmentation)

	// Parameter 4: Vendor Identifier - Context tag 3
	vendor := uint16(vendorId)
	log.Printf("Vendor ID: %d", vendor)

	// BACnet Unsigned encoding: tag + 2 bytes of data
	packet.WriteByte(0x39) // Context tag 3, 2 bytes following
	packet.WriteByte(byte(vendor >> 8))
	packet.WriteByte(byte(vendor))

	// Update BVLC length field
	totalLength := packet.Len()
	packetBytes := packet.Bytes()
	binary.BigEndian.PutUint16(packetBytes[2:4], uint16(totalLength))

	log.Printf("Final packet length: %d bytes", totalLength)

	return packetBytes
}

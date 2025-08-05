package main

// import (
// 	"bytes"
// 	"encoding/binary"
// 	"fmt"
// 	"log"
// 	"net"
// 	"strings"
// )

// const BACnetPort = 47808

// // BACnet Application Tag constants
// const (
// 	APP_TAG_NULL             = 0x00
// 	APP_TAG_BOOLEAN          = 0x10
// 	APP_TAG_UNSIGNED_INT     = 0x20
// 	APP_TAG_SIGNED_INT       = 0x30
// 	APP_TAG_REAL             = 0x40
// 	APP_TAG_DOUBLE           = 0x50
// 	APP_TAG_OCTET_STRING     = 0x60
// 	APP_TAG_CHARACTER_STRING = 0x70
// 	APP_TAG_BIT_STRING       = 0x80
// 	APP_TAG_ENUMERATED       = 0x90
// 	APP_TAG_DATE             = 0xA0
// 	APP_TAG_TIME             = 0xB0
// 	APP_TAG_BACNET_OBJECT_ID = 0xC0
// )

// type BLVC struct {
// 	Type     byte
// 	Function byte
// 	Length   uint16
// }

// type NPDU struct {
// 	Version     byte
// 	ControlFlag byte
// }

// func main() {
// 	// Force binding to the specific target interface
// 	var conn *net.UDPConn
// 	var err error

// 	// Find the target interface
// 	log.Println("Attempting to find and bind to 10.10.10.x interface...")
// 	localIP := findBestInterface()

// 	if localIP != "" {
// 		// Try binding to specific interface first
// 		targetAddr := &net.UDPAddr{IP: net.ParseIP(localIP), Port: BACnetPort}
// 		conn, err = net.ListenUDP("udp4", targetAddr)
// 		if err != nil {
// 			log.Printf("Failed to bind to %s:%d - %v", localIP, BACnetPort, err)
// 			log.Println("This might be because another BACnet application is already running.")

// 			// Check if port is already in use
// 			log.Println("Checking if port 47808 is already in use...")
// 			testConn, testErr := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(localIP), Port: BACnetPort})
// 			if testErr != nil {
// 				log.Printf("Port 47808 on %s is already in use: %v", localIP, testErr)
// 				log.Println("Please stop other BACnet applications (like Yabe server mode) and try again.")

// 				// Try a different port for testing
// 				log.Println("Trying port 47809 for testing...")
// 				testAddr := &net.UDPAddr{IP: net.ParseIP(localIP), Port: 47809}
// 				conn, err = net.ListenUDP("udp4", testAddr)
// 				if err != nil {
// 					log.Fatalf("Failed to bind to test port: %v", err)
// 				}
// 				log.Printf("Successfully bound to %s:%d for testing", localIP, 47809)
// 			} else {
// 				testConn.Close()
// 				log.Println("Port should be available, retrying...")
// 				conn, err = net.ListenUDP("udp4", targetAddr)
// 			}
// 		} else {
// 			log.Printf("Successfully bound to %s:%d", localIP, BACnetPort)
// 		}
// 	}

// 	// Only fall back if we couldn't bind to specific interface
// 	if conn == nil {
// 		log.Println("Falling back to binding to all interfaces...")
// 		allAddr := &net.UDPAddr{IP: net.IPv4zero, Port: BACnetPort}
// 		conn, err = net.ListenUDP("udp4", allAddr)
// 		if err != nil {
// 			log.Fatalf("Failed to listen on any interface: %v", err)
// 		}
// 		log.Printf("Successfully bound to all interfaces: %s", conn.LocalAddr())
// 	}

// 	defer conn.Close()

// 	fmt.Printf("Listening for BACnet packets on UDP %d...\n", BACnetPort)
// 	fmt.Printf("Bound to: %s\n", conn.LocalAddr())

// 	// Print routing table info
// 	log.Println("Network interface summary:")
// 	findBestInterface()

// 	requestCount := 0
// 	buf := make([]byte, 65535)
// 	for {
// 		n, clientAddr, err := conn.ReadFromUDP(buf)
// 		if err != nil {
// 			log.Println("Read error:", err)
// 			continue
// 		}

// 		requestCount++
// 		log.Printf("=== Request #%d ===", requestCount)

// 		// Copy the buffer data to avoid race conditions
// 		packetData := make([]byte, n)
// 		copy(packetData, buf[:n])

// 		handleUDPPacket(packetData, clientAddr, conn)
// 	}
// }

// func handleUDPPacket(buf []byte, clientAddr *net.UDPAddr, conn *net.UDPConn) {
// 	fmt.Printf("\nPacket from %s (%d bytes):\n", clientAddr, len(buf))

// 	// Basic BACnet/IP validation
// 	if len(buf) < 8 || buf[0] != 0x81 {
// 		log.Println("Not a valid BACnet/IP packet")
// 		return
// 	}

// 	fmt.Println("BACnet/IP protocol detected")

// 	// Validate packet length
// 	length := binary.BigEndian.Uint16(buf[2:4])
// 	if int(length) != len(buf) {
// 		log.Printf("Length mismatch: header says %d, actual %d", length, len(buf))
// 		return
// 	}

// 	log.Printf("Packet length: %d bytes", length)
// 	log.Printf("Raw packet: %x", buf)

// 	bvlcFunction := buf[1]

// 	// Handle Who-Is (Broadcast)
// 	if bvlcFunction == 0x0B { // Original-Broadcast-NPDU
// 		handleWhoIs(buf, clientAddr, conn)
// 		return
// 	}

// 	// Handle ReadProperty (Unicast)
// 	if bvlcFunction == 0x0A { // Original-Unicast-NPDU
// 		handleReadProperty(buf, clientAddr, conn)
// 		return
// 	}

// 	log.Printf("Unhandled BVLC function: 0x%02x", bvlcFunction)
// }

// func handleWhoIs(buf []byte, clientAddr *net.UDPAddr, conn *net.UDPConn) {
// 	// Parse Who-Is request
// 	npduIdx := 4
// 	npduControl := buf[npduIdx+1]

// 	apduIdx := npduIdx + 2

// 	// Parse NPDU fields
// 	if (npduControl & 0x20) != 0 { // Destination network present
// 		apduIdx += 2 // Skip destination network
// 		if apduIdx < len(buf) {
// 			dlen := buf[apduIdx]
// 			apduIdx++            // Skip DLEN
// 			apduIdx += int(dlen) // Skip destination address
// 		}
// 	}

// 	if (npduControl & 0x08) != 0 { // Source network present
// 		apduIdx += 2 // Skip source network
// 		if apduIdx < len(buf) {
// 			slen := buf[apduIdx]
// 			apduIdx++            // Skip SLEN
// 			apduIdx += int(slen) // Skip source address
// 		}
// 	}

// 	if (npduControl & 0x20) != 0 { // Hop count present
// 		apduIdx++
// 	}

// 	// Check for Who-Is service
// 	if apduIdx+1 < len(buf) && buf[apduIdx] == 0x10 && buf[apduIdx+1] == 0x08 {
// 		log.Println("Got a Who-Is Broadcast request")

// 		packet := prepareIAmResponse(1223, 12)
// 		log.Printf("I-Am response packet: %x", packet)
// 		log.Println("Sending I-Am Response")

// 		_, err := conn.WriteToUDP(packet, clientAddr)
// 		if err != nil {
// 			log.Printf("Error sending I-Am response: %v", err)
// 		} else {
// 			log.Println("I-Am response sent successfully")
// 		}
// 	}
// }

// func handleReadProperty(buf []byte, clientAddr *net.UDPAddr, conn *net.UDPConn) {
// 	log.Println("Processing ReadProperty request...")

// 	npduIdx := 4
// 	if npduIdx+1 >= len(buf) {
// 		log.Println("Buffer too short for NPDU")
// 		return
// 	}

// 	npduVersion := buf[npduIdx]
// 	npduControl := buf[npduIdx+1]

// 	log.Printf("NPDU Version: %02x, Control: %02x", npduVersion, npduControl)

// 	apduIdx := npduIdx + 2

// 	// Skip NPDU addressing if present
// 	if (npduControl & 0x20) != 0 { // Destination network present
// 		if apduIdx+1 >= len(buf) {
// 			log.Println("Buffer too short for destination network")
// 			return
// 		}
// 		apduIdx += 2 // Skip DNET
// 		if apduIdx >= len(buf) {
// 			log.Println("Buffer too short for DLEN")
// 			return
// 		}
// 		dlen := buf[apduIdx]
// 		apduIdx++            // Skip DLEN
// 		apduIdx += int(dlen) // Skip DADR
// 	}

// 	if (npduControl & 0x08) != 0 { // Source network present
// 		if apduIdx+1 >= len(buf) {
// 			log.Println("Buffer too short for source network")
// 			return
// 		}
// 		apduIdx += 2 // Skip SNET
// 		if apduIdx >= len(buf) {
// 			log.Println("Buffer too short for SLEN")
// 			return
// 		}
// 		slen := buf[apduIdx]
// 		apduIdx++            // Skip SLEN
// 		apduIdx += int(slen) // Skip SADR
// 	}

// 	if (npduControl & 0x20) != 0 { // Hop count present
// 		apduIdx++
// 	}

// 	log.Printf("APDU starts at index: %d, buffer length: %d", apduIdx, len(buf))

// 	if apduIdx+3 >= len(buf) {
// 		log.Printf("Buffer too short for APDU header. Need %d bytes, have %d", apduIdx+4, len(buf))
// 		return
// 	}

// 	// Parse APDU header
// 	apduTypeAndFlags := buf[apduIdx]
// 	apduType := apduTypeAndFlags >> 4 // Upper 4 bits
// 	pduFlags := buf[apduIdx+1]
// 	invokeID := buf[apduIdx+2]
// 	serviceChoice := buf[apduIdx+3]

// 	log.Printf("APDU Type: 0x%01x, Type+Flags: 0x%02x, PDU Flags: 0x%02x, Invoke ID: %d, Service Choice: %d",
// 		apduType, apduTypeAndFlags, pduFlags, invokeID, serviceChoice)

// 	// Check for Confirmed Request (0x0) and ReadProperty service (0x0C = 12)
// 	if apduType != 0x0 {
// 		log.Printf("Not a confirmed request. APDU Type: 0x%01x", apduType)
// 		return
// 	}

// 	if serviceChoice != 0x0C {
// 		log.Printf("Not a ReadProperty request. Service Choice: %d", serviceChoice)
// 		// Send error for unsupported service
// 		errorPacket := prepareErrorResponse(invokeID, serviceChoice, 9, 23) // Service error, Service request denied
// 		log.Printf("Sending error response for unsupported service: %x", errorPacket)
// 		conn.WriteToUDP(errorPacket, clientAddr)
// 		return
// 	}

// 	log.Printf("Got a ReadProperty request, invokeID: %d", invokeID)

// 	// Parse ReadProperty parameters
// 	paramIdx := apduIdx + 4

// 	// Parse Object Identifier (Context Tag 0)
// 	if paramIdx >= len(buf) {
// 		log.Println("No object identifier found")
// 		return
// 	}

// 	if buf[paramIdx] != 0x0C { // Context tag 0, length 4
// 		log.Printf("Expected object identifier context tag 0x0C, got 0x%02x at index %d", buf[paramIdx], paramIdx)
// 		return
// 	}

// 	if paramIdx+4 >= len(buf) {
// 		log.Println("Buffer too short for object identifier")
// 		return
// 	}

// 	// Extract object type and instance
// 	objData := binary.BigEndian.Uint32(buf[paramIdx+1 : paramIdx+5])
// 	objType := (objData >> 22) & 0x3FF
// 	objInstance := objData & 0x3FFFFF

// 	log.Printf("Object Type: %d, Instance: %d", objType, objInstance)

// 	// Parse Property Identifier (Context Tag 1)
// 	propIdx := paramIdx + 5
// 	if propIdx >= len(buf) || buf[propIdx] != 0x19 { // Context tag 1, length 1
// 		log.Printf("Expected property identifier context tag 0x19, got 0x%02x at index %d", buf[propIdx], propIdx)
// 		return
// 	}

// 	if propIdx+1 >= len(buf) {
// 		log.Println("Buffer too short for property identifier")
// 		return
// 	}

// 	propertyID := buf[propIdx+1]
// 	log.Printf("Property ID: %d", propertyID)

// 	// Check for array index (Context Tag 2) - optional
// 	arrayIdx := propIdx + 2
// 	var arrayIndex *uint32 = nil
// 	if arrayIdx < len(buf) && buf[arrayIdx] == 0x29 { // Context tag 2, length 1
// 		if arrayIdx+1 < len(buf) {
// 			index := uint32(buf[arrayIdx+1])
// 			arrayIndex = &index
// 			log.Printf("Array Index: %d", index)
// 		}
// 	}

// 	// Handle the request based on object type, instance, and property
// 	if objType == 8 && objInstance == 1223 { // Device object, instance 1223
// 		switch propertyID {
// 		case 76: // object-list
// 			log.Println("Handling object-list property")
// 			packet := prepareReadPropertyAck(invokeID, 1223, propertyID, arrayIndex)
// 			log.Printf("ReadProperty response packet: %x", packet)
// 			log.Println("Sending ReadProperty ComplexAck Response")

// 			_, err := conn.WriteToUDP(packet, clientAddr)
// 			if err != nil {
// 				log.Printf("Error sending ReadProperty response: %v", err)
// 			} else {
// 				log.Println("ReadProperty response sent successfully")
// 			}
// 			return

// 		case 77: // object-name
// 			log.Println("Handling object-name property")
// 			packet := prepareObjectNameResponse(invokeID, 1223)
// 			_, err := conn.WriteToUDP(packet, clientAddr)
// 			if err != nil {
// 				log.Printf("Error sending object-name response: %v", err)
// 			} else {
// 				log.Println("Object-name response sent successfully")
// 			}
// 			return

// 		default:
// 			log.Printf("Property %d not supported", propertyID)
// 		}
// 	} else {
// 		log.Printf("Object %d:%d not supported", objType, objInstance)
// 	}

// 	// Send error response for unsupported requests
// 	log.Println("Sending Error Response - Property or object not supported")
// 	errorPacket := prepareErrorResponse(invokeID, serviceChoice, 2, 32) // Object error, Unknown property
// 	log.Printf("Error response packet: %x", errorPacket)
// 	_, err := conn.WriteToUDP(errorPacket, clientAddr)
// 	if err != nil {
// 		log.Printf("Error sending error response: %v", err)
// 	} else {
// 		log.Println("Error response sent successfully")
// 	}
// }

// // Helper functions remain the same
// func encodeObjectID(objectType, instanceNumber uint32) uint32 {
// 	return (objectType << 22) | (instanceNumber & 0x3FFFFF)
// }

// func encodeBACnetApplicationTag(tagNumber byte, data []byte) []byte {
// 	var result []byte
// 	dataLen := len(data)

// 	if dataLen < 5 {
// 		tagByte := tagNumber | byte(dataLen)
// 		result = append(result, tagByte)
// 	} else {
// 		tagByte := tagNumber | 0x05
// 		result = append(result, tagByte)
// 		if dataLen < 254 {
// 			result = append(result, byte(dataLen))
// 		} else {
// 			result = append(result, 254)
// 			result = append(result, byte(dataLen>>8), byte(dataLen&0xFF))
// 		}
// 	}

// 	result = append(result, data...)
// 	return result
// }

// func encodeUnsignedInt(value uint32) []byte {
// 	if value < 256 {
// 		return []byte{byte(value)}
// 	} else if value < 65536 {
// 		return []byte{byte(value >> 8), byte(value & 0xFF)}
// 	} else if value < 16777216 {
// 		return []byte{byte(value >> 16), byte(value >> 8), byte(value & 0xFF)}
// 	} else {
// 		return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value & 0xFF)}
// 	}
// }

// func encodeBACnetObjectID(objectID uint32) []byte {
// 	return []byte{
// 		byte(objectID >> 24),
// 		byte(objectID >> 16),
// 		byte(objectID >> 8),
// 		byte(objectID & 0xFF),
// 	}
// }

// func prepareIAmResponse(deviceId int, vendorId int) []byte {
// 	var apduData []byte

// 	apduData = append(apduData, 0x10) // Unconfirmed Request
// 	apduData = append(apduData, 0x00) // I-Am service

// 	objectID := encodeObjectID(8, uint32(deviceId))
// 	objectIDBytes := encodeBACnetObjectID(objectID)
// 	objectIDTag := encodeBACnetApplicationTag(APP_TAG_BACNET_OBJECT_ID, objectIDBytes)
// 	apduData = append(apduData, objectIDTag...)

// 	maxAPDUBytes := encodeUnsignedInt(1476)
// 	maxAPDUTag := encodeBACnetApplicationTag(APP_TAG_UNSIGNED_INT, maxAPDUBytes)
// 	apduData = append(apduData, maxAPDUTag...)

// 	segmentationBytes := []byte{0x00}
// 	segmentationTag := encodeBACnetApplicationTag(APP_TAG_ENUMERATED, segmentationBytes)
// 	apduData = append(apduData, segmentationTag...)

// 	vendorIDBytes := encodeUnsignedInt(uint32(vendorId))
// 	vendorIDTag := encodeBACnetApplicationTag(APP_TAG_UNSIGNED_INT, vendorIDBytes)
// 	apduData = append(apduData, vendorIDTag...)

// 	npduData := []byte{0x01, 0x00}
// 	totalLength := 4 + len(npduData) + len(apduData)

// 	bvlc := BLVC{
// 		Type:     0x81,
// 		Function: 0x0A,
// 		Length:   uint16(totalLength),
// 	}

// 	var packet []byte
// 	buf := new(bytes.Buffer)
// 	binary.Write(buf, binary.BigEndian, bvlc.Type)
// 	binary.Write(buf, binary.BigEndian, bvlc.Function)
// 	binary.Write(buf, binary.BigEndian, bvlc.Length)
// 	packet = append(packet, buf.Bytes()...)
// 	packet = append(packet, npduData...)
// 	packet = append(packet, apduData...)

// 	return packet
// }

// func prepareReadPropertyAck(invokeID byte, deviceID int, propertyID byte, arrayIndex *uint32) []byte {
// 	var apduData []byte

// 	// ComplexAck PDU
// 	apduData = append(apduData, 0x30)     // PDU Type: ComplexAck
// 	apduData = append(apduData, invokeID) // Invoke ID
// 	apduData = append(apduData, 0x0C)     // Service Choice: ReadProperty

// 	// Object ID (Device:deviceID) - Context Tag 0
// 	objectID := encodeObjectID(8, uint32(deviceID))
// 	objectIDBytes := encodeBACnetObjectID(objectID)
// 	apduData = append(apduData, 0x0C) // Context tag 0, length 4
// 	apduData = append(apduData, objectIDBytes...)

// 	// Property Identifier - Context Tag 1
// 	apduData = append(apduData, 0x19)       // Context tag 1, length 1
// 	apduData = append(apduData, propertyID) // Property ID

// 	// Array Index - Context Tag 2 (if present)
// 	if arrayIndex != nil {
// 		apduData = append(apduData, 0x29) // Context tag 2, length 1
// 		apduData = append(apduData, byte(*arrayIndex))
// 	}

// 	// Opening tag for property value (Context Tag 3)
// 	apduData = append(apduData, 0x3E) // Opening tag 3

// 	// Handle object-list property specifically
// 	if propertyID == 76 { // object-list
// 		if arrayIndex != nil && *arrayIndex == 0 {
// 			// Return array length
// 			arrayLengthBytes := encodeUnsignedInt(1) // We have 1 object in our list
// 			arrayLengthTag := encodeBACnetApplicationTag(APP_TAG_UNSIGNED_INT, arrayLengthBytes)
// 			apduData = append(apduData, arrayLengthTag...)
// 		} else if arrayIndex != nil && *arrayIndex == 1 {
// 			// Return first object in the list
// 			objListID := encodeObjectID(8, uint32(deviceID))
// 			objListIDBytes := encodeBACnetObjectID(objListID)
// 			objListIDTag := encodeBACnetApplicationTag(APP_TAG_BACNET_OBJECT_ID, objListIDBytes)
// 			apduData = append(apduData, objListIDTag...)
// 		} else if arrayIndex == nil {
// 			// Return entire array
// 			objListID := encodeObjectID(8, uint32(deviceID))
// 			objListIDBytes := encodeBACnetObjectID(objListID)
// 			objListIDTag := encodeBACnetApplicationTag(APP_TAG_BACNET_OBJECT_ID, objListIDBytes)
// 			apduData = append(apduData, objListIDTag...)
// 		}
// 	}

// 	// Closing tag for property value (Context Tag 3)
// 	apduData = append(apduData, 0x3F) // Closing tag 3

// 	// NPDU
// 	npduData := []byte{0x01, 0x00}

// 	// Calculate total length
// 	totalLength := 4 + len(npduData) + len(apduData)

// 	// BVLC
// 	bvlc := BLVC{
// 		Type:     0x81,
// 		Function: 0x0A,
// 		Length:   uint16(totalLength),
// 	}

// 	// Build final packet
// 	var packet []byte
// 	buf := new(bytes.Buffer)
// 	binary.Write(buf, binary.BigEndian, bvlc.Type)
// 	binary.Write(buf, binary.BigEndian, bvlc.Function)
// 	binary.Write(buf, binary.BigEndian, bvlc.Length)
// 	packet = append(packet, buf.Bytes()...)
// 	packet = append(packet, npduData...)
// 	packet = append(packet, apduData...)

// 	return packet
// }

// func prepareObjectNameResponse(invokeID byte, deviceID int) []byte {
// 	var apduData []byte

// 	// ComplexAck PDU
// 	apduData = append(apduData, 0x30)     // PDU Type: ComplexAck
// 	apduData = append(apduData, invokeID) // Invoke ID
// 	apduData = append(apduData, 0x0C)     // Service Choice: ReadProperty

// 	// Object ID (Device:deviceID) - Context Tag 0
// 	objectID := encodeObjectID(8, uint32(deviceID))
// 	objectIDBytes := encodeBACnetObjectID(objectID)
// 	apduData = append(apduData, 0x0C) // Context tag 0, length 4
// 	apduData = append(apduData, objectIDBytes...)

// 	// Property Identifier (object-name = 77) - Context Tag 1
// 	apduData = append(apduData, 0x19) // Context tag 1, length 1
// 	apduData = append(apduData, 77)   // Property ID: object-name

// 	// Opening tag for property value (Context Tag 3)
// 	apduData = append(apduData, 0x3E) // Opening tag 3

// 	// Object name as character string
// 	deviceName := fmt.Sprintf("Device-%d", deviceID)
// 	nameBytes := []byte(deviceName)
// 	// Add character set encoding (0 = ANSI X3.4)
// 	nameWithEncoding := append([]byte{0}, nameBytes...)
// 	nameTag := encodeBACnetApplicationTag(APP_TAG_CHARACTER_STRING, nameWithEncoding)
// 	apduData = append(apduData, nameTag...)

// 	// Closing tag for property value (Context Tag 3)
// 	apduData = append(apduData, 0x3F) // Closing tag 3

// 	// NPDU
// 	npduData := []byte{0x01, 0x00}

// 	// Calculate total length
// 	totalLength := 4 + len(npduData) + len(apduData)

// 	// BVLC
// 	bvlc := BLVC{
// 		Type:     0x81,
// 		Function: 0x0A,
// 		Length:   uint16(totalLength),
// 	}

// 	// Build final packet
// 	var packet []byte
// 	buf := new(bytes.Buffer)
// 	binary.Write(buf, binary.BigEndian, bvlc.Type)
// 	binary.Write(buf, binary.BigEndian, bvlc.Function)
// 	binary.Write(buf, binary.BigEndian, bvlc.Length)
// 	packet = append(packet, buf.Bytes()...)
// 	packet = append(packet, npduData...)
// 	packet = append(packet, apduData...)

// 	return packet
// }

// func prepareErrorResponse(invokeID byte, serviceChoice byte, errorClass byte, errorCode byte) []byte {
// 	var apduData []byte

// 	// Error PDU
// 	apduData = append(apduData, 0x50)          // PDU Type: Error
// 	apduData = append(apduData, invokeID)      // Invoke ID
// 	apduData = append(apduData, serviceChoice) // Service Choice

// 	// Error Class (Application Tag)
// 	errorClassTag := encodeBACnetApplicationTag(APP_TAG_ENUMERATED, []byte{errorClass})
// 	apduData = append(apduData, errorClassTag...)

// 	// Error Code (Application Tag)
// 	errorCodeTag := encodeBACnetApplicationTag(APP_TAG_ENUMERATED, []byte{errorCode})
// 	apduData = append(apduData, errorCodeTag...)

// 	// NPDU
// 	npduData := []byte{0x01, 0x00}

// 	// Calculate total length
// 	totalLength := 4 + len(npduData) + len(apduData)

// 	// BVLC
// 	bvlc := BLVC{
// 		Type:     0x81,
// 		Function: 0x0A,
// 		Length:   uint16(totalLength),
// 	}

// 	// Build final packet
// 	var packet []byte
// 	buf := new(bytes.Buffer)
// 	binary.Write(buf, binary.BigEndian, bvlc.Type)
// 	binary.Write(buf, binary.BigEndian, bvlc.Function)
// 	binary.Write(buf, binary.BigEndian, bvlc.Length)
// 	packet = append(packet, buf.Bytes()...)
// 	packet = append(packet, npduData...)
// 	packet = append(packet, apduData...)

// 	return packet
// }

// func findBestInterface() string {
// 	ifaces, err := net.Interfaces()
// 	if err != nil {
// 		log.Printf("Error getting interfaces: %v", err)
// 		return ""
// 	}

// 	log.Println("Available network interfaces:")
// 	for _, iface := range ifaces {
// 		addrs, err := iface.Addrs()
// 		if err != nil {
// 			continue
// 		}

// 		for _, addr := range addrs {
// 			var ip net.IP
// 			switch v := addr.(type) {
// 			case *net.IPNet:
// 				ip = v.IP
// 			case *net.IPAddr:
// 				ip = v.IP
// 			}

// 			if ip == nil || ip.IsLoopback() {
// 				continue
// 			}

// 			if ipv4 := ip.To4(); ipv4 != nil {
// 				log.Printf("  %s: %s", iface.Name, ipv4.String())

// 				// Prefer 10.10.10.x network
// 				if strings.HasPrefix(ipv4.String(), "10.10.10.") {
// 					log.Printf("Found target network interface: %s", ipv4.String())
// 					return ipv4.String()
// 				}
// 			}
// 		}
// 	}

// 	return ""
// }

// func getLocalIP() string {
// 	ifaces, err := net.Interfaces()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, iface := range ifaces {
// 		addrs, err := iface.Addrs()
// 		if err != nil {
// 			continue
// 		}
// 		for _, addr := range addrs {
// 			var ip net.IP
// 			switch v := addr.(type) {
// 			case *net.IPNet:
// 				ip = v.IP
// 			case *net.IPAddr:
// 				ip = v.IP
// 			}
// 			if ip == nil || ip.IsLoopback() {
// 				continue
// 			}
// 			if ipv4 := ip.To4(); ipv4 != nil && strings.HasPrefix(ipv4.String(), "172.20.") {
// 				return ipv4.String()
// 			}
// 		}
// 	}
// 	return "0.0.0.0"
// }

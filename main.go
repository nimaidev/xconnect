package main

import (
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
		log.Println(buf[:n])
		//remove the zero'd buffer
		packet, err := ParseAPDUPackets(buf)
		if err != nil {
			log.Fatal(err)
		}
		// handle UDP COnnection
		go handleUDPConnection(packet, conn, clientAddr)

	}
}

func handleUDPConnection(packet *BACnetPacket, conn *net.UDPConn, addr *net.UDPAddr) {
	log.Printf("Packet Type: %#x | Request Type: %#x \n",
		packet.ServiceChoice, packet.PDUType)
	// Handle if Request is Un-confirmed service request
	if packet.PDUType == SERV_UN_CONFIRM_REQ {
		// Handle if Service Choice is WHO_IS
		if packet.ServiceChoice == CNCTX_CMD_WHO_IS {
			log.Println("Got Who-Is Request")
			iamPacket := PrepareIAmResponse(1223, 12)
			conn.WriteToUDP(iamPacket, addr)
		}
	}

	if packet.PDUType == SERV_CONFIRM_REQ {
		log.Println("Got one Confirmed Req")
	}

}

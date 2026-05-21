package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const serverAddr = "localhost:42069"

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatalf("could not open tcp port. error: %s\n", err.Error())
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("could not create connection. error: %s\n", err.Error())
	}
	defer conn.Close()

	fmt.Printf("Sending to %s. Type your message and press Enter to send. Press Ctrl+C to exit.\n", serverAddr)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("could not read line. error: %s\n", err)
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Printf("could not write line. error: %s\n", err)
		}
	}

}

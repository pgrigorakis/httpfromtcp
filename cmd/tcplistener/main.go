package main

import (
	"fmt"
	"log"
	"net"

	"github.com/pgrigorakis/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("could not open tcp port. error: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Server listening on", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("could not accept connection. error: %s\n", err.Error())
			return
		}
		fmt.Println("Connection has been accepted from", conn.RemoteAddr())

		parsedRequest, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("couldn't parse request. error: %s", err.Error())
			return
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", parsedRequest.RequestLine.Method)
		fmt.Printf("- Target: %s\n", parsedRequest.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", parsedRequest.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range parsedRequest.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed.")

	}
}

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		linesChan := getLinesChannel(conn)
		for line := range linesChan {
			fmt.Println(line)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed.")

	}
}

func getLinesChannel(conn io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer conn.Close()
		currentLine := ""
		for {
			data := make([]byte, 8)
			n, err := conn.Read(data)
			if err != nil {
				if currentLine != "" {
					lines <- currentLine
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			parts := strings.Split(string(data[:n]), "\n")
			for i := 0; i < len(parts)-1; i++ {
				currentLine += parts[i]
				lines <- currentLine
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()
	return lines
}

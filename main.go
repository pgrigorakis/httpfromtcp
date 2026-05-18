package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFile = "messages.txt"

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer f.Close()
		currentLine := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
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

func main() {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("could not open file %v: %v\n", inputFile, err)
	}

	for line := range getLinesChannel(file) {
		fmt.Printf("read: %s\n", line)
	}
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pgrigorakis/httpfromtcp/internal/request"
	"github.com/pgrigorakis/httpfromtcp/internal/response"
	"github.com/pgrigorakis/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		message := "<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>"
		w.WriteResponse(response.Status400, []byte(message))
	case "/myproblem":
		message := "<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>"
		w.WriteResponse(response.Status500, []byte(message))
	default:
		message := "<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>"
		w.WriteResponse(response.Status200, []byte(message))
	}
}

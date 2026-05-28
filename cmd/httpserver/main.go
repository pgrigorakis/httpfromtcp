package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pgrigorakis/httpfromtcp/internal/headers"
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
	target := req.RequestLine.RequestTarget
	switch {
	case target == "/yourproblem":
		message := "<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>"
		w.WriteResponse(response.Status400, []byte(message))
	case target == "/myproblem":
		message := "<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>"
		w.WriteResponse(response.Status500, []byte(message))
	case strings.HasPrefix(target, "/httpbin/"):
		proxyHandler(w, req)
	default:
		message := "<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>"
		w.WriteResponse(response.Status200, []byte(message))
	}
}

func proxyHandler(w *response.Writer, req *request.Request) {
	url := "https://httpbin.org/" + strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	res, err := http.Get(url)
	if err != nil {
		w.WriteResponse(response.Status500, []byte("upstream request failed"))
		return
	}
	defer res.Body.Close()

	w.WriteStatusLine(response.Status200)
	headers := headers.Headers{
		"Connection":        "close",
		"Content-Type":      "text/html",
		"Transfer-Encoding": "chunked",
	}
	w.WriteHeaders(headers)

	buf := make([]byte, 1024)
	for {
		n, err := res.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.WriteChunkedBody(buf[:n]); writeErr != nil {
				log.Printf("%s\n", writeErr)
				break
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("%s\n", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Printf("%s\n", err)
	}
}

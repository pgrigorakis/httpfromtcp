package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	case target == "/video":
		videoHandler(w)
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
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	h.Remove("Content-Length")
	w.WriteHeaders(h)

	buf := make([]byte, 1024)
	var body bytes.Buffer
	for {
		n, err := res.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.WriteChunkedBody(buf[:n]); writeErr != nil {
				log.Printf("%s\n", writeErr)
				break
			}
			body.Write(buf[:n])
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

	trailers := headers.Headers{
		"X-Content-SHA256": fmt.Sprintf("%x", sha256.Sum256(body.Bytes())),
		"X-Content-Length": fmt.Sprintf("%d", body.Len()),
	}

	err = w.WriteTrailers(trailers)
	if err != nil {
		log.Printf("%s\n", err)
	}
}

func videoHandler(w *response.Writer) {
	video, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	w.WriteStatusLine(response.Status200)
	h := response.GetDefaultHeaders(0)
	h.Override("Content-Type", "video/mp4")
	h.Override("Content-Length", strconv.Itoa(len(video)))
	w.WriteHeaders(h)
	w.WriteBody(video)
}

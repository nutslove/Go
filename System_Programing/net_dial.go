package main

import (
	"io"
	"net"
	"net/http"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "example.com:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// io.WriteString(conn, "GET / HTTP/1.1\r\nHost: example.com:80\r\n\r\n")

	req, err := http.NewRequest("GET", "http://example.com", nil)
	req.Write(conn)
	io.Copy(os.Stdout, conn)
}

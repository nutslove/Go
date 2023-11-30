package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("r.URL.Path: %s\n", r.URL.Path)
		fmt.Printf("r.URL.RawPath: %s\n", r.URL.RawPath)
		fmt.Printf("r.URL.RawQuery: %s\n", r.URL.RawQuery)
		fmt.Printf("r.URL.Fragment: %s\n", r.URL.Fragment)
		fmt.Printf("r.URL.Host: %s\n", r.URL.Host)
		fmt.Printf("r.URL.User: %s\n", r.URL.User)
		fmt.Printf("r.URL.Opaque: %s\n", r.URL.Opaque)
		fmt.Printf("r.URL.Scheme: %s\n", r.URL.Scheme)
		fmt.Printf("r.URL.String(): %s\n", r.URL.String())
		fmt.Printf("r.URL.RequestURI(): %s\n", r.URL.RequestURI())
		fmt.Printf("r.URL.EscapedPath(): %s\n", r.URL.EscapedPath())
		fmt.Printf("r.URL.IsAbs(): %t\n", r.URL.IsAbs())
		fmt.Printf("r.URL.Hostname(): %s\n", r.URL.Hostname())
		fmt.Printf("r.URL.Port(): %s\n", r.URL.Port())
		fmt.Printf("r.method: %s\n", r.Method)
		fmt.Fprintf(w, "Hello, World")
	})
	http.ListenAndServe(":8080", nil) // nilは、DefaultServeMuxを指す
}

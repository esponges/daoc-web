// Command serve hosts the viewer over HTTP so the browser can fetch the
// converted zone data (file:// URLs block fetch of the binary heightmap).
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", "localhost:8777", "listen address")
	dir := flag.String("dir", "web", "directory to serve")
	flag.Parse()

	fs := http.FileServer(http.Dir(*dir))
	mux := http.NewServeMux()
	mux.Handle("/", noCache(fs))

	fmt.Printf("serving %s at http://%s/\n", *dir, *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// noCache keeps the browser from holding on to stale converter output.
func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}

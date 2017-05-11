package api

import (
	"net"
	"net/http"
)

var (
	listener net.Listener
	quit     chan struct{}
)

const (
	resourceDir = "dist/"
	devDir      = "dev/"
	indexPage   = "index.html"
)

func LaunchWebInterface(host string, g *Gateway) (e error) {
	quit = make(chan struct{})
	listener, e = net.Listen("tcp", host)
	if e != nil {
		return
	}
	serve(listener, NewServeMux(g), quit)
	return
}

func serve(listener net.Listener, mux *http.ServeMux, q chan struct{}) {
	go func() {
		for {
			if e := http.Serve(listener, mux); e != nil {
				select {
				case <-q:
					return
				default:
				}
				continue
			}
		}
	}()
}

// Shutdown closes the http service.
func Shutdown() {
	if quit != nil {
		// must close quit first
		close(quit)
		listener.Close()
		listener = nil
	}
}

// NewServeMux creates a http.ServeMux with handlers registered.
func NewServeMux(g *Gateway) *http.ServeMux {
	// Register objects.
	jsonAPI := NewJsonAPI(g)

	// Prepare mux.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/boards", jsonAPI.BoardListHandler)
	mux.HandleFunc("/api/boards/", jsonAPI.BoardHandler)
	return mux
}
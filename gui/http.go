package gui

import (
	"github.com/evanlinjin/bbs/cxo"
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

func LaunchWebInterface(host, staticDir string, g *cxo.Gateway) (e error) {
	quit = make(chan struct{})
	//appLoc, e := util.DetermineResourcePath(staticDir, resourceDir, devDir)
	//if e != nil {
	//	return e
	//}
	listener, e = net.Listen("tcp", host)
	if e != nil {
		return
	}
	serve(listener, NewJsonMux(g), quit)
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

// NewJsonMux creates a http.ServeMux with handlers registered.
func NewJsonMux(g *cxo.Gateway) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterApiHandlers(mux, g)
	return mux
}

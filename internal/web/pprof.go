package web

import (
	"net/http"
	pprofstd "net/http/pprof"
)

func newPprofMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprofstd.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprofstd.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprofstd.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprofstd.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprofstd.Trace)
	return mux
}

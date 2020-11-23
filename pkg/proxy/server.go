package proxy

import (
	"net/http"
)

// A Server is responsible to serve http requests and proxy them to the direct target
// or to another proxy based on the active Profile configuration
type Server struct {
	// Addr represents the Server address
	Addr  string
	proxy *proxy
}

// ListenAndServe calls the http.ListenAndServe function
// with the proxy handler
func (s *Server) ListenAndServe() error {
	if s.proxy == nil {
		s.proxy = newProxy()
	}
	return http.ListenAndServe(s.Addr, s.proxy)

}

// SetupProfile sets the active profile
func (s *Server) SetupProfile(profile *Profile) {
	if s.proxy == nil {
		s.proxy = newProxy()
	}
	s.proxy.SetProfile(profile)
}

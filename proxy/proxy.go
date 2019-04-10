package proxy

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

// Disclaimer: This part is mainly copied from the excellent https://github.com/elazarl/goproxy/
// all the credit goes to this project. I only made some adaptations to dynamically set (or not) an
// http proxy for CONNECT operations (used for HTTPS connections). I was not able to do it with the
// original goproxy library.

// A proxy is responsible to handle requests and to forward them to the right proxy or directly
// to the requested site.
type proxy struct {
	Tr              *http.Transport
	profile         *Profile
	requestsCounter uint64
}

// SetProfile sets up the active profile
func (p *proxy) SetProfile(profile *Profile) {
	p.profile = profile
	p.Tr.Proxy = profile.chooseProxy
}

// newProxy creates a Proxy with a properly configured http.Transport
func newProxy() *proxy {
	return &proxy{
		Tr: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

var hasPort = regexp.MustCompile(`:\d+$`)

func removeProxyHeaders(r *http.Request) {
	r.RequestURI = "" // this must be reset when serving a request with the client
	// If no Accept-Encoding header exists, Transport will add the headers it can accept
	// and would wrap the response body with the relevant reader.
	r.Header.Del("Accept-Encoding")
	// curl can add that, see
	// https://jdebp.eu./FGA/web-proxy-connection-header.html
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")
	// Connection, Authenticate and Authorization are single hop Header:
	// http://www.w3.org/Protocols/rfc2616/rfc2616.txt
	// 14.10 Connection
	//   The Connection general-header field allows the sender to specify
	//   options that are desired for that particular connection and MUST NOT
	//   be communicated by proxies over further connections.
	r.Header.Del("Connection")
}

func copyHeaders(dst, src http.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := atomic.AddUint64(&p.requestsCounter, 1)
	logger := log.WithField("requestID", reqID)

	ctx := withLogger(r.Context(), logger)
	r = r.WithContext(ctx)
	//r.Header["X-Forwarded-For"] = w.RemoteAddr()
	if r.Method == "CONNECT" {
		p.handleHTTPS(w, r)
	} else {
		var err error
		if !r.URL.IsAbs() {
			http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", 500)
			return
		}

		removeProxyHeaders(r)
		resp, err := p.Tr.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		origBody := resp.Body
		defer origBody.Close()
		// http.ResponseWriter will take care of filling the correct response length
		// Setting it now, might impose wrong value, contradicting the actual new
		// body the user returned.
		// We keep the original body to remove the header only if things changed.
		// This will prevent problems with HEAD requests where there's no body, yet,
		// the Content-Length header should be set.
		if origBody != resp.Body {
			resp.Header.Del("Content-Length")
		}
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		nr, err := io.Copy(w, resp.Body)
		if err := resp.Body.Close(); err != nil {
			logger.WithError(err).Warn("Can't close response body")
		}
		logger.WithField("bytes", nr).WithError(err).Trace("bytes copied to client")
	}

}

func httpError(logger *log.Entry, w io.WriteCloser, err error) {
	if _, err := io.WriteString(w, "HTTP/1.1 502 Bad Gateway\r\n\r\n"); err != nil {
		logger.WithError(err).Warn("Error responding to client")
	}
	if err := w.Close(); err != nil {
		logger.WithError(err).Warn("Error closing client connection")
	}
}

func copyOrWarn(logger *log.Entry, dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	if _, err := io.Copy(dst, src); err != nil {
		logger.WithError(err).Warn("Error copying to client")
	}
	wg.Done()
}

func copyAndClose(logger *log.Entry, dst, src *net.TCPConn) {
	if _, err := io.Copy(dst, src); err != nil {
		logger.WithError(err).Warn("Error copying to client")
	}

	dst.CloseWrite()
	src.CloseRead()
}

func (p *proxy) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	logger := ensureLoggerFromContext(r.Context())
	hij, ok := w.(http.Hijacker)
	if !ok {
		panic("httpserver does not support hijacking")
	}

	proxyClient, _, e := hij.Hijack()
	if e != nil {
		panic("Cannot hijack connection " + e.Error())
	}
	host := r.URL.Host

	if !hasPort.MatchString(host) {
		host += ":80"
	}
	targetSiteCon, err := p.profile.dial(r, "tcp", host)
	if err != nil {
		httpError(logger, proxyClient, err)
		return
	}
	logger.WithField("host", host).Trace("Accepting CONNECT to host")
	proxyClient.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	targetTCP, targetOK := targetSiteCon.(*net.TCPConn)
	proxyClientTCP, clientOK := proxyClient.(*net.TCPConn)
	if targetOK && clientOK {
		go copyAndClose(logger, targetTCP, proxyClientTCP)
		go copyAndClose(logger, proxyClientTCP, targetTCP)
	} else {
		go func() {
			var wg sync.WaitGroup
			wg.Add(2)
			go copyOrWarn(logger, targetSiteCon, proxyClient, &wg)
			go copyOrWarn(logger, proxyClient, targetSiteCon, &wg)
			wg.Wait()
			proxyClient.Close()
			targetSiteCon.Close()

		}()
	}

}

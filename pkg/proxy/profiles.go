package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	xproxy "golang.org/x/net/proxy"
)

// A Rule allows to match an URL pattern to a proxy URL
//
// The Proxy URL may be nil indicating that no proxy should be used (direct connection)
type Rule struct {
	Pattern string
	Proxy   *url.URL
}

// A Profile is set of Rules and a Default proxy URL if none of the rules match
//
// The Default proxy URL may be nil indicating that no proxy should be used (direct connection)
type Profile struct {
	Default *url.URL
	Rules   []Rule
}

func (p *Profile) chooseProxy(req *http.Request) (*url.URL, error) {
	for _, r := range p.Rules {
		hostname := stripPort(req.URL)
		logger := log.WithFields(log.Fields{
			"hostname": hostname,
			"pattern":  r.Pattern,
			"proxy":    r.Proxy,
		})
		logger.Trace("check matching hostname against rule pattern")
		rePattern := strings.Replace(r.Pattern, ".", `\.`, -1)
		rePattern = strings.Replace(rePattern, "*", ".*", -1)
		rePattern = "^" + rePattern + "$"
		if ok, err := regexp.MatchString(rePattern, hostname); err == nil && ok {
			logger.Trace("matched!")
			return r.Proxy, nil
		}
	}

	return p.Default, nil
}

// Modified from url/url.go credit goes to the Go team
func stripPort(hostportURL *url.URL) string {
	var hostport string
	if hostportURL.Host != "" {
		hostport = hostportURL.Host
	} else {
		hostport = hostportURL.String()
	}

	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	if i := strings.IndexByte(hostport, ']'); i != -1 {
		return strings.TrimPrefix(hostport[:i], "[")
	}
	return hostport[:colon]
}

func (p *Profile) dial(r *http.Request, network, addr string) (net.Conn, error) {
	proxy, err := p.chooseProxy(r)
	if err != nil {
		return nil, err
	}
	if proxy == nil {
		return net.Dial(network, addr)
	}

	switch proxy.Scheme {
	case "http", "https":
		return p.dialHTTP(proxy, network, addr)
	case "socks5":
		return p.dialSocks5(proxy, network, addr)
	default:
		return nil, fmt.Errorf("unsupported scheme %q for proxy %s", proxy.Scheme, proxy.String())
	}
}

func (p *Profile) dialSocks5(proxy *url.URL, network, addr string) (net.Conn, error) {
	d, err := xproxy.SOCKS5(network, proxy.Host, nil, nil)
	if err != nil {
		return nil, err
	}
	return d.Dial(network, addr)
}

func (p *Profile) dialHTTP(proxy *url.URL, network, addr string) (net.Conn, error) {
	c, err := net.Dial(network, proxy.Host)
	if err != nil {
		return nil, err
	}

	if proxy.Scheme == "https" {
		c = tls.Client(c, &tls.Config{InsecureSkipVerify: true})
	}

	connectReq := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}
	connectReq.Write(c)
	// Read response.
	// Okay to use and discard buffered reader here, because
	// TLS server will not speak until spoken to.
	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, connectReq)
	if err != nil {
		c.Close()
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		resp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		c.Close()
		return nil, errors.New("proxy refused connection" + string(resp))
	}
	return c, nil
}

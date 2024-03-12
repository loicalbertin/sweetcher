package proxy

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	expectedContent := `Hello that's all folks!`
	s := Server{Addr: ":9988"}
	s.SetupProfile(&Profile{
		Default: makeURL(t, "http://127.0.0.1:9989"),
		Rules: []Rule{
			{Pattern: "something.*", Proxy: makeURL(t, "http://127.0.0.1:9998")},
			{Pattern: "wrongproxy.*", Proxy: makeURL(t, "wrongproto://127.0.0.1:9998")},
		},
	})

	go s.ListenAndServe()

	s2 := Server{Addr: ":9989"}
	s2.SetupProfile(&Profile{
		Default: nil,
	})

	go s2.ListenAndServe()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedContent))
	}))

	ts2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedContent))
	}))

	defer ts.Close()

	proxyUrl, err := url.Parse("http://127.0.0.1:9988")
	if err != nil {
		t.Fatalf("failed to parse proxy URL: %v", err)
	}
	httpClient := &http.Client{Transport: &http.Transport{
		Proxy:           http.ProxyURL(proxyUrl),
		IdleConnTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	resp, err := httpClient.Get(ts.URL)
	if err != nil {
		t.Errorf("Use proxy in http mode: %v", err)
	}
	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Errorf("Use proxy in http mode, read response body error: %v", err)
	}
	if string(b) != expectedContent {
		t.Errorf("Use proxy in http mode, expected content %q received %q", expectedContent, b)
	}

	resp, err = httpClient.Get(ts2.URL)
	if err != nil {
		t.Errorf("Use proxy in https mode: %v", err)
	}
	b, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Errorf("Use proxy in https mode, read response body error: %v", err)
	}
	if string(b) != expectedContent {
		t.Errorf("Use proxy in https mode, expected content %q received %q", expectedContent, b)
	}

	resp, err = httpClient.Get("http://something.missing")
	if err != nil {
		t.Errorf("Use proxy in http mode with missing forward proxy: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Error("Use proxy in http mode with missing forward proxy: expecting an error")
	}

	resp, err = httpClient.Get("http://wrongproxy.com")
	if err != nil {
		t.Errorf("Use proxy in http mode with missing forward proxy: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Error("Use proxy in http mode with missing forward proxy: expecting an error")
	}

	<-time.After(60 * time.Second)
}

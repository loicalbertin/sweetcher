package proxy

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_proxyNotHalfClosableConnection(t *testing.T) {

	s, c := net.Pipe()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		c.Write([]byte("\r\n\r\n"))
		c.Close()
		wg.Done()
	}()

	go func() {
		_, err := io.ReadAll(s)
		assert.NilError(t, err)
		s.Close()
		wg.Done()
	}()

	proxyNotHalfClosableConnection(context.Background(), slog.Default(), c, s)
	wg.Wait()
}

type noopStrBuilderCloser struct {
	*strings.Builder
}

func (n noopStrBuilderCloser) Close() error {
	return nil
}

func Test_httpError(t *testing.T) {

	b := &noopStrBuilderCloser{new(strings.Builder)}
	httpError(b, errors.New("some error"))

	if !(strings.Contains(b.String(), "502 Bad Gateway") && strings.Contains(b.String(), "some error")) {
		t.Error("missing 502 error")
	}
}

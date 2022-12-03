package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// SequentialRoundTripper wraps a normal roundtripper by adding a mutex
// which forces all requests to be completed sequentially.
type SequentialRoundTripper struct {
	mu    sync.Mutex
	rt    http.RoundTripper
	sleep time.Duration
}

func NewSequentialRoundTripper(rt http.RoundTripper, sleep time.Duration) *SequentialRoundTripper {
	return &SequentialRoundTripper{
		rt:    rt,
		sleep: sleep,
	}
}

func (srt *SequentialRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	srt.mu.Lock()
	defer func() {
		time.Sleep(srt.sleep)
		srt.mu.Unlock()
	}()

	return srt.rt.RoundTrip(req)
}

// SequentialReverseProxy is like httputil.ReverseProxy but sequential using SequentialRoundTripper
type SequentialReverseProxy struct {
	from string
	to   *url.URL
	p    *httputil.ReverseProxy
	s    *http.Server
}

func NewSequentialReverseProxy(from string, to *url.URL, sleep time.Duration) *SequentialReverseProxy {
	p := httputil.NewSingleHostReverseProxy(to)
	p.Transport = NewSequentialRoundTripper(http.DefaultTransport, sleepDuration)
	return &SequentialReverseProxy{
		from: from,
		to:   to,
		p:    p,
	}
}

func (p *SequentialReverseProxy) Run() error {
	s := &http.Server{
		Addr:              p.from,
		Handler:           p.p,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
	p.s = s

	log.Printf("starting proxy from %s to %s", p.from, p.to)
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serving proxy from %s to %s: %w", p.from, p.to, err)
	}
	return nil
}

func (p *SequentialReverseProxy) Shutdown(ctx context.Context) error {
	if err := p.s.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down proxy from %s to %s: %w", p.from, p.to, err)
	}
	return nil
}

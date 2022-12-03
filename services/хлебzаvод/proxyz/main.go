package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	sleepDuration   = time.Millisecond * 10
	shutdownTimeout = time.Millisecond * 500
)

func main() {
	configs, err := LoadConfigs()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	proxies := make([]*SequentialReverseProxy, len(configs))
	for i, config := range configs {
		proxies[i] = NewSequentialReverseProxy(config.From, config.To, sleepDuration)
	}

	var wg sync.WaitGroup
	for _, proxy := range proxies {
		wg.Add(1)
		proxy := proxy

		go func() {
			defer wg.Done()
			if err := proxy.Run(); err != nil {
				log.Print(err)
			}
		}()
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	<-ch
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	for _, proxy := range proxies {
		proxy := proxy

		go func() {
			if err := proxy.Shutdown(shutdownCtx); err != nil {
				log.Print(err)
			}
		}()
	}

	wg.Wait()
}

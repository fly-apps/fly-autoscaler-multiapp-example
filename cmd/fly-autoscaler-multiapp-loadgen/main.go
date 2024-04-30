package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	clientN := flag.Int("n", 0, "number of clients")
	flag.Parse()
	if flag.NArg() == 0 {
		return fmt.Errorf("usage: fly-autoscaler-multiapp-loadgen -n N https://APPNAME.fly.dev")
	} else if flag.NArg() > 1 {
		return fmt.Errorf("too many arguments")
	}

	// Parse and build out connection URL.
	connectURL, err := url.Parse(flag.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	connectURL.Path = "/connect"

	// Initialize clients in separate goroutines.
	slog.Info("starting clients", slog.Int("n", *clientN))
	var wg sync.WaitGroup
	for i := 0; i < *clientN; i++ {
		i := i
		wg.Add(1)
		go func() { defer wg.Done(); monitorClient(ctx, i, connectURL.String()) }()
	}
	wg.Wait()

	return nil
}

func monitorClient(ctx context.Context, index int, connectURL string) {
	for {
		select {
		case <-ctx.Done():
		default:
			slog.Info("connecting", slog.Int("i", index))
			if err := connect(ctx, connectURL); err != nil {
				slog.Error("connection failed", slog.Any("err", err))
			}
		}
	}
}

func connect(ctx context.Context, connectURL string) error {
	req, err := http.NewRequest(http.MethodGet, connectURL, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if code := resp.StatusCode; code != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", code)
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("read http body: %w", err)
	}

	return nil
}

package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// getFreePort finds an available TCP port for testing.
func getFreePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			t.Logf("failed to close listener: %v", err)
		}
	}()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func TestStartServer_And_GracefulShutdown(t *testing.T) {
	// Simple handler
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("pong")) })

	port := getFreePort(t)

	srv := startServer(mux, port)

	// Wait briefly for server to start
	time.Sleep(100 * time.Millisecond)

	// Verify server responds
	resp, err := http.Get("http://127.0.0.1:" + port + "/ping")
	if err != nil {
		t.Fatalf("get /ping: %v", err)
	}
	_ = resp.Body.Close()

	// Fire graceful shutdown after sending a signal
	cleaned := make(chan struct{}, 1)
	go func() {
		gracefulShutdown(context.Background(), srv, func() { close(cleaned) })
	}()

	// Give goroutine time to set up signal.Notify
	time.Sleep(100 * time.Millisecond)

	// Send SIGTERM to trigger shutdown path
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)

	select {
	case <-cleaned:
		// ok
	case <-time.After(5 * time.Second):
		t.Fatal("graceful shutdown did not complete in time")
	}
}

package httpserver

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestServerStartAndShutdown(t *testing.T) {
	t.Parallel()

	server := New(Config{Address: "127.0.0.1:0"}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	if err := server.Start(); err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}

	var response *http.Response
	var err error
	for range 20 {
		response, err = http.Get("http://" + server.Addr())
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("http.Get() returned error: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll() returned error: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("body = %q, want %q", string(body), "ok")
	}

	if err := server.Reload(Config{
		Address:         server.Addr(),
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		IdleTimeout:     time.Second,
		ShutdownTimeout: time.Second,
	}); err != nil {
		t.Fatalf("Reload() returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() returned error: %v", err)
	}
}

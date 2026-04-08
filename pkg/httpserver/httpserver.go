package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

// AsyncSubmitter 描述可异步提交任务的窄接口。
type AsyncSubmitter interface {
	SubmitDefault(context.Context, func()) error
}

// Config 描述 HTTPServer 配置。
type Config struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// Server 提供最小 HTTP 服务生命周期控制能力。
type Server struct {
	mu       sync.RWMutex
	cfg      Config
	handler  http.Handler
	server   *http.Server
	listener net.Listener
	addr     string
	executor AsyncSubmitter
}

// New 创建一个新的 HTTPServer 包装器。
func New(cfg Config, handler http.Handler) *Server {
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 5 * time.Second
	}

	return &Server{
		cfg:     cfg,
		handler: handler,
	}
}

// Start 启动 HTTP 服务。
func (s *Server) Start() error {
	s.mu.Lock()
	if s.server != nil {
		s.mu.Unlock()
		return errors.New("http server already started")
	}

	listener, err := net.Listen("tcp", normalizeAddress(s.cfg.Address))
	if err != nil {
		s.mu.Unlock()
		return err
	}

	server := &http.Server{
		Addr:         listener.Addr().String(),
		Handler:      s.handler,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
		IdleTimeout:  s.cfg.IdleTimeout,
	}

	s.server = server
	s.listener = listener
	s.addr = listener.Addr().String()
	executor := s.executor
	s.mu.Unlock()

	serve := func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return
		}
	}

	if executor != nil {
		if err := executor.SubmitDefault(context.Background(), serve); err != nil {
			_ = listener.Close()
			s.mu.Lock()
			s.server = nil
			s.listener = nil
			s.addr = ""
			s.mu.Unlock()
			return err
		}
		return nil
	}

	go serve()
	return nil
}

// Shutdown 优雅关闭 HTTP 服务。
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	server := s.server
	timeout := s.cfg.ShutdownTimeout
	s.server = nil
	s.listener = nil
	s.addr = ""
	s.mu.Unlock()

	if server == nil {
		return nil
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	return server.Shutdown(ctx)
}

// Reload 更新运行时配置。
func (s *Server) Reload(cfg Config) error {
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 5 * time.Second
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil && cfg.Address != "" && normalizeAddress(cfg.Address) != s.addr {
		return errors.New("cannot reload listening address while server is running")
	}

	s.cfg = cfg
	if s.server != nil {
		s.server.ReadTimeout = cfg.ReadTimeout
		s.server.WriteTimeout = cfg.WriteTimeout
		s.server.IdleTimeout = cfg.IdleTimeout
	}

	return nil
}

// SetExecutor 设置可选的异步执行器。
func (s *Server) SetExecutor(executor AsyncSubmitter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executor = executor
}

// Addr 返回实际监听地址。
func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.addr
}

func normalizeAddress(address string) string {
	if address == "" {
		return "127.0.0.1:0"
	}
	return address
}

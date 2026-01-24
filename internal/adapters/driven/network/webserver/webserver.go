package webserver

import (
	"fmt"
	"net"
	"net/http"
)

type WebServerAdapter struct {
	isOn bool
	srv  *http.Server
}

func NewWebServer() *WebServerAdapter {
	return &WebServerAdapter{}
}

func (s *WebServerAdapter) Start(addr string, port int, handler http.Handler) error {
	if s.isOn {
		return nil
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return err
	}

	s.srv = &http.Server{Handler: handler}
	s.isOn = true

	go func() {
		if err := s.srv.Serve(ln); err != http.ErrServerClosed {
			fmt.Println("Web server error:", err)
			s.Stop()
		}
	}()
	return nil
}

func (s *WebServerAdapter) Stop() {
	if s.srv != nil {
		s.srv.Close()
		s.srv = nil
	}

	s.isOn = false
}

func (s *WebServerAdapter) IsRunning() bool {
	return s.isOn
}

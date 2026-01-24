package services

import (
	"net/http"
	"strconv"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
)

type WebServerService struct {
	config    ports.Configuration
	webServer ports.WebServer
}

func NewWebServerService(c ports.Configuration, w ports.WebServer) *WebServerService {
	return &WebServerService{
		config:    c,
		webServer: w,
	}
}

func (s *WebServerService) StartWebServer(serveStr string, onStop func()) {
	addr := ""
	if s.config.SrvLocalhostOnly().Get() {
		addr = "127.0.0.1"
	}
	port := s.config.SrvPort().Get()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(404)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(serveStr)))
		w.Write([]byte(serveStr))

		if s.config.AutoStopSrv().Get() {
			s.StopWebServer(onStop)
		}
	})

	if err := s.webServer.Start(addr, port, handler); err != nil {
		s.StopWebServer(onStop)
	}
}

func (s *WebServerService) StopWebServer(onStop func()) {
	s.webServer.Stop()
	if onStop != nil {
		onStop()
	}
}

func (s *WebServerService) IsWebServerRunning() bool {
	return s.webServer.IsRunning()
}

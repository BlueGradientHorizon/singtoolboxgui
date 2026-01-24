package downloader

import (
	"io"
	"net/http"
	"time"
)

type HttpDownloaderAdapter struct{}

func NewHttpDownloader() *HttpDownloaderAdapter {
	return &HttpDownloaderAdapter{}
}

func (d *HttpDownloaderAdapter) Download(url string, timeout time.Duration) (string, error) {
	client := &http.Client{Timeout: timeout * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

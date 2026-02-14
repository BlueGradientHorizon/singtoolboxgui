package ports

import (
	"context"
	"net/http"
	"time"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
)

// Driven

type Configuration interface {
	Subscriptions() domain.ConfigurationValue[[]domain.Subscription]
	SubscriptionDlTimeout() domain.ConfigurationValue[int]
	ProfilesFoundTotal() domain.ConfigurationValue[int]
	ProfilesDuplicatedTotal() domain.ConfigurationValue[int]
	ParsingErrorsTotal() domain.ConfigurationValue[int]
	ValidationErrorsTotal() domain.ConfigurationValue[int]
	WorkingProfilesTotal() domain.ConfigurationValue[int]
	BatchSize() domain.ConfigurationValue[int]
	DedupEnabled() domain.ConfigurationValue[bool]
	EnableBatches() domain.ConfigurationValue[bool]
	RecheckRounds() domain.ConfigurationValue[int]
	RoundTimeout() domain.ConfigurationValue[int]
	AutoStartSrv() domain.ConfigurationValue[bool]
	AutoStopSrv() domain.ConfigurationValue[bool]
	SrvPort() domain.ConfigurationValue[int]
	SrvLocalhostOnly() domain.ConfigurationValue[bool]
}

type Downloader interface {
	Download(url string, timeout time.Duration) (string, error)
}

type WebServer interface {
	Start(addr string, port int, handler http.Handler) error
	Stop()
	IsRunning() bool
}

type CoreAdapter interface {
	ValidateOutbound(outbound interface{}) error
	CreateInstance(ctx context.Context, outbounds interface{}) (interface{}, error)
	StartInstance(instance interface{}) error
	CloseInstance(instance interface{}) error
	GetOutbounds(instance interface{}) (interface{}, error)
	GetOutboundsCount(outbounds interface{}) int
	SliceOutbounds(outbounds interface{}, start, end int) interface{}
	BuildOutboundsFromResults(results interface{}) interface{}
	CreateLatencyTest(ctx context.Context, settings interface{}, outbounds interface{}) (interface{}, error)
	RunLatencyTest(test interface{}, resChan chan<- interface{})
}

// Driving

type SubscriptionsService interface {
	DownloadSubscriptions(timeout time.Duration, updateChans ...chan<- domain.DownloadSubscriptionResult)
	GetWorkingSubscriptionsProfiles(sortByLatency bool) []domain.Profile
	ExportWorkingProfiles() string
}

type TestService interface {
	ValidateSubscriptions() int
	GetTestParameters() domain.LatencyTestParameters
	RunLatencyTest(testCtx context.Context, updateChans ...chan<- domain.LatencyTestUpdate)
}

type WebServerService interface {
	StartWebServer(serveStr string, onStop func())
	StopWebServer(onStop func())
	IsWebServerRunning() bool
}

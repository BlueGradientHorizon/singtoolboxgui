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
	GetSupportedCores() []domain.CoreInfo
	SetActiveCore(coreType any) error
	// ParseProfile parses a connection URI into a ProxyProfile
	ParseProfile(connURI string) (*domain.ProxyProfile, error)
	// SetProfileTag sets a unique tag on the profile's config
	SetProfileTag(profile *domain.ProxyProfile, tag string)
	// ValidateOutbound validates a single outbound configuration
	ValidateOutbound(outbound any) error
	// CreateInstance creates a test runner instance with the given profiles
	CreateInstance(ctx context.Context, profiles []domain.ProxyProfile) (any, error)
	StartInstance(instance any) error
	CloseInstance(instance any) error
	GetOutbounds(instance any) (any, error)
	GetOutboundsCount(outbounds any) int
	SliceOutbounds(outbounds any, start, end int) any
	// BuildOutboundsFromResults builds outbounds from latency test results
	BuildOutboundsFromResults(results any) any
	CreateLatencyTest(ctx context.Context, settings domain.LatencyTestSettings, outbounds any) (any, error)
	RunLatencyTest(test any, resChan chan<- any)
	// FindProfileByTag finds a profile by its config tag
	FindProfileByTag(profiles []domain.ProxyProfile, tag string) *domain.ProxyProfile
	// CreateLatencyTestResultsMap creates a new results map for latency tests
	CreateLatencyTestResultsMap() any
	// AddToResultsMap adds a result to the results map
	AddToResultsMap(resultsMap any, profile domain.ProxyProfile, result any)
	// GetResultsCount returns the count of results in the results map
	GetResultsCount(resultsMap any) int
	// GetResultTag gets the tag from a latency test result
	GetResultTag(result any) string
	// GetResultDelay gets the delay from a latency test result
	GetResultDelay(result any) int32
	// GetResultError gets the error from a latency test result
	GetResultError(result any) error
	// MergeResultsMaps merges source map into destination
	MergeResultsMaps(dst, src any)
	// NewLatencyTestSettings creates default latency test settings
	NewLatencyTestSettings() domain.LatencyTestSettings
	// IterateResults iterates over results map and calls the callback for each result
	IterateResults(resultsMap any, callback func(profile domain.ProxyProfile, tag string, delay int32) bool)
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

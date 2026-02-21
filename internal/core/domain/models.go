package domain

import (
	"time"
)

type ConfigurationValue[T any] struct {
	Get     func() T
	Set     func(T)
	Default T
}

type Profile struct {
	URI     string `json:"uri"`
	Latency int    `json:"latency"`
}

// ProxyProfile represents a generic proxy profile with its configuration and connection URI
type ProxyProfile struct {
	Config  any    // Core-specific outbound configuration
	ConnURI string // Original connection URI
}

type Subscription struct {
	ID               string    `json:"id"`
	Note             string    `json:"note"`
	URL              string    `json:"url"`
	ParsingErrors    int       `json:"parsing_errors"`
	ValidationErrors int       `json:"validation_errors"`
	ProfilesURIs     []string  `json:"profiles"`
	WorkingProfiles  []Profile `json:"working_profiles"`
}

type DownloadSubscriptionResult struct {
	Index   int
	Success bool
}

type LatencyTestSettings struct {
	TestURL string
	Timeout time.Duration
}

type LatencyTestParameters struct {
	BatchSize  int
	Batches    int
	Rounds     int
	LTSettings LatencyTestSettings
}

type LatencyTestStatusUpdate int

const (
	LTStatusStarted LatencyTestStatusUpdate = iota
	LTStatusRunning
	LTStatusWaiting
	LTStatusFinished
)

type LatencyTestProgressUpdate struct {
	ProgressValue float64
}

type LatencyTestInfoUpdate struct {
	DeltaMode  bool
	BatchIndex int
	RoundIndex int
	Total      int
	Running    int
	Failed     int
	Succeeded  int
}

type LatencyTestUpdate struct {
	Status   LatencyTestStatusUpdate
	Progress *LatencyTestProgressUpdate
	Info     *LatencyTestInfoUpdate
}

type CoreInfo struct {
	Name    string
	Version string
	Type    any
}

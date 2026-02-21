package stb

import (
	"context"
	"fmt"
	"maps"

	"github.com/bluegradienthorizon/singtoolbox/core"
	"github.com/bluegradienthorizon/singtoolbox/parsers"
	"github.com/bluegradienthorizon/singtoolbox/testers"
	"github.com/bluegradienthorizon/singtoolbox/testrunner"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
)

// STBCore implements CoreAdapter using singtoolbox's testrunner API
type STBCore struct {
	instance  *testrunner.TestRunner
	outbounds []core.Outbound
	profiles  []domain.ProxyProfile
	coreType  testrunner.CoreType
}

// NewSTBCore creates a new STBCore adapter without a core type.
// SetActiveCore must be called before using the adapter.
func NewSTBCore() *STBCore {
	return &STBCore{}
}

// GetSupportedCores returns a list of all supported proxy cores
func (s *STBCore) GetSupportedCores() []domain.CoreInfo {
	cores := testrunner.GetSupportedCores()
	result := make([]domain.CoreInfo, len(cores))
	for i, c := range cores {
		result[i] = domain.CoreInfo{
			Name:    c.Name,
			Version: c.Version,
			Type:    c.Type,
		}
	}
	return result
}

// SetActiveCore sets the active core type to use for testing
func (s *STBCore) SetActiveCore(coreType any) error {
	ct, ok := coreType.(testrunner.CoreType)
	if !ok {
		return fmt.Errorf("invalid core type: expected testrunner.CoreType")
	}
	s.coreType = ct
	return nil
}

// ParseProfile parses a connection URI into a ProxyProfile
func (s *STBCore) ParseProfile(connURI string) (*domain.ProxyProfile, error) {
	p, err := parsers.ParseProfile(connURI)
	if err != nil {
		return nil, err
	}
	return &domain.ProxyProfile{
		Config:  p.Config,
		ConnURI: p.ConnURI,
	}, nil
}

// SetProfileTag sets a unique tag on the profile's config
func (s *STBCore) SetProfileTag(profile *domain.ProxyProfile, tag string) {
	if config, ok := profile.Config.(*core.OutboundConfig); ok {
		config.Tag = tag
	}
}

// ValidateOutbound validates a single outbound configuration
func (s *STBCore) ValidateOutbound(outbound any) error {
	config, ok := outbound.(*core.OutboundConfig)
	if !ok {
		return fmt.Errorf("invalid outbound type: expected *core.OutboundConfig")
	}

	// Create a temporary runner to validate
	runner, err := testrunner.NewTestRunner(testrunner.TestRunnerConfig{
		CoreType:    s.coreType,
		LogLevel:    "panic",
		AutoCleanup: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create test runner for validation: %w", err)
	}
	defer runner.Close()

	// Create a profile with the config for validation
	profile := parsers.ProxyProfile{
		Config: config,
	}

	validationErrors, err := runner.CreateCore(context.Background(), []parsers.ProxyProfile{profile})
	if err != nil {
		return err
	}

	if len(validationErrors) > 0 {
		for errMsg := range validationErrors {
			return fmt.Errorf("validation error: %s", errMsg)
		}
	}

	runner.StopCore()
	return nil
}

// toSTBProfile converts a domain.ProxyProfile to parsers.ProxyProfile
func toSTBProfile(p domain.ProxyProfile) parsers.ProxyProfile {
	return parsers.ProxyProfile{
		Config:  p.Config.(*core.OutboundConfig),
		ConnURI: p.ConnURI,
	}
}

// toSTBProfiles converts []domain.ProxyProfile to []parsers.ProxyProfile
func toSTBProfiles(profiles []domain.ProxyProfile) []parsers.ProxyProfile {
	result := make([]parsers.ProxyProfile, len(profiles))
	for i, p := range profiles {
		result[i] = toSTBProfile(p)
	}
	return result
}

// CreateInstance creates a new test runner instance with the given profiles
func (s *STBCore) CreateInstance(ctx context.Context, profiles []domain.ProxyProfile) (any, error) {
	domainProfiles := profiles

	runner, err := testrunner.NewTestRunner(testrunner.TestRunnerConfig{
		CoreType:    s.coreType,
		LogLevel:    "panic",
		AutoCleanup: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create test runner: %w", err)
	}

	// Convert domain profiles to STB profiles
	stbProfiles := toSTBProfiles(domainProfiles)

	// Create core with profiles
	validationErrors, err := runner.CreateCore(ctx, stbProfiles)
	if err != nil {
		runner.Close()
		return nil, fmt.Errorf("failed to create core: %w", err)
	}

	// Log validation errors but don't fail
	if len(validationErrors) > 0 {
		for errMsg, count := range validationErrors {
			fmt.Printf("validation error (%d): %s\n", count, errMsg)
		}
	}

	s.instance = runner
	s.profiles = domainProfiles
	s.outbounds = runner.GetOutbounds()

	return runner, nil
}

// StartInstance starts the test runner's core
func (s *STBCore) StartInstance(instance any) error {
	runner, ok := instance.(*testrunner.TestRunner)
	if !ok {
		return fmt.Errorf("invalid instance type: expected *testrunner.TestRunner")
	}
	return runner.StartCore()
}

// CloseInstance closes the test runner instance
func (s *STBCore) CloseInstance(instance any) error {
	runner, ok := instance.(*testrunner.TestRunner)
	if !ok {
		return fmt.Errorf("invalid instance type: expected *testrunner.TestRunner")
	}
	runner.Close()
	return nil
}

// GetOutbounds returns the outbounds from the test runner
func (s *STBCore) GetOutbounds(instance any) (any, error) {
	runner, ok := instance.(*testrunner.TestRunner)
	if !ok {
		return nil, fmt.Errorf("invalid instance type: expected *testrunner.TestRunner")
	}
	return runner.GetOutbounds(), nil
}

// GetOutboundsCount returns the count of outbounds
func (s *STBCore) GetOutboundsCount(outbounds any) int {
	obs, ok := outbounds.([]core.Outbound)
	if !ok {
		return 0
	}
	return len(obs)
}

// SliceOutbounds returns a slice of outbounds
func (s *STBCore) SliceOutbounds(outbounds any, start, end int) any {
	obs, ok := outbounds.([]core.Outbound)
	if !ok {
		return nil
	}
	if end > len(obs) {
		end = len(obs)
	}
	return obs[start:end]
}

// BuildOutboundsFromResults builds outbounds slice from test results
func (s *STBCore) BuildOutboundsFromResults(results any) any {
	resultsMap, ok := results.(map[domain.ProxyProfile]testers.LatencyTestResult)
	if !ok {
		return []core.Outbound{}
	}

	var outbounds []core.Outbound
	for _, r := range resultsMap {
		if r.Error == nil && s.instance != nil {
			outbound, err := s.instance.GetCoreInstance().GetOutboundByTag(r.Tag)
			if err == nil {
				outbounds = append(outbounds, outbound)
			}
		}
	}
	return outbounds
}

// LatencyTest wraps the latency test data
type LatencyTest struct {
	runner    *testrunner.TestRunner
	outbounds []core.Outbound
	settings  testers.LatencyTestSettings
	ctx       context.Context
}

// CreateLatencyTest creates a latency test with the given settings and outbounds
func (s *STBCore) CreateLatencyTest(ctx context.Context, settings domain.LatencyTestSettings, outbounds any) (any, error) {
	lts := testers.LatencyTestSettings{
		TestURL: settings.TestURL,
		Timeout: settings.Timeout,
	}

	obs, ok := outbounds.([]core.Outbound)
	if !ok {
		return nil, fmt.Errorf("invalid outbounds type: expected []core.Outbound")
	}

	return &LatencyTest{
		runner:    s.instance,
		outbounds: obs,
		settings:  lts,
		ctx:       ctx,
	}, nil
}

// RunLatencyTest runs the latency test and sends results to the channel
func (s *STBCore) RunLatencyTest(test any, resChan chan<- any) {
	lt, ok := test.(*LatencyTest)
	if !ok {
		return
	}

	// Run the latency test round with progress callback
	results, err := lt.runner.RunLatencyTestRound(lt.ctx, lt.outbounds, lt.settings, 0, func(result testers.LatencyTestResult) {
		resChan <- result
	})
	if err != nil {
		return
	}

	// Send any remaining results that weren't sent via callback
	for _, r := range results {
		select {
		case resChan <- r:
		default:
		}
	}
}

// FindProfileByTag finds a profile by its config tag
func (s *STBCore) FindProfileByTag(profiles []domain.ProxyProfile, tag string) *domain.ProxyProfile {
	for i, p := range profiles {
		if config, ok := p.Config.(*core.OutboundConfig); ok {
			if config.Tag == tag {
				return &profiles[i]
			}
		}
	}
	return nil
}

// CreateLatencyTestResultsMap creates a new results map for latency tests
func (s *STBCore) CreateLatencyTestResultsMap() any {
	return make(map[domain.ProxyProfile]testers.LatencyTestResult)
}

// AddToResultsMap adds a result to the results map
func (s *STBCore) AddToResultsMap(resultsMap any, profile domain.ProxyProfile, result any) {
	rm, ok := resultsMap.(map[domain.ProxyProfile]testers.LatencyTestResult)
	if !ok {
		return
	}
	res, ok := result.(testers.LatencyTestResult)
	if !ok {
		return
	}
	rm[profile] = res
}

// GetResultsCount returns the count of results in the results map
func (s *STBCore) GetResultsCount(resultsMap any) int {
	rm, ok := resultsMap.(map[domain.ProxyProfile]testers.LatencyTestResult)
	if !ok {
		return 0
	}
	return len(rm)
}

// GetResultTag gets the tag from a latency test result
func (s *STBCore) GetResultTag(result any) string {
	res, ok := result.(testers.LatencyTestResult)
	if !ok {
		return ""
	}
	return res.Tag
}

// GetResultDelay gets the delay from a latency test result
func (s *STBCore) GetResultDelay(result any) int32 {
	res, ok := result.(testers.LatencyTestResult)
	if !ok {
		return 0
	}
	return res.Delay
}

// GetResultError gets the error from a latency test result
func (s *STBCore) GetResultError(result any) error {
	res, ok := result.(testers.LatencyTestResult)
	if !ok {
		return nil
	}
	return res.Error
}

// MergeResultsMaps merges source map into destination
func (s *STBCore) MergeResultsMaps(dst, src any) {
	dstMap, ok := dst.(map[domain.ProxyProfile]testers.LatencyTestResult)
	if !ok {
		return
	}
	srcMap, ok := src.(map[domain.ProxyProfile]testers.LatencyTestResult)
	if !ok {
		return
	}
	maps.Copy(dstMap, srcMap)
}

// NewLatencyTestSettings creates default latency test settings
func (s *STBCore) NewLatencyTestSettings() domain.LatencyTestSettings {
	lts := testers.NewLatencyTestSettings()
	return domain.LatencyTestSettings{
		TestURL: lts.TestURL,
		Timeout: lts.Timeout,
	}
}

// IterateResults iterates over results map and calls the callback for each result
func (s *STBCore) IterateResults(resultsMap any, callback func(profile domain.ProxyProfile, tag string, delay int32) bool) {
	rm, ok := resultsMap.(map[domain.ProxyProfile]testers.LatencyTestResult)
	if !ok {
		return
	}
	for p, r := range rm {
		if !callback(p, r.Tag, r.Delay) {
			break
		}
	}
}

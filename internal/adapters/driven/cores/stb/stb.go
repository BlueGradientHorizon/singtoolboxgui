package stb

import (
	"context"
	"fmt"

	"github.com/bluegradienthorizon/singtoolbox/core"
	"github.com/bluegradienthorizon/singtoolbox/parsers"
	"github.com/bluegradienthorizon/singtoolbox/testers"
	"github.com/bluegradienthorizon/singtoolbox/testrunner"
)

// STBCore implements CoreAdapter using singtoolbox's testrunner API
type STBCore struct {
	instance  *testrunner.TestRunner
	outbounds []core.Outbound
	profiles  []parsers.ProxyProfile
	coreType  testrunner.CoreType
}

// NewSTBCore creates a new STBCore adapter with the specified core type
func NewSTBCore(coreType testrunner.CoreType) *STBCore {
	return &STBCore{
		coreType: coreType,
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

// CreateInstance creates a new test runner instance with the given profiles
func (s *STBCore) CreateInstance(ctx context.Context, outbounds any) (any, error) {
	profiles, ok := outbounds.([]parsers.ProxyProfile)
	if !ok {
		return nil, fmt.Errorf("invalid outbounds type: expected []parsers.ProxyProfile")
	}

	runner, err := testrunner.NewTestRunner(testrunner.TestRunnerConfig{
		CoreType:    s.coreType,
		LogLevel:    "panic",
		AutoCleanup: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create test runner: %w", err)
	}

	// Create core with profiles
	validationErrors, err := runner.CreateCore(ctx, profiles)
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
	s.profiles = profiles
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
	resultsMap, ok := results.(map[parsers.ProxyProfile]testers.LatencyTestResult)
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
func (s *STBCore) CreateLatencyTest(ctx context.Context, settings any, outbounds any) (any, error) {
	lts, ok := settings.(testers.LatencyTestSettings)
	if !ok {
		return nil, fmt.Errorf("invalid settings type: expected testers.LatencyTestSettings")
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

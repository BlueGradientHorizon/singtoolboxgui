package singbox

import (
	"context"
	"fmt"

	"github.com/bluegradienthorizon/singtoolbox/parsers"
	"github.com/bluegradienthorizon/singtoolbox/testers"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
)

type SingBoxCore struct{}

func NewSingBoxCore() *SingBoxCore {
	return &SingBoxCore{}
}

func (s *SingBoxCore) ValidateOutbound(outbound interface{}) error {
	opt, ok := outbound.(*option.Outbound)
	if !ok {
		return fmt.Errorf("invalid outbound type")
	}
	ctx := include.Context(context.Background())
	instance, err := box.New(box.Options{
		Context: ctx,
		Options: option.Options{Outbounds: []option.Outbound{*opt}},
	})
	if err != nil {
		return err
	}
	instance.Close()
	return nil
}

func (s *SingBoxCore) CreateInstance(ctx context.Context, outbounds interface{}) (interface{}, error) {
	profiles, ok := outbounds.([]parsers.ProxyProfile)
	if !ok {
		return nil, fmt.Errorf("invalid outbounds type")
	}

	var optionOutbounds []option.Outbound
	for _, p := range profiles {
		optionOutbounds = append(optionOutbounds, *p.Outbound)
	}

	boxCtx := include.Context(ctx)
	opts := option.Options{
		Log:       &option.LogOptions{Level: "panic"},
		Outbounds: optionOutbounds,
	}

	instance, err := box.New(box.Options{Context: boxCtx, Options: opts})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (s *SingBoxCore) StartInstance(instance interface{}) error {
	boxInstance, ok := instance.(*box.Box)
	if !ok {
		return fmt.Errorf("invalid instance type")
	}
	return boxInstance.Start()
}

func (s *SingBoxCore) CloseInstance(instance interface{}) error {
	boxInstance, ok := instance.(*box.Box)
	if !ok {
		return fmt.Errorf("invalid instance type")
	}
	return boxInstance.Close()
}

func (s *SingBoxCore) GetOutbounds(instance interface{}) (interface{}, error) {
	boxInstance, ok := instance.(*box.Box)
	if !ok {
		return nil, fmt.Errorf("invalid instance type")
	}
	outbounds := boxInstance.Outbound().Outbounds()
	// Return as []adapter.Outbound
	return outbounds, nil
}

func (s *SingBoxCore) GetOutboundsCount(outbounds interface{}) int {
	obs, ok := outbounds.([]adapter.Outbound)
	if !ok {
		return 0
	}
	return len(obs)
}

func (s *SingBoxCore) SliceOutbounds(outbounds interface{}, start, end int) interface{} {
	obs, ok := outbounds.([]adapter.Outbound)
	if !ok {
		return nil
	}
	return obs[start:end]
}

func (s *SingBoxCore) ConvertToAdapterOutbounds(outbounds interface{}) interface{} {
	// This is a pass-through since the outbounds are already []adapter.Outbound
	return outbounds
}

func (s *SingBoxCore) BuildOutboundsFromResults(results interface{}) interface{} {
	resultsMap, ok := results.(map[parsers.ProxyProfile]testers.LatencyTestResult)
	if !ok {
		return []adapter.Outbound{}
	}
	var outbounds []adapter.Outbound
	for _, r := range resultsMap {
		outbounds = append(outbounds, r.Outbound)
	}
	return outbounds
}

func (s *SingBoxCore) CreateLatencyTest(ctx context.Context, settings interface{}, outbounds interface{}) (interface{}, error) {
	lts, ok := settings.(testers.LatencyTestSettings)
	if !ok {
		return nil, fmt.Errorf("invalid settings type")
	}
	obs, ok := outbounds.([]adapter.Outbound)
	if !ok {
		return nil, fmt.Errorf("invalid outbounds type")
	}
	return testers.NewLatencyTest(ctx, lts, obs)
}

func (s *SingBoxCore) RunLatencyTest(test interface{}, resChan chan<- interface{}) {
	lt, ok := test.(*testers.LatencyTest)
	if !ok {
		return
	}
	// Create a typed channel for the actual test
	typedChan := make(chan testers.LatencyTestResult, cap(resChan))

	// Forward results from typed channel to interface channel
	go func() {
		for res := range typedChan {
			resChan <- res
		}
	}()

	// Run the test (this will write to typedChan and return when done)
	lt.Run(typedChan)
}

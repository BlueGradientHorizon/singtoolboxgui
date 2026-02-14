package services

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/bluegradienthorizon/singtoolbox/parsers"
	"github.com/bluegradienthorizon/singtoolbox/testers"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/common"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
)

type TestService struct {
	config        ports.Configuration
	coreAdapter   ports.CoreAdapter
	validProfiles []parsers.ProxyProfile
}

func NewTestService(
	c ports.Configuration,
	ca ports.CoreAdapter,
) *TestService {
	return &TestService{
		config:      c,
		coreAdapter: ca,
	}
}

type validateSubscriptionResult struct {
	validProfiles    *[]parsers.ProxyProfile
	parsingErrors    int
	validationErrors int
}

func (s *TestService) validateSubscription(sub domain.Subscription) validateSubscriptionResult {
	res := validateSubscriptionResult{}
	// var sub *domain.Subscription = nil

	// for _, subSearch := range subs {
	// 	if subSearch.ID == ID {
	// 		sub = &subSearch
	// 	}
	// }
	// if sub == nil {
	// 	return nil, errors.New("ValidateSubscription: ID not found")
	// }

	var validProfiles []parsers.ProxyProfile
	var parsingErrors int
	var validationErrors int

	// Parse
	var subProfiles []parsers.ProxyProfile
	for _, u := range sub.ProfilesURIs {
		p, err := parsers.ParseProfile(u)
		if err != nil {
			sub.ParsingErrors++
			continue
		}
		subProfiles = append(subProfiles, *p)
	}
	parsingErrors += sub.ParsingErrors

	// Validate
	var validSubProfiles []parsers.ProxyProfile
	for i, p := range subProfiles {
		p.Outbound.Tag = fmt.Sprintf("%s-outbound-%d", sub.ID, i)
		err := s.coreAdapter.ValidateOutbound(p.Outbound)
		if err != nil {
			sub.ValidationErrors++
			continue
		}
		validSubProfiles = append(validSubProfiles, p)
	}
	validationErrors += sub.ValidationErrors
	validProfiles = append(validProfiles, validSubProfiles...)

	res.validProfiles = &validProfiles
	res.parsingErrors = parsingErrors
	res.validationErrors = validationErrors

	return res
}

func (s *TestService) ValidateSubscriptions() int {
	var validProfiles []parsers.ProxyProfile
	parsingErrorsTotal, validationErrorsTotal := 0, 0

	subs := s.config.Subscriptions().Get()
	for _, sub := range subs {
		p := s.validateSubscription(sub)
		validProfiles = append(validProfiles, *p.validProfiles...)
		parsingErrorsTotal += p.parsingErrors
		validationErrorsTotal += p.validationErrors
	}
	s.config.ParsingErrorsTotal().Set(parsingErrorsTotal)
	s.config.ValidationErrorsTotal().Set(validationErrorsTotal)

	s.validProfiles = validProfiles

	return len(validProfiles)
}

func (s *TestService) GetTestParameters() domain.LatencyTestParameters {
	batchSize := s.config.BatchSize().Get()

	var batches int
	if s.config.EnableBatches().Get() {
		batches = (len(s.validProfiles) + batchSize - 1) / batchSize
	} else {
		batches = 1
		batchSize = len(s.validProfiles)
	}

	rounds := s.config.RecheckRounds().Get()
	timeoutSec := s.config.RoundTimeout().Get()

	lts := testers.NewLatencyTestSettings()
	lts.Timeout = time.Duration(timeoutSec) * time.Second

	return domain.LatencyTestParameters{
		BatchSize: batchSize,
		Batches:   batches,
		Rounds:    rounds,
		LTSettings: domain.LatencyTestSettings{
			TestURL: lts.TestURL,
			Timeout: lts.Timeout,
		},
	}
}

func (s *TestService) RunLatencyTest(testCtx context.Context, updateChans ...chan<- domain.LatencyTestUpdate) {
	common.SendChans(domain.LatencyTestUpdate{
		Status: domain.LTStatusStarted,
		Progress: &domain.LatencyTestProgressUpdate{
			ProgressValue: float64(0),
		},
		Info: nil,
	}, updateChans...)
	// Sing-box Instance for Testing
	instance, err := s.coreAdapter.CreateInstance(testCtx, s.validProfiles)
	if err != nil {
		return
	}
	s.coreAdapter.StartInstance(instance)
	defer s.coreAdapter.CloseInstance(instance)

	tp := s.GetTestParameters()

	totalWorkingProfilesMap := make(map[parsers.ProxyProfile]testers.LatencyTestResult)

	lts := testers.NewLatencyTestSettings()
	lts.TestURL = tp.LTSettings.TestURL
	lts.Timeout = tp.LTSettings.Timeout

	instanceOutboundsRaw, err := s.coreAdapter.GetOutbounds(instance)
	if err != nil {
		return
	}

	for iB := range tp.Batches {
		start := iB * tp.BatchSize
		end := min(start+tp.BatchSize, s.coreAdapter.GetOutboundsCount(instanceOutboundsRaw))
		batchAdapterOutbounds := s.coreAdapter.SliceOutbounds(instanceOutboundsRaw, start, end)
		batchWorkingProfilesMap := make(map[parsers.ProxyProfile]testers.LatencyTestResult)

		for iR := range tp.Rounds {
			if testCtx.Err() != nil {
				break
			}

			if s.coreAdapter.GetOutboundsCount(batchAdapterOutbounds) == 0 {
				break
			}

			singleBatchValue := float64(1) / float64(tp.Batches)
			batchValue := float64(1) / float64(tp.Batches) * float64(iB)
			roundValue := float64(1) / float64(tp.Rounds) * float64(iR)
			curProgressValue := batchValue + singleBatchValue*roundValue
			// println(fmt.Sprintf("singleBatchValue %.2f", singleBatchValue))
			// println(fmt.Sprintf("batchValue %.2f", batchValue))
			// println(fmt.Sprintf("roundValue %.2f", roundValue))
			// println(fmt.Sprintf("curProgressValue %.2f", curProgressValue))

			common.SendChans(domain.LatencyTestUpdate{
				Status: domain.LTStatusWaiting,
				Progress: &domain.LatencyTestProgressUpdate{
					ProgressValue: curProgressValue,
				},
				Info: &domain.LatencyTestInfoUpdate{
					BatchIndex: iB,
					RoundIndex: iR,
					Total:      s.coreAdapter.GetOutboundsCount(batchAdapterOutbounds),
					Running:    s.coreAdapter.GetOutboundsCount(batchAdapterOutbounds),
					Succeeded:  0,
					Failed:     0,
				},
			}, updateChans...)

			resChan := make(chan interface{}, s.coreAdapter.GetOutboundsCount(batchAdapterOutbounds))
			// t1 := time.Now()
			lt, err := s.coreAdapter.CreateLatencyTest(testCtx, lts, batchAdapterOutbounds)
			// println("t1", time.Since(t1).Milliseconds())
			if err != nil {
				common.SendChans(domain.LatencyTestUpdate{
					Status:   domain.LTStatusWaiting,
					Progress: nil,
					Info: &domain.LatencyTestInfoUpdate{
						BatchIndex: iB,
						RoundIndex: iR,
						Total:      -1,
						Running:    -1,
						Succeeded:  -1,
						Failed:     -1,
					},
				}, updateChans...)
				continue
			}
			// t2 := time.Now()
			s.coreAdapter.RunLatencyTest(lt, resChan)
			// println("t2", time.Since(t2).Milliseconds())
			common.SendChans(domain.LatencyTestUpdate{
				Status:   domain.LTStatusRunning,
				Progress: nil,
				Info:     nil,
			}, updateChans...)

			roundWorkingProfilesMap := make(map[parsers.ProxyProfile]testers.LatencyTestResult)

			processed := 0
			for processed < s.coreAdapter.GetOutboundsCount(batchAdapterOutbounds) {
				resRaw := <-resChan
				res := resRaw.(testers.LatencyTestResult)
				processed++
				success, fail := 0, 1
				if res.Error == nil {
					success, fail = 1, 0
					idx := slices.IndexFunc(s.validProfiles, func(p parsers.ProxyProfile) bool {
						return p.Outbound.Tag == res.Tag
					})
					if idx != -1 {
						roundWorkingProfilesMap[s.validProfiles[idx]] = res
					}
				}
				common.SendChans(domain.LatencyTestUpdate{
					Status:   domain.LTStatusRunning,
					Progress: nil,
					Info: &domain.LatencyTestInfoUpdate{
						DeltaMode:  true,
						BatchIndex: iB,
						RoundIndex: iR,
						Running:    -1,
						Succeeded:  success,
						Failed:     fail,
					},
				}, updateChans...)
			}

			batchAdapterOutbounds = s.coreAdapter.BuildOutboundsFromResults(roundWorkingProfilesMap)
			batchWorkingProfilesMap = roundWorkingProfilesMap
			println(fmt.Sprintf("iR %d %d", iR, len(roundWorkingProfilesMap)))
		}

		maps.Copy(totalWorkingProfilesMap, batchWorkingProfilesMap)
	}

	subs := s.config.Subscriptions().Get()
	for i := range subs {
		subs[i].WorkingProfiles = nil
		for p, r := range totalWorkingProfilesMap {
			if strings.Contains(r.Tag, subs[i].ID) {
				subs[i].WorkingProfiles = append(subs[i].WorkingProfiles, domain.Profile{
					URI:     p.ConnURI,
					Latency: int(r.Delay),
				})
			}
		}
	}

	s.config.Subscriptions().Set(subs)
	s.config.WorkingProfilesTotal().Set(len(totalWorkingProfilesMap))

	common.SendChans(domain.LatencyTestUpdate{
		Status: domain.LTStatusFinished,
		Progress: &domain.LatencyTestProgressUpdate{
			ProgressValue: float64(1),
		},
		Info: nil,
	}, updateChans...)
}

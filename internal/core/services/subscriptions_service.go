package services

import (
	"cmp"
	"encoding/base64"
	"slices"
	"strings"
	"time"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/common"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
)

type SubscriptionsService struct {
	config     ports.Configuration
	downloader ports.Downloader
}

func NewSubscriptionsService(
	c ports.Configuration,
	d ports.Downloader,
) *SubscriptionsService {
	return &SubscriptionsService{
		config:     c,
		downloader: d,
	}
}

func (s *SubscriptionsService) DownloadSubscriptions(timeout time.Duration, updateChans ...chan<- domain.DownloadSubscriptionResult) {
	var profilesFound int
	var profilesDuplicated int

	subs := s.config.Subscriptions().Get()
	dedupEnabled := s.config.DedupEnabled().Get() // Check config

	for i := range subs {
		res := domain.DownloadSubscriptionResult{Index: i, Success: false}
		content, err := s.downloader.Download(subs[i].URL, timeout)
		if err != nil {
			common.SendChans(res, updateChans...)
			continue
		}

		if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
			content = string(decoded)
		} else if decoded, err := base64.RawStdEncoding.DecodeString(content); err == nil {
			content = string(decoded)
		}

		lines := strings.Split(content, "\n")
		var validURIs []string

		// Map to track unique URIs within this subscription
		seen := make(map[string]bool)

		for _, l := range lines {
			l = strings.TrimSpace(l)
			if strings.Contains(l, "://") && !strings.HasPrefix(l, "#") {

				// Deduplication Logic
				if dedupEnabled {
					key := l
					// Calculate comparison key: ignore '#' if it appears after '?'
					qIdx := strings.Index(key, "?")
					startSearch := 0
					if qIdx != -1 {
						startSearch = qIdx
					}

					// Find first '#' starting from after the query parameters
					if hIdx := strings.Index(key[startSearch:], "#"); hIdx != -1 {
						key = key[:startSearch+hIdx]
					}

					// If we have seen this key before, skip adding the real URI
					if seen[key] {
						profilesDuplicated++
						continue
					}
					seen[key] = true
				}

				validURIs = append(validURIs, l)
			}
		}
		subs[i].ProfilesURIs = validURIs
		profilesFound += len(validURIs)
		res.Success = true
		common.SendChans(res, updateChans...)
	}

	s.config.ProfilesFoundTotal().Set(profilesFound)
	s.config.ProfilesDuplicatedTotal().Set(profilesDuplicated)
	s.config.Subscriptions().Set(subs)
}

func (s *SubscriptionsService) GetWorkingSubscriptionsProfiles(sortByLatency bool) []domain.Profile {
	var wp []domain.Profile

	subs := s.config.Subscriptions().Get()
	for _, sub := range subs {
		for _, p := range sub.WorkingProfiles {
			wp = append(wp, p)
		}
	}

	if sortByLatency {
		slices.SortFunc(wp, func(a, b domain.Profile) int {
			return cmp.Compare(a.Latency, b.Latency)
		})
	}

	return wp
}

func (s *SubscriptionsService) ExportWorkingProfiles() string {
	p := s.GetWorkingSubscriptionsProfiles(true)
	ps := common.NewlineJoinedString(p, func(p domain.Profile) string { return p.URI })
	return ps
}

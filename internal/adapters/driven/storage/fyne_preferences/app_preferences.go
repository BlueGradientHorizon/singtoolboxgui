package fyne_preferences

import (
	"fmt"
	"reflect"

	"fyne.io/fyne/v2"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
)

type FynePreferences struct {
	PDedupEnabled          FynePreference[bool]
	PAutoStartSrv          FynePreference[bool]
	PAutoStopSrv           FynePreference[bool]
	PSrvLocalhostOnly      FynePreference[bool]
	PEnableBatches         FynePreference[bool]
	PProfilesFoundTotal    FynePreference[int]
	PParsingErrorsTotal    FynePreference[int]
	PValidationErrorsTotal FynePreference[int]
	PWorkingProfilesTotal  FynePreference[int]
	PSubscriptionDlTimeout FynePreference[int]
	PRecheckRounds         FynePreference[int]
	PRoundTimeout          FynePreference[int]
	PBatchSize             FynePreference[int]
	PSrvPort               FynePreference[int]
	PSubscriptions         FynePreference[[]domain.Subscription]
}

func NewFynePreferences(p fyne.Preferences) *FynePreferences {
	ap := &FynePreferences{
		PDedupEnabled:          NewFynePreference(p, "dedup_enabled", true),
		PAutoStartSrv:          NewFynePreference(p, "auto_start_srv", true),
		PAutoStopSrv:           NewFynePreference(p, "auto_stop_srv", true),
		PSrvLocalhostOnly:      NewFynePreference(p, "srv_localhost_only", true),
		PEnableBatches:         NewFynePreference(p, "enable_batches", true),
		PProfilesFoundTotal:    NewFynePreference(p, "profiles_found_total", -1),
		PParsingErrorsTotal:    NewFynePreference(p, "parsing_errors_total", -1),
		PValidationErrorsTotal: NewFynePreference(p, "validation_errors_total", -1),
		PWorkingProfilesTotal:  NewFynePreference(p, "working_profiles_total", -1),
		PSubscriptionDlTimeout: NewFynePreference(p, "subscription_dl_timeout", 10),
		PRecheckRounds:         NewFynePreference(p, "recheck_rounds", 3),
		PRoundTimeout:          NewFynePreference(p, "round_timeout", 15),
		PBatchSize:             NewFynePreference(p, "batch_size", 5000),
		PSrvPort:               NewFynePreference(p, "srv_port", 35240),
		PSubscriptions:         NewFynePreference(p, "subscriptions", []domain.Subscription{}),
	}

	ap.initAll()
	return ap
}

func (p *FynePreferences) initAll() {
	pElem := reflect.ValueOf(p).Elem()

	for i := 0; i < pElem.NumField(); i++ {
		f := pElem.Field(i)
		fName := pElem.Type().Field(i).Name
		if f.CanAddr() {
			ptr := f.Addr()
			if initializer, ok := ptr.Interface().(interface{ EnsureDefault() }); ok {
				initializer.EnsureDefault()
				println(fName)
				continue
			}
		}
		panic(fmt.Sprintf(
			"field %s (type %s) is likely not a kind of %s",
			fName,
			f.Type(),
			reflect.TypeFor[FynePreference[any]]().Name(),
		))
	}
}

// Implements [ports.Configuration]

func (p FynePreferences) Subscriptions() domain.ConfigurationValue[[]domain.Subscription] {
	return domain.ConfigurationValue[[]domain.Subscription]{
		Get:     p.PSubscriptions.Get,
		Set:     p.PSubscriptions.Set,
		Default: p.PSubscriptions.DefaultValue,
	}
}
func (p FynePreferences) SubscriptionDlTimeout() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PSubscriptionDlTimeout.Get,
		Set:     p.PSubscriptionDlTimeout.Set,
		Default: p.PSubscriptionDlTimeout.DefaultValue,
	}
}
func (p FynePreferences) ProfilesFoundTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PProfilesFoundTotal.Get,
		Set:     p.PProfilesFoundTotal.Set,
		Default: p.PProfilesFoundTotal.DefaultValue,
	}
}
func (p FynePreferences) ParsingErrorsTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PParsingErrorsTotal.Get,
		Set:     p.PParsingErrorsTotal.Set,
		Default: p.PParsingErrorsTotal.DefaultValue,
	}
}
func (p FynePreferences) ValidationErrorsTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PValidationErrorsTotal.Get,
		Set:     p.PValidationErrorsTotal.Set,
		Default: p.PValidationErrorsTotal.DefaultValue,
	}
}
func (p FynePreferences) WorkingProfilesTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PWorkingProfilesTotal.Get,
		Set:     p.PWorkingProfilesTotal.Set,
		Default: p.PWorkingProfilesTotal.DefaultValue,
	}
}
func (p FynePreferences) BatchSize() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PBatchSize.Get,
		Set:     p.PBatchSize.Set,
		Default: p.PBatchSize.DefaultValue,
	}
}
func (p FynePreferences) DedupEnabled() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PDedupEnabled.Get,
		Set:     p.PDedupEnabled.Set,
		Default: p.PDedupEnabled.DefaultValue,
	}
}
func (p FynePreferences) EnableBatches() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PEnableBatches.Get,
		Set:     p.PEnableBatches.Set,
		Default: p.PEnableBatches.DefaultValue,
	}
}
func (p FynePreferences) RecheckRounds() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PRecheckRounds.Get,
		Set:     p.PRecheckRounds.Set,
		Default: p.PRecheckRounds.DefaultValue,
	}
}
func (p FynePreferences) RoundTimeout() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PRoundTimeout.Get,
		Set:     p.PRoundTimeout.Set,
		Default: p.PRoundTimeout.DefaultValue,
	}
}
func (p FynePreferences) AutoStartSrv() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PAutoStartSrv.Get,
		Set:     p.PAutoStartSrv.Set,
		Default: p.PAutoStartSrv.DefaultValue,
	}
}
func (p FynePreferences) AutoStopSrv() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PAutoStopSrv.Get,
		Set:     p.PAutoStopSrv.Set,
		Default: p.PAutoStopSrv.DefaultValue,
	}
}
func (p FynePreferences) SrvPort() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PSrvPort.Get,
		Set:     p.PSrvPort.Set,
		Default: p.PSrvPort.DefaultValue,
	}
}
func (p FynePreferences) SrvLocalhostOnly() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PSrvLocalhostOnly.Get,
		Set:     p.PSrvLocalhostOnly.Set,
		Default: p.PSrvLocalhostOnly.DefaultValue,
	}
}

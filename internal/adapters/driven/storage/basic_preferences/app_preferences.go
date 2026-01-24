package basic_preferences

import (
	"fmt"
	"reflect"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
)

type BasicPreferences struct {
	PDedupEnabled            BasicPreference[bool]
	PAutoStartSrv            BasicPreference[bool]
	PAutoStopSrv             BasicPreference[bool]
	PSrvLocalhostOnly        BasicPreference[bool]
	PEnableBatches           BasicPreference[bool]
	PProfilesFoundTotal      BasicPreference[int]
	PProfilesDuplicatedTotal BasicPreference[int]
	PParsingErrorsTotal      BasicPreference[int]
	PValidationErrorsTotal   BasicPreference[int]
	PWorkingProfilesTotal    BasicPreference[int]
	PSubscriptionDlTimeout   BasicPreference[int]
	PRecheckRounds           BasicPreference[int]
	PRoundTimeout            BasicPreference[int]
	PBatchSize               BasicPreference[int]
	PSrvPort                 BasicPreference[int]
	PSubscriptions           BasicPreference[[]domain.Subscription]
}

func NewBasicPreferences(appID string) *BasicPreferences {
	p := NewJSONStore(appID)
	ap := &BasicPreferences{
		PDedupEnabled:            NewBasicPreference(p, "dedup_enabled", true),
		PAutoStartSrv:            NewBasicPreference(p, "auto_start_srv", true),
		PAutoStopSrv:             NewBasicPreference(p, "auto_stop_srv", true),
		PSrvLocalhostOnly:        NewBasicPreference(p, "srv_localhost_only", true),
		PEnableBatches:           NewBasicPreference(p, "enable_batches", true),
		PProfilesFoundTotal:      NewBasicPreference(p, "profiles_found_total", -1),
		PProfilesDuplicatedTotal: NewBasicPreference(p, "profiles_duplicated_total", -1),
		PParsingErrorsTotal:      NewBasicPreference(p, "parsing_errors_total", -1),
		PValidationErrorsTotal:   NewBasicPreference(p, "validation_errors_total", -1),
		PWorkingProfilesTotal:    NewBasicPreference(p, "working_profiles_total", -1),
		PSubscriptionDlTimeout:   NewBasicPreference(p, "subscription_dl_timeout", 10),
		PRecheckRounds:           NewBasicPreference(p, "recheck_rounds", 3),
		PRoundTimeout:            NewBasicPreference(p, "round_timeout", 15),
		PBatchSize:               NewBasicPreference(p, "batch_size", 5000),
		PSrvPort:                 NewBasicPreference(p, "srv_port", 35240),
		PSubscriptions:           NewBasicPreference(p, "subscriptions", []domain.Subscription{}),
	}

	ap.initAll()
	return ap
}

func (p *BasicPreferences) initAll() {
	pElem := reflect.ValueOf(p).Elem()

	for i := 0; i < pElem.NumField(); i++ {
		f := pElem.Field(i)
		fName := pElem.Type().Field(i).Name
		if f.CanAddr() {
			ptr := f.Addr()
			if initializer, ok := ptr.Interface().(interface{ EnsureDefault() }); ok {
				initializer.EnsureDefault()
				continue
			}
		}
		panic(fmt.Sprintf(
			"field %s (type %s) is likely not a kind of %s",
			fName,
			f.Type(),
			reflect.TypeFor[BasicPreference[any]]().Name(),
		))
	}
}

func (p BasicPreferences) Subscriptions() domain.ConfigurationValue[[]domain.Subscription] {
	return domain.ConfigurationValue[[]domain.Subscription]{
		Get:     p.PSubscriptions.Get,
		Set:     p.PSubscriptions.Set,
		Default: p.PSubscriptions.DefaultValue,
	}
}
func (p BasicPreferences) SubscriptionDlTimeout() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PSubscriptionDlTimeout.Get,
		Set:     p.PSubscriptionDlTimeout.Set,
		Default: p.PSubscriptionDlTimeout.DefaultValue,
	}
}
func (p BasicPreferences) ProfilesFoundTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PProfilesFoundTotal.Get,
		Set:     p.PProfilesFoundTotal.Set,
		Default: p.PProfilesFoundTotal.DefaultValue,
	}
}
func (p BasicPreferences) ProfilesDuplicatedTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PProfilesDuplicatedTotal.Get,
		Set:     p.PProfilesDuplicatedTotal.Set,
		Default: p.PProfilesDuplicatedTotal.DefaultValue,
	}
}
func (p BasicPreferences) ParsingErrorsTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PParsingErrorsTotal.Get,
		Set:     p.PParsingErrorsTotal.Set,
		Default: p.PParsingErrorsTotal.DefaultValue,
	}
}
func (p BasicPreferences) ValidationErrorsTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PValidationErrorsTotal.Get,
		Set:     p.PValidationErrorsTotal.Set,
		Default: p.PValidationErrorsTotal.DefaultValue,
	}
}
func (p BasicPreferences) WorkingProfilesTotal() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PWorkingProfilesTotal.Get,
		Set:     p.PWorkingProfilesTotal.Set,
		Default: p.PWorkingProfilesTotal.DefaultValue,
	}
}
func (p BasicPreferences) BatchSize() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PBatchSize.Get,
		Set:     p.PBatchSize.Set,
		Default: p.PBatchSize.DefaultValue,
	}
}
func (p BasicPreferences) DedupEnabled() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PDedupEnabled.Get,
		Set:     p.PDedupEnabled.Set,
		Default: p.PDedupEnabled.DefaultValue,
	}
}
func (p BasicPreferences) EnableBatches() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PEnableBatches.Get,
		Set:     p.PEnableBatches.Set,
		Default: p.PEnableBatches.DefaultValue,
	}
}
func (p BasicPreferences) RecheckRounds() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PRecheckRounds.Get,
		Set:     p.PRecheckRounds.Set,
		Default: p.PRecheckRounds.DefaultValue,
	}
}
func (p BasicPreferences) RoundTimeout() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PRoundTimeout.Get,
		Set:     p.PRoundTimeout.Set,
		Default: p.PRoundTimeout.DefaultValue,
	}
}
func (p BasicPreferences) AutoStartSrv() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PAutoStartSrv.Get,
		Set:     p.PAutoStartSrv.Set,
		Default: p.PAutoStartSrv.DefaultValue,
	}
}
func (p BasicPreferences) AutoStopSrv() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PAutoStopSrv.Get,
		Set:     p.PAutoStopSrv.Set,
		Default: p.PAutoStopSrv.DefaultValue,
	}
}
func (p BasicPreferences) SrvPort() domain.ConfigurationValue[int] {
	return domain.ConfigurationValue[int]{
		Get:     p.PSrvPort.Get,
		Set:     p.PSrvPort.Set,
		Default: p.PSrvPort.DefaultValue,
	}
}
func (p BasicPreferences) SrvLocalhostOnly() domain.ConfigurationValue[bool] {
	return domain.ConfigurationValue[bool]{
		Get:     p.PSrvLocalhostOnly.Get,
		Set:     p.PSrvLocalhostOnly.Set,
		Default: p.PSrvLocalhostOnly.DefaultValue,
	}
}

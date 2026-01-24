package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/fyne_ui/appwidgets"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
)

func numEntryContainerPref(label string, p domain.ConfigurationValue[int]) *fyne.Container {
	e := appwidgets.NewPreferenceNumericalEntry(p)
	return numEntryContainer(label, e)
}

func numEntryContainer(label string, e *appwidgets.NumericalEntry) *fyne.Container {
	return container.NewBorder(nil, nil, widget.NewLabel(label), nil, e)
}

func ShowSettings(a fyne.App, p ports.Configuration) {
	w := a.NewWindow("Settings")

	subDlTimeout := numEntryContainerPref("Subscription download timeout (s):", p.SubscriptionDlTimeout())
	dedup := appwidgets.NewPreferenceCheck("Perform deduplication", p.DedupEnabled())
	rounds := numEntryContainerPref("Recheck rounds:", p.RecheckRounds())
	timeout := numEntryContainerPref("Round timeout (s):", p.RoundTimeout())
	batches := appwidgets.NewPreferenceCheck("Test by batches", p.EnableBatches())
	batchSize := numEntryContainerPref("Batch size:", p.BatchSize())
	autoStart := appwidgets.NewPreferenceCheck("Auto-start web-server after test", p.AutoStartSrv())
	portEntry := appwidgets.NewPreferenceNumericalEntry(p.SrvPort())
	portEntry.Max = 65535
	port := numEntryContainer("Web-server port:", portEntry)
	autoStop := appwidgets.NewPreferenceCheck("Auto-stop web-server after first request", p.AutoStopSrv())
	localhostOnly := appwidgets.NewPreferenceCheck("Web-server listens only localhost", p.SrvLocalhostOnly())

	w.SetContent(container.NewVBox(
		subDlTimeout,
		dedup,
		rounds,
		timeout,
		batches,
		batchSize,
		autoStart,
		port,
		autoStop,
		localhostOnly,
	))
	w.Resize(fyne.NewSize(400, 300))
	w.Show()
}

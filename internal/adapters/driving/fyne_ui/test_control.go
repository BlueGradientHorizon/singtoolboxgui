package fyne_ui

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
)

func (g *GUI) StartTest() {
	if g.TestCtx != nil && g.TestCtxCancel != nil {
		return
	}

	fyne.Do(func() {
		g.BtnTest.SetText("PREPARING")
		g.BtnTest.Importance = widget.DangerImportance
		g.BtnTest.Disable()
		g.BtnTest.Refresh()
	})

	if g.TestService.ValidateSubscriptions() == 0 {
		return
	} else {
		g.UpdateStatsLabels()
	}

	tp := g.TestService.GetTestParameters()

	g.CleanStatsTables()
	g.AddStatsTables(tp.Batches, tp.Rounds)

	animValueOverride, animStart, animStop, animEnd := g.StartProgressBar(tp.LTSettings.Timeout * time.Duration(tp.Batches) * time.Duration(tp.Rounds))
	defer close(animValueOverride)
	defer close(animStart)
	defer close(animStop)
	defer close(animEnd)

	testCtx, testCtxCancel := context.WithCancel(context.Background())
	defer testCtxCancel()

	g.TestCtx = &testCtx
	g.TestCtxCancel = &testCtxCancel
	updateChan := make(chan domain.LatencyTestUpdate, 100)
	defer close(updateChan)
	go g.TestService.RunLatencyTest(*g.TestCtx, updateChan)

	type roundStats struct {
		Running   int
		Succeeded int
		Failed    int
		Total     int
	}

	type batchStats struct {
		rounds []roundStats
	}

	type testStats struct {
		batches []batchStats
	}

	stats := testStats{}
	stats.batches = make([]batchStats, tp.Batches)
	for b := range stats.batches {
		stats.batches[b].rounds = make([]roundStats, tp.Rounds)
	}

	lastSeenStats := make([][]roundStats, tp.Batches)
	for i := range lastSeenStats {
		lastSeenStats[i] = make([]roundStats, tp.Rounds)
	}

	statsTablesUpdateInterval := 250 * time.Millisecond
	stopStatsTablesUpdater := make(chan struct{}, 1)

	go func() {
		ticker := time.NewTicker(statsTablesUpdateInterval)
	T:
		for range ticker.C {
			for iB := range tp.Batches {
				for iR := range tp.Rounds {
					current := stats.batches[iB].rounds[iR]
					if current == lastSeenStats[iB][iR] {
						continue
					}
					g.StatsTables[iB].Stats[iR].Total.Set(stats.batches[iB].rounds[iR].Total)
					g.StatsTables[iB].Stats[iR].Running.Set(stats.batches[iB].rounds[iR].Running)
					g.StatsTables[iB].Stats[iR].Failed.Set(stats.batches[iB].rounds[iR].Failed)
					g.StatsTables[iB].Stats[iR].Succeeded.Set(stats.batches[iB].rounds[iR].Succeeded)
					lastSeenStats[iB][iR] = current
				}
			}
			select {
			case <-stopStatsTablesUpdater:
				break T
			default:
			}
		}
	}()

	oldLTStatus := domain.LTStatusWaiting
F:
	for {
		upd := <-updateChan

		if upd.Progress != nil {
			// println(fmt.Sprintf("g.ProgressBar.Value %.2f", g.ProgressBar.Value/g.ProgressBar.Max))
			select {
			case animValueOverride <- upd.Progress.ProgressValue:
			default:
			}
		}

		if upd.Info != nil {
			iB := upd.Info.BatchIndex
			iR := upd.Info.RoundIndex
			ttl := upd.Info.Total
			rnn := upd.Info.Running
			fld := upd.Info.Failed
			scd := upd.Info.Succeeded
			if upd.Info.DeltaMode {
				if ttl != 0 {
					stats.batches[iB].rounds[iR].Total += ttl
				}
				if rnn != 0 {
					stats.batches[iB].rounds[iR].Running += rnn
				}
				if fld != 0 {
					stats.batches[iB].rounds[iR].Failed += fld
				}
				if scd != 0 {
					stats.batches[iB].rounds[iR].Succeeded += scd
				}
			} else {
				stats.batches[iB].rounds[iR].Total = ttl
				stats.batches[iB].rounds[iR].Running = rnn
				stats.batches[iB].rounds[iR].Failed = fld
				stats.batches[iB].rounds[iR].Succeeded += scd
			}
		}

		if upd.Status == oldLTStatus {
			continue
		} else {
			oldLTStatus = upd.Status
		}

		switch upd.Status {
		case domain.LTStatusStarted:
			fyne.Do(func() {
				g.BtnTest.SetText("STOP")
				g.BtnTest.Importance = widget.DangerImportance
				g.BtnTest.Enable()
				g.BtnTest.Refresh()
			})
		case domain.LTStatusRunning:
			select {
			case animStart <- struct{}{}:
			default:
			}
		case domain.LTStatusWaiting:
			select {
			case animStop <- struct{}{}:
			default:
			}
		case domain.LTStatusFinished:
			select {
			case stopStatsTablesUpdater <- struct{}{}:
			default:
			}
			break F
		}
	}
}

func (g *GUI) StopTest() {
	if g.TestCtx != nil && g.TestCtxCancel != nil {
		(*g.TestCtxCancel)()
		g.TestCtx = nil
		g.TestCtxCancel = nil
	}
	fyne.Do(func() {
		g.BtnTest.SetText("TEST")
		g.BtnTest.Importance = widget.HighImportance
		g.BtnTest.Enable()
		g.BtnTest.Refresh()
	})
}

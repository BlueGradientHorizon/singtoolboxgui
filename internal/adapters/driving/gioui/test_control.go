package gioui

import (
	"context"
	"time"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/appwidgets"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
)

func (g *GUI) StartTest() {
	if g.TestCtx != nil && g.TestCtxCancel != nil {
		return
	}

	testCtx, testCtxCancel := context.WithCancel(context.Background())
	defer testCtxCancel()

	g.TestCtx = &testCtx
	g.TestCtxCancel = &testCtxCancel

	tp := g.TestService.GetTestParameters()

	g.CleanStatsTables()
	g.AddStatsTables(tp.Batches, tp.Rounds)

	updateChan := make(chan domain.LatencyTestUpdate, 100)
	defer close(updateChan)
	go g.TestService.RunLatencyTest(*g.TestCtx, updateChan)

	animValueOverride, animStart, animStop, animEnd := g.StartProgressBar(tp.LTSettings.Timeout * time.Duration(tp.Batches) * time.Duration(tp.Rounds))
	g.ProgressBarAnimEnd = animEnd
	defer close(animValueOverride)
	defer close(animStart)
	defer close(animStop)
	defer func() {
		close(animEnd)
		g.ProgressBarAnimEnd = nil
	}()

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
			updated := false
			for iB := range tp.Batches {
				for iR := range tp.Rounds {
					current := stats.batches[iB].rounds[iR]
					if current == lastSeenStats[iB][iR] {
						continue
					}

					rs := appwidgets.RoundStats{
						Total:     stats.batches[iB].rounds[iR].Total,
						Running:   stats.batches[iB].rounds[iR].Running,
						Failed:    stats.batches[iB].rounds[iR].Failed,
						Succeeded: stats.batches[iB].rounds[iR].Succeeded,
					}
					g.StatsTables[iB].UpdateRow(iR, rs)

					lastSeenStats[iB][iR] = current
					updated = true
				}
			}
			if updated {
				g.Window.Invalidate()
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
				stats.batches[iB].rounds[iR].Succeeded = scd
			}
		}

		if upd.Status == oldLTStatus {
			continue
		} else {
			oldLTStatus = upd.Status
		}

		switch upd.Status {
		case domain.LTStatusStarted:
			g.Window.Invalidate()
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
	if g.ProgressBarAnimEnd != nil {
		select {
		case g.ProgressBarAnimEnd <- struct{}{}:
		default:
		}
		g.ProgressBarAnimEnd = nil
	}
	g.Window.Invalidate()
}

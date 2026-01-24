package fyne_ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
)

func (g *GUI) ReadyProgressBarFormatter() string {
	return "Ready"
}

func (g *GUI) RunningProgressBarFormatter() string {
	remaining := g.ProgressBar.Max - g.ProgressBar.Value
	return fmt.Sprintf("%.1fs", remaining)
}

// animValueOverride, animStart, animStop, animEnd
func (g *GUI) StartProgressBar(duration time.Duration) (chan<- float64, chan<- struct{}, chan<- struct{}, chan<- struct{}) {
	g.ProgressBar.TextFormatter = g.RunningProgressBarFormatter
	g.ProgressBar.Value = 0
	fyne.Do(func() { g.ProgressBar.Refresh() })

	g.ProgressBar.Min = 0
	g.ProgressBar.Max = duration.Seconds()

	animValue := binding.NewFloat()
	animValueOverride := make(chan float64, 1)
	animStart := make(chan struct{}, 1)
	animStop := make(chan struct{}, 1)
	animEnd := make(chan struct{}, 1)

	running := false

	g.ProgressBar.Bind(animValue)

	go func() {
		start := time.Now()
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		defer g.StopProgressBar()
		defer g.ProgressBar.Unbind()

	T:
		for range ticker.C {
			select {
			case <-animEnd:
				break T
			default:
			}

			select {
			case <-animStop:
				running = false
			default:
			}

			select {
			case <-animStart:
				running = true
			default:
			}

			if !running {
				continue
			}

			select {
			case vOverride := <-animValueOverride:
				if vOverride >= 1.0 {
					animValue.Set(g.ProgressBar.Max)
					break T
				}

				elapsedNeeded := time.Duration(vOverride * float64(duration))
				start = time.Now().Add(-elapsedNeeded)
				animValue.Set(g.ProgressBar.Max - elapsedNeeded.Seconds())
			default:
			}

			elapsed := time.Since(start)
			v := elapsed.Seconds() / duration.Seconds()

			if v >= 1.0 {
				break
			}

			animValue.Set(v * g.ProgressBar.Max)
		}
	}()

	return animValueOverride, animStart, animStop, animEnd
}

func (g *GUI) StopProgressBar() {
	g.ProgressBar.TextFormatter = g.ReadyProgressBarFormatter
	fyne.Do(func() {
		g.ProgressBar.SetValue(g.ProgressBar.Min)
		g.ProgressBar.Refresh()
	})
}

package gioui

import (
	"fmt"
	"time"
)

func (g *GUI) ReadyProgressBarFormatter() string {
	return "Ready"
}

func (g *GUI) RunningProgressBarFormatter() string {
	remaining := g.ProgressBarMax - (g.ProgressBarVal * g.ProgressBarMax)
	return fmt.Sprintf("%.1fs", remaining)
}

func (g *GUI) StartProgressBar(duration time.Duration) (chan<- float64, chan<- struct{}, chan<- struct{}, chan<- struct{}) {
	g.ProgressBarFmt = g.RunningProgressBarFormatter
	g.ProgressBarVal = 0
	g.ProgressBarMax = float32(duration.Seconds())
	g.Window.Invalidate()

	animValueOverride := make(chan float64, 1)
	animStart := make(chan struct{}, 1)
	animStop := make(chan struct{}, 1)
	animEnd := make(chan struct{}, 1)

	running := false

	go func() {
		start := time.Now()
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		defer g.StopProgressBar()

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
					g.ProgressBarVal = 1.0
					g.Window.Invalidate()
					break T
				}
				elapsedNeeded := time.Duration(vOverride * float64(duration))
				start = time.Now().Add(-elapsedNeeded)
				g.ProgressBarVal = float32(1.0 - (elapsedNeeded.Seconds() / duration.Seconds()))
			default:
			}

			elapsed := time.Since(start)
			v := elapsed.Seconds() / duration.Seconds()

			if v >= 1.0 {
				g.ProgressBarVal = 1.0
				g.Window.Invalidate()
				break
			}

			g.ProgressBarVal = float32(v)
			g.Window.Invalidate()
		}
	}()

	return animValueOverride, animStart, animStop, animEnd
}

func (g *GUI) StopProgressBar() {
	g.ProgressBarFmt = g.ReadyProgressBarFormatter
	g.ProgressBarVal = 0
	g.Window.Invalidate()
}

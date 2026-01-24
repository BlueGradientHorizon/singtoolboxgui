package common

func SendChans[T any](v T, chans ...chan<- T) {
	for _, c := range chans {
		if c != nil {
			c <- v
		}
	}
}

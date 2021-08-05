package common

import "sync"

// MergeChannelsString multiplexes string channels into returned channel
func MergeChannelsString(cs ...<-chan string) <-chan string {
	out := make(chan string)
	wg := sync.WaitGroup{}
	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan string) {
			defer wg.Done()
			for n := range c {
				out <- n
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

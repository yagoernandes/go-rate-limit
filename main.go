package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yagoernandes/go-rate-limit/pkg/rate_limiter"
)

const (
	rate = 10
)

func main() {
	ctx := context.Background()
	limiter := rate_limiter.NewLimiter(ctx, rate)
	limiter.SetTick(200 * time.Millisecond)

	var wg sync.WaitGroup
	start := time.Now()
	counter := 0
	listPartialTimes := make([]time.Time, 0)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				// time.Sleep(200 * time.Millisecond)
				if err := limiter.Wait(context.Background()); err != nil {
					fmt.Println("err limiter:", err)
					return
				}
				counter++
				partial := time.Since(start)
				speedPerSecond := float64(counter) / partial.Seconds()

				listPartialTimes = append(listPartialTimes, time.Now())
				if len(listPartialTimes) > rate {
					listPartialTimes = listPartialTimes[1:]
				}
				partialInstant := time.Since(listPartialTimes[0])
				speedPerSecondInstant := float64(rate) / partialInstant.Seconds()

				fmt.Printf("counter %d time %v, speed avg: %.2f/s \t instant speed: %.2f/s \n", counter, time.Since(start), speedPerSecond, speedPerSecondInstant)
				if counter%rate == 0 {
					fmt.Printf("________________________________________ (%d) %v --- %v \n", counter, time.Since(start), time.Now())
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("final")
}

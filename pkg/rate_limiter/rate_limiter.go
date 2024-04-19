package rate_limiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	defaultInterval = 1 * time.Second
	defaultTick     = 5 * time.Millisecond
)

type Limiter interface {
	WaitN(ctx context.Context, qtd float64) error
	Wait(ctx context.Context) error
	SetMaxAccumulationInterval(d time.Duration) *limiter
	SetRate(rate float64) *limiter
	SetBurstMax(burstMax float64) *limiter
	SetInterval(interval time.Duration) *limiter
	SetTick(tick time.Duration) *limiter
	GetRate() float64
	GetBurstMax() float64
	GetInterval() time.Duration
	GetMaxAccumulationInterval() time.Duration
	GetTickets() float64
	GetLastLiberation() time.Time
	GetTick() time.Duration
}

type limiter struct {
	tickets                 float64
	rate                    float64
	burstMax                float64
	lastLiberation          time.Time
	interval                time.Duration
	maxAccumulationInterval time.Duration
	tick                    time.Duration
	mu                      sync.Mutex
}

func NewLimiter(
	ctx context.Context,
	rate float64,
) Limiter {
	l := limiter{
		rate:                    rate,
		tickets:                 rate,
		lastLiberation:          time.Now(),
		interval:                defaultInterval,
		maxAccumulationInterval: defaultInterval,
		burstMax:                rate,
		tick:                    defaultTick,
	}

	go l.liberate(ctx)
	return &l
}

func (l *limiter) liberate(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(l.interval)
			l.mu.Lock()

			elapsed := time.Since(l.lastLiberation)
			if elapsed > l.maxAccumulationInterval {
				elapsed = l.maxAccumulationInterval
			}
			l.tickets += l.rate * elapsed.Seconds()
			if l.tickets > l.burstMax {
				l.tickets += l.burstMax
			}
			l.lastLiberation = time.Now()
			l.mu.Unlock()
		}
	}
}

// WaitN waits for N tickets
func (l *limiter) WaitN(ctx context.Context, qtd float64) error {
	if qtd > l.burstMax {
		return errors.New("burst max exceeded")
	}

	if l.tickets >= qtd {
		l.mu.Lock()
		l.tickets -= qtd
		l.mu.Unlock()
		return nil
	}

	var counterTickets float64 = 0

	// wait for more tickets
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(l.tick)
			if l.tickets >= (qtd - counterTickets) {
				l.mu.Lock()
				l.tickets -= (qtd - counterTickets)
				l.mu.Unlock()
				return nil
			} else {
				l.mu.Lock()
				counterTickets = counterTickets + l.tickets
				l.tickets = 0
				l.mu.Unlock()
			}
		}
	}
}

// Wait waits for one ticket
func (l *limiter) Wait(ctx context.Context) error {
	return l.WaitN(ctx, 1)
}

// Setters

// SetMaxAccumulationInterval sets the maximum accumulation interval
func (l *limiter) SetMaxAccumulationInterval(d time.Duration) *limiter {
	l.mu.Lock()
	l.maxAccumulationInterval = d
	l.mu.Unlock()
	return l
}

// SetRate sets the rate
func (l *limiter) SetRate(rate float64) *limiter {
	l.mu.Lock()
	l.rate = rate
	l.mu.Unlock()
	return l
}

// SetBurstMax sets the burst max
func (l *limiter) SetBurstMax(burstMax float64) *limiter {
	l.mu.Lock()
	l.burstMax = burstMax
	l.mu.Unlock()
	return l
}

// SetInterval sets the interval
func (l *limiter) SetInterval(interval time.Duration) *limiter {
	l.mu.Lock()
	l.interval = interval
	l.mu.Unlock()
	return l
}

func (l *limiter) SetTick(tick time.Duration) *limiter {
	l.mu.Lock()
	l.tick = tick
	l.mu.Unlock()
	return l
}

// Getters

// GetRate returns the rate
func (l *limiter) GetRate() float64 {
	return l.rate
}

// GetBurstMax returns the burst max
func (l *limiter) GetBurstMax() float64 {
	return l.burstMax
}

// GetInterval returns the interval
func (l *limiter) GetInterval() time.Duration {
	return l.interval
}

// GetMaxAccumulationInterval returns the maximum accumulation interval
func (l *limiter) GetMaxAccumulationInterval() time.Duration {
	return l.maxAccumulationInterval
}

// GetTickets returns the tickets
func (l *limiter) GetTickets() float64 {
	return l.tickets
}

// GetLastLiberation returns the last liberation
func (l *limiter) GetLastLiberation() time.Time {
	return l.lastLiberation
}

// GetTick returns the tick
func (l *limiter) GetTick() time.Duration {
	return l.tick
}

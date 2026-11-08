package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func newTestCounter(now *time.Time) *DailyCounter {
	c := NewDailyCounter()
	c.now = func() time.Time { return *now }
	return c
}

func TestDailyCounter_BlocksAfterLimit(t *testing.T) {
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	c := newTestCounter(&now)

	for i := range 3 {
		if !c.Allow(Rule{Key: "user:1", Limit: 3}) {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}
	if c.Allow(Rule{Key: "user:1", Limit: 3}) {
		t.Fatal("4th call should be blocked")
	}
	if !c.Allow(Rule{Key: "user:2", Limit: 3}) {
		t.Fatal("other key should not be affected")
	}
}

func TestDailyCounter_ResetsOnNewUTCDay(t *testing.T) {
	now := time.Date(2026, 11, 8, 23, 59, 0, 0, time.UTC)
	c := newTestCounter(&now)

	if !c.Allow(Rule{Key: "k", Limit: 1}) {
		t.Fatal("first call should be allowed")
	}
	if c.Allow(Rule{Key: "k", Limit: 1}) {
		t.Fatal("second call on same day should be blocked")
	}

	now = now.Add(2 * time.Minute)
	if !c.Allow(Rule{Key: "k", Limit: 1}) {
		t.Fatal("call on next UTC day should be allowed")
	}
}

func TestDailyCounter_DoesNotCountWhenAnyRuleBlocks(t *testing.T) {
	now := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)
	c := newTestCounter(&now)

	if !c.Allow(Rule{Key: "global", Limit: 1}) {
		t.Fatal("first global call should be allowed")
	}
	// global は上限に達しているので user は増えないこと
	if c.Allow(Rule{Key: "user:1", Limit: 1}, Rule{Key: "global", Limit: 1}) {
		t.Fatal("should be blocked by global rule")
	}
	if c.counts["user:1"] != 0 {
		t.Fatalf("user count = %d, want 0", c.counts["user:1"])
	}
}

func TestDailyCounter_ZeroLimitIsUnlimited(t *testing.T) {
	now := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)
	c := newTestCounter(&now)
	for range 100 {
		if !c.Allow(Rule{Key: "k", Limit: 0}) {
			t.Fatal("zero limit should never block")
		}
	}
}

func TestDailyCounter_ConcurrentCallsRespectLimit(t *testing.T) {
	now := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)
	c := newTestCounter(&now)

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	for range 50 {
		wg.Go(func() {
			if c.Allow(Rule{Key: "k", Limit: 10}) {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	if allowed != 10 {
		t.Fatalf("allowed = %d, want 10", allowed)
	}
}

// Package ratelimit はキーごとの回数を UTC の日付単位で数える簡易レートリミッタを提供する。
//
// v0.0.1 は API が単一インスタンスで動く前提のインメモリ実装。
// 複数インスタンス構成にする時は DB / Redis などの共有ストアに置き換える。
package ratelimit

import (
	"sync"
	"time"
)

// Rule は 1 つのキーと、そのキーに許す 1 日あたりの上限回数。
// Limit が 0 以下のルールは無制限として扱う。
type Rule struct {
	Key   string
	Limit int
}

// DailyCounter はキーごとの当日の利用回数を保持する。
type DailyCounter struct {
	mu     sync.Mutex
	day    string
	counts map[string]int
	now    func() time.Time
}

func NewDailyCounter() *DailyCounter {
	return &DailyCounter{counts: map[string]int{}, now: time.Now}
}

// Allow はすべてのルールが上限未満なら全キーを 1 増やして true を返す。
// どれか 1 つでも上限に達していれば何も増やさず false を返す。
func (c *DailyCounter) Allow(rules ...Rule) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	today := c.now().UTC().Format(time.DateOnly)
	if c.day != today {
		c.day = today
		c.counts = map[string]int{}
	}

	for _, r := range rules {
		if r.Limit > 0 && c.counts[r.Key] >= r.Limit {
			return false
		}
	}
	for _, r := range rules {
		if r.Limit > 0 {
			c.counts[r.Key]++
		}
	}
	return true
}

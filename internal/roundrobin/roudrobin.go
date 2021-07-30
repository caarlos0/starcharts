package roundrobin

import "sync/atomic"

type RoundRobin struct {
	items []string
	next  int64
}

func New(items []string) *RoundRobin {
	return &RoundRobin{items, 0}
}

func (rr *RoundRobin) Pick() string {
	idx := atomic.LoadInt64(&rr.next)
	atomic.StoreInt64(&rr.next, (idx+1)%int64(len(rr.items)))
	return rr.items[rr.next]
}

type Token struct {
	token string
}

func (t *Token) String() string {
	return t.token
}

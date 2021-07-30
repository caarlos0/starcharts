// Package roundrobin provides round robin invalidation-aware load balancing of github tokens.
package roundrobin

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/apex/log"
)

// RoundRobiner can pick a token from a list of tokens.
type RoundRobiner interface {
	Pick() (*Token, error)
}

// New round robin implementation with the given list of tokens.
func New(tokens []string) RoundRobiner {
	log.Debugf("creating round robin with %d tokens", len(tokens))
	if len(tokens) == 0 {
		return &noTokensRoundRobin{}
	}
	result := make([]*Token, 0, len(tokens))
	for _, item := range tokens {
		result = append(result, NewToken(item))
	}
	return &realRoundRobin{tokens: result}
}

type realRoundRobin struct {
	tokens []*Token
	next   int64
}

func (rr *realRoundRobin) Pick() (*Token, error) {
	return rr.doPick(0)
}

func (rr *realRoundRobin) doPick(try int) (*Token, error) {
	if try > len(rr.tokens) {
		return nil, fmt.Errorf("no valid tokens left")
	}
	idx := atomic.LoadInt64(&rr.next)
	atomic.StoreInt64(&rr.next, (idx+1)%int64(len(rr.tokens)))
	if pick := rr.tokens[idx]; pick.OK() {
		log.Debugf("picked %s", pick.Key())
		return pick, nil
	}
	return rr.doPick(try + 1)
}

type noTokensRoundRobin struct{}

func (rr *noTokensRoundRobin) Pick() (*Token, error) {
	return nil, nil
}

// Token is a github token.
type Token struct {
	token string
	valid bool
	lock  sync.RWMutex
}

// NewToken from its string representation.
func NewToken(token string) *Token {
	return &Token{
		token: token,
		valid: true,
	}
}

// String returns the last 3 chars for the token.
func (t *Token) String() string {
	return t.token[len(t.token)-3:]
}

// Key returns the actual token.
func (t *Token) Key() string {
	return t.token
}

// OK returns true if the token is valid.
func (t *Token) OK() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.valid
}

// Invalidate invalidates the token.
func (t *Token) Invalidate() {
	log.Warnf("invalidated token '...%s'", t)
	t.lock.Lock()
	defer t.lock.Unlock()
	t.valid = false
}

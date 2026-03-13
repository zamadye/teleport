/*
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package limiter

import (
	"cmp"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"

	"github.com/gravitational/teleport/lib/limiter/internal/ratelimit"
	"github.com/gravitational/teleport/lib/utils"
)

// RateLimiter controls connection rate using the token bucket algorithm.
// See: https://en.wikipedia.org/wiki/Token_bucket
type RateLimiter struct {
	*ratelimit.TokenLimiter
	rateLimits *utils.FnCache
	mu         sync.Mutex
	rates      *ratelimit.RateSet
	clock      clockwork.Clock
}

// Rate defines connection rate
type Rate struct {
	Period  time.Duration
	Average int64
	Burst   int64
}

// NewRateLimiter returns new request rate limiter.
func NewRateLimiter(config Config) (*RateLimiter, error) {
	limiter := RateLimiter{}

	limiter.rates = ratelimit.NewRateSet()
	for _, rate := range config.Rates {
		err := limiter.rates.Add(rate.Period, rate.Average, rate.Burst)
		if err != nil {
			return nil, trace.Wrap(err)
		}
	}

	if config.Clock == nil {
		config.Clock = clockwork.NewRealClock()
	}

	limiter.clock = config.Clock

	var err error
	limiter.TokenLimiter, err = ratelimit.New(ratelimit.TokenLimiterConfig{
		Clock: config.Clock,
		Rates: limiter.rates,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	limiter.rateLimits, err = utils.NewFnCache(utils.FnCacheConfig{
		// The default TTL here is not super important because we set
		// the TTL explicitly for each entry we insert.
		TTL:   10 * time.Second,
		Clock: config.Clock,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &limiter, nil
}

// rateLimitKey combines a client token with the rate period so that
// different rate configurations maintain independent buckets for the
// same client IP. Without this, a call with a custom rate and a call
// with the default rate would share a single TokenBucketSet, and
// Update would constantly swap bucket periods - effectively making
// per-endpoint custom rates useless.
type rateLimitKey struct {
	token     string
	maxPeriod time.Duration
}

// IsRateLimited checks if the provided token is currently
// rate-limited under the default rates, without consuming any
// tokens. It returns true if the token would be rate-limited and
// false otherwise.
func (l *RateLimiter) IsRateLimited(token string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := rateLimitKey{token: token, maxPeriod: l.rates.MaxPeriod()}
	bucketSet, ok := l.rateLimits.GetIfExists(key)
	if !ok {
		return false
	}
	bucket, ok := bucketSet.(*ratelimit.TokenBucketSet)
	if !ok {
		return false
	}
	return bucket.IsRateLimited()
}

// RegisterRequest increases number of requests for the provided token.
// If neither the default rates nor a custom rate are configured, the
// request passes through without any limiting.
func (l *RateLimiter) RegisterRequest(token string, customRate *ratelimit.RateSet) error {
	// cmp.Or returns the first non-zero value. A non-nil RateSet
	// with zero entries is still selected over l.rates, which is
	// fine because Len() == 0 catches that case below.
	rate := cmp.Or(customRate, l.rates)
	if rate.Len() == 0 {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// We set the TTL as 10 times the rate period. E.g. if rate is 100 requests/second
	// per client IP, the counters for this IP will expire after 10 seconds of inactivity.
	ttl := rate.MaxPeriod()*10 + 1
	key := rateLimitKey{token: token, maxPeriod: rate.MaxPeriod()}
	bucketSet, err := utils.FnCacheGetWithTTL(context.TODO(), l.rateLimits, key, ttl,
		func(ctx context.Context) (*ratelimit.TokenBucketSet, error) {
			return ratelimit.NewTokenBucketSet(rate, l.clock), nil
		},
	)
	if err != nil {
		return trace.Wrap(err)
	}

	bucketSet.Update(rate)

	delay, err := bucketSet.Consume(1)
	if err != nil {
		return err
	}
	if delay > 0 {
		return trace.LimitExceeded("rate limit exceeded, try again in %v", delay)
	}
	return nil
}

// WrapHandle wraps the given HTTP handler with the rate limiter.
func (l *RateLimiter) WrapHandle(h http.Handler) {
	l.TokenLimiter.Wrap(h)
}

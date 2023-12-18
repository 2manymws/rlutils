package rlutils

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/2manymws/rl"
	"github.com/2manymws/rl/counter"
)

type BaseLimiter struct {
	reqLimit         int `mapstructure:"req_limit"`
	windowLen        time.Duration
	targetExtensions []string
	onRequestLimit   func(*rl.Context, string) http.HandlerFunc
	rl.Counter
}

func NewBaseLimiter(
	reqLimit int,
	windowLen time.Duration,
	targetExtensions []string,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
) BaseLimiter {
	ttl := windowLen * 2 // 最低2回分のウィンドウ分のカウンタを維持する
	return BaseLimiter{
		reqLimit:         reqLimit,
		windowLen:        windowLen,
		Counter:          counter.New(ttl),
		onRequestLimit:   onRequestLimit,
		targetExtensions: targetExtensions,
	}
}

func (l *BaseLimiter) ShouldSetXRateLimitHeaders(r *rl.Context) bool {
	return false
}

func (l *BaseLimiter) Name() string {
	return "base_limiter"
}

func (l *BaseLimiter) isTargetRequest(r *http.Request) bool {
	return l.isTargetExtensions(r)
}

func (l *BaseLimiter) isTargetExtensions(r *http.Request) bool {
	if len(l.targetExtensions) == 0 {
		return true
	}
	extension := filepath.Ext(r.URL.Path)
	for _, ext := range l.targetExtensions {
		if strings.EqualFold(ext, extension) {
			return true
		}
	}
	return false
}
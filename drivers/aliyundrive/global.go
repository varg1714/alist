package aliyundrive

import (
	"crypto/ecdsa"
	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/pkg/generic_sync"
	"golang.org/x/time/rate"
	"time"
)

type State struct {
	deviceID   string
	signature  string
	retry      int
	privateKey *ecdsa.PrivateKey
}

var global = generic_sync.MapOf[string, *State]{}

var shareTokenCache = cache.NewMemCache(cache.WithShards[string](128))
var fileListRespCache = cache.NewMemCache(cache.WithShards[Files](128))
var limiter = rate.NewLimiter(rate.Every(1000*time.Millisecond), 1)

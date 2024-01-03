package baidu_share

import "github.com/Xhofe/go-cache"

// do others that not defined in Driver interface

var shareTokenCache = cache.NewMemCache(cache.WithShards[ShareInfo](128))

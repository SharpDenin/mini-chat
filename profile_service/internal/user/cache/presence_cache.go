package cache

import (
	"github.com/jellydator/ttlcache/v3"
	"profile_service/pkg/grpc_generated/chat"
	"strconv"
	"time"
)

type PresenceCache struct {
	cache *ttlcache.Cache[string, *chat.GetPresenceResponse]
}

func NewPresenceCache(ttl time.Duration) *PresenceCache {
	c := ttlcache.New[string, *chat.GetPresenceResponse](
		ttlcache.WithTTL[string, *chat.GetPresenceResponse](ttl),
	)
	go c.Start()
	return &PresenceCache{cache: c}
}

func (pc *PresenceCache) Get(userId int64) (*chat.GetPresenceResponse, bool) {
	key := strconv.FormatInt(userId, 10)
	item := pc.cache.Get(key)
	if item != nil {
		return item.Value(), true
	}

	return nil, false
}

func (pc *PresenceCache) Set(userId int64, resp *chat.GetPresenceResponse) {
	key := strconv.FormatInt(userId, 10)
	pc.cache.Set(key, resp, ttlcache.DefaultTTL)
}

func (pc *PresenceCache) Stop() {
	pc.cache.Stop()
}

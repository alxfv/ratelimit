package inmem

import (
	"math/rand"

	"github.com/coocood/freecache"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/envoyproxy/ratelimit/src/config"
	"github.com/envoyproxy/ratelimit/src/limiter"
	"github.com/envoyproxy/ratelimit/src/server"
	"github.com/envoyproxy/ratelimit/src/settings"
	"github.com/envoyproxy/ratelimit/src/utils"
	"golang.org/x/net/context"
)

type rateLimitCacheImpl struct {
	baseRateLimiter *limiter.BaseRateLimiter
}

func (r *rateLimitCacheImpl) DoLimit(ctx context.Context, request *pb.RateLimitRequest, limits []*config.RateLimit) []*pb.RateLimitResponse_DescriptorStatus {
	responseDescriptorStatuses := make([]*pb.RateLimitResponse_DescriptorStatus, len(request.Descriptors))

	return responseDescriptorStatuses
}

func (r *rateLimitCacheImpl) Flush() {
}

var _ limiter.RateLimitCache = (*rateLimitCacheImpl)(nil)

func NewRateLimiterCacheImplFromSettings(
	s settings.Settings,
	localCache *freecache.Cache,
	srv server.Server,
	timeSource utils.TimeSource,
	jitterRand *rand.Rand,
	expirationJitterMaxSeconds int64,
) limiter.RateLimitCache {
	return NewFixedRateLimitCacheImpl(
		timeSource,
		jitterRand,
		expirationJitterMaxSeconds,
		localCache,
		s.NearLimitRatio,
		s.CacheKeyPrefix,
	)
}

func NewFixedRateLimitCacheImpl(
	timeSource utils.TimeSource,
	jitterRand *rand.Rand,
	expirationJitterMaxSeconds int64,
	localCache *freecache.Cache,
	nearLimitRatio float32,
	cacheKeyPrefix string,
) limiter.RateLimitCache {
	return &rateLimitCacheImpl{
		baseRateLimiter: limiter.NewBaseRateLimit(timeSource, jitterRand, expirationJitterMaxSeconds, localCache, nearLimitRatio, cacheKeyPrefix),
	}
}

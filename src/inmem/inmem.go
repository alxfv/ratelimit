package inmem

import (
	"math/rand"
	"time"

	"github.com/coocood/freecache"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/envoyproxy/ratelimit/src/config"
	"github.com/envoyproxy/ratelimit/src/limiter"
	"github.com/envoyproxy/ratelimit/src/settings"
	"github.com/envoyproxy/ratelimit/src/stats"
	"github.com/envoyproxy/ratelimit/src/utils"
	"github.com/juju/ratelimit"
	logger "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type rateLimitInMemImpl struct {
	baseRateLimiter *limiter.BaseRateLimiter
	buckets         map[string]*ratelimit.Bucket
}

func (this *rateLimitInMemImpl) DoLimit(
	ctx context.Context,
	request *pb.RateLimitRequest,
	limits []*config.RateLimit,
) []*pb.RateLimitResponse_DescriptorStatus {

	logger.Debugf("starting cache lookup")

	// request.HitsAddend could be 0 (default value) if not specified by the caller in the RateLimit request.
	hitsAddend := utils.Max(1, request.HitsAddend)

	logger.Debugf("hitsAddend = %v", hitsAddend)

	cacheKeys := make([]string, len(request.Descriptors))
	for i := 0; i < len(request.Descriptors); i++ {
		if limits[i] == nil {
			continue
		}
		limits[i].Stats.TotalHits.Add(uint64(hitsAddend))

		cacheKeys[i] = limits[i].FullKey
	}

	logger.Debugf("cacheKeys = %v", cacheKeys)

	responseDescriptorStatuses := make([]*pb.RateLimitResponse_DescriptorStatus,
		len(request.Descriptors))


	for i, key := range cacheKeys {
		logger.Debugf("key: %v", key)
		logger.Debugf("limit: %+v", limits[i])

		limit := limits[i]

		if limit == nil {
			// default response
			responseDescriptorStatuses[i] = this.baseRateLimiter.GetResponseDescriptorStatus(
				key,
				&limiter.LimitInfo{},
				false,
				hitsAddend,
			)

			continue
		}

		logger.Debugf("bucket key: %s", key)

		if _, ok := this.buckets[key]; !ok {
			capacity := int64(limit.Limit.RequestsPerUnit)

			// unit = second, perUnit = 10, fillInterval = 1 * 1000 / 10 = 0.1 100ms
			// unit = minute, perUnit = 10, fillInterval = 60 * 1000 / 10 = 600 ms
			divider := utils.UnitToDivider(limit.Limit.Unit)
			interval := divider * 1000 / int64(limit.Limit.RequestsPerUnit)
			fillInterval := time.Duration(interval) * time.Millisecond

			this.buckets[key] = ratelimit.NewBucket(fillInterval, capacity)

			logger.Debugf(
				"create new bucket with fillInterval: %v, capacity: %v",
				fillInterval,
				capacity,
			)
		} else {
			logger.Debugf("available tokens in bucket %s: %v", key, this.buckets[key].Available())
		}

		limitBeforeIncrease := uint32(this.buckets[key].Capacity() - this.buckets[key].Available())

		taken := this.buckets[key].TakeAvailable(int64(hitsAddend))
		logger.Debugf("tokens have been taken: %v", taken)

		available := this.buckets[key].Available()

		logger.Debugf("available tokens in bucket %s after increase: %v", key, available)

		limitAfterIncrease := limitBeforeIncrease + hitsAddend

		logger.Debugf("limitBeforeIncrease: %v", limitBeforeIncrease)
		logger.Debugf("limitAfterIncrease: %v", limitAfterIncrease)

		limitInfo := limiter.NewRateLimitInfo(
			limits[i],
			limitBeforeIncrease,
			limitAfterIncrease,
			0,
			0,
		)

		logger.Debugf("limitInfo: %v", limitInfo)

		responseDescriptorStatuses[i] = this.baseRateLimiter.GetResponseDescriptorStatus(
			key,
			limitInfo,
			false,
			hitsAddend,
		)
	}

	return responseDescriptorStatuses
}

// Waits for any unfinished asynchronous work. This may be used by unit tests,
// since the memcache cache does increments in a background gorountine.
func (this *rateLimitInMemImpl) Flush() {}

var _ limiter.RateLimitCache = (*rateLimitInMemImpl)(nil)

func NewRateLimiterInMemImplFromSettings(
	s settings.Settings,
	localCache *freecache.Cache,
	timeSource utils.TimeSource,
	jitterRand *rand.Rand,
	expirationJitterMaxSeconds int64,
	statsManager stats.Manager,
) limiter.RateLimitCache {
	return NewRateLimitInMemImpl(
		timeSource,
		jitterRand,
		expirationJitterMaxSeconds,
		localCache,
		s.NearLimitRatio,
		s.CacheKeyPrefix,
		statsManager,
	)
}

func NewRateLimitInMemImpl(
	timeSource utils.TimeSource,
	jitterRand *rand.Rand,
	expirationJitterMaxSeconds int64,
	localCache *freecache.Cache,
	nearLimitRatio float32,
	cacheKeyPrefix string,
	statsManager stats.Manager,
) limiter.RateLimitCache {
	return &rateLimitInMemImpl{
		baseRateLimiter: limiter.NewBaseRateLimit(
			timeSource,
			jitterRand,
			expirationJitterMaxSeconds,
			localCache,
			nearLimitRatio,
			cacheKeyPrefix,
			statsManager,
		),
		buckets: make(map[string]*ratelimit.Bucket),
	}
}

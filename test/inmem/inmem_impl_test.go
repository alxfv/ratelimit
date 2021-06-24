package inmem_test

import (
	"math/rand"
	"testing"

	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/envoyproxy/ratelimit/src/config"
	"github.com/envoyproxy/ratelimit/src/inmem"
	"github.com/envoyproxy/ratelimit/src/utils"
	"github.com/envoyproxy/ratelimit/test/common"
	"github.com/envoyproxy/ratelimit/test/mocks/stats"
	mock_utils "github.com/envoyproxy/ratelimit/test/mocks/utils"
	"github.com/golang/mock/gomock"
	gostats "github.com/lyft/gostats"
	"github.com/stretchr/testify/assert"
)

func TestInMemSimple(t *testing.T) {
	a := assert.New(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	statsStore := gostats.NewStore(gostats.NewNullSink(), false)
	sm := stats.NewMockStatManager(statsStore)

	timeSource := mock_utils.NewMockTimeSource(controller)

	rl := inmem.NewRateLimitInMemImpl(
		timeSource,
		rand.New(rand.NewSource(1)),
		0,
		nil,
		0.8,
		"",
		sm,
	)

	request := common.NewRateLimitRequest("domain", [][][2]string{{{"key", "value"}}}, 1)
	limits := []*config.RateLimit{config.NewRateLimit(
		2,
		pb.RateLimitResponse_RateLimit_SECOND,
		sm.NewStats("key_value"),
	)}

	timeSource.EXPECT().UnixNow().Return(int64(1234)).MaxTimes(3)

	a.Equal(
		[]*pb.RateLimitResponse_DescriptorStatus{{
			Code:               pb.RateLimitResponse_OK,
			CurrentLimit:       limits[0].Limit,
			LimitRemaining:     1,
			DurationUntilReset: utils.CalculateReset(limits[0].Limit, timeSource),
		}},
		rl.DoLimit(nil, request, limits),
	)
	a.Equal(uint64(1), limits[0].Stats.TotalHits.Value())
	a.Equal(uint64(0), limits[0].Stats.OverLimit.Value())
	a.Equal(uint64(0), limits[0].Stats.NearLimit.Value())
	a.Equal(uint64(1), limits[0].Stats.WithinLimit.Value())

	timeSource.EXPECT().UnixNow().Return(int64(1234)).MaxTimes(3)

	a.Equal(
		[]*pb.RateLimitResponse_DescriptorStatus{{
			Code:               pb.RateLimitResponse_OK,
			CurrentLimit:       limits[0].Limit,
			LimitRemaining:     0,
			DurationUntilReset: utils.CalculateReset(limits[0].Limit, timeSource),
		}},
		rl.DoLimit(nil, request, limits),
	)
	a.Equal(uint64(2), limits[0].Stats.TotalHits.Value())
	a.Equal(uint64(0), limits[0].Stats.OverLimit.Value())
	a.Equal(uint64(1), limits[0].Stats.NearLimit.Value())
	a.Equal(uint64(2), limits[0].Stats.WithinLimit.Value())

	timeSource.EXPECT().UnixNow().Return(int64(1234)).MaxTimes(3)

	a.Equal(
		[]*pb.RateLimitResponse_DescriptorStatus{{
			Code:               pb.RateLimitResponse_OVER_LIMIT,
			CurrentLimit:       limits[0].Limit,
			LimitRemaining:     0,
			DurationUntilReset: utils.CalculateReset(limits[0].Limit, timeSource),
		}},
		rl.DoLimit(nil, request, limits),
	)
	a.Equal(uint64(3), limits[0].Stats.TotalHits.Value())
	a.Equal(uint64(1), limits[0].Stats.OverLimit.Value())
	a.Equal(uint64(1), limits[0].Stats.NearLimit.Value())
	a.Equal(uint64(2), limits[0].Stats.WithinLimit.Value())

	timeSource.EXPECT().UnixNow().Return(int64(1234)).MaxTimes(3)

	a.Equal(
		[]*pb.RateLimitResponse_DescriptorStatus{{
			Code:               pb.RateLimitResponse_OVER_LIMIT,
			CurrentLimit:       limits[0].Limit,
			LimitRemaining:     0,
			DurationUntilReset: utils.CalculateReset(limits[0].Limit, timeSource),
		}},
		rl.DoLimit(nil, request, limits),
	)
	a.Equal(uint64(4), limits[0].Stats.TotalHits.Value())
	a.Equal(uint64(2), limits[0].Stats.OverLimit.Value())
	a.Equal(uint64(1), limits[0].Stats.NearLimit.Value())
	a.Equal(uint64(2), limits[0].Stats.WithinLimit.Value())
}

func TestInMemTwoDescriptors(t *testing.T) {
	a := assert.New(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	statsStore := gostats.NewStore(gostats.NewNullSink(), false)
	sm := stats.NewMockStatManager(statsStore)

	timeSource := mock_utils.NewMockTimeSource(controller)

	rl := inmem.NewRateLimitInMemImpl(
		timeSource,
		rand.New(rand.NewSource(1)),
		0,
		nil,
		0.8,
		"",
		sm,
	)

	descriptors := [][][2]string{{{"key1", "value1"}, {"key2", ""}}}
	request := common.NewRateLimitRequest("domain", descriptors, 1)
	limits := []*config.RateLimit{
		config.NewRateLimit(
			2,
			pb.RateLimitResponse_RateLimit_SECOND,
			sm.NewStats("key_value"),
		)}

	timeSource.EXPECT().UnixNow().Return(int64(1234)).MaxTimes(3)

	a.Equal(
		[]*pb.RateLimitResponse_DescriptorStatus{{
			Code:               pb.RateLimitResponse_OK,
			CurrentLimit:       limits[0].Limit,
			LimitRemaining:     1,
			DurationUntilReset: utils.CalculateReset(limits[0].Limit, timeSource),
		}},
		rl.DoLimit(nil, request, limits),
	)
	a.Equal(uint64(1), limits[0].Stats.TotalHits.Value())
	a.Equal(uint64(0), limits[0].Stats.OverLimit.Value())
	a.Equal(uint64(0), limits[0].Stats.NearLimit.Value())
	a.Equal(uint64(1), limits[0].Stats.WithinLimit.Value())

	timeSource.EXPECT().UnixNow().Return(int64(1234)).MaxTimes(3)

	a.Equal(
		[]*pb.RateLimitResponse_DescriptorStatus{{
			Code:               pb.RateLimitResponse_OK,
			CurrentLimit:       limits[0].Limit,
			LimitRemaining:     0,
			DurationUntilReset: utils.CalculateReset(limits[0].Limit, timeSource),
		}},
		rl.DoLimit(nil, request, limits),
	)
	a.Equal(uint64(2), limits[0].Stats.TotalHits.Value())
	a.Equal(uint64(0), limits[0].Stats.OverLimit.Value())
	a.Equal(uint64(1), limits[0].Stats.NearLimit.Value())
	a.Equal(uint64(2), limits[0].Stats.WithinLimit.Value())
}

func TestInMemNoLimits(t *testing.T) {
	a := assert.New(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	statsStore := gostats.NewStore(gostats.NewNullSink(), false)
	sm := stats.NewMockStatManager(statsStore)

	timeSource := mock_utils.NewMockTimeSource(controller)

	rl := inmem.NewRateLimitInMemImpl(
		timeSource,
		rand.New(rand.NewSource(1)),
		0,
		nil,
		0.8,
		"",
		sm,
	)

	descriptors := [][][2]string{{{"key1", "value1"}, {"key2", ""}}}
	request := common.NewRateLimitRequest("domain", descriptors, 1)
	limits := []*config.RateLimit{nil}

	timeSource.EXPECT().UnixNow().Return(int64(1234)).MaxTimes(3)

	a.Equal(
		[]*pb.RateLimitResponse_DescriptorStatus{{
			Code: pb.RateLimitResponse_OK,
		}},
		rl.DoLimit(nil, request, limits),
	)
	timeSource.EXPECT().UnixNow().Return(int64(1234)).MaxTimes(3)

	a.Equal(
		[]*pb.RateLimitResponse_DescriptorStatus{{
			Code: pb.RateLimitResponse_OK,
		}},
		rl.DoLimit(nil, request, limits),
	)
}

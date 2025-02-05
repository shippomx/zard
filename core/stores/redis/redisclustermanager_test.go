package redis

import (
	"context"
	"os"
	"testing"

	"github.com/shippomx/zard/core/breaker"
	"github.com/alicebob/miniredis/v2"
	red "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestSplitClusterAddrs(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "single address",
			input:    "127.0.0.1:8000",
			expected: []string{"127.0.0.1:8000"},
		},
		{
			name:     "multiple addresses with duplicates",
			input:    "127.0.0.1:8000,127.0.0.1:8001, 127.0.0.1:8000",
			expected: []string{"127.0.0.1:8000", "127.0.0.1:8001"},
		},
		{
			name:     "multiple addresses without duplicates",
			input:    "127.0.0.1:8000, 127.0.0.1:8001, 127.0.0.1:8002",
			expected: []string{"127.0.0.1:8000", "127.0.0.1:8001", "127.0.0.1:8002"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.expected, splitClusterAddrs(tc.input))
		})
	}
}

func TestGetCluster(t *testing.T) {
	r := miniredis.RunT(t)
	defer r.Close()
	c, err := getCluster(&Redis{
		Addr:  r.Addr(),
		Type:  ClusterType,
		tls:   true,
		brk:   breaker.NewBreaker(),
		hooks: []red.Hook{defaultDurationHook},
	})
	if assert.NoError(t, err) {
		assert.NotNil(t, c)
	}
}

func TestGetRedisVersion(t *testing.T) {
	r := miniredis.RunT(t)
	defer r.Close()
	store := red.NewClient(&red.Options{
		Addr: r.Addr(),
	})
	infoMap := store.InfoMap(context.TODO(), "Server")
	version := infoMap.Item("Server", "redis_version")
	assert.Equal(t, version, "")
}

func TestDiscoverResolverClient_SetOnUpdate(t *testing.T) {
	dc := &DiscoverResolverClient{}

	// Test case: Setting the onUpdate function
	t.Run("SetOnUpdate", func(t *testing.T) {
		fn := func(_ context.Context) {}
		dc.SetOnUpdate(fn)
		if dc.onUpdate == nil {
			t.Error("onUpdate function not set")
		}
	})

	// Test case: Updating the onUpdate function
	t.Run("UpdateOnUpdate", func(t *testing.T) {
		a := 0
		fn1 := func(_ context.Context) { a = 1 }
		fn2 := func(_ context.Context) { a = 2 }
		dc.SetOnUpdate(fn1)
		dc.SetOnUpdate(fn2)
		if dc.onUpdate != nil {
			dc.onUpdate(context.Background())
			assert.Equal(t, 2, a)
		}
	})
}

// TestGetClusterWithDB tests getCluster with a database
// TODO: use miniredis instead
// Note: miniredis does not support custom info cmd, which causes the discover client to not work properly.
func TestGetClusterWithDB(t *testing.T) {
	// skip the test if LOCAL_TEST_REDIS_HOST is not set
	if os.Getenv("LOCAL_TEST_REDIS_HOST") == "" {
		t.Skip("LOCAL_TEST_REDIS_HOST is not set, eg: 127.0.0.1:6379")
	}

	// start a miniredis server and get its address
	r := miniredis.RunT(t)
	defer r.Close()

	// create a redis client with the address and db
	c, err := getCluster(&Redis{
		Addr:             r.Addr() + ", " + os.Getenv("LOCAL_TEST_REDIS_HOST"),
		Type:             ClusterType,
		brk:              breaker.NewBreaker(),
		SingleReplicaSet: true,
		RouteRandomly:    true,
		hooks:            []red.Hook{defaultDurationHook},
		DB:               2,
	})
	assert.NoError(t, err)

	// set a key and value in the redis
	res, err := c.Set(context.Background(), "key", "value2", 0).Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", res)

	// get the value in the redis and assert it is equal to "value2"
	testRedisClient := red.NewClient(&red.Options{
		Addr: os.Getenv("LOCAL_TEST_REDIS_HOST"),
		DB:   2,
	})
	assert.Equal(t, "value2", testRedisClient.Get(context.Background(), "key").Val())
}

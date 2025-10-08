package unit_tests

import (
	"context"
	"testing"
	"time"

	"aegis-api/cache"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient2 mocks the RedisClient interface
type MockRedisClient2 struct {
	mock.Mock
}

func (m *MockRedisClient2) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient2) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient2) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient2) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	args := m.Called(ctx, cursor, match, count)
	return args.Get(0).(*redis.ScanCmd)
}

func TestRedisAdapter_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("key exists", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Get", ctx, "existing-key").Return(redis.NewStringResult("value", nil))
		value, exists, err := adapter.Get(ctx, "existing-key")

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "value", value)
		mockClient.AssertExpectations(t)
	})

	t.Run("key does not exist", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Get", ctx, "non-existent-key").Return(redis.NewStringResult("", redis.Nil))
		value, exists, err := adapter.Get(ctx, "non-existent-key")

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, "", value)
		mockClient.AssertExpectations(t)
	})

	t.Run("redis error", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Get", ctx, "error-key").Return(redis.NewStringResult("", assert.AnError))
		value, exists, err := adapter.Get(ctx, "error-key")

		assert.Error(t, err)
		assert.False(t, exists)
		assert.Equal(t, "", value)
		mockClient.AssertExpectations(t)
	})
}

func TestRedisAdapter_Set(t *testing.T) {
	ctx := context.Background()

	t.Run("successful set", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Set", ctx, "key", "value", 5*time.Minute).Return(redis.NewStatusResult("OK", nil))
		err := adapter.Set(ctx, "key", "value", 5*time.Minute)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("set error", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Set", ctx, "key", "value", 5*time.Minute).Return(redis.NewStatusResult("", assert.AnError))
		err := adapter.Set(ctx, "key", "value", 5*time.Minute)

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestRedisAdapter_Del(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete single key", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Del", ctx, []string{"key1"}).Return(redis.NewIntResult(1, nil))
		count, err := adapter.Del(ctx, "key1")

		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		mockClient.AssertExpectations(t)
	})

	t.Run("delete multiple keys", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Del", ctx, []string{"key1", "key2", "key3"}).Return(redis.NewIntResult(3, nil))
		count, err := adapter.Del(ctx, "key1", "key2", "key3")

		assert.NoError(t, err)
		assert.Equal(t, 3, count)
		mockClient.AssertExpectations(t)
	})

	t.Run("delete error", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Del", ctx, []string{"key1"}).Return(redis.NewIntResult(0, assert.AnError))
		count, err := adapter.Del(ctx, "key1")

		assert.Error(t, err)
		assert.Equal(t, 0, count)
		mockClient.AssertExpectations(t)
	})

	t.Run("delete non-existent key returns 0", func(t *testing.T) {
		mockClient := new(MockRedisClient2)
		adapter := cache.NewRedisWithClient(mockClient)

		mockClient.On("Del", ctx, []string{"non-existent"}).Return(redis.NewIntResult(0, nil))
		count, err := adapter.Del(ctx, "non-existent")

		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		mockClient.AssertExpectations(t)
	})
}

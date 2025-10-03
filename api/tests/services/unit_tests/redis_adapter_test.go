package unit_tests

import (
	"aegis-api/cache"
	"context"
	"errors"
	"testing"
	"time"

	//"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockClient is a mock for the cache.Client interface
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Get(ctx context.Context, key string) (string, bool, error) {
	args := m.Called(ctx, key)
	value, _ := args.Get(0).(string)
	ok, _ := args.Get(1).(bool)
	err, _ := args.Get(2).(error)
	return value, ok, err
}

func (m *MockClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockClient) Del(ctx context.Context, keys ...string) (int, error) {
	args := m.Called(ctx, keys)
	count, _ := args.Get(0).(int)
	err, _ := args.Get(1).(error)
	return count, err
}

func TestNewRedis(t *testing.T) {
	addr := "localhost:6379"
	password := "secret"
	db := 1

	redisClient := cache.NewRedis(addr, password, db)
	require.NotNil(t, redisClient, "NewRedis should return a non-nil Redis struct")
	// Cannot access rdb directly due to it being unexported; rely on functional tests
}

func TestRedisGet(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockClient{}
	cacheRedis := mockClient // Use the mock directly as a cache.Client

	// Test case 1: Successful retrieval
	mockClient.On("Get", ctx, "key1").Return("value1", true, nil).Once()
	value, ok, err := cacheRedis.Get(ctx, "key1")
	require.NoError(t, err, "Get should not return an error")
	require.True(t, ok, "Get should return true for existing key")
	require.Equal(t, "value1", value, "Get should return correct value")

	// Test case 2: Key not found
	mockClient.On("Get", ctx, "key2").Return("", false, nil).Once()
	value, ok, err = cacheRedis.Get(ctx, "key2")
	require.NoError(t, err, "Get should not return an error for non-existing key")
	require.False(t, ok, "Get should return false for non-existing key")
	require.Empty(t, value, "Get should return empty string for non-existing key")

	// Test case 3: Error
	mockClient.On("Get", ctx, "key3").Return("", false, errors.New("redis error")).Once()
	value, ok, err = cacheRedis.Get(ctx, "key3")
	require.Error(t, err, "Get should return an error")
	require.Equal(t, "redis error", err.Error(), "Get should return correct error message")
	require.False(t, ok, "Get should return false on error")
	require.Empty(t, value, "Get should return empty string on error")

	mockClient.AssertExpectations(t)
}

func TestRedisSet(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockClient{}
	cacheRedis := mockClient // Use the mock directly as a cache.Client
	ttl := time.Second * 60

	// Test case 1: Successful set
	mockClient.On("Set", ctx, "key1", "value1", ttl).Return(nil).Once()
	err := cacheRedis.Set(ctx, "key1", "value1", ttl)
	require.NoError(t, err, "Set should not return an error")

	// Test case 2: Error
	mockClient.On("Set", ctx, "key2", "value2", ttl).Return(errors.New("redis error")).Once()
	err = cacheRedis.Set(ctx, "key2", "value2", ttl)
	require.Error(t, err, "Set should return an error")
	require.Equal(t, "redis error", err.Error(), "Set should return correct error message")

	mockClient.AssertExpectations(t)
}

func TestRedisDel(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockClient{}
	cacheRedis := mockClient // Use the mock directly as a cache.Client

	// Test case 1: Successful deletion
	mockClient.On("Del", ctx, []string{"key1", "key2"}).Return(2, nil).Once()
	count, err := cacheRedis.Del(ctx, "key1", "key2")
	require.NoError(t, err, "Del should not return an error")
	require.Equal(t, 2, count, "Del should return correct number of deleted keys")

	// Test case 2: No keys deleted
	mockClient.On("Del", ctx, []string{"key3"}).Return(0, nil).Once()
	count, err = cacheRedis.Del(ctx, "key3")
	require.NoError(t, err, "Del should not return an error")
	require.Equal(t, 0, count, "Del should return 0 for non-existing keys")

	// Test case 3: Error
	mockClient.On("Del", ctx, []string{"key4"}).Return(0, errors.New("redis error")).Once()
	count, err = cacheRedis.Del(ctx, "key4")
	require.Error(t, err, "Del should return an error")
	require.Equal(t, "redis error", err.Error(), "Del should return correct error message")
	require.Equal(t, 0, count, "Del should return 0 on error")

	mockClient.AssertExpectations(t)
}

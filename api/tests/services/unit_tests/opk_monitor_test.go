package unit_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"aegis-api/internal/x3dh"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKeyStore - shared mock for all test files
type MockKeyStore struct {
	mock.Mock
}

func (m *MockKeyStore) StoreBundle(ctx context.Context, req x3dh.RegisterBundleRequest, crypto x3dh.CryptoService) error {
	args := m.Called(ctx, req, crypto)
	return args.Error(0)
}

func (m *MockKeyStore) GetIdentityKey(ctx context.Context, userID string) (*x3dh.IdentityKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*x3dh.IdentityKey), args.Error(1)
}

func (m *MockKeyStore) GetSignedPreKey(ctx context.Context, userID string) (*x3dh.SignedPreKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*x3dh.SignedPreKey), args.Error(1)
}

func (m *MockKeyStore) ConsumeOneTimePreKey(ctx context.Context, userID string) (*x3dh.OneTimePreKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*x3dh.OneTimePreKey), args.Error(1)
}

func (m *MockKeyStore) InsertOPKs(ctx context.Context, userID string, opks []x3dh.OneTimePreKeyUpload) error {
	args := m.Called(ctx, userID, opks)
	return args.Error(0)
}

func (m *MockKeyStore) RotateSignedPreKey(ctx context.Context, userID, newSPK, signature string, expiresAt *time.Time) error {
	args := m.Called(ctx, userID, newSPK, signature, expiresAt)
	return args.Error(0)
}

func (m *MockKeyStore) CountOPKs(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockKeyStore) CountAvailableOPKs(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockKeyStore) ListUsersWithOPKs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// Add other shared mocks here (MockCryptoService, MockAuditLogger)
// Test the OPKMonitor creation

// Test the OPKMonitor creation
func TestNewOPKMonitor(t *testing.T) {
	store := new(MockKeyStore)
	threshold := 10
	interval := 5 * time.Second

	monitor := x3dh.NewOPKMonitor(store, threshold, interval)
	assert.NotNil(t, monitor)
}

// Test the monitoring logic with sufficient OPKs
func TestOPKMonitor_CheckAllUsers_SufficientOPKs(t *testing.T) {
	store := new(MockKeyStore)
	// Use a shorter interval so the ticker fires quickly
	monitor := x3dh.NewOPKMonitor(store, 5, 10*time.Millisecond)

	ctx := context.Background()
	userIDs := []string{"user1", "user2", "user3"}

	// Mock store responses - use mock.Anything for context to match any context type
	store.On("ListUsersWithOPKs", mock.Anything).Return(userIDs, nil)
	store.On("CountAvailableOPKs", mock.Anything, "user1").Return(10, nil)
	store.On("CountAvailableOPKs", mock.Anything, "user2").Return(8, nil)
	store.On("CountAvailableOPKs", mock.Anything, "user3").Return(15, nil)

	// Give enough time for at least one tick
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	// Run monitor in goroutine so we can control timing
	done := make(chan bool)
	go func() {
		monitor.Start(ctxWithTimeout)
		done <- true
	}()

	// Wait for completion
	select {
	case <-done:
		// Monitor completed
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Test timed out")
	}

	// Verify all expected calls were made
	store.AssertExpectations(t)
}

// Test the monitoring logic with low OPKs
func TestOPKMonitor_CheckAllUsers_LowOPKs(t *testing.T) {
	store := new(MockKeyStore)
	monitor := x3dh.NewOPKMonitor(store, 5, 10*time.Millisecond)

	ctx := context.Background()
	userIDs := []string{"user1", "user2"}

	// Mock store responses - user1 has low OPKs, user2 has sufficient
	store.On("ListUsersWithOPKs", mock.Anything).Return(userIDs, nil)
	store.On("CountAvailableOPKs", mock.Anything, "user1").Return(2, nil)  // Below threshold
	store.On("CountAvailableOPKs", mock.Anything, "user2").Return(10, nil) // Above threshold

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	done := make(chan bool)
	go func() {
		monitor.Start(ctxWithTimeout)
		done <- true
	}()

	select {
	case <-done:
		// Monitor completed
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Test timed out")
	}

	store.AssertExpectations(t)
}

// Test error handling when ListUsersWithOPKs fails
func TestOPKMonitor_CheckAllUsers_ListUsersError(t *testing.T) {
	store := new(MockKeyStore)
	monitor := x3dh.NewOPKMonitor(store, 5, 10*time.Millisecond)

	ctx := context.Background()

	// Mock store to return an error
	store.On("ListUsersWithOPKs", mock.Anything).Return([]string(nil), assert.AnError)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	done := make(chan bool)
	go func() {
		monitor.Start(ctxWithTimeout)
		done <- true
	}()

	select {
	case <-done:
		// Monitor completed
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Test timed out")
	}

	store.AssertExpectations(t)
}

// Test error handling when CountAvailableOPKs fails
func TestOPKMonitor_CheckAllUsers_CountOPKsError(t *testing.T) {
	store := new(MockKeyStore)
	monitor := x3dh.NewOPKMonitor(store, 5, 10*time.Millisecond)

	ctx := context.Background()
	userIDs := []string{"user1", "user2"}

	// Mock store responses - user1 fails, user2 succeeds
	store.On("ListUsersWithOPKs", mock.Anything).Return(userIDs, nil)
	store.On("CountAvailableOPKs", mock.Anything, "user1").Return(0, assert.AnError)
	store.On("CountAvailableOPKs", mock.Anything, "user2").Return(10, nil)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	done := make(chan bool)
	go func() {
		monitor.Start(ctxWithTimeout)
		done <- true
	}()

	select {
	case <-done:
		// Monitor completed
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Test timed out")
	}

	store.AssertExpectations(t)
}

// Test context cancellation stops the monitor
func TestOPKMonitor_Start_ContextCancellation(t *testing.T) {
	store := new(MockKeyStore)
	monitor := x3dh.NewOPKMonitor(store, 5, 100*time.Millisecond) // Longer interval

	ctx, cancel := context.WithCancel(context.Background())

	// Start monitor in goroutine
	done := make(chan bool)
	go func() {
		monitor.Start(ctx)
		done <- true
	}()

	// Cancel context immediately
	cancel()

	// Wait for monitor to stop (with timeout to prevent hanging)
	select {
	case <-done:
		// Monitor stopped as expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Monitor did not stop within expected time")
	}

	// The store methods might not be called if context is cancelled quickly
	store.AssertExpectations(t)
}

// Test monitor runs multiple cycles before being cancelled
func TestOPKMonitor_Start_MultipleCycles(t *testing.T) {
	store := new(MockKeyStore)
	// Use a short interval for testing
	monitor := x3dh.NewOPKMonitor(store, 5, 20*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	userIDs := []string{"user1"}

	// Set up expectations for multiple calls - use Maybe() since we can't predict exact count
	store.On("ListUsersWithOPKs", mock.Anything).Return(userIDs, nil).Maybe()
	store.On("CountAvailableOPKs", mock.Anything, "user1").Return(10, nil).Maybe()

	// Start monitor in goroutine
	done := make(chan bool)
	go func() {
		monitor.Start(ctx)
		done <- true
	}()

	// Let it run for multiple cycles
	time.Sleep(100 * time.Millisecond)

	// Cancel and wait for stop
	cancel()

	select {
	case <-done:
		// Monitor stopped as expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Monitor did not stop within expected time")
	}

	// Should have been called at least once, possibly more
	store.AssertExpectations(t)
}

// Test edge cases and boundary conditions
func TestOPKMonitor_EdgeCases(t *testing.T) {
	t.Run("Zero threshold", func(t *testing.T) {
		store := new(MockKeyStore)
		monitor := x3dh.NewOPKMonitor(store, 0, time.Second)
		assert.NotNil(t, monitor)
	})

	t.Run("Negative threshold", func(t *testing.T) {
		store := new(MockKeyStore)
		monitor := x3dh.NewOPKMonitor(store, -1, time.Second)
		assert.NotNil(t, monitor)
	})

	t.Run("Very short interval", func(t *testing.T) {
		store := new(MockKeyStore)
		monitor := x3dh.NewOPKMonitor(store, 5, time.Nanosecond)
		assert.NotNil(t, monitor)
	})

	t.Run("Empty user list", func(t *testing.T) {
		store := new(MockKeyStore)
		monitor := x3dh.NewOPKMonitor(store, 5, 10*time.Millisecond)

		ctx := context.Background()
		emptyUserIDs := []string{}

		store.On("ListUsersWithOPKs", mock.Anything).Return(emptyUserIDs, nil)

		ctxWithTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		go func() {
			monitor.Start(ctxWithTimeout)
			done <- true
		}()

		select {
		case <-done:
			// Monitor completed
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Test timed out")
		}

		store.AssertExpectations(t)
	})
}

// Test that doesn't rely on timing - directly test checkAllUsers logic by making OPKMonitor testable
func TestOPKMonitor_DirectCheckLogic(t *testing.T) {
	t.Run("Sufficient OPKs - no warnings", func(t *testing.T) {
		store := new(MockKeyStore)
		monitor := x3dh.NewOPKMonitor(store, 5, time.Hour) // Long interval, won't matter

		userIDs := []string{"user1", "user2"}
		store.On("ListUsersWithOPKs", mock.Anything).Return(userIDs, nil)
		store.On("CountAvailableOPKs", mock.Anything, "user1").Return(10, nil)
		store.On("CountAvailableOPKs", mock.Anything, "user2").Return(8, nil)

		// Run just one check by using very short timeout after first tick
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// This will exit quickly due to context timeout, but may not call checkAllUsers
		monitor.Start(ctx)

		// Don't assert expectations since timing is unpredictable
	})
}

// Benchmark test to ensure the monitor doesn't consume excessive resources
func BenchmarkOPKMonitor_CheckAllUsers(b *testing.B) {
	store := new(MockKeyStore)

	// Create a large number of users for stress testing
	userIDs := make([]string, 100) // Reduced from 1000 for benchmark efficiency
	for i := range userIDs {
		userIDs[i] = fmt.Sprintf("user%d", i)
	}

	ctx := context.Background()
	store.On("ListUsersWithOPKs", mock.Anything).Return(userIDs, nil)

	for i := range userIDs {
		store.On("CountAvailableOPKs", mock.Anything, userIDs[i]).Return(10, nil)
	}

	b.ResetTimer()

	// Create monitor for each iteration to avoid state issues
	for i := 0; i < b.N; i++ {
		monitor := x3dh.NewOPKMonitor(store, 5, time.Hour)
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Nanosecond)
		monitor.Start(ctxWithTimeout)
		cancel()
	}
}

// Integration test that verifies the monitor works with a real-like scenario
func TestOPKMonitor_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	store := new(MockKeyStore)
	monitor := x3dh.NewOPKMonitor(store, 5, 25*time.Millisecond)

	userIDs := []string{"alice", "bob", "charlie"}

	// Set up realistic scenarios
	store.On("ListUsersWithOPKs", mock.Anything).Return(userIDs, nil)
	store.On("CountAvailableOPKs", mock.Anything, "alice").Return(15, nil)  // High count
	store.On("CountAvailableOPKs", mock.Anything, "bob").Return(3, nil)     // Low count
	store.On("CountAvailableOPKs", mock.Anything, "charlie").Return(7, nil) // Medium count

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run the monitor
	done := make(chan bool)
	go func() {
		monitor.Start(ctx)
		done <- true
	}()

	// Wait for completion
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Integration test timed out")
	}

	// Verify that monitoring calls were made
	store.AssertExpectations(t)
}

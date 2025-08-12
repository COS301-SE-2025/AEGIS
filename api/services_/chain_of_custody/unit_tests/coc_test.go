package unit_tests

import (
	coc "aegis-api/services_/chain_of_custody"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for the Repo interface
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Insert(ctx context.Context, p coc.LogParams) (string, error) {
	args := m.Called(ctx, p)
	return args.String(0), args.Error(1)
}

func (m *MockRepo) ListByEvidence(ctx context.Context, evidenceID string, f coc.ListFilters) ([]coc.Entry, error) {
	args := m.Called(ctx, evidenceID, f)
	return args.Get(0).([]coc.Entry), args.Error(1)
}

// Mock for the Auditor interface
type MockAuditor struct {
	mock.Mock
}

func (m *MockAuditor) Log(ctx context.Context, typ string, fields map[string]any) {
	// Expecting specific parameters and return mock result
	m.Called(ctx, typ, fields)
}

// Test the Log method of coc.Service
func TestCoCService_Log(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAuditor := new(MockAuditor)

	// Initialize the service with mock dependencies
	cocService := coc.Service{
		Repo:  mockRepo,
		Authz: coc.SimpleAuthz{},
		Audit: mockAuditor,
	}

	// Prepare test data
	params := coc.LogParams{
		CaseID:     "case-123",
		EvidenceID: "evidence-456",
		ActorID:    nil,
		Action:     coc.ActionUpload,
		Reason:     nil,
		Location:   nil,
		HashMD5:    nil,
		HashSHA1:   nil,
		HashSHA256: nil,
		OccurredAt: time.Now(),
	}

	// Set the expected behavior for Insert
	mockRepo.On("Insert", mock.Anything, params).Return("log-789", nil)

	// Expect the Log function of Auditor to be called with the correct arguments
	mockAuditor.On("Log", mock.Anything, "CHAIN_OF_CUSTODY_LOG", mock.Anything).Return(nil)

	// Test the Log function
	id, err := cocService.Log(context.Background(), params)

	// Assertions
	assert.Nil(t, err)
	assert.Equal(t, "log-789", id) // Expecting "log-789" as the return value

	// Assert Repo.Insert was called
	mockRepo.AssertExpectations(t)

	// Assert Auditor.Log was called with expected parameters
	mockAuditor.AssertExpectations(t)
}

// Test ListByEvidence method
func TestCoCService_ListByEvidence(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAuditor := new(MockAuditor)

	// Initialize the service with mock dependencies
	cocService := coc.Service{
		Repo:  mockRepo,
		Authz: coc.SimpleAuthz{},
		Audit: mockAuditor,
	}

	// Prepare test data
	filter := coc.ListFilters{
		Action: nil,
		Limit:  100,
		Offset: 0,
	}

	expectedEntries := []coc.Entry{
		{
			ID:         "log-1",
			CaseID:     "case-123",
			EvidenceID: "evidence-456",
			ActorID:    nil,
			Action:     coc.ActionUpload,
			Reason:     nil,
			Location:   nil,
			HashMD5:    nil,
			HashSHA1:   nil,
			HashSHA256: nil,
			OccurredAt: time.Now(),
			CreatedAt:  time.Now(),
		},
	}

	// Set the expected behavior for ListByEvidence
	mockRepo.On("ListByEvidence", mock.Anything, "evidence-456", filter).Return(expectedEntries, nil)

	// Test the ListByEvidence function
	entries, err := cocService.ListByEvidence(context.Background(), "evidence-456", filter)

	// Assertions
	assert.Nil(t, err)
	assert.Equal(t, expectedEntries, entries)

	// Assert ListByEvidence was called
	mockRepo.AssertExpectations(t)
}

// Test the ToCSV function
func TestCoCService_ToCSV(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAuditor := new(MockAuditor)

	// Initialize the service with mock dependencies
	cocService := coc.Service{
		Repo:  mockRepo,
		Authz: coc.SimpleAuthz{},
		Audit: mockAuditor,
	}

	// Prepare test data
	entries := []coc.Entry{
		{
			ID:         "log-1",
			CaseID:     "case-123",
			EvidenceID: "evidence-456",
			ActorID:    nil,
			Action:     coc.ActionUpload,
			Reason:     nil,
			Location:   nil,
			HashMD5:    nil,
			HashSHA1:   nil,
			HashSHA256: nil,
			OccurredAt: time.Now(),
			CreatedAt:  time.Now(),
		},
	}

	// Test the ToCSV function
	csvBytes, err := cocService.ToCSV(entries)

	// Assertions
	assert.Nil(t, err)
	assert.True(t, len(csvBytes) > 0)
}

// Test RBAC: Authorization check for logging CoC
func TestCoCService_Authorization(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAuditor := new(MockAuditor)

	// Initialize the service with mock dependencies
	cocService := coc.Service{
		Repo:  mockRepo,
		Authz: coc.SimpleAuthz{}, // This always returns true for testing
		Audit: mockAuditor,
	}

	// Prepare test data
	params := coc.LogParams{
		CaseID:     "case-123",
		EvidenceID: "evidence-456",
		ActorID:    nil,
		Action:     coc.ActionUpload,
		Reason:     nil,
		Location:   nil,
		HashMD5:    nil,
		HashSHA1:   nil,
		HashSHA256: nil,
		OccurredAt: time.Now(),
	}

	// Set the expected behavior for Insert
	mockRepo.On("Insert", mock.Anything, params).Return("log-789", nil)

	// Expect the Log function of Auditor to be called with the correct arguments
	mockAuditor.On("Log", mock.Anything, "CHAIN_OF_CUSTODY_LOG", mock.Anything).Return(nil)

	// Test the Log function with authorization check
	id, err := cocService.Log(context.Background(), params)

	// Assertions
	assert.Nil(t, err)
	assert.Equal(t, "log-789", id) // Expecting "log-789" as the return value
}

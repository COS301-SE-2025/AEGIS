package isolation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"aegis-api/services_/evidence/evidence_viewer"
	
)

func TestFindEvidenceWithMock(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockCursor := new(evidence_viewer.MockCursor)

	mockDB.On("Find", mock.Anything, mock.Anything).Return(mockCursor, nil)
	mockCursor.On("All", mock.Anything, mock.Anything).Return(nil)

	cursor, err := mockDB.Find(context.Background(), nil)
	assert.NoError(t, err)

	err = cursor.All(context.Background(), nil)
	assert.NoError(t, err)
}

func TestFindEvidenceByCaseWithMock(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockCursor := new(evidence_viewer.MockCursor)

	// Set expectations
	mockDB.On("Find", mock.Anything, mock.Anything).Return(mockCursor, nil)
	mockCursor.On("Close", mock.Anything).Return(nil)
	
	// Fix: Set up the mock to return data and ensure the slice is populated
	mockCursor.On("All", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]evidence_viewer.EvidenceResponse)
		*arg = []evidence_viewer.EvidenceResponse{
			{
				ID:        "ev123",
				CaseID:    "case456",
				Filename:  "photo.jpg",
				FileType:  "image",
				IPFSCID:   "cid123",
				UploadedAt: time.Now().Format(time.RFC3339),
			},
		}
	})

	repo := &evidence_viewer.MongoEvidenceRepository{Collection: mockDB}

	evidences, err := repo.GetEvidenceByCase("case123")
	assert.NoError(t, err)
	assert.NotNil(t, evidences)
	assert.Len(t, evidences, 1) // Add this to verify we got the expected data
	assert.Equal(t, "ev123", evidences[0].ID)
}

func TestFindEvidenceByIDWithMock(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockSingleResult := new(evidence_viewer.MockSingleResult)

	expected := evidence_viewer.EvidenceResponse{
		ID:        "ev123",
		CaseID:    "case456",
		Filename:  "file1.jpg",
		FileType:  "image",
		IPFSCID:   "cid456",
		UploadedAt: time.Now().Format(time.RFC3339),
	}

	mockDB.On("FindOne", mock.Anything, mock.Anything).Return(mockSingleResult)
	mockSingleResult.On("Decode", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*evidence_viewer.EvidenceResponse)
		*arg = expected
	})

	repo := &evidence_viewer.MongoEvidenceRepository{Collection: mockDB}

	result, err := repo.GetEvidenceByID("ev123")
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Filename, result.Filename)
}

func TestSearchEvidenceWithMock(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockCursor := new(evidence_viewer.MockCursor)

	mockDB.On("Find", mock.Anything, mock.Anything).Return(mockCursor, nil)
	mockCursor.On("All", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]evidence_viewer.EvidenceResponse)
		*arg = []evidence_viewer.EvidenceResponse{
			{
				ID:        "ev123",
				CaseID:    "case456",
				Filename:  "photo.jpg",
				FileType:  "image",
				IPFSCID:   "cid123",
				UploadedAt: time.Now().Format(time.RFC3339),
			},
		}
	})
	mockCursor.On("Close", mock.Anything).Return(nil)

	repo := &evidence_viewer.MongoEvidenceRepository{Collection: mockDB}

	results, err := repo.SearchEvidence("photo")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)
	assert.Equal(t, "ev123", results[0].ID)
}

func TestGetFilteredEvidenceWithMock(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockCursor := new(evidence_viewer.MockCursor)

	mockDB.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(mockCursor, nil)
	mockCursor.On("All", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]evidence_viewer.EvidenceResponse)
		*arg = []evidence_viewer.EvidenceResponse{
			{
				ID:        "ev456",
				CaseID:    "case123",
				Filename:  "filtered_photo.jpg",
				FileType:  "image",
				IPFSCID:   "cid456",
				UploadedAt: time.Now().Format(time.RFC3339),
			},
		}
	})
	mockCursor.On("Close", mock.Anything).Return(nil)

	repo := &evidence_viewer.MongoEvidenceRepository{Collection: mockDB}

	filters := map[string]interface{}{
		"file_type": "image",
	}
	results, err := repo.GetFilteredEvidence("case123", filters, "uploaded_at", "desc")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)
	assert.Equal(t, "ev456", results[0].ID)
}
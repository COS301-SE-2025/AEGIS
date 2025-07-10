package unit_tests

import (
	"context"
	"testing"

	"aegis-api/services_/evidence/evidence_viewer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestGetEvidenceFilesByCaseID(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockCursor := new(evidence_viewer.MockCursor)

	mockDB.On("Find", mock.Anything, mock.Anything).Return(mockCursor, nil)
	mockCursor.On("Close", mock.Anything).Return(nil)

	mockCursor.On("All", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]evidence_viewer.EvidenceFile)
		*arg = []evidence_viewer.EvidenceFile{
			{
				ID:   "ev123",
				Data: []byte("mock data"),
			},
		}
	})

	repo := &evidence_viewer.MongoEvidenceRepository{Collection: mockDB}

	files, err := repo.GetEvidenceFilesByCaseID("case123")
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "ev123", files[0].ID)
	assert.Equal(t, []byte("mock data"), files[0].Data)
}

func TestGetEvidenceFileByID(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockSingleResult := new(evidence_viewer.MockSingleResult)

	expected := evidence_viewer.EvidenceFile{
		ID:   "ev123",
		Data: []byte("file bytes"),
	}

	mockDB.On("FindOne", mock.Anything, mock.Anything).Return(mockSingleResult)
	mockSingleResult.On("Decode", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*evidence_viewer.EvidenceFile)
		*arg = expected
	})

	repo := &evidence_viewer.MongoEvidenceRepository{Collection: mockDB}

	result, err := repo.GetEvidenceFileByID("ev123")
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Data, result.Data)
}

func TestSearchEvidenceFiles(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockCursor := new(evidence_viewer.MockCursor)

	mockDB.On("Find", mock.Anything, mock.Anything).Return(mockCursor, nil)
	mockCursor.On("Close", mock.Anything).Return(nil)

	mockCursor.On("All", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]evidence_viewer.EvidenceFile)
		*arg = []evidence_viewer.EvidenceFile{
			{
				ID:   "ev789",
				Data: []byte("search hit"),
			},
		}
	})

	repo := &evidence_viewer.MongoEvidenceRepository{Collection: mockDB}

	results, err := repo.SearchEvidenceFiles("photo")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ev789", results[0].ID)
}

func TestGetFilteredEvidenceFiles(t *testing.T) {
	mockDB := new(evidence_viewer.MockCollection)
	mockCursor := new(evidence_viewer.MockCursor)

	mockDB.On("Find", mock.Anything, mock.Anything, mock.Anything).Return(mockCursor, nil)
	mockCursor.On("Close", mock.Anything).Return(nil)

	mockCursor.On("All", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*[]evidence_viewer.EvidenceFile)
		*arg = []evidence_viewer.EvidenceFile{
			{
				ID:   "ev456",
				Data: []byte("filtered content"),
			},
		}
	})

	repo := &evidence_viewer.MongoEvidenceRepository{Collection: mockDB}

	filters := map[string]interface{}{
		"file_type": "image",
	}
	results, err := repo.GetFilteredEvidenceFiles("case456", filters, "uploaded_at", "desc")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ev456", results[0].ID)
}

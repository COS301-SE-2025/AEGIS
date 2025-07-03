package evidence_viewer

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo/options"
	
)

type MockCursor struct {
	mock.Mock
}

func (m *MockCursor) All(ctx context.Context, result interface{}) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(Cursor), args.Error(1)
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(SingleResult)
}

package evidence_viewer

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	

)

type EvidenceCollection interface {
	Find(context.Context, interface{}, ...*options.FindOptions) (Cursor, error)
	FindOne(context.Context, interface{}, ...*options.FindOneOptions) SingleResult
}

type Cursor interface {
	All(context.Context, interface{}) error
	Close(context.Context) error
}

type SingleResult interface {
	Decode(v interface{}) error
}

type RealCollection struct {
	Collection *mongo.Collection
}

func (r *RealCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	cursor, err := r.Collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &RealCursor{Cursor: cursor}, nil
}



func (r *RealCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return &RealSingleResult{SingleResult: r.Collection.FindOne(ctx, filter, opts...)}
}

type RealCursor struct {
	Cursor *mongo.Cursor
}

func (c *RealCursor) All(ctx context.Context, results interface{}) error {
	return c.Cursor.All(ctx, results)
}

func (c *RealCursor) Close(ctx context.Context) error {
	return c.Cursor.Close(ctx)
}

type RealSingleResult struct {
	SingleResult *mongo.SingleResult
}

func (r *RealSingleResult) Decode(v interface{}) error {
	return r.SingleResult.Decode(v)
}

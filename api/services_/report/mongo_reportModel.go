package report

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Each subsection of a report
// Combines Postgres metadata and Mongo content sections
type ReportWithContent struct {
	Metadata *Report         `json:"metadata"` // Postgres metadata
	Content  []ReportSection `json:"content"`  // Sections from Mongo
}

// Mongo section struct (if not already defined)
type ReportSection struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	Order     int                `bson:"order" json:"order"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// Full report content document
type ReportContentMongo struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // MongoDB ID for the report content
	ReportID  string             `bson:"report_id"`     // Link to Postgres report metadata
	Sections  []ReportSection    `bson:"sections"`      // Array of subsections
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

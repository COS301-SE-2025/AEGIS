package report

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReportMongoRepository interface {
	SaveReportContent(ctx context.Context, content *ReportContentMongo) error
	GetReportContent(ctx context.Context, reportID primitive.ObjectID) (*ReportContentMongo, error)
	UpdateSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newContent string) error
	AddSection(ctx context.Context, reportID primitive.ObjectID, section ReportSection) error
	DeleteSection(ctx context.Context, reportID, sectionID primitive.ObjectID) error
	UpdateSections(ctx context.Context, reportID primitive.ObjectID, sections []ReportSection) error // for reorder
	FindByReportUUID(ctx context.Context, reportUUID uuid.UUID) (*ReportContentMongo, error)         // for mapping
	UpdateSectionTitle(ctx context.Context, reportID, sectionID primitive.ObjectID, newTitle string) error
	ReorderSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newOrder int) error
	BulkUpdateSections(ctx context.Context, reportID primitive.ObjectID, sections []ReportSection) error
	// NEW: for a batch of reports, return max(sections.updated_at) per report_id (string UUID)
	LatestUpdateByReportIDs(ctx context.Context, reportIDs []string) (map[string]time.Time, error)
}
type ReportMongoRepoImpl struct {
	collection *mongo.Collection
}

// Constructor
func NewReportMongoRepo(coll *mongo.Collection) ReportMongoRepository {
	return &ReportMongoRepoImpl{
		collection: coll,
	}
}

// SaveReportContent saves a new report content document in Mongo
func (r *ReportMongoRepoImpl) SaveReportContent(ctx context.Context, content *ReportContentMongo) error {
	if content.ID.IsZero() {
		content.ID = primitive.NewObjectID()
	}
	if content.CreatedAt.IsZero() {
		content.CreatedAt = time.Now()
	}
	content.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, content)
	return err
}

// GetReportContent fetches the content by Mongo report ID
func (r *ReportMongoRepoImpl) GetReportContent(ctx context.Context, mongoID primitive.ObjectID) (*ReportContentMongo, error) {
	var reportContent ReportContentMongo
	err := r.collection.FindOne(ctx, bson.M{"_id": mongoID}).Decode(&reportContent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve report content from MongoDB: %w", err)
	}
	return &reportContent, nil
}

// UpdateSection updates the content of a subsection
func (r *ReportMongoRepoImpl) UpdateSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newContent string) error {
	filter := bson.M{"_id": reportID, "sections._id": sectionID} // <-- _id, not report_id
	update := bson.M{
		"$set": bson.M{
			"sections.$.content":    newContent,
			"sections.$.updated_at": time.Now(),
			"updated_at":            time.Now(),
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("section not found")
	}
	return nil
}

// AddSection appends a new subsection
func (r *ReportMongoRepoImpl) AddSection(ctx context.Context, reportID primitive.ObjectID, section ReportSection) error {
	if section.ID.IsZero() {
		section.ID = primitive.NewObjectID()
	}
	if section.CreatedAt.IsZero() {
		section.CreatedAt = time.Now()
	}
	section.UpdatedAt = time.Now()

	filter := bson.M{"_id": reportID} // <-- _id
	update := bson.M{
		"$push": bson.M{"sections": section},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("report not found")
	}
	return nil
}

// DeleteSection removes a subsection
func (r *ReportMongoRepoImpl) DeleteSection(ctx context.Context, reportID, sectionID primitive.ObjectID) error {
	filter := bson.M{"_id": reportID}
	update := bson.M{
		"$pull": bson.M{"sections": bson.M{"_id": sectionID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("report not found")
	}
	return nil
}

func (s *ReportServiceImpl) AddCustomSection(ctx context.Context, reportUUID uuid.UUID, title, content string, order int) error {
	// 1. Map reportUUID → Mongo ObjectID
	mongoID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}

	// 2. Create section
	section := ReportSection{
		ID:        primitive.NewObjectID(),
		Title:     title,
		Content:   content,
		Order:     order,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 3. Call Mongo repo
	return s.mongoRepo.AddSection(ctx, mongoID, section)
}

func (s *ReportServiceImpl) DeleteCustomSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID) error {
	mongoID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.DeleteSection(ctx, mongoID, sectionID)
}

func (s *ReportServiceImpl) getMongoID(ctx context.Context, reportUUID uuid.UUID) (primitive.ObjectID, error) {
	// Fetch the Postgres report metadata first
	_, err := s.repo.GetByID(ctx, reportUUID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("report not found: %w", err)
	}

	// Find the Mongo document linked to this report UUID
	mongoReport, err := s.mongoRepo.FindByReportUUID(ctx, reportUUID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("mongo report not found: %w", err)
	}
	if mongoReport == nil {
		return primitive.NilObjectID, fmt.Errorf("mongo report not found")
	}

	return mongoReport.ID, nil
}

func (s *ReportServiceImpl) ReorderSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newOrder int) error {
	mongoDoc, err := s.mongoRepo.FindByReportUUID(ctx, reportUUID)
	if err != nil {
		return err
	}

	var section *ReportSection
	for i := range mongoDoc.Sections {
		if mongoDoc.Sections[i].ID == sectionID {
			section = &mongoDoc.Sections[i]
			break
		}
	}
	if section == nil {
		return fmt.Errorf("section not found")
	}

	section.Order = newOrder
	// Sort sections by Order
	sections := mongoDoc.Sections
	sort.SliceStable(sections, func(i, j int) bool { return sections[i].Order < sections[j].Order })

	return s.mongoRepo.UpdateSections(ctx, mongoDoc.ID, sections)
}

func (r *ReportMongoRepoImpl) FindByReportUUID(ctx context.Context, reportUUID uuid.UUID) (*ReportContentMongo, error) {
	var result ReportContentMongo
	// Assuming you store the Postgres UUID in Mongo as a string
	filter := bson.M{"report_id": reportUUID.String()}

	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (r *ReportMongoRepoImpl) UpdateSections(ctx context.Context, reportID primitive.ObjectID, sections []ReportSection) error {
	filter := bson.M{"_id": reportID}
	update := bson.M{
		"$set": bson.M{
			"sections":   sections,
			"updated_at": time.Now(),
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("report not found")
	}
	return nil
}
func (r *ReportMongoRepoImpl) UpdateSectionTitle(ctx context.Context, reportID, sectionID primitive.ObjectID, newTitle string) error {
	filter := bson.M{"_id": reportID, "sections._id": sectionID}
	update := bson.M{
		"$set": bson.M{
			"sections.$.title":      newTitle,
			"sections.$.updated_at": time.Now(),
			"updated_at":            time.Now(),
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("section not found")
	}
	return nil
}

func (r *ReportMongoRepoImpl) ReorderSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newOrder int) error {
	reportContent, err := r.GetReportContent(ctx, reportID) // uses _id
	if err != nil {
		return err
	}
	var target *ReportSection
	for i := range reportContent.Sections {
		if reportContent.Sections[i].ID == sectionID {
			target = &reportContent.Sections[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("section not found")
	}

	oldOrder := target.Order
	target.Order = newOrder

	for i := range reportContent.Sections {
		if reportContent.Sections[i].ID == sectionID {
			continue
		}
		if oldOrder < newOrder {
			if reportContent.Sections[i].Order > oldOrder && reportContent.Sections[i].Order <= newOrder {
				reportContent.Sections[i].Order--
			}
		} else if oldOrder > newOrder {
			if reportContent.Sections[i].Order >= newOrder && reportContent.Sections[i].Order < oldOrder {
				reportContent.Sections[i].Order++
			}
		}
	}

	return r.UpdateSections(ctx, reportID, reportContent.Sections) // uses _id
}

func (r *ReportMongoRepoImpl) BulkUpdateSections(ctx context.Context, reportID primitive.ObjectID, sections []ReportSection) error {
	// If you keep this method, make it consistent with _id as well:
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": reportID}, bson.M{
		"$set": bson.M{"sections": sections, "updated_at": time.Now()},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("report not found")
	}
	return nil
}

// Ensure this type implements the interface at compile time
var _ ReportMongoRepository = (*ReportMongoRepoImpl)(nil)

// LatestUpdateByReportIDs returns max(updated_at, max(sections.updated_at)) per ReportID (string UUID).
func (r *ReportMongoRepoImpl) LatestUpdateByReportIDs(ctx context.Context, reportIDs []string) (map[string]time.Time, error) {
	out := make(map[string]time.Time, len(reportIDs))
	if len(reportIDs) == 0 {
		return out, nil
	}

	pipeline := mongo.Pipeline{
		// Only consider the requested report_ids
		bson.D{{Key: "$match", Value: bson.D{{Key: "report_id", Value: bson.D{{Key: "$in", Value: reportIDs}}}}}},
		// Keep the top-level updated_at so we can compare later
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "docUpdated", Value: "$updated_at"}}}},
		// Unwind sections (but keep docs even if there are none)
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$sections"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		// Compute per-doc max of sections.updated_at and carry along docUpdated
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$report_id"},
			{Key: "lastSectionsUpdate", Value: bson.D{{Key: "$max", Value: "$sections.updated_at"}}},
			{Key: "docUpdated", Value: bson.D{{Key: "$max", Value: "$docUpdated"}}},
		}}},
		// Final "lastUpdate" = max(lastSectionsUpdate, docUpdated)
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "lastUpdate", Value: bson.D{{Key: "$max", Value: bson.A{"$lastSectionsUpdate", "$docUpdated"}}}},
		}}},
	}

	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var row struct {
			ID         string    `bson:"_id"`
			LastUpdate time.Time `bson:"lastUpdate"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		out[row.ID] = row.LastUpdate
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

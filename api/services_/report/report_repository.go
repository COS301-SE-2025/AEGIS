package report

import (
	reportshared "aegis-api/services_/report/shared"
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
	GetReportContent(ctx context.Context, reportID primitive.ObjectID, tenantID, teamID string) (*ReportContentMongo, error)
	UpdateSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newContent, tenantID, teamID string) error
	AddSection(ctx context.Context, reportID primitive.ObjectID, section ReportSection, tenantID, teamID string) error

	DeleteSection(ctx context.Context, reportID, sectionID primitive.ObjectID, tenantID, teamID string) error
	UpdateSections(ctx context.Context, reportID primitive.ObjectID, sections []ReportSection, tenantID, teamID string) error // for reorder
	FindByReportUUID(ctx context.Context, reportUUID uuid.UUID) (*ReportContentMongo, error)                                  // for mapping
	UpdateSectionTitle(ctx context.Context, reportID, sectionID primitive.ObjectID, newTitle string, tenantID, teamID string) error
	ReorderSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newOrder int, tenantID, teamID string) error
	BulkUpdateSections(ctx context.Context, reportID primitive.ObjectID, sections []ReportSection) error
	// NEW: for a batch of reports, return max(sections.updated_at) per report_id (string UUID)
	LatestUpdateByReportIDs(ctx context.Context, reportIDs []string, tenantID, teamID string) (map[string]time.Time, error)
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
func (r *ReportMongoRepoImpl) ReorderSection(
	ctx context.Context,
	reportID, sectionID primitive.ObjectID,
	newOrder int,
	tenantID, teamID string,
) error {
	// Load scoped doc
	filter := bson.M{"_id": reportID}
	if tenantID != "" {
		filter["tenant_id"] = tenantID
	}
	if teamID != "" {
		filter["team_id"] = teamID
	}

	var doc ReportContentMongo
	if err := r.collection.FindOne(ctx, filter).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrMongoReportNotFound
		}
		return err
	}
	if len(doc.Sections) == 0 {
		return ErrSectionNotFound
	}

	// Find source index
	from := -1
	for i, s := range doc.Sections {
		if s.ID == sectionID {
			from = i
			break
		}
	}
	if from == -1 {
		return ErrSectionNotFound
	}

	// Clamp target (1-based -> 0-based)
	n := len(doc.Sections)
	if newOrder < 1 {
		newOrder = 1
	}
	if newOrder > n {
		newOrder = n
	}
	to := newOrder - 1
	if from == to {
		return nil
	}

	// Move item and renumber 1..N
	moved := doc.Sections[from]
	secs := append(append([]ReportSection{}, doc.Sections[:from]...), doc.Sections[from+1:]...)
	if to > len(secs) {
		to = len(secs)
	}
	secs = append(secs[:to], append([]ReportSection{moved}, secs[to:]...)...)
	for i := range secs {
		secs[i].Order = i + 1
	}

	// Persist — NOTE: POSitional args (Go has no named args)
	return r.UpdateSections(ctx, reportID, secs, tenantID, teamID)
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

func ttFilter(tenantID, teamID string) bson.M {
	f := bson.M{}
	if tenantID != "" {
		f["tenant_id"] = tenantID
	}
	if teamID != "" {
		f["team_id"] = teamID
	}
	return f
}

// GetReportContent fetches the content by Mongo report ID
// GetReportContent
func (r *ReportMongoRepoImpl) GetReportContent(ctx context.Context, mongoID primitive.ObjectID, tenantID, teamID string) (*ReportContentMongo, error) {
	filter := bson.M{"_id": mongoID}
	for k, v := range ttFilter(tenantID, teamID) {
		filter[k] = v
	}

	var reportContent ReportContentMongo
	err := r.collection.FindOne(ctx, filter).Decode(&reportContent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve report content from MongoDB: %w", err)
	}
	return &reportContent, nil
}

// UpdateSection
func (r *ReportMongoRepoImpl) UpdateSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newContent string, tenantID, teamID string) error {
	filter := bson.M{"_id": reportID, "sections._id": sectionID}
	for k, v := range ttFilter(tenantID, teamID) {
		filter[k] = v
	}

	update := bson.M{"$set": bson.M{
		"sections.$.content":    newContent,
		"sections.$.updated_at": time.Now(),
		"updated_at":            time.Now(),
	}}
	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("section not found")
	}
	return nil
}

// AddSection
func (r *ReportMongoRepoImpl) AddSection(ctx context.Context, reportID primitive.ObjectID, section ReportSection, tenantID, teamID string) error {
	if section.ID.IsZero() {
		section.ID = primitive.NewObjectID()
	}
	if section.CreatedAt.IsZero() {
		section.CreatedAt = time.Now()
	}
	section.UpdatedAt = time.Now()

	filter := bson.M{"_id": reportID}
	for k, v := range ttFilter(tenantID, teamID) {
		filter[k] = v
	}

	update := bson.M{"$push": bson.M{"sections": section}, "$set": bson.M{"updated_at": time.Now()}}
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
func (r *ReportMongoRepoImpl) DeleteSection(
	ctx context.Context,
	reportID, sectionID primitive.ObjectID,
	tenantID, teamID string, // NEW
) error {
	filter := bson.M{
		"_id":          reportID,
		"sections._id": sectionID,
	}
	if tenantID != "" {
		filter["tenant_id"] = tenantID
	}
	if teamID != "" {
		filter["team_id"] = teamID
	}

	update := bson.M{
		"$pull": bson.M{"sections": bson.M{"_id": sectionID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrSectionNotFound // use your sentinel error
	}
	return nil
}

func (s *ReportServiceImpl) AddCustomSection(
	ctx context.Context,
	reportUUID uuid.UUID,
	title, content string,
	order int,
) error {
	mongoID, tenantID, teamID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}

	section := ReportSection{
		ID:        primitive.NewObjectID(),
		Title:     title,
		Content:   content,
		Order:     order,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Insert into MongoDB
	errMongo := s.mongoRepo.AddSection(ctx, mongoID, section, tenantID, teamID)
	if errMongo != nil {
		return errMongo
	}
	// Insert into Postgres
	if s.pgSectionRepo != nil {
		sectionPG := reportshared.ReportSection{
			ID:        section.ID.Hex(),
			ReportID:  reportUUID.String(),
			Title:     section.Title,
			Content:   section.Content,
			Order:     section.Order,
			CreatedAt: section.CreatedAt,
			UpdatedAt: section.UpdatedAt,
		}
		if errPG := s.pgSectionRepo.CreateSection(ctx, &sectionPG); errPG != nil {
			return errPG
		}
	}
	return nil
}

func (s *ReportServiceImpl) DeleteCustomSection(
	ctx context.Context,
	reportUUID uuid.UUID,
	sectionID primitive.ObjectID,
) error {
	mongoID, tenantID, teamID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.DeleteSection(ctx, mongoID, sectionID, tenantID, teamID)
}

func (s *ReportServiceImpl) getMongoID(
	ctx context.Context,
	reportUUID uuid.UUID,
) (oid primitive.ObjectID, tenant string, team string, err error) {
	meta, err := s.repo.GetByID(ctx, reportUUID.String())
	if err != nil {
		return primitive.NilObjectID, "", "", fmt.Errorf("%w", ErrReportNotFound)
	}
	if meta.MongoID == "" {
		return primitive.NilObjectID, "", "", fmt.Errorf("%w", ErrMongoReportNotFound)
	}
	oid, err = primitive.ObjectIDFromHex(meta.MongoID)
	if err != nil {
		return primitive.NilObjectID, "", "", fmt.Errorf("invalid mongo id: %w", err)
	}
	return oid, meta.TenantID.String(), meta.TeamID.String(), nil
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

	return s.mongoRepo.UpdateSections(ctx, mongoDoc.ID, sections, mongoDoc.TenantID, mongoDoc.TeamID)

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

// In your repo file

// UpdateSections replaces the entire sections array for a given report document.
// It enforces tenant/team scoping and normalizes timestamps/ids on each section.

func (r *ReportMongoRepoImpl) UpdateSections(
	ctx context.Context,
	reportID primitive.ObjectID,
	sections []ReportSection,
	tenantID, teamID string, // multitenancy guards
) error {
	now := time.Now()

	// Normalize: ensure IDs & timestamps; keep caller-provided Order
	for i := range sections {
		if sections[i].ID.IsZero() {
			sections[i].ID = primitive.NewObjectID()
		}
		if sections[i].CreatedAt.IsZero() {
			sections[i].CreatedAt = now
		}
		sections[i].UpdatedAt = now
	}

	// Keep a deterministic order
	sort.SliceStable(sections, func(i, j int) bool { return sections[i].Order < sections[j].Order })

	// Filter with defense-in-depth
	filter := bson.M{"_id": reportID}
	if tenantID != "" {
		filter["tenant_id"] = tenantID
	}
	if teamID != "" {
		filter["team_id"] = teamID
	}

	update := bson.M{
		"$set": bson.M{
			"sections":   sections,
			"updated_at": now,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrMongoReportNotFound
	}
	return nil
}

func (r *ReportMongoRepoImpl) UpdateSectionTitle(
	ctx context.Context,
	reportID, sectionID primitive.ObjectID,
	newTitle string,
	tenantID, teamID string,
) error {
	filter := bson.M{
		"_id":          reportID,
		"sections._id": sectionID,
	}
	if tenantID != "" {
		filter["tenant_id"] = tenantID
	}
	if teamID != "" {
		filter["team_id"] = teamID
	}

	update := bson.M{"$set": bson.M{
		"sections.$.title":      newTitle,
		"sections.$.updated_at": time.Now(),
		"updated_at":            time.Now(),
	}}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrSectionNotFound
	}
	return nil
}

// var _ ReportMongoRepository = (*ReportMongoRepoImpl)(nil)

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
// var _ ReportMongoRepository = (*ReportMongoRepoImpl)(nil)

// LatestUpdateByReportIDs returns, per report_id, the max of:
//   - top-level updated_at
//   - max(sections.updated_at)
//
// It scopes by tenant/team and tolerates docs with no sections.
func (r *ReportMongoRepoImpl) LatestUpdateByReportIDs(
	ctx context.Context,
	reportIDs []string,
	tenantID string,
	teamID string,
) (map[string]time.Time, error) {
	out := make(map[string]time.Time, len(reportIDs))
	if len(reportIDs) == 0 {
		return out, nil
	}

	// Precise multitenant match
	match := bson.D{{Key: "report_id", Value: bson.D{{Key: "$in", Value: reportIDs}}}}
	if tenantID != "" {
		match = append(match, bson.E{Key: "tenant_id", Value: tenantID})
	}
	if teamID != "" {
		match = append(match, bson.E{Key: "team_id", Value: teamID})
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "docUpdated", Value: "$updated_at"}}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$sections"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$report_id"},
			{Key: "lastSectionsUpdate", Value: bson.D{{Key: "$max", Value: "$sections.updated_at"}}},
			{Key: "docUpdated", Value: bson.D{{Key: "$max", Value: "$docUpdated"}}},
		}}},
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

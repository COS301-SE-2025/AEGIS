// services_/report/repo_mongo_recent.go
package report

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoReportRepository struct {
	col *mongo.Collection // points to the "report_contents" collection
}

// LatestUpdateByReportIDs returns max(section.updated_at) per ReportID (string UUID).
func (m *MongoReportRepository) LatestUpdateByReportIDs(
	ctx context.Context,
	reportIDs []string,
	tenantID string,
	teamID string,
) (map[string]time.Time, error) {
	if len(reportIDs) == 0 {
		return map[string]time.Time{}, nil
	}

	// Build a precise match to leverage index: { tenant_id, team_id, report_id: { $in: [...] } }
	match := bson.M{"report_id": bson.M{"$in": reportIDs}}
	if tenantID != "" {
		match["tenant_id"] = tenantID
	}
	if teamID != "" {
		match["team_id"] = teamID
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$unwind", Value: "$sections"}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":        "$report_id",
			"lastUpdate": bson.M{"$max": "$sections.updated_at"},
		}}},
	}

	cur, err := m.col.Aggregate(ctx, pipeline) // NOTE: use m.collection if that's your field name
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make(map[string]time.Time, len(reportIDs))
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
	return out, cur.Err()
}

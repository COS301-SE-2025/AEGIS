package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenerateReport_EndToEnd(t *testing.T) {
	caseID := uuid.New() // <-- keep as UUID
	w := doRequest("POST", "/reports/cases/"+caseID.String(), "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp struct {
		ID           string    `json:"id"`
		Name         string    `json:"name"`
		Status       string    `json:"status"`
		Version      int       `json:"version"`
		LastModified time.Time `json:"last_modified"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.ID)
	require.Equal(t, "draft", resp.Status)
	require.NotZero(t, resp.LastModified)

	// Use the same collection handle as the rest of your tests
	var doc bson.M
	err := mongoColl.FindOne(context.Background(), bson.M{
		"report_id": resp.ID,
	}).Decode(&doc)
	require.NoError(t, err, "report_contents doc should exist")
}

func TestDownloadReportPDF_ReturnsPDF(t *testing.T) {
	// Create a report (correct path)
	caseID := uuid.New()
	w := doRequest("POST", "/reports/cases/"+caseID.String(), "")
	require.Equal(t, http.StatusOK, w.Code)

	var reportResp struct{ ID string }
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &reportResp))

	// Download PDF (correct path)
	w = doRequest("GET", "/reports/"+reportResp.ID+"/download/pdf", "")
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	require.True(t, bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")))
	require.Greater(t, w.Body.Len(), 100)
}

func TestGetRecentReports_Basic(t *testing.T) {
	// Create two reports (correct path)
	for i := 0; i < 2; i++ {
		w := doRequest("POST", "/reports/cases/"+uuid.New().String(), "")
		require.Equal(t, http.StatusOK, w.Code)
	}

	// Ask for recent (limit=1)
	w := doRequest("GET", "/reports/recent?limit=1", "")
	require.Equal(t, http.StatusOK, w.Code)

	var list []struct {
		ID           string `json:"id"`
		Title        string `json:"title"`
		Status       string `json:"status"`
		LastModified string `json:"lastModified"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list, 1)
	require.NotEmpty(t, list[0].ID)
	require.NotEmpty(t, list[0].LastModified)
}

// Seeds a report row in Postgres and a matching report_contents doc in Mongo.
// Returns reportID + section ObjectID that certainly exist.
func seedReportWithOneSection(t *testing.T) (reportID uuid.UUID, sectionID primitive.ObjectID) {
	t.Helper()

	reportID = uuid.New()
	caseID := uuid.New()
	sectionID = primitive.NewObjectID()
	now := time.Now().UTC()

	// one ObjectID for BOTH Postgres mongo_id and Mongo _id
	docID := primitive.NewObjectID()
	mongoHex := docID.Hex()

	reportNumber := uuid.NewString()[:12]

	// Postgres row: mongo_id MUST equal the Mongo doc's _id hex
	_, err := pgSQL.Exec(`
	  INSERT INTO reports
	    (id, case_id, examiner_id, tenant_id, team_id,
	     name, mongo_id, report_number, status, version, file_path)
	  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'draft',1,$9)
	`, reportID, caseID, FixedUserID, FixedTenantID, FixedTeamID,
		"Seeded Report", mongoHex, reportNumber, "")
	if err != nil {
		t.Fatalf("seed pg reports: %v", err)
	}

	// Mongo doc: _id MUST be that same ObjectID (docID)
	doc := bson.D{
		{Key: "_id", Value: docID},
		{Key: "report_id", Value: reportID.String()},
		{Key: "tenant_id", Value: FixedTenantID.String()},
		{Key: "team_id", Value: FixedTeamID.String()},
		{Key: "sections", Value: bson.A{
			bson.D{
				{Key: "_id", Value: sectionID},
				{Key: "title", Value: "Temp Title"},
				{Key: "content", Value: "initial content"},
				{Key: "order", Value: 1},
				{Key: "created_at", Value: now},
				{Key: "updated_at", Value: now},
			},
		}},
		{Key: "created_at", Value: now},
		{Key: "updated_at", Value: now},
	}
	if _, err := mongoColl.InsertOne(tcCtx, doc); err != nil {
		t.Fatalf("seed mongo report_contents: %v", err)
	}

	return reportID, sectionID
}

// creates a report via API, adds a section via API, then finds that section's ObjectID in Mongo
func createReportAndSection(t *testing.T, title, content string) (reportID uuid.UUID, sectionID primitive.ObjectID) {
	t.Helper()

	// 1) Generate report
	caseID := uuid.New()
	w := doRequest("POST", "/reports/cases/"+caseID.String(), "")

	require.Equal(t, 200, w.Code, w.Body.String())

	var genResp struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &genResp))
	reportID = genResp.ID

	// 2) Add section
	body := fmt.Sprintf(`{"title":%q,"content":%q,"order":1}`, title, content)
	w = doRequest("POST", "/reports/"+reportID.String()+"/sections", body)
	require.Equal(t, 200, w.Code, w.Body.String())

	// 3) Find section in Mongo by title (same tenant/team/report)
	var doc bson.M
	err := mongoColl.FindOne(tcCtx, bson.M{
		"report_id":      reportID.String(),
		"tenant_id":      fixedTenantID.String(),
		"team_id":        fixedTeamID.String(),
		"sections.title": title,
	}).Decode(&doc)
	require.NoError(t, err)

	arr, _ := doc["sections"].(primitive.A)
	require.NotEmpty(t, arr)

	for _, raw := range arr {
		m := raw.(bson.M)
		if m["title"] == title {
			oid, ok := m["_id"].(primitive.ObjectID)
			require.True(t, ok)
			sectionID = oid
			break
		}
	}
	require.False(t, sectionID.IsZero())
	return
}

var (
	fixedTenantID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fixedTeamID   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	//fixedUserID   = uuid.MustParse("33333333-3333-3333-3333-333333333333")
)

func TestUpdateSectionContent_EmptyAllowed(t *testing.T) {
	repID, secID := seedReportWithOneSection(t)

	// PUT /reports/:reportID/sections/:sectionID/content
	url := fmt.Sprintf("/reports/%s/sections/%s/content", repID, secID.Hex())
	w := doRequest("PUT", url, `{"content": ""}`)
	require.Equal(t, 204, w.Code, w.Body.String())

	// verify in Mongo with same tenant/team/report
	var out bson.M
	err := mongoColl.FindOne(tcCtx, bson.M{
		"report_id":    repID.String(),
		"tenant_id":    FixedTenantID.String(),
		"team_id":      FixedTeamID.String(),
		"sections._id": secID,
	}).Decode(&out)
	require.NoError(t, err)

	sections := out["sections"].(primitive.A)
	var found bson.M
	for _, x := range sections {
		m := x.(bson.M)
		if m["_id"] == secID {
			found = m
			break
		}
	}
	require.NotNil(t, found)
	require.Equal(t, "", found["content"])
}

package integration

// import (
// 	"aegis-api/handlers"
// 	"aegis-api/routes"
// 	"aegis-api/services_/auditlog"
// 	chain_of_custody "aegis-api/services_/chain_of_custody"
// 	"context"
// 	"testing"
// 	"time"

// 	"bytes"
// 	"encoding/json"
// 	"io/ioutil"
// 	"net/http"
// 	"net/http/httptest"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// )

// // Helper function to set up the database and router for testing
// func setupTestEnvironment(t *testing.T) (*gin.Engine, *gorm.DB, *mongo.Database) {
// 	// Initialize PostgreSQL DB for testing
// 	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
// 	require.NoError(t, err)
// 	err = db.AutoMigrate(&chain_of_custody.CoCEntryModel{})
// 	require.NoError(t, err)

// 	// Initialize MongoDB for audit logs
// 	mongoDB, err := db.ConnectMongo()
// 	require.NoError(t, err)

// 	// Set up the Gin router
// 	router := gin.Default()
// 	// Setup the necessary services, handlers, and routes
// 	// Initialize necessary services
// 	cocRepo := chain_of_custody.GormRepo(db)
// 	cocSvc := &chain_of_custody.Service{
// 		Repo:  cocRepo,
// 		Authz: chain_of_custody.SimpleAuthz{}, // Assuming all actions are allowed for now
// 		Audit: auditlog.NewAuditLogger(mongoDB, auditlog.NewZapLogger()),
// 	}

// 	cocHandler := handlers.NewCoCHandler(cocSvc, cocSvc.Audit)
// 	// Register the routes
// 	routes.RegisterCoCRoutes(router, cocHandler)

// 	return router, db, mongoDB
// }

// // Test POST /coc/log (Logging a CoC action)
// func TestCoCService_Log_Integration(t *testing.T) {
// 	// Set up the test environment (DB, MongoDB, router)
// 	router, db, mongoDB := setupTestEnvironment(t)

// 	// Prepare test data for the request
// 	params := chain_of_custody.LogParams{
// 		CaseID:     "case-123",
// 		EvidenceID: "evidence-456",
// 		ActorID:    nil,
// 		Action:     chain_of_custody.ActionUpload,
// 		Reason:     nil,
// 		Location:   nil,
// 		HashMD5:    nil,
// 		HashSHA1:   nil,
// 		HashSHA256: nil,
// 		OccurredAt: time.Now(),
// 	}

// 	// Create a mock HTTP request
// 	reqBody := gin.H{
// 		"caseId":     params.CaseID,
// 		"evidenceId": params.EvidenceID,
// 		"action":     params.Action,
// 		"reason":     params.Reason,
// 		"location":   params.Location,
// 		"hashMd5":    params.HashMD5,
// 		"hashSha1":   params.HashSHA1,
// 		"hashSha256": params.HashSHA256,
// 		"occurredAt": params.OccurredAt.Format(time.RFC3339),
// 	}

// 	body, err := json.Marshal(reqBody)
// 	require.NoError(t, err)

// 	req, err := http.NewRequest("POST", "/api/v1/coc/log", nil)
// 	require.NoError(t, err)

// 	req.Body = ioutil.NopCloser(bytes.NewReader(body)) // Setting the body for the request

// 	// Create a response recorder
// 	rr := httptest.NewRecorder()

// 	// Call the handler (This will trigger the actual route logic)
// 	router.ServeHTTP(rr, req)

// 	// Check that the response status code is 200 OK
// 	assert.Equal(t, http.StatusOK, rr.Code)

// 	// Verify data was inserted into PostgreSQL (check the `chain_of_custody` table)
// 	var coCEntry chain_of_custody.CoCEntryModel
// 	err = db.First(&coCEntry, "evidence_id = ?", params.EvidenceID).Error
// 	require.NoError(t, err)
// 	assert.Equal(t, params.EvidenceID, coCEntry.EvidenceID)
// 	assert.Equal(t, "upload", string(coCEntry.Action)) // Ensure action is "upload"

// 	// Verify the audit log in MongoDB
// 	var auditLogs []auditlog.AuditLog
// 	cur, err := mongoDB.Collection("audit_logs").Find(context.Background(), bson.M{"target.id": params.EvidenceID})
// 	require.NoError(t, err)
// 	defer cur.Close(context.Background())

// 	for cur.Next(context.Background()) {
// 		var log auditlog.AuditLog
// 		err := cur.Decode(&log)
// 		require.NoError(t, err)
// 		auditLogs = append(auditLogs, log)
// 	}

// 	require.NoError(t, cur.Err())
// 	require.Len(t, auditLogs, 1)
// 	assert.Equal(t, "CHAIN_OF_CUSTODY_LOG", auditLogs[0].Action)
// 	assert.Equal(t, "SUCCESS", auditLogs[0].Status)
// }

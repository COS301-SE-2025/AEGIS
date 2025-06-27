package integration

// import (
// 	"aegis-api/services_/annotation_threads/messages"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/require"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// )

// func setupRouterAndServices(t *testing.T) (*gin.Engine, messages.MessageService, uuid.UUID) {
// 	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
// 	require.NoError(t, err)
// 	require.NoError(t, db.AutoMigrate(&messages.ThreadMessage{}, &messages.MessageMention{}, &messages.MessageReaction{}))

// 	// seed thread & user
// 	threadID := uuid.New()
// 	userID := uuid.New()
// 	require.NoError(t, db.Exec("CREATE TABLE threads (id UUID PRIMARY KEY)").Error)
// 	require.NoError(t, db.Exec("INSERT INTO threads (id) VALUES (?)", threadID).Error)
// 	require.NoError(t, db.Exec("CREATE TABLE users (id UUID PRIMARY KEY)").Error)
// 	require.NoError(t, db.Exec("INSERT INTO users (id) VALUES (?)", userID).Error)

// 	// repo := messages.NewMessageRepository(db)
// 	// hub := messages.NewMessageService(repo, nil) // no websocket hub for integration
// 	// handler := handlers.NewMessageHandler(hub)

// 	//router := gin.New()
// 	//group := router.Group("/api/v1")
// 	//handlers.RegisterMessageRoutes(group, hub)

// 	//return router, hub, threadID
// }

// func TestSendAndFetchAndApproveReactionFlow(t *testing.T) {
// 	// router, svc, threadID := setupRouterAndServices(t)

// 	// userID := uuid.New()
// 	// // Insert same user into users table so fk passes
// 	// //db := svc.(*messages.MessageServiceImpl).repo.DB
// 	// require.NoError(t, db.Exec("INSERT INTO users (id) VALUES (?)", userID).Error)

// 	// 1. Send message
// 	//messageBody := `{"user_id":"` + userID.String() + `","message":"Hi!","parent_message_id":null,"mentions":[]}`
// 	//req := httptest.NewRequest(http.MethodPost, "/api/v1/threads/"+threadID.String()+"/messages",
// 	//	strings.NewReader(messageBody))
// 	//req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()
// 	//router.ServeHTTP(w, req)
// 	require.Equal(t, http.StatusOK, w.Code)

// 	// parse returned message id
// 	var msg messages.ThreadMessage
// 	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &msg))

// 	// 2. Fetch messages in thread
// 	//req2 := httptest.NewRequest(http.MethodGet, "/api/v1/threads/"+threadID.String()+"/messages", nil)
// 	w2 := httptest.NewRecorder()
// 	//router.ServeHTTP(w2, req2)
// 	require.Equal(t, http.StatusOK, w2.Code)
// 	require.Contains(t, w2.Body.String(), `"Hi!"`)

// 	// 3. Approve message
// 	approveBody := `{"approver_id":"` + userID.String() + `"}`
// 	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/messages/"+msg.ID.String()+"/approve",
// 		strings.NewReader(approveBody))
// 	req3.Header.Set("Content-Type", "application/json")
// 	w3 := httptest.NewRecorder()
// 	router.ServeHTTP(w3, req3)
// 	require.Equal(t, http.StatusOK, w3.Code)

// 	// 4. Add reaction
// 	reactionBody := `{"user_id":"` + userID.String() + `","reaction":"👍"}`
// 	req4 := httptest.NewRequest(http.MethodPost, "/api/v1/messages/"+msg.ID.String()+"/reactions",
// 		strings.NewReader(reactionBody))
// 	req4.Header.Set("Content-Type", "application/json")
// 	w4 := httptest.NewRecorder()
// 	router.ServeHTTP(w4, req4)
// 	require.Equal(t, http.StatusOK, w4.Code)

// 	// 5. Fetch replies (none yet)
// 	req5 := httptest.NewRequest(http.MethodGet, "/api/v1/messages/"+msg.ID.String()+"/replies", nil)
// 	w5 := httptest.NewRecorder()
// 	router.ServeHTTP(w5, req5)
// 	require.Equal(t, http.StatusOK, w5.Code)
// 	require.Equal(t, "[]", w5.Body.String())
// }

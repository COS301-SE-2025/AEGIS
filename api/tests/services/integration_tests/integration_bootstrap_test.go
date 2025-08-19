// file: services_/report/integration_tests/integration_bootstrap_test.go
package integration_test

import (
	"aegis-api/handlers"
	routesPkg "aegis-api/routes"
	report "aegis-api/services_/report"
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// at top-level (same package: integration_test)
type RouteMount func(root *gin.RouterGroup)

var routeMounts []RouteMount

// Call this in test files to add routes into the shared Gin router:
func RegisterRoutes(m RouteMount) { routeMounts = append(routeMounts, m) }

//go:embed testdata/schema.sql
var embeddedSchema []byte

var (
	tcCtx       context.Context
	pgC         *postgres.PostgresContainer
	mongoC      testcontainers.Container
	pgSQL       *sql.DB
	pgDB        *gorm.DB
	mongoClient *mongo.Client
	mongoDB     *mongo.Database
	mongoColl   *mongo.Collection
	router      *gin.Engine
)

func buildRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// existing report wiring...
	pgRepo := report.NewReportRepository(pgDB)
	mRepo := report.NewReportMongoRepo(mongoColl)
	svc := report.NewReportService(pgRepo, mRepo)
	h := handlers.NewReportHandler(svc)

	r := gin.New()
	r.Use(gin.Recovery(), stubAuth())

	routesPkg.RegisterReportRoutes(r.Group(""), h) // your existing line

	// NEW: mount case endpoints for tests
	registerCaseTestEndpoints(r)
	registerCaseAssignmentTestEndpoints(r)

	return r
}

func writeSchemaToTemp() (string, error) {
	dir, err := os.MkdirTemp("", "schema-*")
	if err != nil {
		return "", err
	}
	fp := filepath.Join(dir, "schema.sql")
	if err := os.WriteFile(fp, embeddedSchema, 0o644); err != nil {
		return "", err
	}
	return fp, nil
}

func startPostgres(ctx context.Context) (*postgres.PostgresContainer, *sql.DB, *gorm.DB, error) {
	schemaPath, err := writeSchemaToTemp()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("write schema: %w", err)
	}

	// Use Debian variant to keep contrib/pgcrypto smooth
	container, err := postgres.Run(
		ctx,
		"postgres:16", // <- not alpine
		postgres.WithDatabase("app_database"),
		postgres.WithUsername("app_user"),
		postgres.WithPassword("password"),
		postgres.WithSQLDriver("pgx"),
		postgres.WithInitScripts(schemaPath),
		testcontainers.WithWaitStrategy(
			wait.ForSQL("5432/tcp", "pgx", func(host string, port nat.Port) string {
				return fmt.Sprintf(
					"host=%s port=%s user=app_user password=password dbname=app_database sslmode=disable",
					host, port.Port(),
				)
			}).WithStartupTimeout(2*time.Minute),
		),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("start pg: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("conn string: %w", err)
	}

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("sql open: %w", err)
	}

	// Small retry – sometimes first ping gets a reset during final readiness flips
	deadline := time.Now().Add(20 * time.Second)
	for {
		if err := sqlDB.Ping(); err == nil {
			break
		}
		if time.Now().After(deadline) {
			return nil, nil, nil, fmt.Errorf("sql ping: %w", err)
		}
		time.Sleep(300 * time.Millisecond)
	}

	gdb, err := gorm.Open(gormpg.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gorm open: %w", err)
	}

	return container, sqlDB, gdb, nil
}
func startMongo(ctx context.Context) (testcontainers.Container, *mongo.Client, *mongo.Database, *mongo.Collection, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:6",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor: wait.
			ForLog("Waiting for connections").
			WithOccurrence(1).
			WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("mongo start: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("mongo host: %w", err)
	}
	port, err := c.MappedPort(ctx, "27017/tcp")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("mongo port: %w", err)
	}
	uri := fmt.Sprintf("mongodb://%s:%s", host, port.Port())

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("mongo ping: %w", err)
	}
	db := client.Database("aegis_test")
	return c, client, db, db.Collection("report_contents"), nil
}

// top-level (same package: integration_test)
var (
	FixedUserID   = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	FixedTenantID = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	FixedTeamID   = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
)

func stubAuth() gin.HandlerFunc {
	uid := FixedUserID.String()
	tid := FixedTenantID.String()
	gid := FixedTeamID.String()
	return func(c *gin.Context) {
		c.Set("userID", uid)
		c.Set("tenantID", tid)
		c.Set("teamID", gid)
		c.Next()
	}
}

func TestMain(m *testing.M) {
	tcCtx = context.Background()

	var err error
	pgC, pgSQL, pgDB, err = startPostgres(tcCtx)
	if err != nil {
		fmt.Println("startPostgres:", err)
		os.Exit(1)
	}

	mongoC, mongoClient, mongoDB, mongoColl, err = startMongo(tcCtx)
	if err != nil {
		fmt.Println("startMongo:", err)
		_ = pgC.Terminate(tcCtx)
		os.Exit(1)
	}

	router = buildRouter()

	// ✅ Seed without *testing.T
	if err := seedCoreFixtures(); err != nil {
		fmt.Println("seedCoreFixtures:", err)
		_ = mongoClient.Disconnect(tcCtx)
		_ = mongoC.Terminate(tcCtx)
		_ = pgSQL.Close()
		_ = pgC.Terminate(tcCtx)
		os.Exit(1)
	}

	code := m.Run()

	// teardown...
	if mongoClient != nil {
		_ = mongoClient.Disconnect(tcCtx)
	}
	if mongoC != nil {
		_ = mongoC.Terminate(tcCtx)
	}
	if pgSQL != nil {
		_ = pgSQL.Close()
	}
	if pgC != nil {
		_ = pgC.Terminate(tcCtx)
	}
	os.Exit(code)
}

// Simple helper (reused by tests)
func doRequest(method, url, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

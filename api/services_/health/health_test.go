package health

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"encoding/json"
    "io"
    "net/http"



	 _ "github.com/lib/pq" 
	"github.com/testcontainers/testcontainers-go"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// containers
var (
	pgContainer    testcontainers.Container
	mongoContainer testcontainers.Container
	ipfsContainer  testcontainers.Container
)

// clients
var (
	pgDB     *sql.DB
	mongoCli *mongo.Client
	ipfsAPI  string
)

// // mock IPFS client for tests (replace with real client if you have one)
// type fakeIPFS struct{}

// func (f *fakeIPFS) ID(ctx context.Context) (string, error) {
// 	return "fake-ipfs-id", nil
// }

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error

	// Start PostgreSQL
	pgContainer, pgDB, err = startPostgres(ctx)
	if err != nil {
		fmt.Println("failed to start postgres:", err)
		os.Exit(1)
	}

	// Start MongoDB
	mongoContainer, mongoCli, err = startMongo(ctx)
	if err != nil {
		fmt.Println("failed to start mongo:", err)
		terminate(pgContainer, ctx)
		os.Exit(1)
	}

	// Optionally start IPFS (for now we’ll use a fake client)
	// ipfsContainer, ipfsAPI, err = startIPFS(ctx)

	// Start IPFS container
		ipfsContainer, ipfsAPI, err = startIPFS(ctx)
		if err != nil {
			fmt.Println("failed to start ipfs:", err)
			terminate(pgContainer, ctx)
			terminate(mongoContainer, ctx)
			os.Exit(1)
		}

	code := m.Run()

	terminate(pgContainer, ctx)
	terminate(mongoContainer, ctx)
	terminate(ipfsContainer, ctx)

	os.Exit(code)
}

func startPostgres(ctx context.Context) (testcontainers.Container, *sql.DB, error) {
	container, err := tcPostgres.Run(ctx,
		"postgres:15-alpine",
		tcPostgres.WithDatabase("testdb"),
		tcPostgres.WithUsername("postgres"),
		tcPostgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := container.ConnectionString(ctx,
    "sslmode=disable",
	)
	if err != nil {
		return container, nil, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return container, nil, err
	}

	// verify connectivity
	if err := db.PingContext(ctx); err != nil {
		return container, nil, err
	}

	return container, db, nil
}

func startMongo(ctx context.Context) (testcontainers.Container, *mongo.Client, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:6",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(30 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	if err != nil {
		return nil, nil, err
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "27017")
	uri := fmt.Sprintf("mongodb://%s:%s", host, port.Port())

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return container, nil, err
	}

	// verify connectivity
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return container, nil, err
	}

	return container, client, nil
}
func startIPFS(ctx context.Context) (testcontainers.Container, string, error) {
    req := testcontainers.ContainerRequest{
        Image:        "ipfs/kubo:latest", // Official IPFS image
        ExposedPorts: []string{"5001/tcp"},
        WaitingFor:   wait.ForLog("Daemon is ready").WithStartupTimeout(30 * time.Second),
    }
    
    container, err := testcontainers.GenericContainer(ctx, 
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    if err != nil {
        return nil, "", err
    }

    host, err := container.Host(ctx)
    if err != nil {
        return container, "", err
    }
    port, err := container.MappedPort(ctx, "5001")
    if err != nil {
        return container, "", err
    }

    apiURL := fmt.Sprintf("http://%s:%s", host, port.Port())
    return container, apiURL, nil
}

func terminate(container testcontainers.Container, ctx context.Context) {
	if container != nil {
		_ = container.Terminate(ctx)
	}
}



type simpleIPFSClient struct {
    apiURL string
    client *http.Client
}

func NewSimpleIPFSClient(apiURL string) IPFSClient {
    return &simpleIPFSClient{
        apiURL: apiURL,
        client: &http.Client{Timeout: 10 * time.Second},
    }
}

func (s *simpleIPFSClient) ID(ctx context.Context) (string, error) {
    // Call IPFS API: /api/v0/id
    url := fmt.Sprintf("%s/api/v0/id", s.apiURL)
    req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
    if err != nil {
        return "", err
    }

    resp, err := s.client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("IPFS API error: %s", body)
    }

    // Parse response
    var result struct {
        ID string `json:"ID"`
    }
    if err := json.Unmarshal(body, &result); err != nil {
        return "", err
    }

    return result.ID, nil
}

func TestHealthService_GetHealth(t *testing.T) {
	// build repo & service
	repo := &Repository{
		Postgres: pgDB,
		Mongo:    mongoCli,
		IPFS:     NewSimpleIPFSClient(ipfsAPI), // swap with real IPFS client later
	}
	svc := &Service{Repo: repo}

	// run health check
	result := svc.GetHealth()

// 	if runtime.GOOS == "windows" {
//     return HealthStatus{Name: "disk", Status: "ok"} // or skip entirely
// }

	if result.Status != "ok" {
		t.Errorf("expected status ok, got %s", result.Status)
	}

	// ensure components are returned
	if len(result.Components) == 0 {
		t.Errorf("expected components, got none")
	}

	for _, c := range result.Components {
		if c.Status != "ok" {
			t.Errorf("component %s is unhealthy: %s", c.Name, c.Error)
		}
	}
}

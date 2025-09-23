package health

import (
	"context"
	"database/sql"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"



)

// IPFSClient is an interface placeholder for your IPFS client.
type IPFSClient interface {
	ID(ctx context.Context) (string, error) // simple health check
}

type Repository struct {
	Postgres *sql.DB
	Mongo    *mongo.Client
	IPFS     IPFSClient
}



func (r *Repository) CheckPostgres() ComponentStatus {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := r.Postgres.PingContext(ctx)
	latency := time.Since(start)
	DependencyLatency.WithLabelValues("postgres").Observe(latency.Seconds())

	status := "ok"
	errMsg := ""
	if err != nil {
		status = "unhealthy"
		errMsg = err.Error()
		DependencyErrors.WithLabelValues("postgres").Inc()
	}

	return ComponentStatus{
		Name: "postgres", Status: status, Latency: latency,
		Error: errMsg, Timestamp: time.Now(),
	}
}

func (r *Repository) CheckMongo() ComponentStatus {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := r.Mongo.Ping(ctx, readpref.Primary())
	latency := time.Since(start)
	DependencyLatency.WithLabelValues("mongodb").Observe(latency.Seconds())

	status := "ok"
	errMsg := ""
	if err != nil {
		status = "unhealthy"
		errMsg = err.Error()
		DependencyErrors.WithLabelValues("mongodb").Inc()
	}

	return ComponentStatus{
		Name: "mongodb", Status: status, Latency: latency,
		Error: errMsg, Timestamp: time.Now(),
	}
}

func (r *Repository) CheckIPFS() ComponentStatus {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := r.IPFS.ID(ctx)
	latency := time.Since(start)
	DependencyLatency.WithLabelValues("ipfs").Observe(latency.Seconds())

	status := "ok"
	errMsg := ""
	if err != nil {
		status = "unhealthy"
		errMsg = err.Error()
		DependencyErrors.WithLabelValues("ipfs").Inc()
	}

	return ComponentStatus{
		Name: "ipfs", Status: status, Latency: latency,
		Error: errMsg, Timestamp: time.Now(),
	}
}


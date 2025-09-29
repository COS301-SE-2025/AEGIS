package health

import "time"

type Service struct {
	Repo Checker
}

func (s *Service) GetHealth() HealthResponse {
	components := []ComponentStatus{
		s.Repo.CheckPostgres(),
		s.Repo.CheckMongo(),
		s.Repo.CheckIPFS(),
		s.Repo.CheckDisk(),
		s.Repo.CheckMemory(),
	}

	// Update Prometheus gauges
	UpdateResourceMetrics()

	overall := "ok"
	for _, c := range components {
		if c.Status != "ok" {
			overall = "unhealthy"
			break
		}
	}

	return HealthResponse{
		Status:     overall,
		Timestamp:  time.Now(),
		Components: components,
	}
}


func (s *Service) GetReadiness() bool {
	// readiness = only critical deps
	pg := s.Repo.CheckPostgres()
	mongo := s.Repo.CheckMongo()
	ipfs := s.Repo.CheckIPFS()

	return pg.Status == "ok" && mongo.Status == "ok" && ipfs.Status == "ok"
}

func (s *Service) GetLiveness() bool {
	// If process is alive, always true unless crash
	return true
}

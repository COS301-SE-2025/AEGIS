// services_/report/service_recent.go
package report

import (
	"context"
	"sort"
	"time"
)

func (s *ReportServiceImpl) ListRecentReports(ctx context.Context, opts RecentReportsOptions) ([]RecentReport, error) {
	// 1) Pull a “candidate” window from Postgres (fast)
	candidateLimit := opts.Limit
	if candidateLimit <= 0 {
		candidateLimit = 10
	}
	window := candidateLimit * 3
	if window < 30 {
		window = 30
	}

	candidates, err := s.repo.ListRecentCandidates(ctx, opts, window)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return []RecentReport{}, nil
	}

	// 2) Ask Mongo for max(section.updated_at) per report, scoped by tenant/team
	ids := make([]string, 0, len(candidates))
	for _, r := range candidates {
		ids = append(ids, r.ID.String())
	}
	team := ""
	if opts.TeamID != nil {
		team = opts.TeamID.String()
	}
	mx, err := s.mongoRepo.LatestUpdateByReportIDs(ctx, ids, opts.TenantID.String(), team)
	if err != nil {
		return nil, err
	}

	// 3) Merge: lastModified = max(pg.updated_at, mongoMax[id])
	out := make([]RecentReport, 0, len(candidates))
	for _, r := range candidates {
		last := r.UpdatedAt
		if t, ok := mx[r.ID.String()]; ok && t.After(last) {
			last = t
		}
		out = append(out, RecentReport{
			ID:           r.ID,
			Title:        r.Name,
			Status:       r.Status,
			LastModified: last,
		})
	}

	// 4) Sort by lastModified desc and cut to Limit
	sort.Slice(out, func(i, j int) bool { return out[i].LastModified.After(out[j].LastModified) })
	if len(out) > candidateLimit {
		out = out[:candidateLimit]
	}

	// 5) Present times in Africa/Johannesburg
	if loc, _ := time.LoadLocation("Africa/Johannesburg"); loc != nil {
		for i := range out {
			out[i].LastModified = out[i].LastModified.In(loc)
		}
	}

	return out, nil
}

package unit_tests

import (
	"testing"
	"time"

	graphicalmapping "aegis-api/services_/GraphicalMapping"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockIOCRepo struct {
	iocs []*graphicalmapping.IOC
}

func (m *mockIOCRepo) Create(ioc *graphicalmapping.IOC) error {
	m.iocs = append(m.iocs, ioc)
	return nil
}
func (m *mockIOCRepo) GetByID(id string) (*graphicalmapping.IOC, error) {
	for _, i := range m.iocs {
		if i.ID == id {
			return i, nil
		}
	}
	return nil, nil
}
func (m *mockIOCRepo) ListByCase(caseID string) ([]*graphicalmapping.IOC, error) {
	var out []*graphicalmapping.IOC
	for _, i := range m.iocs {
		if i.CaseID == caseID {
			out = append(out, i)
		}
	}
	return out, nil
}
func (m *mockIOCRepo) ListByTenant(tenantID string) ([]*graphicalmapping.IOC, error) {
	var out []*graphicalmapping.IOC
	for _, i := range m.iocs {
		if i.TenantID == tenantID {
			out = append(out, i)
		}
	}
	return out, nil
}
func (m *mockIOCRepo) FindSimilar(tenantID, iocType, value string) ([]*graphicalmapping.IOC, error) {
	var out []*graphicalmapping.IOC
	for _, i := range m.iocs {
		if i.TenantID == tenantID && i.Type == iocType && i.Value == value {
			out = append(out, i)
		}
	}
	return out, nil
}

func TestAddAndGetIOC(t *testing.T) {
	repo := &mockIOCRepo{}
	service := graphicalmapping.NewIOCService(repo)

	ioc := &graphicalmapping.IOC{
		ID:        uuid.New().String(),
		TenantID:  uuid.New().String(),
		CaseID:    uuid.New().String(),
		Type:      "IP",
		Value:     "192.168.1.1",
		CreatedAt: time.Now().UTC(),
	}
	created, err := service.AddIOC(ioc)
	require.NoError(t, err)
	require.Equal(t, ioc.ID, created.ID)

	fetched, err := service.GetIOC(ioc.ID)
	require.NoError(t, err)
	require.Equal(t, ioc.Value, fetched.Value)
}

func TestListIOCsByCase(t *testing.T) {
	repo := &mockIOCRepo{}
	service := graphicalmapping.NewIOCService(repo)
	caseID := uuid.New().String()
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: "t1", CaseID: caseID, Type: "IP", Value: "1.1.1.1", CreatedAt: time.Now().UTC()})
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: "t1", CaseID: caseID, Type: "Domain", Value: "example.com", CreatedAt: time.Now().UTC()})
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: "t1", CaseID: uuid.New().String(), Type: "IP", Value: "2.2.2.2", CreatedAt: time.Now().UTC()})

	iocs, err := service.ListIOCsByCase(caseID)
	require.NoError(t, err)
	require.Len(t, iocs, 2)
}

func TestBuildIOCGraph(t *testing.T) {
	repo := &mockIOCRepo{}
	service := graphicalmapping.NewIOCService(repo)
	tenantID := "tenant1"
	case1 := "case1"
	case2 := "case2"
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: tenantID, CaseID: case1, Type: "IP", Value: "8.8.8.8", CreatedAt: time.Now().UTC()})
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: tenantID, CaseID: case2, Type: "IP", Value: "8.8.8.8", CreatedAt: time.Now().UTC()})
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: tenantID, CaseID: case1, Type: "Domain", Value: "test.com", CreatedAt: time.Now().UTC()})

	nodes, edges, err := service.BuildIOCGraph(tenantID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(nodes), 3) // 2 cases + 1 shared IOC
	require.GreaterOrEqual(t, len(edges), 3)
}

func TestBuildIOCGraphByCase(t *testing.T) {
	repo := &mockIOCRepo{}
	service := graphicalmapping.NewIOCService(repo)
	tenantID := "tenant1"
	case1 := "case1"
	case2 := "case2"
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: tenantID, CaseID: case1, Type: "IP", Value: "8.8.8.8", CreatedAt: time.Now().UTC()})
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: tenantID, CaseID: case2, Type: "IP", Value: "8.8.8.8", CreatedAt: time.Now().UTC()})
	repo.Create(&graphicalmapping.IOC{ID: uuid.New().String(), TenantID: tenantID, CaseID: case1, Type: "Domain", Value: "test.com", CreatedAt: time.Now().UTC()})

	nodes, edges, err := service.BuildIOCGraphByCase(tenantID, case1)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(nodes), 3) // case1, case2, shared IOC
	require.GreaterOrEqual(t, len(edges), 2)
}

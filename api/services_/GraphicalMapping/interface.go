package graphicalmapping

type IOCRepository interface {
	Create(ioc *IOC) error
	GetByID(id string) (*IOC, error)
	ListByTenant(tenantID string) ([]*IOC, error)
	ListByCase(caseID string) ([]*IOC, error)
	FindSimilar(tenantID string, iocType, value string) ([]*IOC, error)
}

type IOCService interface {
	AddIOC(ioc *IOC) (*IOC, error)
	GetIOC(id string) (*IOC, error)
	ListIOCsForTenant(tenantID string) ([]*IOC, error)
	BuildIOCGraph(tenantID string) (nodes []GraphNode, edges []GraphEdge, err error)
	BuildIOCGraphByCase(tenantID, caseID string) ([]GraphNode, []GraphEdge, error)
	ListIOCsByCase(caseID string) ([]*IOC, error)
}

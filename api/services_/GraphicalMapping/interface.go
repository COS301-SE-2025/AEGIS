package graphicalmapping

type IOCRepository interface {
	Create(ioc *IOC) error
	GetByID(id uint) (*IOC, error)
	ListByTenant(tenantID uint) ([]*IOC, error)
	ListByCase(caseID uint) ([]*IOC, error)
	FindSimilar(tenantID uint, iocType, value string) ([]*IOC, error)
}

type IOCService interface {
	AddIOC(ioc *IOC) error
	GetIOC(id uint) (*IOC, error)
	ListIOCsForTenant(tenantID uint) ([]*IOC, error)
	BuildIOCGraph(tenantID uint) (nodes []GraphNode, edges []GraphEdge, err error)
}

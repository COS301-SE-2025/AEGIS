package graphicalmapping

import (
	"fmt"
)

// GraphNode represents a node in the graph (case or IOC)
type GraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"` // "case" or "ioc"
}

// GraphEdge represents an edge between nodes
type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

type iocService struct {
	repo IOCRepository
}

type CytoscapeElement struct {
	Data map[string]string `json:"data"`
}

func NewIOCService(repo IOCRepository) IOCService {
	return &iocService{repo: repo}
}

func (s *iocService) AddIOC(ioc *IOC) (*IOC, error) {
	err := s.repo.Create(ioc)
	if err != nil {
		return nil, err
	}
	return ioc, nil
}

func (s *iocService) GetIOC(id string) (*IOC, error) {
	return s.repo.GetByID(id)
}
func (s *iocService) ListIOCsByCase(caseID string) ([]*IOC, error) {
	return s.repo.ListByCase(caseID)
}

func (s *iocService) ListIOCsForTenant(tenantID string) ([]*IOC, error) {
	return s.repo.ListByTenant(tenantID)
}

// BuildIOCGraph builds nodes and edges based on IOCs similarity for a tenant
// IOCs that appear in multiple cases will be shared nodes
func (s *iocService) BuildIOCGraph(tenantID string) ([]GraphNode, []GraphEdge, error) {
	iocs, err := s.repo.ListByTenant(tenantID)
	if err != nil {
		return nil, nil, err
	}

	nodes := []GraphNode{}
	edges := []GraphEdge{}

	caseNodeMap := map[string]bool{}
	iocNodeMap := map[string]bool{} // Key: Type:Value, not individual IOC ID

	// Add case nodes
	for _, ioc := range iocs {
		if !caseNodeMap[ioc.CaseID] {
			nodes = append(nodes, GraphNode{
				ID:    fmt.Sprintf("case-%s", ioc.CaseID),
				Label: fmt.Sprintf("Case %s", ioc.CaseID),
				Type:  "case",
			})
			caseNodeMap[ioc.CaseID] = true
		}
	}

	// Group IOCs by Type+Value to create shared nodes
	iocGroups := map[string][]*IOC{}
	for i := range iocs {
		key := iocs[i].Type + ":" + iocs[i].Value
		iocGroups[key] = append(iocGroups[key], iocs[i])
	}

	// Add IOC nodes (one per unique Type:Value combination) and edges from cases to IOCs
	for typeValue, group := range iocGroups {
		// Create single node for this IOC type:value combination
		if !iocNodeMap[typeValue] {
			// Use the first IOC's data for the node (they all have same Type and Value)
			firstIOC := group[0]
			nodes = append(nodes, GraphNode{
				ID:    fmt.Sprintf("ioc-%s", typeValue), // Use Type:Value as unique identifier
				Label: fmt.Sprintf("%s: %s", firstIOC.Type, firstIOC.Value),
				Type:  "ioc",
			})
			iocNodeMap[typeValue] = true
		}

		// Add edges from all cases that contain this IOC to the shared IOC node
		casesSeen := map[string]bool{}
		for _, ioc := range group {
			if !casesSeen[ioc.CaseID] {
				edges = append(edges, GraphEdge{
					Source: fmt.Sprintf("case-%s", ioc.CaseID),
					Target: fmt.Sprintf("ioc-%s", typeValue),
					Label:  "contains",
				})
				casesSeen[ioc.CaseID] = true
			}
		}
	}

	return nodes, edges, nil
}

// ConvertToCytoscapeElements converts graph nodes and edges to Cytoscape format expected by the frontend
func ConvertToCytoscapeElements(nodes []GraphNode, edges []GraphEdge) []CytoscapeElement {
	var elements []CytoscapeElement

	for _, n := range nodes {
		elements = append(elements, CytoscapeElement{
			Data: map[string]string{
				"id":    n.ID,
				"label": n.Label,
				"type":  n.Type,
			},
		})
	}

	for _, e := range edges {
		elements = append(elements, CytoscapeElement{
			Data: map[string]string{
				"source": e.Source,
				"target": e.Target,
				"label":  e.Label,
			},
		})
	}

	return elements
}

// BuildIOCGraphByCase builds a graph for a specific case, including similar IOCs from other cases
// IOCs that appear in multiple cases will be shared nodes
func (s *iocService) BuildIOCGraphByCase(tenantID, caseID string) ([]GraphNode, []GraphEdge, error) {
	// Fetch all IOCs for this tenant
	allIOCs, err := s.repo.ListByTenant(tenantID)
	if err != nil {
		return nil, nil, err
	}

	// Filter to only those for the requested case
	caseIOCs := []*IOC{}
	for _, i := range allIOCs {
		if i.CaseID == caseID {
			caseIOCs = append(caseIOCs, i)
		}
	}
	if len(caseIOCs) == 0 {
		return nil, nil, nil // No IOCs found for this case
	}

	nodes := []GraphNode{}
	edges := []GraphEdge{}

	// Add node for the selected case
	nodes = append(nodes, GraphNode{
		ID:    fmt.Sprintf("case-%s", caseID),
		Label: fmt.Sprintf("Case %s", caseID),
		Type:  "case",
	})

	iocNodeMap := map[string]bool{} // Key: Type:Value
	caseNodeMap := map[string]bool{caseID: true}

	// Process each IOC in the case
	for _, ioc := range caseIOCs {
		typeValue := ioc.Type + ":" + ioc.Value
		iocNodeID := fmt.Sprintf("ioc-%s", typeValue)

		// Add IOC node if not exists (shared node for this Type:Value combination)
		if !iocNodeMap[typeValue] {
			nodes = append(nodes, GraphNode{
				ID:    iocNodeID,
				Label: fmt.Sprintf("%s: %s", ioc.Type, ioc.Value),
				Type:  "ioc",
			})
			iocNodeMap[typeValue] = true
		}

		// Add edge from case to IOC (if not already added)
		caseToIOCExists := false
		for _, e := range edges {
			if e.Source == fmt.Sprintf("case-%s", caseID) && e.Target == iocNodeID && e.Label == "contains" {
				caseToIOCExists = true
				break
			}
		}
		if !caseToIOCExists {
			edges = append(edges, GraphEdge{
				Source: fmt.Sprintf("case-%s", caseID),
				Target: iocNodeID,
				Label:  "contains",
			})
		}

		// Find similar IOCs in other cases
		similarIOCs, err := s.repo.FindSimilar(tenantID, ioc.Type, ioc.Value)
		if err != nil {
			return nil, nil, err
		}

		for _, sim := range similarIOCs {
			// Skip IOCs in the same case (already handled)
			if sim.CaseID == caseID {
				continue
			}

			simCaseID := fmt.Sprintf("case-%s", sim.CaseID)

			// Add similar case node if not exists
			if !caseNodeMap[sim.CaseID] {
				nodes = append(nodes, GraphNode{
					ID:    simCaseID,
					Label: fmt.Sprintf("Case %s", sim.CaseID),
					Type:  "case",
				})
				caseNodeMap[sim.CaseID] = true
			}

			// Add edge from similar case to the SAME IOC node (shared node)
			// Check if edge already exists
			caseToIOCExists := false
			for _, e := range edges {
				if e.Source == simCaseID && e.Target == iocNodeID && e.Label == "contains" {
					caseToIOCExists = true
					break
				}
			}
			if !caseToIOCExists {
				edges = append(edges, GraphEdge{
					Source: simCaseID,
					Target: iocNodeID,
					Label:  "contains",
				})
			}
		}
	}

	return nodes, edges, nil
}

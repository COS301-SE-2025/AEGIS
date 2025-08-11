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
func (s *iocService) BuildIOCGraph(tenantID string) ([]GraphNode, []GraphEdge, error) {
	iocs, err := s.repo.ListByTenant(tenantID)
	if err != nil {
		return nil, nil, err
	}

	nodes := []GraphNode{}
	edges := []GraphEdge{}

	caseNodeMap := map[string]bool{}
	iocNodeIDs := map[string]bool{}

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

	// Add IOC nodes and edges from case -> IOC
	for _, ioc := range iocs {
		iocID := fmt.Sprintf("ioc-%s", ioc.ID)
		if !iocNodeIDs[iocID] {
			nodes = append(nodes, GraphNode{
				ID:    iocID,
				Label: fmt.Sprintf("%s: %s", ioc.Type, ioc.Value),
				Type:  "ioc",
			})
			iocNodeIDs[iocID] = true
		}
		edges = append(edges, GraphEdge{
			Source: fmt.Sprintf("case-%s", ioc.CaseID),
			Target: iocID,
			Label:  "contains",
		})
	}

	// Group IOCs by Type+Value to find similarities
	iocGroups := map[string][]*IOC{}
	for i := range iocs {
		key := iocs[i].Type + ":" + iocs[i].Value
		iocGroups[key] = append(iocGroups[key], iocs[i])
	}

	// Add similarity edges between IOCs of different cases with same Type+Value
	for _, group := range iocGroups {
		if len(group) < 2 {
			continue
		}
		// Connect all pairs
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				edges = append(edges, GraphEdge{
					Source: fmt.Sprintf("ioc-%s", group[i].ID),
					Target: fmt.Sprintf("ioc-%s", group[j].ID),
					Label:  "similar",
				})
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

// BuildIOCGraphByCase builds a graph for a specific case, including similar IOCs from other cases(cross-case relationships)
func (s *iocService) BuildIOCGraphByCase(tenantID, caseID string) ([]GraphNode, []GraphEdge, error) {
	// First we fetch all IOCs for this tenant and case only
	iocs, err := s.repo.ListByTenant(tenantID)
	if err != nil {
		return nil, nil, err
	}

	// Filter to only those for the requested case, same IOCs
	caseIOCs := []*IOC{}
	for _, i := range iocs {
		if i.CaseID == caseID {
			caseIOCs = append(caseIOCs, i)
		}
	}
	if len(caseIOCs) == 0 {
		return nil, nil, nil // IOCS not found for this case, not added yet
	}

	nodes := []GraphNode{}
	edges := []GraphEdge{}

	// Add node for the selected case
	nodes = append(nodes, GraphNode{
		ID:    fmt.Sprintf("case-%s", caseID),
		Label: fmt.Sprintf("Case %s", caseID),
		Type:  "case",
	})

	iocNodeIDs := map[string]bool{}

	// Add IOC nodes + edges from case -> IOC
	for _, ioc := range caseIOCs {
		iocID := fmt.Sprintf("ioc-%s", ioc.ID)
		if !iocNodeIDs[iocID] {
			nodes = append(nodes, GraphNode{
				ID:    iocID,
				Label: fmt.Sprintf("%s: %s", ioc.Type, ioc.Value),
				Type:  "ioc",
			})
			iocNodeIDs[iocID] = true
		}
		edges = append(edges, GraphEdge{
			Source: fmt.Sprintf("case-%s", caseID),
			Target: iocID,
			Label:  "contains",
		})
	}

	// Step 2: For each IOC in the case, find similar IOCs in other cases
	for _, ioc := range caseIOCs {
		similarIOCs, err := s.repo.FindSimilar(tenantID, ioc.Type, ioc.Value)
		if err != nil {
			return nil, nil, err
		}

		for _, sim := range similarIOCs {
			// Skip the IOC itself and those in the same case (already added)
			if sim.ID == ioc.ID || sim.CaseID == caseID {
				continue
			}

			simIOCId := fmt.Sprintf("ioc-%s", sim.ID)
			simCaseId := fmt.Sprintf("case-%s", sim.CaseID)

			// Add similar IOC node if not exists
			if !iocNodeIDs[simIOCId] {
				nodes = append(nodes, GraphNode{
					ID:    simIOCId,
					Label: fmt.Sprintf("%s: %s", sim.Type, sim.Value),
					Type:  "ioc",
				})
				iocNodeIDs[simIOCId] = true
			}

			// Add similar case node if not exists
			caseExists := false
			for _, n := range nodes {
				if n.ID == simCaseId {
					caseExists = true
					break
				}
			}
			if !caseExists {
				nodes = append(nodes, GraphNode{
					ID:    simCaseId,
					Label: fmt.Sprintf("Case %s", sim.CaseID),
					Type:  "case",
				})
			}

			// Add edges: similar IOC belongs to its case
			edges = append(edges, GraphEdge{
				Source: simCaseId,
				Target: simIOCId,
				Label:  "contains",
			})

			// Add similarity edge between the IOC in current case and this similar IOC
			edges = append(edges, GraphEdge{
				Source: fmt.Sprintf("ioc-%s", ioc.ID),
				Target: simIOCId,
				Label:  "similar",
			})
		}
	}

	return nodes, edges, nil
}

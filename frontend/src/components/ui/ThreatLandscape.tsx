import cytoscape from "cytoscape";
import { useEffect, useRef, useState } from "react";
import {jwtDecode} from "jwt-decode";

const ThreatLandscape = ({ userCases }: { userCases: any[] }) => {
  const cyRef = useRef<HTMLDivElement | null>(null);
  const [mode, setMode] = useState<"case" | "network">("case");
  const [selectedCase, setSelectedCase] = useState(userCases[0]?.id || "");
  const [graphData, setGraphData] = useState({ nodes: [], edges: [] });
  const [tenantId, setTenantId] = useState<string>("");

  // Extract tenantId from JWT token on mount
  useEffect(() => {
    const token = sessionStorage.getItem("authToken") || "";
    if (!token) {
      console.error("No auth token found");
      return;
    }

    try {
      const decoded = jwtDecode<any>(token);
      if (decoded?.tenant_id) {
        setTenantId(decoded.tenant_id);
      } else {
        console.error("tenant_id claim missing from token");
      }
    } catch (err) {
      console.error("Failed to decode token", err);
    }
  }, []);

// Update the state to store cytoscape elements directly
const [elements, setElements] = useState<cytoscape.ElementDefinition[]>([]);

// Update the fetch handler
useEffect(() => {
  if (!tenantId) return;

  const fetchGraph = async () => {
    const token = sessionStorage.getItem("authToken") || "";
    let endpoint = "";

    if (mode === "case") {
      if (!selectedCase) return;
      endpoint = `https://localhost/api/v1/tenants/${tenantId}/cases/${selectedCase}/ioc-graph`;
    } else {
      endpoint = `https://localhost/api/v1/tenants/${tenantId}/ioc-graph`;
    }

    try {
      const res = await fetch(endpoint, {
        headers: {
          "Content-Type": "application/json",
          "Authorization": token ?`Bearer ${token}` : "",
        },
      });
      if (!res.ok) throw new Error(res.statusText);

      const data = await res.json();
      
      // Ensure we're getting properly formatted elements
      if (Array.isArray(data)) {
        // If backend returns Cytoscape elements directly
        setElements(data);
      } else if (data.nodes && data.edges) {
        // If backend returns { nodes, edges }, convert them
        const cyElements = [
          ...data.nodes.map((n: any) => ({
            data: {
              id: n.id,
              label: n.label,
              type: n.type,
            },
            group: 'nodes',
          })),
          ...data.edges.map((e: any) => ({
            data: {
              id: `${e.source}-${e.target}-${e.label}`,
              source: e.source,
              target: e.target,
              label: e.label,
            },
            group: 'edges',
          })),
        ];
        setElements(cyElements);
      }
    } catch (err) {
      console.error("Error fetching graph data", err);
    }
  };

  fetchGraph();
}, [mode, selectedCase, tenantId]);

// Update Cytoscape initialization
useEffect(() => {
  if (!cyRef.current || elements.length === 0) return;

  const cy = cytoscape({
    container: cyRef.current,
    elements: elements,
    style: [
      {
        selector: 'node[type="case"]',
        style: {
          'background-color': '#3b82f6',
          shape: 'rectangle',
        },
      },
      {
        selector: 'node[type="ioc"]',
        style: {
          'background-color': '#ef4444',
          shape: 'ellipse',
        },
      },
      {
        selector: 'node',
        style: {
          label: 'data(label)',
          color: '#a08d8dff',
          'text-valign': 'center',
          'text-halign': 'center',
          'font-size': '10px',
          width: 30,
          height: 30,
        },
      },
      {
        selector: 'edge',
        style: {
          'line-color': '#6b7280',
          width: 1.5,
          'curve-style': 'straight',
          'target-arrow-shape': 'triangle',
          'arrow-scale': 0.5,
        },
      },
    ],
    layout: {
      name: 'cose',
      animate: true,
      randomize: true,
    },
  });

  return () => {
    cy.destroy();
  };
}, [elements]);

  return (
    <div className="overflow-hidden w-[550px] h-[366px] rounded-lg border border-border bg-card p-6">
      <div className="flex justify-between items-center mb-2">
        <h2 className="font-bold text-foreground text-lg">Threat Landscape</h2>
        <div className="flex gap-2">
          <button
            className={`px-2 py-1 text-sm rounded transition-colors ${
              mode === "case"
                ? "bg-primary text-primary-foreground"
                : "bg-primary/10 text-primary border border-primary"
            }`}
            onClick={() => setMode("case")}
          >
            Case View
          </button>
          <button
            className={`px-2 py-1 text-sm rounded transition-colors ${
              mode === "network"
                ? "bg-primary text-primary-foreground"
                : "bg-primary/10 text-primary border border-primary"
            }`}
            onClick={() => setMode("network")}
          >
            My Cases Network
          </button>
        </div>
      </div>

      {mode === "case" && (
        <select
          value={selectedCase}
          onChange={(e) => setSelectedCase(e.target.value)}
          className="mb-3 w-full bg-primary/10 text-foreground border border-primary rounded p-1 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
        >
          {userCases.map((c) => (
            <option key={c.id} value={c.id}>
              {c.title}
            </option>
          ))}
        </select>
      )}

      <div ref={cyRef} className="w-full h-[265px]" />
    </div>
  );
};

export { ThreatLandscape };

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

  // Fetch graph data when mode, selectedCase, or tenantId changes
  useEffect(() => {
    if (!tenantId) return; // Wait until tenantId is set

    const fetchGraph = async () => {
      const token = sessionStorage.getItem("authToken") || "";

      let endpoint = "";
      if (mode === "case") {
        if (!selectedCase) return; // no case selected
        endpoint = `http://localhost:8080/api/v1/tenants/${tenantId}/cases/${selectedCase}/ioc-graph`;
      } else {
        endpoint = `http://localhost:8080/api/v1/tenants/${tenantId}/ioc-graph`;
      }

      try {
        const res = await fetch(endpoint, {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!res.ok) {
          console.error("Failed to fetch graph data", res.statusText);
          return;
        }

        const data = await res.json();
        setGraphData(data);
      } catch (err) {
        console.error("Error fetching graph data", err);
      }
    };

    fetchGraph();
  }, [mode, selectedCase, tenantId]);

  // Initialize Cytoscape
  useEffect(() => {
    if (!cyRef.current) return;

    const cy = cytoscape({
      container: cyRef.current,
      elements: graphData,
      style: [
        {
          selector: "node",
          style: {
            "background-color": "#3b82f6",
            label: "data(label)",
            color: "#fff",
            "text-valign": "center",
            "text-halign": "center",
            "font-size": "10px",
          },
        },
        {
          selector: "edge",
          style: {
            "line-color": "#6b7280",
            width: 1.5,
            "curve-style": "straight",
          },
        },
      ],
      layout: { name: "cose" },
    });

    return () => {
      cy.destroy();
    };
  }, [graphData]);

  return (
    <div className="overflow-hidden w-[550px] h-[366px] rounded-lg border bg-card p-6">
      <div className="flex justify-between items-center mb-2">
        <h2 className="font-bold text-white text-lg">Threat Landscape</h2>
        <div className="flex gap-2">
          <button
            className={`px-2 py-1 text-sm rounded ${
              mode === "case" ? "bg-blue-600 text-white" : "bg-muted text-foreground"
            }`}
            onClick={() => setMode("case")}
          >
            Case View
          </button>
          <button
            className={`px-2 py-1 text-sm rounded ${
              mode === "network" ? "bg-blue-600 text-white" : "bg-muted text-foreground"
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
          className="mb-3 w-full bg-muted text-white p-1 rounded border border-border text-sm"
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

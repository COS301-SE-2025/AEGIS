import { useEffect, useState, useRef } from "react";
import { Shield, Plus, Search, Filter, X } from "lucide-react";
import { useNavigate } from "react-router-dom";
import cytoscape from "cytoscape";

interface IOC {
  id: string;
  type: string;
  value: string;
  caseId: string;
  createdAt: string;
}

export const IOCPage: React.FC = () => {
  const [iocs, setIocs] = useState<IOC[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [typeFilter, setTypeFilter] = useState("");
  const [showAddForm, setShowAddForm] = useState(false);
  const [newIOC, setNewIOC] = useState({ type: "", value: "" });
  const [selectedIOC, setSelectedIOC] = useState<IOC | null>(null);
  const cyRef = useRef<HTMLDivElement | null>(null);
  const navigate = useNavigate();
  const token = sessionStorage.getItem("authToken") || "";

  useEffect(() => {
    fetchIOCs();
  }, []);

  const fetchIOCs = async () => {
    const res = await fetch("http://localhost:8080/api/v1/iocs", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const data = await res.json();
      setIocs(data.iocs || []);
    }
  };

  const handleAddIOC = async () => {
    const res = await fetch("http://localhost:8080/api/v1/iocs", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(newIOC),
    });
    if (res.ok) {
      setShowAddForm(false);
      setNewIOC({ type: "", value: "" });
      fetchIOCs();
    }
  };

  // Filtered list for table & graph
  const filteredIOCs = iocs.filter((ioc) => {
    return (
      (!searchTerm || ioc.value.toLowerCase().includes(searchTerm.toLowerCase())) &&
      (!typeFilter || ioc.type === typeFilter)
    );
  });

  // Build graph elements
  const graphElements = [
    ...filteredIOCs.map((ioc) => ({
      data: { id: `ioc-${ioc.id}`, label: `${ioc.type}: ${ioc.value}`, type: "ioc" },
    })),
    ...filteredIOCs.map((ioc) => ({
      data: { id: `case-${ioc.caseId}`, label: `Case ${ioc.caseId}`, type: "case" },
    })),
    ...filteredIOCs.map((ioc) => ({
      data: { source: `ioc-${ioc.id}`, target: `case-${ioc.caseId}` },
    })),
  ];

  // Initialize Cytoscape
  useEffect(() => {
    if (!cyRef.current) return;

    const cy = cytoscape({
      container: cyRef.current,
      elements: graphElements,
      style: [
        {
          selector: "node[type='ioc']",
          style: {
            "background-color": "#3b82f6",
            label: "data(label)",
            color: "#fff",
            "text-valign": "center",
            "text-halign": "center",
            "font-size": "9px",
            "border-width": 2,
            "border-color": "#1e3a8a",
            "shape": "ellipse",
          },
        },
        {
          selector: "node[type='case']",
          style: {
            "background-color": "#10b981",
            label: "data(label)",
            color: "#fff",
            "text-valign": "center",
            "text-halign": "center",
            "font-size": "9px",
            "border-width": 2,
            "border-color": "#064e3b",
            "shape": "round-rectangle",
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

    // Click handler
    cy.on("tap", "node", (evt) => {
      const nodeData = evt.target.data();

      if (nodeData.type === "case") {
        const caseId = nodeData.id.replace("case-", "");
        navigate(`/case-management/${caseId}`);
      } else if (nodeData.type === "ioc") {
        const iocId = nodeData.id.replace("ioc-", "");
        const ioc = iocs.find((item) => item.id === iocId);
        if (ioc) setSelectedIOC(ioc);
      }
    });

    return () => {
      cy.destroy();
    };
  }, [searchTerm, typeFilter, iocs]);

  return (
    <div className="min-h-screen bg-background text-white p-8 flex gap-6">
      {/* Main Content */}
      <div className="flex-1">
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Shield className="text-blue-500" /> IOC Management
          </h1>
          <button
            onClick={() => setShowAddForm(!showAddForm)}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-md"
          >
            <Plus className="w-4 h-4" /> Add IOC
          </button>
        </div>

        {/* Filters */}
        <div className="flex gap-4 mb-6">
          <div className="relative w-1/3">
            <Search className="absolute left-3 top-2.5 w-4 h-4 text-gray-400" />
            <input
              className="w-full bg-card border border-border rounded-lg pl-10 pr-4 py-2 text-sm"
              placeholder="Search by value..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <div className="relative">
            <Filter className="absolute left-3 top-2.5 w-4 h-4 text-gray-400" />
            <select
              className="bg-card border border-border rounded-lg pl-10 pr-4 py-2 text-sm"
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
            >
              <option value="">All Types</option>
              <option value="IP">IP Address</option>
              <option value="Email">Email</option>
              <option value="Domain">Domain</option>
              <option value="Hash">File Hash</option>
              <option value="URL">URL</option>
            </select>
          </div>
          {typeFilter && (
            <button
              onClick={() => setTypeFilter("")}
              className="flex items-center gap-1 px-3 py-2 border border-gray-500 rounded-lg text-sm hover:bg-gray-700"
            >
              <X className="w-3 h-3" /> Clear Filter
            </button>
          )}
        </div>

        {/* Add Form */}
        {showAddForm && (
          <div className="bg-card border border-border rounded-lg p-6 mb-6">
            <h2 className="text-lg font-semibold mb-4">Add New IOC</h2>
            <div className="grid grid-cols-2 gap-4 mb-4">
              <select
                className="bg-muted border border-border rounded-lg px-3 py-2"
                value={newIOC.type}
                onChange={(e) => setNewIOC({ ...newIOC, type: e.target.value })}
              >
                <option value="">Select Type</option>
                <option value="IP">IP Address</option>
                <option value="Email">Email</option>
                <option value="Domain">Domain</option>
                <option value="Hash">File Hash</option>
                <option value="URL">URL</option>
                <option value="Attack Type">Attack</option>
              </select>
              <input
                type="text"
                placeholder="IOC Value"
                className="bg-muted border border-border rounded-lg px-3 py-2"
                value={newIOC.value}
                onChange={(e) => setNewIOC({ ...newIOC, value: e.target.value })}
              />
            </div>
            <div className="flex gap-3">
              <button
                onClick={handleAddIOC}
                className="bg-green-600 hover:bg-green-700 px-4 py-2 rounded-md"
              >
                Save IOC
              </button>
              <button
                onClick={() => setShowAddForm(false)}
                className="bg-gray-600 hover:bg-gray-700 px-4 py-2 rounded-md"
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* IOC Graph */}
        <div className="bg-card border border-border rounded-lg mb-6">
          <div className="p-3 border-b border-border font-semibold text-sm text-foreground">
            IOC Relationship Graph (Click nodes for details)
          </div>
          <div ref={cyRef} className="w-full h-[300px]" />
        </div>

        {/* IOC Table */}
        <div className="bg-card border border-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="p-3 text-left">Type</th>
                <th className="p-3 text-left">Value</th>
                <th className="p-3 text-left">Case</th>
                <th className="p-3 text-left">Date Added</th>
              </tr>
            </thead>
            <tbody>
              {filteredIOCs.map((ioc) => (
                <tr key={ioc.id} className="border-t border-border hover:bg-muted">
                  <td className="p-3">{ioc.type}</td>
                  <td className="p-3">{ioc.value}</td>
                  <td className="p-3">{ioc.caseId}</td>
                  <td className="p-3">{new Date(ioc.createdAt).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {filteredIOCs.length === 0 && (
            <div className="p-6 text-center text-muted-foreground">
              No IOCs found.
            </div>
          )}
        </div>
      </div>

      {/* Side Panel for IOC Details */}
      {selectedIOC && (
        <div className="w-80 bg-card border border-border rounded-lg p-6 flex flex-col">
          <h2 className="text-xl font-bold mb-4 text-blue-400">IOC Details</h2>
          <p className="text-sm mb-2"><strong>Type:</strong> {selectedIOC.type}</p>
          <p className="text-sm mb-2"><strong>Value:</strong> {selectedIOC.value}</p>
          <p className="text-sm mb-2"><strong>Case ID:</strong> {selectedIOC.caseId}</p>
          <p className="text-sm mb-4"><strong>Date Added:</strong> {new Date(selectedIOC.createdAt).toLocaleString()}</p>
          <button
            onClick={() => setSelectedIOC(null)}
            className="bg-gray-600 hover:bg-gray-700 px-4 py-2 rounded-md mt-auto"
          >
            Close
          </button>
        </div>
      )}
    </div>
  );
};

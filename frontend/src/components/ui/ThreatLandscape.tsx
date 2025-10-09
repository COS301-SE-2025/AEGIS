// // import cytoscape from "cytoscape";
// // import { useEffect, useRef, useState } from "react";
// // import {jwtDecode} from "jwt-decode";

// // const ThreatLandscape = ({ userCases }: { userCases: any[] }) => {
// //   const cyRef = useRef<HTMLDivElement | null>(null);
// //   const [mode, setMode] = useState<"case" | "network">("case");
// //   const [selectedCase, setSelectedCase] = useState(userCases[0]?.id || "");
// //   //const [graphData, setGraphData] = useState({ nodes: [], edges: [] });
// //   const [tenantId, setTenantId] = useState<string>("");

// //   // Extract tenantId from JWT token on mount
// //   useEffect(() => {
// //     const token = sessionStorage.getItem("authToken") || "";
// //     if (!token) {
// //       console.error("No auth token found");
// //       return;
// //     }

// //     try {
// //       const decoded = jwtDecode<any>(token);
// //       if (decoded?.tenant_id) {
// //         setTenantId(decoded.tenant_id);
// //       } else {
// //         console.error("tenant_id claim missing from token");
// //       }
// //     } catch (err) {
// //       console.error("Failed to decode token", err);
// //     }
// //   }, []);

// // // Update the state to store cytoscape elements directly
// // const [elements, setElements] = useState<cytoscape.ElementDefinition[]>([]);

// // // Update the fetch handler
// // useEffect(() => {
// //   if (!tenantId) return;

// //   const fetchGraph = async () => {
// //     const token = sessionStorage.getItem("authToken") || "";
// //     let endpoint = "";

// //     if (mode === "case") {
// //       if (!selectedCase) return;
// //       endpoint = `https://localhost/api/v1/tenants/${tenantId}/cases/${selectedCase}/ioc-graph`;
// //     } else {
// //       endpoint = `https://localhost/api/v1/tenants/${tenantId}/ioc-graph`;
// //     }

// //     try {
// //       const res = await fetch(endpoint, {
// //         headers: {
// //           "Content-Type": "application/json",
// //           "Authorization": token ?`Bearer ${token}` : "",
// //         },
// //       });
// //       if (!res.ok) throw new Error(res.statusText);

// //       const data = await res.json();
      
// //       // Ensure we're getting properly formatted elements
// //       if (Array.isArray(data)) {
// //         // If backend returns Cytoscape elements directly
// //         setElements(data);
// //       } else if (data.nodes && data.edges) {
// //         // If backend returns { nodes, edges }, convert them
// //         const cyElements = [
// //           ...data.nodes.map((n: any) => ({
// //             data: {
// //               id: n.id,
// //               label: n.label,
// //               type: n.type,
// //             },
// //             group: 'nodes',
// //           })),
// //           ...data.edges.map((e: any) => ({
// //             data: {
// //               id: `${e.source}-${e.target}-${e.label}`,
// //               source: e.source,
// //               target: e.target,
// //               label: e.label,
// //             },
// //             group: 'edges',
// //           })),
// //         ];
// //         setElements(cyElements);
// //       }
// //     } catch (err) {
// //       console.error("Error fetching graph data", err);
// //     }
// //   };

// //   fetchGraph();
// // }, [mode, selectedCase, tenantId]);

// // // Update Cytoscape initialization
// // useEffect(() => {
// //   if (!cyRef.current || elements.length === 0) return;

// //   const cy = cytoscape({
// //     container: cyRef.current,
// //     elements: elements,
// //     style: [
// //       {
// //         selector: 'node[type="case"]',
// //         style: {
// //           'background-color': '#3b82f6',
// //           shape: 'rectangle',
// //         },
// //       },
// //       {
// //         selector: 'node[type="ioc"]',
// //         style: {
// //           'background-color': '#ef4444',
// //           shape: 'ellipse',
// //         },
// //       },
// //       {
// //         selector: 'node',
// //         style: {
// //           label: 'data(label)',
// //           color: '#a08d8dff',
// //           'text-valign': 'center',
// //           'text-halign': 'center',
// //           'font-size': '10px',
// //           width: 30,
// //           height: 30,
// //         },
// //       },
// //       {
// //         selector: 'edge',
// //         style: {
// //           'line-color': '#6b7280',
// //           width: 1.5,
// //           'curve-style': 'straight',
// //           'target-arrow-shape': 'triangle',
// //           'arrow-scale': 0.5,
// //         },
// //       },
// //     ],
// //     layout: {
// //       name: 'cose',
// //       animate: true,
// //       randomize: true,
// //     },
// //   });

// //   return () => {
// //     cy.destroy();
// //   };
// // }, [elements]);

// //   return (
// //     <div className="overflow-hidden w-[550px] h-[366px] rounded-lg border border-border bg-card p-6">
// //       <div className="flex justify-between items-center mb-2">
// //         <h2 className="font-bold text-foreground text-lg">Threat Landscape</h2>
// //         <div className="flex gap-2">
// //           <button
// //             className={`px-2 py-1 text-sm rounded transition-colors ${
// //               mode === "case"
// //                 ? "bg-primary text-primary-foreground"
// //                 : "bg-primary/10 text-primary border border-primary"
// //             }`}
// //             onClick={() => setMode("case")}
// //           >
// //             Case View
// //           </button>
// //           <button
// //             className={`px-2 py-1 text-sm rounded transition-colors ${
// //               mode === "network"
// //                 ? "bg-primary text-primary-foreground"
// //                 : "bg-primary/10 text-primary border border-primary"
// //             }`}
// //             onClick={() => setMode("network")}
// //           >
// //             My Cases Network
// //           </button>
// //         </div>
// //       </div>

// //       {mode === "case" && (
// //         <select
// //           value={selectedCase}
// //           onChange={(e) => setSelectedCase(e.target.value)}
// //           className="mb-3 w-full bg-primary/10 text-foreground border border-primary rounded p-1 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
// //         >
// //           {userCases.map((c) => (
// //             <option key={c.id} value={c.id}>
// //               {c.title}
// //             </option>
// //           ))}
// //         </select>
// //       )}

// //       <div ref={cyRef} className="w-full h-[265px]" />
// //     </div>
// //   );
// // };

// // export { ThreatLandscape };

// import cytoscape from "cytoscape";
// import { useEffect, useRef, useState } from "react";
// import {jwtDecode} from "jwt-decode";

// const ThreatLandscape = ({ userCases }: { userCases: any[] }) => {
//   const cyRef = useRef<HTMLDivElement | null>(null);
//   const cyInstanceRef = useRef<cytoscape.Core | null>(null);
//   const containerRef = useRef<HTMLDivElement | null>(null);
//   const [mode, setMode] = useState<"case" | "network">("case");
//   const [selectedCase, setSelectedCase] = useState(userCases[0]?.id || "");
//   const [tenantId, setTenantId] = useState<string>("");
//   const [isHovered, setIsHovered] = useState(false);

//   // Extract tenantId from JWT token on mount
//   useEffect(() => {
//     const token = sessionStorage.getItem("authToken") || "";
//     if (!token) {
//       console.error("No auth token found");
//       return;
//     }

//     try {
//       const decoded = jwtDecode<any>(token);
//       if (decoded?.tenant_id) {
//         setTenantId(decoded.tenant_id);
//       } else {
//         console.error("tenant_id claim missing from token");
//       }
//     } catch (err) {
//       console.error("Failed to decode token", err);
//     }
//   }, []);

//   const [elements, setElements] = useState<cytoscape.ElementDefinition[]>([]);

//   useEffect(() => {
//     if (!tenantId) return;

//     const fetchGraph = async () => {
//       const token = sessionStorage.getItem("authToken") || "";
//       let endpoint = "";

//       if (mode === "case") {
//         if (!selectedCase) return;
//         endpoint = `https://localhost/api/v1/tenants/${tenantId}/cases/${selectedCase}/ioc-graph`;
//       } else {
//         endpoint = `https://localhost/api/v1/tenants/${tenantId}/ioc-graph`;
//       }

//       try {
//         const res = await fetch(endpoint, {
//           headers: {
//             "Content-Type": "application/json",
//             "Authorization": token ? `Bearer ${token}` : "",
//           },
//         });
//         if (!res.ok) throw new Error(res.statusText);

//         const data = await res.json();
        
//         if (Array.isArray(data)) {
//           setElements(data);
//         } else if (data.nodes && data.edges) {
//           const cyElements = [
//             ...data.nodes.map((n: any) => ({
//               data: {
//                 id: n.id,
//                 label: n.label,
//                 type: n.type,
//               },
//               group: 'nodes',
//             })),
//             ...data.edges.map((e: any) => ({
//               data: {
//                 id: `${e.source}-${e.target}-${e.label}`,
//                 source: e.source,
//                 target: e.target,
//                 label: e.label,
//               },
//               group: 'edges',
//             })),
//           ];
//           setElements(cyElements);
//         }
//       } catch (err) {
//         console.error("Error fetching graph data", err);
//       }
//     };

//     fetchGraph();
//   }, [mode, selectedCase, tenantId]);

//   useEffect(() => {
//     if (!cyRef.current || elements.length === 0) return;

//     const cy = cytoscape({
//       container: cyRef.current,
//       elements: elements,
//       style: [
//         {
//           selector: 'node[type="case"]',
//           style: {
//             'background-color': '#3b82f6',
//             shape: 'rectangle',
//           },
//         },
//         {
//           selector: 'node[type="ioc"]',
//           style: {
//             'background-color': '#ef4444',
//             shape: 'ellipse',
//           },
//         },
//         {
//           selector: 'node',
//           style: {
//             label: 'data(label)',
//             color: '#a08d8dff',
//             'text-valign': 'center',
//             'text-halign': 'center',
//             'font-size': '10px',
//             width: 30,
//             height: 30,
//           },
//         },
//         {
//           selector: 'edge',
//           style: {
//             'line-color': '#6b7280',
//             width: 1.5,
//             'curve-style': 'straight',
//             'target-arrow-shape': 'triangle',
//             'arrow-scale': 0.5,
//           },
//         },
//       ],
//       layout: {
//         name: 'cose',
//         animate: true,
//         randomize: true,
//       },
//       minZoom: 0.3,
//       maxZoom: 3,
//       wheelSensitivity: 0.2,
//     });

//     cyInstanceRef.current = cy;

//     return () => {
//       cy.destroy();
//       cyInstanceRef.current = null;
//     };
//   }, [elements]);

//   // Handle hover zoom
//   useEffect(() => {
//     if (!cyInstanceRef.current) return;

//     const cy = cyInstanceRef.current;
    
//     if (isHovered) {
//       // Zoom in smoothly
//       cy.animate({
//         zoom: cy.zoom() * 1.8,
//         center: cy.extent(),
//       }, {
//         duration: 300,
//         easing: 'ease-in-out-cubic',
//       });
//     } else {
//       // Zoom out to fit
//       cy.animate({
//         fit: {
//           eles: cy.elements(),
//           padding: 50,
//         },
//       }, {
//         duration: 300,
//         easing: 'ease-in-out-cubic',
//       });
//     }
//   }, [isHovered]);

//   const handleReset = () => {
//     if (cyInstanceRef.current) {
//       cyInstanceRef.current.animate({
//         fit: {
//           eles: cyInstanceRef.current.elements(),
//           padding: 50,
//         },
//       }, {
//         duration: 400,
//         easing: 'ease-in-out-cubic',
//       });
//     }
//   };

//   const handleMouseEnter = () => {
//     setIsHovered(true);
//   };

//   const handleMouseLeave = () => {
//     setIsHovered(false);
//   };

//   return (
//     <div 
//       ref={containerRef}
//       className="relative w-[550px] h-[366px] rounded-lg border border-border bg-card p-6 transition-all duration-300"
//       style={{
//         transform: isHovered ? 'scale(1.5)' : 'scale(1)',
//         transformOrigin: 'center center',
//         zIndex: isHovered ? 50 : 1,
//       }}
//       onMouseEnter={handleMouseEnter}
//       onMouseLeave={handleMouseLeave}
//     >
//       <div className="flex justify-between items-center mb-2">
//         <h2 className="font-bold text-foreground text-lg">Threat Landscape</h2>
//         <div className="flex gap-2">
//           <button
//             className={`px-2 py-1 text-sm rounded transition-colors ${
//               mode === "case"
//                 ? "bg-primary text-primary-foreground"
//                 : "bg-primary/10 text-primary border border-primary"
//             }`}
//             onClick={() => setMode("case")}
//           >
//             Case View
//           </button>
//           <button
//             className={`px-2 py-1 text-sm rounded transition-colors ${
//               mode === "network"
//                 ? "bg-primary text-primary-foreground"
//                 : "bg-primary/10 text-primary border border-primary"
//             }`}
//             onClick={() => setMode("network")}
//           >
//             My Cases Network
//           </button>
//         </div>
//       </div>

//       {mode === "case" && (
//         <select
//           value={selectedCase}
//           onChange={(e) => setSelectedCase(e.target.value)}
//           className="mb-3 w-full bg-primary/10 text-foreground border border-primary rounded p-1 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
//         >
//           {userCases.map((c) => (
//             <option key={c.id} value={c.id}>
//               {c.title}
//             </option>
//           ))}
//         </select>
//       )}

//       <div className="relative">
//         <div ref={cyRef} className="w-full h-[265px]" />
        
//         {/* Reset button overlay */}
//         <button
//           onClick={handleReset}
//           className="absolute bottom-2 right-2 bg-primary text-primary-foreground px-3 py-1 rounded text-xs hover:bg-primary/90 transition-colors shadow-md z-10"
//           title="Reset view"
//         >
//           Reset View
//         </button>
//       </div>
//     </div>
//   );
// };

// export { ThreatLandscape };

import cytoscape from "cytoscape";
//@ts-ignore
import coseBilkent from "cytoscape-cose-bilkent";
import { useEffect, useRef, useState } from "react";
import { jwtDecode } from "jwt-decode";

cytoscape.use(coseBilkent);

const ThreatLandscape = ({ userCases }: { userCases: any[] }) => {
  const cyRef = useRef<HTMLDivElement | null>(null);
  const cyInstanceRef = useRef<cytoscape.Core | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [mode, setMode] = useState<"case" | "network">("case");
  const [selectedCase, setSelectedCase] = useState(userCases[0]?.id || "");
  const [tenantId, setTenantId] = useState<string>("");
  const [isExpanded, setIsExpanded] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

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

  const [elements, setElements] = useState<cytoscape.ElementDefinition[]>([]);

  useEffect(() => {
    if (!tenantId) return;

    const fetchGraph = async () => {
      setIsLoading(true);
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
            "Authorization": token ? `Bearer ${token}` : "",
          },
        });
        if (!res.ok) throw new Error(res.statusText);

        const data = await res.json();
        
        if (Array.isArray(data)) {
          setElements(data);
        } else if (data.nodes && data.edges) {
          const cyElements = [
            ...data.nodes.map((n: any) => ({
              data: {
                id: n.id,
                label: n.label,
                type: n.type,
                // Add additional properties for styling
                severity: n.severity || "medium",
                confidence: n.confidence || "medium",
              },
              group: 'nodes',
            })),
            ...data.edges.map((e: any) => ({
              data: {
                id: `${e.source}-${e.target}-${e.label}`,
                source: e.source,
                target: e.target,
                label: e.label,
                relationship: e.relationship || "related",
              },
              group: 'edges',
            })),
          ];
          setElements(cyElements);
        }
      } catch (err) {
        console.error("Error fetching graph data", err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchGraph();
  }, [mode, selectedCase, tenantId]);

  useEffect(() => {
    if (!cyRef.current || elements.length === 0) return;

    const cy = cytoscape({
      container: cyRef.current,
      elements: elements,
      style: [
        // Case nodes - blue rounded rectangle (more visible)
        {
          selector: 'node[type="case"]',
          style: {
            'background-color': '#3b82f6',
            'background-opacity': 1,
            shape: 'roundrectangle',  // Changed from 'diamond'
            width: 50,                 // Increased from 45
            height: 50,
            'border-width': 3,         // Increased from 2
            'border-color': '#1e40af',
            'border-opacity': 1,
          },
        },
        // IOC nodes - red/orange circles (more visible)
        {
          selector: 'node[type="ioc"]',
          style: {
            'background-color': '#ef4444',  // Default red
            'background-opacity': 1,
            shape: 'ellipse',           // Changed from 'star' to simple circle
            width: 40,                   // Increased from 35
            height: 40,
            'border-width': 2,
            'border-color': '#991b1b',   // Dark red border
            'border-opacity': 1,
          },
        },
        // Specific severity colors for IOC nodes
        {
          selector: 'node[severity="high"]',
          style: {
            'background-color': '#dc2626',  // Darker red
            'border-color': '#7f1d1d',
          },
        },
        {
          selector: 'node[severity="medium"]',
          style: {
            'background-color': '#f59e0b',  // Orange
            'border-color': '#92400e',
          },
        },
        {
          selector: 'node[severity="low"]',
          style: {
            'background-color': '#10b981',  // Green
            'border-color': '#065f46',
          },
        },
        {
          selector: 'node[severity="critical"]',
          style: {
            'background-color': '#991b1b',  // Very dark red
            'border-color': '#450a0a',
            width: 45,
            height: 45,
          },
        },
        // Unknown type nodes - gray hexagon
        {
          selector: 'node',
          style: {
            'background-color': '#6b7280',
            shape: 'hexagon',
            width: 30,
            height: 30,
          },
        },
        // Node labels
        {
          selector: 'node',
          style: {
            label: 'data(label)',
            color: '#1f2937',
            'text-valign': 'center',
            'text-halign': 'center',
            'font-size': '8px',
            'font-weight': 'bold',
            'text-outline-width': 1,
            'text-outline-color': '#ffffff',
            'text-outline-opacity': 0.8,
            'text-wrap': 'wrap',
            'text-max-width': '40px',
          },
        },
        // Edge styles
        {
          selector: 'edge',
          style: {
            'line-color': '#9ca3af',
            width: 1.5,
            'curve-style': 'bezier',
            'target-arrow-shape': 'triangle',
            'target-arrow-color': '#9ca3af',
            'arrow-scale': 0.8,
            'line-style': 'solid',
          },
        },
        // Different line styles for relationships
        {
          selector: 'edge[relationship="strong"]',
          style: {
            width: 3,
            'line-color': '#3b82f6',
            'target-arrow-color': '#3b82f6',
          },
        },
        {
          selector: 'edge[relationship="weak"]',
          style: {
            width: 1,
            'line-style': 'dashed',
            'line-dash-pattern': [5, 5],
          },
        },
      ],
      layout: {
        name: 'cose-bilkent',
        animate: true,
        animationDuration: 1000,
        randomize: true,
        nodeRepulsion: 4500,
        idealEdgeLength: 100,
        nestingFactor: 0.8,
        gravity: 0.25,
        numIter: 2500,
      } as any,
      minZoom: 0.1,
      maxZoom: 5,
      wheelSensitivity: 0.1,
      boxSelectionEnabled: true,
      autounselectify: false,
    });

    // Add node and edge interactions
    cy.on('tap', 'node', function(evt) {
      const node = evt.target;
      const neighborhood = node.neighborhood().add(node);
      
      cy.elements().addClass('faded');
      neighborhood.removeClass('faded');
      
      cy.animate({
        fit: {
          eles: neighborhood,
          padding: 100,
        },
      }, {
        duration: 800,
      });
    });

    cy.on('tap', function(evt) {
      if (evt.target === cy) {
        cy.elements().removeClass('faded');
        cy.animate({
          fit: {
            eles: cy.elements(),
            padding: 50,
          },
        }, {
          duration: 800,
        });
      }
    });

    // Add hover effects - lighten nodes on hover
    cy.on('mouseover', 'node', function(evt) {
      const node = evt.target;
      node.animate({
        style: {
          'border-width': 4,
          'background-opacity': 0.7,  // Make lighter by reducing opacity
        },
      }, {
        duration: 200,
      });
    });

    cy.on('mouseout', 'node', function(evt) {
      const node = evt.target;
      node.animate({
        style: {
          'border-width': 3,  // Back to original (or 2 depending on your base style)
          'background-opacity': 1,  // Back to full opacity
        },
      }, {
        duration: 200,
      });
    });

    cyInstanceRef.current = cy;

    return () => {
      cy.destroy();
      cyInstanceRef.current = null;
    };
  }, [elements]);

  const handleReset = () => {
    if (cyInstanceRef.current) {
      cyInstanceRef.current.elements().removeClass('faded');
      cyInstanceRef.current.animate({
        fit: {
          eles: cyInstanceRef.current.elements(),
          padding: 50,
        },
      }, {
        duration: 800,
        easing: 'ease-in-out-cubic',
      });
    }
  };

  const handleExpand = () => {
    setIsExpanded(!isExpanded);
    // Give a moment for the DOM to update before resetting the view
    setTimeout(handleReset, 100);
  };

  const handleFullscreen = () => {
    if (!containerRef.current) return;
    
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      containerRef.current.requestFullscreen();
    }
  };

  return (
    <div 
      ref={containerRef}
      className={`relative rounded-lg border border-border bg-card p-6 transition-all duration-300 ${
        isExpanded 
          ? "fixed inset-4 z-50 bg-background/95 backdrop-blur-sm" 
          : "w-[550px] h-[366px]"
      }`}
    >
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

      <div className="relative">
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-10">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        )}
        
        <div 
          ref={cyRef} 
          className={`w-full bg-muted/20 rounded border ${
            isExpanded ? "h-[calc(100vh-200px)]" : "h-[265px]"
          }`} 
        />
        
        {/* Control buttons overlay */}
        <div className="absolute bottom-2 right-2 flex gap-2 z-10">
          <button
            onClick={handleReset}
            className="bg-primary text-primary-foreground px-3 py-1 rounded text-xs hover:bg-primary/90 transition-colors shadow-md flex items-center gap-1"
            title="Reset view"
          >
            <span>↺</span> Reset
          </button>
          <button
            onClick={handleExpand}
            className="bg-primary text-primary-foreground px-3 py-1 rounded text-xs hover:bg-primary/90 transition-colors shadow-md flex items-center gap-1"
            title={isExpanded ? "Collapse" : "Expand"}
          >
            <span>{isExpanded ? "⊖" : "⊕"}</span> {isExpanded ? "Collapse" : "Expand"}
          </button>
          <button
            onClick={handleFullscreen}
            className="bg-primary text-primary-foreground px-3 py-1 rounded text-xs hover:bg-primary/90 transition-colors shadow-md"
            title="Fullscreen"
          >
            ⛶
          </button>
        </div>

        {/* Legend */}
        <div className="absolute top-2 left-2 bg-background/80 backdrop-blur-sm rounded-lg p-2 text-xs border z-10">
          <div className="font-semibold mb-1">Legend</div>
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-blue-500 rounded-sm transform rotate-45"></div>
              <span>Case</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-red-500 rounded-full"></div>
              <span>High Severity</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
              <span>Medium Severity</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 bg-green-500 rounded-full"></div>
              <span>Low Severity</span>
            </div>
          </div>
        </div>
      </div>

      {/* Close button for expanded mode */}
      {isExpanded && (
        <button
          onClick={handleExpand}
          className="absolute top-4 right-4 bg-destructive text-destructive-foreground px-3 py-1 rounded text-sm hover:bg-destructive/90 transition-colors z-10"
        >
          ✕ Close
        </button>
      )}
    </div>
  );
};

export { ThreatLandscape };
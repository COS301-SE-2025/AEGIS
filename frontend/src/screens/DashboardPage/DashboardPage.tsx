import {
  Bell,
  FileText,
  Folder,
  Home,
  MessageSquare,
  Search,
  Settings,
  Briefcase,
  CheckCircle,
  Database,
  AlertTriangle,
  Pencil,
  Trash2,
} from "lucide-react";

import { Link } from "react-router-dom";
import { useState, useEffect } from "react";
import { Progress } from "../../components/ui/progress";
import { cn } from "../../lib/utils";
import { SidebarToggleButton } from "../../context/SidebarToggleContext";

import { ClipboardList } from "lucide-react";

import { ThreatLandscape } from "../../components/ui/ThreatLandscape";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import React from "react";



interface CaseCard {
  id: string;
  title: string;
  team_name: string;
  creator: string;
  priority: string;
  description: string;
  lastActivity: string;
  progress: number;
  image: string;
  attackType?: string;
  status: string;
  investigation_stage: string;
}

export const DashBoardPage = () => {
  const [caseCards, setCaseCards] = useState<CaseCard[]>([]);
  const [recentActivities, setRecentActivities] = useState<any[]>([]);
  const [activeTab, setActiveTab] = useState("active");
  const [, setProfile] = useState<{ name: string; email: string; role: string; image: string } | null>(null);
  const storedUser = sessionStorage.getItem("user");
  const user = storedUser ? JSON.parse(storedUser) : null;
  const displayName = user?.name || user?.email?.split("@")[0] || "Agent User";
  const initials = displayName
    .split(" ")
    .map((part: string) => part[0])
    .join("")
    .toUpperCase();

const [editingCase, setEditingCase] = useState<CaseCard | null>(null);
const [updatedStatus, setUpdatedStatus] = useState("");
const [updatedStage, setUpdatedStage] = useState("");
const [] = useState<File | null>(null);
const [updatedTitle, setUpdatedTitle] = useState("");
const [updatedDescription, setUpdatedDescription] = useState("");
interface Notification {
  id: string;
  message: string;
  read: boolean;
  archived: boolean;
  // Add other properties as needed
}

const [openCases, setOpenCases] = useState([]);
const [closedCases, setClosedCases] = useState([]);
const [archivedCases] = useState([]); // <-- Add this line
const [evidenceCount, setEvidenceCount] = useState(0);
const [evidenceError, setEvidenceError] = useState<string | null>(null);
const [notifications] = useState<Notification[]>([]);
const [searchQuery, setSearchQuery] = useState("");
  interface DashboardTile {
  id: string;
  value: string;
  label: string;
  color: string;
  icon: React.ReactElement;
  isVisible: boolean;
}
// Add these new state variables after your existing useState declarations
const [availableTiles, setAvailableTiles] = useState<DashboardTile[]>(() => {
  // Define the default tiles configuration
  const getDefaultTiles = () => [
    {
      id: "ongoing-cases",
      value: "0", // Will be updated by other effects
      label: "Cases ongoing",
      color: "text-[#636ae8]",
      icon: <Briefcase className="w-[75px] h-[52px] text-[#636ae8] flex-shrink-0" />,
      isVisible: true,
    },
    {
      id: "closed-cases",
      value: "0", // Will be updated by other effects
      label: "Cases Closed",
      color: "text-green-500",
      icon: <CheckCircle className="w-[75px] h-[52px] text-green-500 flex-shrink-0" />,
      isVisible: true,
    },
    {
      id: "evidence-count",
      value: "0", // Will be updated by other effects
      label: "Evidence Collected",
      color: "text-sky-500",
      icon: <Database className="w-[75px] h-[52px] text-sky-500 flex-shrink-0" />,
      isVisible: true,
    },
    {
      id: "total-alerts",
      value: "12", // Replace with actual data
      label: "Active Alerts",
      color: "text-red-500",
      icon: <AlertTriangle className="w-[75px] h-[52px] text-red-500 flex-shrink-0" />,
      isVisible: false,
    },
  ];



  const defaultTiles = getDefaultTiles();

  // Check if there are saved preferences in localStorage
  const savedTiles = localStorage.getItem("dashboardTiles");
  if (savedTiles) {
    try {
      const savedConfig: Array<{id: string; isVisible: boolean}> = JSON.parse(savedTiles);
      return defaultTiles.map(tile => {
        const saved = savedConfig.find((s) => s.id === tile.id);
        return saved ? { ...tile, isVisible: saved.isVisible } : tile;
      });
    } catch (e) {
      console.error("Error parsing saved tiles:", e);
      return defaultTiles;
    }
  }
  
  return defaultTiles;
});

// Dashboard Tile Customization persistence
useEffect(() => {
  // Only save the visibility and order, not the React elements
  const tileConfig = availableTiles.map(tile => ({
    id: tile.id,
    isVisible: tile.isVisible
  }));
  console.log('Saving to localStorage:', tileConfig); // Add this for debugging
  localStorage.setItem("dashboardTiles", JSON.stringify(tileConfig));
}, [availableTiles]);


// ✅ these are outside of the array
const unreadCount = notifications.filter((n) => !n.read && !n.archived).length;
const [role, setRole] = useState<string>(user?.role || "");
const isDFIRAdmin = role === "DFIR Admin";
const [showTileCustomizer, setShowTileCustomizer] = useState(false);

useEffect(() => {
  const fetchCases = async () => {
    try {
      const token = sessionStorage.getItem("authToken") || "";

      let endpoint = "";
      let responseKey = "";

      if (activeTab === "active") {
        endpoint = "http://localhost:8080/api/v1/cases/active";
        responseKey = "cases";
      } else if (activeTab === "closed") {
        endpoint = "http://localhost:8080/api/v1/cases/closed";
        responseKey = "closed_cases";
      }

      const res = await fetch(endpoint, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!res.ok) throw new Error(`Failed to load ${activeTab} cases`);
      const data = await res.json();

      console.log(`Fetched ${activeTab} cases:`, data);

      const rawCases = data[responseKey] || [];

      const mappedCases = rawCases.map((c: any) => ({
        id: c.id,
        title: c.title,
        team_name: c.team_name,
        creator: c.created_by,
        priority: c.priority,
        description: c.description,
        lastActivity: c.created_at,
        progress: c.progress || 0,
        image:
          c.image ||
          "https://www.cwilson.com/app/uploads/2022/11/iStock-962094400-1024x565.jpg",
      }));

      setCaseCards(mappedCases.reverse());
    } catch (err) {
      console.error(`Error fetching ${activeTab} cases:`, err);
      setCaseCards([]);
    }
  };

  fetchCases();
}, [activeTab]);

useEffect(() => {
  const token = sessionStorage.getItem("authToken");
  if (!token) return;

  // Decode JWT to get tenantId
  function getTenantIdFromJWT(jwt: string): string | null {
    try {
      const payload = jwt.split('.')[1];
      const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
      return decoded.tenant_id || decoded.tenantId || null;
    } catch {
      return null;
    }
  }

  const tenantId = getTenantIdFromJWT(token);
  if (!tenantId) return;

  const fetchEvidenceCount = async () => {
    try {
      const res = await fetch(`http://localhost:8080/api/v1/evidence/count/${tenantId}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      const data = await res.json();
      console.log("Evidence count API response:", data, "Status:", res.status);
      if (!res.ok || typeof data.count !== "number") {
        setEvidenceError(`Failed to fetch evidence count. Status: ${res.status}`);
        setEvidenceCount(0);
        setAvailableTiles((prev) =>
          prev.map((tile) =>
            tile.id === "evidence-count"
              ? { ...tile, value: "0" }
              : tile
          )
        );
        return;
      }
      setEvidenceCount(data.count);
      setEvidenceError(null);
      setAvailableTiles((prev) =>
        prev.map((tile) =>
          tile.id === "evidence-count"
            ? { ...tile, value: data.count.toString() }
            : tile
        )
      );
    } catch (error) {
      console.error("Error fetching evidence count:", error);
      setEvidenceError("Error fetching evidence count. See console for details.");
      setEvidenceCount(0);
      setAvailableTiles((prev) =>
        prev.map((tile) =>
          tile.id === "evidence-count"
            ? { ...tile, value: "0" }
            : tile
        )
      );
    }
  };

  fetchEvidenceCount();
}, []);


const handleDragEnd = (result: DropResult) => {
  if (!result.destination) return;

  const items = Array.from(availableTiles);
  const [reorderedItem] = items.splice(result.source.index, 1);
  items.splice(result.destination.index, 0, reorderedItem);

  setAvailableTiles(items);
};

const toggleTileVisibility = (tileId: string) => {
  console.log('Toggling tile:', tileId); // Add this for debugging
  setAvailableTiles(prev => {
    const updated = prev.map(tile =>
      tile.id === tileId ? { ...tile, isVisible: !tile.isVisible } : tile
    );
    console.log('Updated tiles:', updated); // Add this for debugging
    return updated;
  });
};
useEffect(() => {
  const fetchRecentActivities = async () => {
    try {
      const token = sessionStorage.getItem("authToken") || "";
      const res = await fetch(`http://localhost:8080/api/v1/auditlogs/recent/${user.id}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const json = await res.json();
      console.log("Fetched recent activities:", json); // Should contain { data: [...], success: true }

      const activities = json.data || []; // ✅ Use the correct key
      const sorted = activities.sort((a: any, b: any) =>
        new Date(b.Timestamp).getTime() - new Date(a.Timestamp).getTime()
      );

      setRecentActivities(sorted.slice(0, 20));
    } catch (err) {
      console.error("Error fetching recent activities:", err);
      setRecentActivities([]);
    }
  };

  fetchRecentActivities();
}, []);


useEffect(() => {
  const fetchCasesCount = async () => {
    const token = sessionStorage.getItem("authToken") || "";

    try {
      const [openRes, closedRes] = await Promise.all([
        fetch("http://localhost:8080/api/v1/cases/filter?status=open", {
          headers: { "Authorization": `Bearer ${token}` }
        }),
        fetch("http://localhost:8080/api/v1/cases/filter?status=closed", {
          headers: { "Authorization": `Bearer ${token}` }
        }),
      ]);

      const openData = await openRes.json();
      const closedData = await closedRes.json();

      setOpenCases(openData.cases || []);
      setClosedCases(closedData.cases || []);

      // Update tile values
      setAvailableTiles(prev => prev.map(tile => {
        switch(tile.id) {
          case 'ongoing-cases':
            return { ...tile, value: (openData.cases || []).length.toString() };
          case 'closed-cases':
            return { ...tile, value: (closedData.cases || []).length.toString() };
          case 'evidence-count':
            // Only update if evidenceCount is still zero (not set by backend)
            return tile.value === "0"
              ? { ...tile, value: evidenceCount.toString() }
              : tile;
          default:
            return tile;
        }
      }));

    } catch (error) {
      console.error("Failed to fetch cases:", error);
    }
  };

  fetchCasesCount();
}, []);


// Define getIcon ABOVE the .map
const getIcon = (action: string) => {
  if (action.toLowerCase().includes("alert")) return AlertTriangle;
  if (action.toLowerCase().includes("case")) return Briefcase;
  if (action.toLowerCase().includes("evidence")) return FileText;
  if (action.toLowerCase().includes("login")) return Pencil;
  return FileText;
};

const handleSaveCase = async () => {
  if (!editingCase) return;
  
  const token = sessionStorage.getItem("authToken") || "";
  
  try {
    const res = await fetch(`http://localhost:8080/api/v1/cases/${editingCase.id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        title: updatedTitle,
        description: updatedDescription,
        status: updatedStatus,
        investigation_stage: updatedStage,
      }),
    });

    if (!res.ok) throw new Error("Failed to update case");

    const data = await res.json();
    console.log("Case updated:", data);

    // ✅ Close the modal
    setEditingCase(null);

    // ✅ Update the list locally without refetch
    setCaseCards(prev =>
      prev.map(c =>
        c.id === editingCase.id
          ? {
              ...c,
              title: updatedTitle,
              description: updatedDescription,
              status: updatedStatus,
              investigation_stage: updatedStage,
            }
          : c
      )
    );

    //alert("Case updated successfully!");
  } catch (err) {
    console.error("Error updating case:", err);
   // alert("Failed to update case");
  }
};

<ul className="space-y-4">
  {recentActivities.map((activity, index) => {
    const Icon = getIcon(activity.Action);
    const timeAgo = activity.Timestamp
      ? new Date(activity.Timestamp).toLocaleString()
      : "unknown time";

    return (
      <li key={index}>
        <div className="flex items-start gap-3 mb-2">
          <Icon className="w-5 h-5 mt-1 text-foreground" />
          <div>
            <p className="text-foreground text-sm">
              <strong>{activity.Actor?.email}</strong> {activity.Description}
            </p>
            <p className="text-muted-foreground text-xs">{timeAgo}</p>
          </div>
        </div>
        {index < recentActivities.length - 1 && (
          <hr className="w-[500px] border-t-[2px] border-[#8C8D8B]" />
        )}
      </li>
    );
  })}
</ul>

    useEffect(() => {
    const fetchProfile = async () => {
      try {
        const token = sessionStorage.getItem("authToken");
        const res = await fetch(`http://localhost:8080/api/v1/profile/${user?.id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!res.ok) throw new Error("Failed to load profile");

        const result = await res.json();
// keep a simple local Profile state if you want, but importantly set role:
      setRole(result.data.role || "");
        // Update both the state and sessionStorage
        setProfile({
          name: result.data.name,
          email: result.data.email,
          role: result.data.role,
          image: result.data.image_url,
        });

        // Update sessionStorage
      sessionStorage.setItem(
        "user",
        JSON.stringify({
          ...user,
          name: result.data.name,
          email: result.data.email,
          image_url: result.data.image_url,
          role: result.data.role,               // <= IMPORTANT
        })
      );
      } catch (err) {
        console.error("Error fetching profile:", err);
      }
    };

    if (user?.id) fetchProfile();
  }, [user?.id]);




// (optional) quick fallback: if your JWT has a "role" claim, you can seed it here
useEffect(() => {
  if (!role) {
    const token = sessionStorage.getItem("authToken");
    if (token) {
      try {
        const [, payloadB64] = token.split(".");
        const json = JSON.parse(
          decodeURIComponent(
            atob(payloadB64.replace(/-/g, "+").replace(/_/g, "/"))
              .split("")
              .map(c => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
              .join("")
          )
        );
        if (json?.role) setRole(json.role);
      } catch { /* ignore */ }
    }
  }
}, [role]);



  return (
    <div className="min-h-screen bg-background text-white">
      {/* Sidebar */}
      <div className="fixed left-0 top-0 h-full w-80 bg-background border-r border-border p-6 flex flex-col z-10">
        {/* Logo */}
        <div className="flex items-center gap-3 mb-8">
          <div className="w-14 h-14 rounded-lg overflow-hidden">
            <img
              src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
              alt="AEGIS Logo"
              className="w-full h-full object-cover"
            />
          </div>
          <span className="font-bold text-white text-2xl">AEGIS</span>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-2">
          <div className="flex items-center gap-3 bg-blue-600 text-white p-3 rounded-lg">
            <Home className="w-6 h-6" />
            <span className="text-lg">Dashboard</span>
          </div>
          <div className="flex items-center gap-3 text-muted-foreground hover:text-white hover:bg-muted p-3 rounded-lg transition-colors cursor-pointer">
            <FileText className="w-6 h-6" />
            <Link to="/case-management"><span className="text-lg">Case Management</span></Link>
          </div>
          <div className="flex items-center gap-3 text-muted-foreground hover:text-white hover:bg-muted p-3 rounded-lg transition-colors cursor-pointer">
            <Folder className="w-6 h-6" />
            <Link to="/evidence-viewer"><span className="text-lg">Evidence Viewer</span></Link>
          </div>
          <div className="flex items-center gap-3 text-muted-foreground hover:text-white hover:bg-muted p-3 rounded-lg transition-colors cursor-pointer">
            <MessageSquare className="w-6 h-6" />
            <span className="text-lg">
              <Link to="/secure-chat">Secure Chat</Link>
            </span>
          </div>
            {isDFIRAdmin && (
              <div className="flex items-center gap-3 text-muted-foreground hover:text-white hover:bg-muted p-3 rounded-lg transition-colors cursor-pointer">
                <ClipboardList className="w-6 h-6" />
                <Link to="/report-dashboard">
                  <span className="text-lg">Case Reports</span>
                </Link>
              </div>
            )}


        </nav>

        {/* User Profile */}
        <div className="border-t border-bg-accent pt-4">
          <div className="flex items-center gap-3">
            <Link to="/profile">
              {user?.image_url ? (
                <img
                  src={
                    user.image_url.startsWith("http") || user.image_url.startsWith("data:")
                      ? user.image_url
                      : `http://localhost:8080${user.image_url}`
                  }
                  alt="Profile"
                  className="w-12 h-12 rounded-full object-cover"
                />
              ) : (
                <div className="w-12 h-12 bg-muted rounded-full flex items-center justify-center">
                  <span className="text-foreground font-medium">{initials}</span>
                </div>
              )}
            </Link>

            <div>
              <p className="font-semibold text-foreground">{displayName}</p>
              <p className="text-muted-foreground text-sm">{user?.email || "user@dfir.com"}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="ml-80 min-h-screen bg-background">
        {/* Topbar */}
        <div className="sticky top-0 bg-background bg-opacity-100 border-b border-border p-4 z-50">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-6">
              <SidebarToggleButton />
              <button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">Dashboard</button>
              <Link to="/case-management">
                <button className="text-muted-foreground hover:text-white px-4 py-2 rounded-lg transition-colors">
                  Case Management
                </button>
              </Link>
              <Link to="/evidence-viewer">
                <button className="text-muted-foreground hover:text-white px-4 py-2 rounded-lg transition-colors">
                  Evidence Viewer
                </button>
              </Link>
              <button className="text-muted-foreground hover:text-white px-4 py-2 rounded-lg transition-colors">
                <Link to="/secure-chat">Secure Chat</Link>
              </button>
            </div>
            

            <div className="flex items-center gap-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
              <input
                className="w-80 h-12 bg-card border border-muted rounded-lg pl-10 pr-4 text-white placeholder-gray-400 focus:outline-none focus:border-blue-500"
                placeholder="Search cases"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
              <Link to="/notifications">
                <button className="relative p-2 text-muted-foreground hover:text-white transition-colors">
                  <Bell className="w-6 h-6" />
                  {unreadCount > 0 && (
                    <span className="absolute -top-1 -right-1 bg-red-600 text-white text-xs px-1.5 py-0.5 rounded-full">
                      {unreadCount}
                    </span>
                  )}
                </button>
              </Link>

              <Link to="/settings">
                <button className="p-2 text-muted-foreground hover:text-white transition-colors">
                  <Settings className="w-6 h-6" />
                </button>
              </Link>
              <Link to="/profile">
                {user?.image_url ? (
                  <img
                    src={
                      user.image_url.startsWith("http") || user.image_url.startsWith("data:")
                        ? user.image_url
                        : `http://localhost:8080${user.image_url}`
                    }
                    alt="Profile"
                    className="w-10 h-10 rounded-full object-cover"
                  />
                ) : (
                  <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center">
                    <span className="text-foreground font-medium text-sm">{initials}</span>
                  </div>
                )}
              </Link>

            </div>
          </div>
        </div>

        {/* Page Content */}
        <main className="p-8">
        {evidenceError && (
          <div className="mb-4 p-3 bg-red-900 text-red-300 rounded">{evidenceError}</div>
        )}
          <h1 className="text-3xl font-semibold mb-6">Dashboard Overview</h1>

          {/* Metric Cards */}
          {/* Customizable Dashboard Tiles */}
          <div className="flex justify-between items-center mb-4">
            <button
              onClick={() => setShowTileCustomizer(true)}
              className="bg-blue-600 text-white text-sm px-4 py-2 rounded-md hover:bg-blue-700"
            >
              Customize Dashboard
            </button>
          </div>
          <DragDropContext onDragEnd={handleDragEnd}>
            <Droppable droppableId="dashboard-tiles" direction="horizontal">
              {(provided) => (
                <div
                  {...provided.droppableProps}
                  ref={provided.innerRef}
                  className="flex gap-6 flex-wrap mb-8"
                >
                  {availableTiles
                    .filter(tile => tile.isVisible)
                    .map((tile, index) => (
                      <Draggable key={tile.id} draggableId={tile.id} index={index}>
                        {(provided, snapshot) => (
                          <div
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                            {...provided.dragHandleProps}
                            className={`w-[266px] h-[123px] flex-shrink-0 bg-card border-[5px] border rounded-[8px] p-4 flex items-center justify-between ${
                              snapshot.isDragging ? 'opacity-50' : ''
                            }`}
                          >
                            <div>
                              <p className={`text-3xl font-bold ${tile.color}`}>{tile.value}</p>
                              <p className="text-foreground text-sm">{tile.label}</p>
                            </div>
                            {tile.icon}
                          </div>
                        )}
                      </Draggable>
                    ))}
                  {provided.placeholder}
                </div>
              )}
            </Droppable>
          </DragDropContext>

          {/* Threat landscape and recent activities */}
          <div className="mt-[100px] flex gap-6">
          <ThreatLandscape userCases={caseCards} />
          {/* Recent Activities */}

            <div className="w-[529px] h-[366px] flex-shrink-0 rounded-lg border border-border bg-card p-6 overflow-hidden">
              <div className="flex items-center justify-between mb-4">
                <h2 className="font-bold text-[#58a6ff] text-lg flex items-center gap-2">
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z"/>
                  </svg>
                  SECURITY ACTIVITY LOG
                </h2>
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-green-500 animate-pulse"></span>
                  <span className="text-xs text-muted-foreground">LIVE</span>
                </div>
              </div>

              <div className="h-[300px] overflow-y-auto pr-2 custom-scrollbar">
                <ul className="space-y-3">
                  {recentActivities.map((activity, index) => {
                    const Icon = getIcon(activity.Action);
                    const timeAgo = activity.Timestamp 
                      ? new Date(activity.Timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                      : "--:--";
                    const date = activity.Timestamp
                      ? new Date(activity.Timestamp).toLocaleDateString([], { month: 'short', day: 'numeric' })
                      : "---";

                    // Color coding based on action type
                    let activityColor = "text-[#58a6ff]"; // Default blue
                    if (activity.Action.toLowerCase().includes("alert") || activity.Action.toLowerCase().includes("threat")) {
                      activityColor = "text-[#f85149]"; // Red for alerts
                    } else if (activity.Action.toLowerCase().includes("login")) {
                      activityColor = "text-[#3fb950]"; // Green for logins
                    }

                    return (
                      <li key={index} className="group">
                        <div className="flex gap-3 p-2 rounded-md hover:bg-[#161b22] transition-colors">
                          <div className={`flex-shrink-0 mt-1 ${activityColor}`}>
                            <Icon className="w-4 h-4" />
                          </div>
                          
                          <div className="flex-1 min-w-0">
                            <div className="flex justify-between items-baseline gap-2">
                              <p className={`text-sm font-mono truncate ${activityColor}`}>
                                {activity.Actor?.email || "system"}
                              </p>
                              <div className="flex items-center gap-2">
                                <span className="text-xs text-muted-foreground">{date}</span>
                                <span className="text-xs text-[#484f58]">{timeAgo}</span>
                              </div>
                            </div>
                            
                            <p className="text-[#c9d1d9] text-sm mt-1 font-mono">
                              {activity.Description}
                            </p>
                            
                            <div className="mt-1 flex gap-2">
                              <span className="text-[10px] px-2 py-0.5 bg-[#21262d] text-[#8b949e] rounded-full">
                                {activity.Action}
                              </span>
                              {activity.Timestamp && (
                                <span className="text-[10px] px-2 py-0.5 bg-[#1e60e5] bg-opacity-20 text-[#58a6ff] rounded-full">
                                  {new Date(activity.Timestamp).toLocaleTimeString([], { hour12: false }) + " UTC"}
                                </span>
                              )}
                            </div>
                          </div>
                        </div>
                        
                        {index < recentActivities.length - 1 && (
                          <div className="h-px bg-[#21262d] mx-2 my-1 group-last:hidden"></div>
                        )}
                      </li>
                    );
                  })}
                </ul>
              </div>
            </div>
          </div>

          {/* Case cards */}
          <div className="w-full bg-card border border-border rounded-lg mt-8 p-6">
            <div className="flex justify-between items-center mb-4">
              <div className="flex gap-2">
                <button
                  onClick={() => setActiveTab("active")}
                  className={cn(
                    "text-sm rounded-lg h-8 px-4",
                    activeTab === "active"
                      ? "bg-muted text-foreground"
                      : "bg-card text-muted-foreground border border-muted"
                  )}
                >
                  Active Cases ({openCases.length})
                </button>
                <button
                  onClick={() => setActiveTab("archived")}
                  className={cn(
                    "text-sm rounded-lg h-8 px-4",
                    activeTab === "archived"
                      ? "bg-muted text-white"
                      : "bg-card text-muted-foreground border border-muted"
                  )}
                >
                  Archived Cases ({archivedCases.length})
                </button>
                <button
                  onClick={() => setActiveTab("closed")}
                  className={cn(
                    "text-sm rounded-lg h-8 px-4",
                    activeTab === "closed"
                      ? "bg-muted text-white"
                      : "bg-card text-muted-foreground border border-muted"
                  )}
                >
                  Closed Cases ({closedCases.length})
                </button>
              </div>
              <Link to="/create-case">
                <button className="bg-blue-600 text-white text-sm px-4 py-2 rounded-md hover:bg-blue-700">
                  Create Case
                </button>
              </Link>
            </div>

            {caseCards.length === 0 ? (
              <div className="text-center text-muted-foreground py-8">
                <p>No cases found. Create your first case to get started!</p>
              </div>
            ) : (
              <div className="flex flex-wrap gap-6">
              {caseCards
                .filter(card => 
                  searchQuery === "" || 
                  card.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                  card.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
                  card.team_name.toLowerCase().includes(searchQuery.toLowerCase())
                )
                .map((card) => {
                  console.log("Evidence Card ID:", card.id);
                  return (
                    <div
                      key={card.id}
                      className="relative flex flex-col justify-between items-center w-[460px] h-[450px] p-4 bg-card border border rounded-[8px]"
                    >
                      <div className="absolute bottom-3 right-3 flex flex-col items-end gap-2 z-10">
                        {/* Edit button (below) */}
                        <button
                          onClick={() => {
                            setEditingCase(card);
                            setUpdatedStatus(card.status || "open");
                            setUpdatedStage(card.investigation_stage || "Triage");
                            setUpdatedTitle(card.title);
                            setUpdatedDescription(card.description);
                          }}
                          className="text-muted-foreground hover:text-blue-500 transition-colors"
                          title="Edit Case"
                        >
                          <Pencil className="w-4 h-4" />
                        </button>
                        {/* Delete button  */}
                        <button
                          onClick={async () => {
                            const confirmed = window.confirm("Are you sure you want to archive this case?");
                            if (!confirmed) return;

                            const token = sessionStorage.getItem("authToken") || "";

                            try {
                              const res = await fetch(`http://localhost:8080/api/v1/cases/${card.id}`, {
                                method: "PATCH",
                                headers: {
                                  "Content-Type": "application/json",
                                  Authorization: `Bearer ${token}`,
                                },
                                body: JSON.stringify({ status: "archived" }),
                              });

                              if (res.ok) {
                                setCaseCards(prev => prev.filter(c => c.id !== card.id));
                              } else {
                               // alert("Failed to archive the case.");
                              }
                            } catch (error) {
                              console.error("Archive error:", error);
                              //alert("An error occurred.");
                            }
                          }}
                          className="text-muted-foreground hover:text-red-500 transition-colors"
                          title="Archive Case"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>

                      </div>

                      <img
                        src={card.image || "https://www.cwilson.com/app/uploads/2022/11/iStock-962094400-1024x565.jpg"}
                        alt={card.description || "Case image"}
                        width={331}
                        height={180}
                        className="rounded-md mb-3"
                      />
                      <h3 className="text-white text-lg font-bold text-center mb-1">
                        {card.title || "Untitled Case"}
                      </h3>
                      <div className="text-sm text-muted-foreground text-center mb-2">
                        Team: {card.team_name} |  Last Activity: {
                        card.lastActivity
                          ? new Date(card.lastActivity).toLocaleString("en-GB", {
                              day: "2-digit",
                              month: "2-digit",
                              year: "numeric",
                              hour: "2-digit",
                              minute: "2-digit",
                              hour12: false
                            })
                          : "Unknown"
                      }
                      </div>
                      <div className="flex justify-between items-center w-full text-xs mb-1">
                        <div className="flex items-center gap-1">
                          <span
                            className={cn(
                              "w-2 h-2 rounded-full",
                              card.priority === "critical"
                                ? "bg-red-500"
                                : card.priority === "high"
                                ? "bg-orange-400"
                                : card.priority === "mid"
                                ? "bg-yellow-400"
                                : "bg-green-400"
                            )}
                          ></span>
                          <span className="text-muted-foreground capitalize">{card.priority}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <span className="w-2 h-2 rounded-full bg-blue-400"></span>
                          <span className="text-muted-foreground">Ongoing</span>
                        </div>
                      </div>
                      <Progress
                        value={card.progress}
                        className="w-full h-3 bg-muted mb-3 [&>div]:bg-green-500"
                      />
                      <Link to={card.id && card.id.length === 36 ? `/evidence-viewer/${card.id}` : "#"}>
                        <button
                          className="bg-blue-600 text-white text-sm px-14 py-2 rounded hover:bg-muted"
                          disabled={!card.id || card.id.length !== 36}
                          title={!card.id || card.id.length !== 36 ? "Invalid Case ID" : "View Evidence Details"}
                        >
                          View Evidence Details
                        </button>
                      </Link>
                      <Link to={`/case-management/${card.id}`}>
                        <button className="bg-blue-600 text-white text-sm px-14 py-2 rounded hover:bg-muted">
                          View Details
                        </button>
                      </Link>

                    </div>
                  );
                })}
                {editingCase && (
                <div className="fixed inset-0 bg-background flex items-center justify-center z-50">
                  <div className="bg-card p-6 rounded-2xl shadow-lg border border-border w-full max-w-md">
                    <h2 className="text-xl font-semibold text-foreground mb-4">Edit Case</h2>

                    {/* Case Title */}
                    <div className="mb-4">
                      <label className="block text-sm font-medium text-muted-foreground mb-1">
                        Case Title
                      </label>
                      <input
                        type="text"
                        value={updatedTitle}
                        onChange={(e) => setUpdatedTitle(e.target.value)}
                        className="w-full bg-muted text-foreground p-2 rounded-md border border-border"
                      />
                    </div>

                    {/* Case Description */}
                    <div className="mb-4">
                      <label className="block text-sm font-medium text-muted-foreground mb-1">
                        Description
                      </label>
                      <textarea
                        rows={3}
                        value={updatedDescription}
                        onChange={(e) => setUpdatedDescription(e.target.value)}
                        className="w-full bg-muted text-foreground p-2 rounded-md border border-border resize-none"
                      />
                    </div>

                    {/* Status Dropdown */}
                    <div className="mb-4">
                      <label className="block text-sm font-medium text-muted-foreground mb-1">
                        Status
                      </label>
                      <select
                        className="w-full bg-muted text-foreground p-2 rounded-md border border-border"
                        value={updatedStatus}
                        onChange={(e) => setUpdatedStatus(e.target.value)}
                      >
                        <option value="open">Open</option>
                        <option value="ongoing">Ongoing</option>
                        <option value="closed">Closed</option>
                        <option value="under_review">Under Review</option>
                      </select>
                    </div>

                    {/* Investigation Stage Dropdown */}
                    <div className="mb-4">
                      <label className="block text-sm font-medium text-muted-foreground mb-1">
                        Investigation Stage
                      </label>
                      <select
                        className="w-full bg-muted text-foreground p-2 rounded-md border border-border"
                        value={updatedStage}
                        onChange={(e) => setUpdatedStage(e.target.value)}
                      >
                        <option value="Triage">Triage</option>
                        <option value="Evidence Collection">Evidence Collection</option>
                        <option value="Analysis">Analysis</option>
                        <option value="Correlation & Threat Intelligence">Correlation & Threat Intelligence</option>
                        <option value="Containment & Eradication">Containment & Eradication</option>
                        <option value="Recovery">Recovery</option>
                        <option value="Reporting & Documentation">Reporting & Documentation</option>
                        <option value="Case Closure & Review">Case Closure & Review</option>
                      </select>
                    </div>

                    {/* Upload Evidence Button */}
                    <div className="mb-4">
                      <label className="block text-sm font-medium text-muted-foreground mb-1">
                        Upload Evidence
                      </label>
                      <Link to={`/upload-evidence/${editingCase.id}`} className="inline-block w-full">
                        <button className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm">
                          Go to Upload Evidence Page
                        </button>
                      </Link>
                    </div>

                    {/* Assign Members Button */}
                    <div className="mb-4">
                      <label className="block text-sm font-medium text-muted-foreground mb-1">
                        Assign Members
                      </label>
                      <Link to={`/assign-case-members/${editingCase.id}`} className="inline-block w-full">
                        <button className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm">
                          Go to Assign Members Page
                        </button>
                      </Link>
                    </div>

                    {/* Action Buttons */}
                    <div className="flex justify-end gap-3 pt-2">
                      <button
                        onClick={() => setEditingCase(null)}
                        className="px-4 py-2 text-sm text-muted-foreground hover:text-white"
                      >
                        Cancel
                      </button>

                      <button
                        onClick={handleSaveCase}
                        className="px-4 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
                      >
                        Save Changes
                      </button>

                    </div>
                  </div>
                </div>
              )}
              </div>
            )}
          </div>
          {/* Tile Customizer Modal */}
            {showTileCustomizer && (
              <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                <div className="bg-card p-6 rounded-2xl shadow-lg border border-border w-full max-w-md">
                  <h2 className="text-xl font-semibold text-foreground mb-4">Customize Dashboard</h2>
                  
                  <div className="space-y-3 max-h-96 overflow-y-auto">
                    {availableTiles.map(tile => (
                      <div key={tile.id} className="flex items-center justify-between p-3 bg-muted rounded">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 flex items-center justify-center">
                            {React.cloneElement(tile.icon, { className: "w-5 h-5" })}
                          </div>
                          <span className="text-foreground">{tile.label}</span>
                        </div>
                        <button
                          onClick={() => toggleTileVisibility(tile.id)}
                          className={`px-3 py-1 rounded text-sm ${
                            tile.isVisible
                              ? 'bg-green-600 text-white'
                              : 'bg-gray-600 text-gray-300'
                          }`}
                        >
                          {tile.isVisible ? 'Visible' : 'Hidden'}
                        </button>
                      </div>
                    ))}
                  </div>

                  <div className="flex justify-end gap-3 mt-6">
                    <button
                      onClick={() => setShowTileCustomizer(false)}
                      className="px-4 py-2 text-sm text-muted-foreground hover:text-white"
                    >
                      Close
                    </button>
                  </div>
                </div>
              </div>
            )}
        </main>
      </div>
    </div>
  );
};

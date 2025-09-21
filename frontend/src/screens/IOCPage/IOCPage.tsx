import { useEffect, useState } from "react";
import { useTheme } from "../../context/ThemeContext";
import { useParams, useNavigate } from "react-router-dom";
import axios from "axios";
import { Button } from "../../components/ui/button";
import { toast } from "react-hot-toast";
import { Shield, AlertTriangle, Database, Hash, Globe, Mail, Link, FileText, Plus, ArrowLeft, Eye, Clock, Zap } from "lucide-react";

interface Case {
  id: string;
  name: string;
}

interface IOC {
  id: string; // changed to string
  type: string;
  value: string;
  created_at: string;
}

const iocTypes = ["IP", "Email", "Domain", "Hash", "URL"];

// Icon mapping for IOC types
const getIOCIcon = (type: string) => {
  switch (type) {
    case "IP": return <Globe className="w-4 h-4 text-slate-300" />;
    case "Email": return <Mail className="w-4 h-4 text-slate-300" />;
    case "Domain": return <Link className="w-4 h-4 text-slate-300" />;
    case "Hash": return <Hash className="w-4 h-4 text-slate-300" />;
    case "URL": return <FileText className="w-4 h-4 text-slate-300" />;
    default: return <Shield className="w-4 h-4 text-slate-300" />;
  }
};

// Threat level based on IOC type (mock logic for visual appeal)
const getThreatLevel = (type: string) => {
  switch (type) {
    case "IP": return { level: "HIGH", colorClass: "bg-red-500/20 text-red-400 border-red-500/30" };
    case "Hash": return { level: "CRITICAL", colorClass: "bg-red-600/20 text-red-300 border-red-600/30" };
    case "Domain": return { level: "MEDIUM", colorClass: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30" };
    case "Email": return { level: "MEDIUM", colorClass: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30" };
    case "URL": return { level: "HIGH", colorClass: "bg-red-500/20 text-red-400 border-red-500/30" };
    default: return { level: "LOW", colorClass: "bg-green-500/20 text-green-400 border-green-500/30" };
  }
};

export const IOCPage = () => {
  const { theme } = useTheme();
  const { case_id } = useParams<{ case_id: string }>(); // match route param naming
  const navigate = useNavigate();

  // Removed unused caseName state
  const [iocs, setIocs] = useState<IOC[]>([]);
  const [type, setType] = useState("IP");
  const [value, setValue] = useState("");
  const [loading, setLoading] = useState(false);

  const token = sessionStorage.getItem("authToken");

  useEffect(() => {
    async function fetchData() {
      if (!case_id) return;

      if (!token) {
        toast.error("You are not authenticated. Please log in.");
        return;
      }

      try {
        setLoading(true);

        // GET Case details
        await axios.get<Case>(`http://localhost:8080/api/v1/cases/${case_id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        // Removed setCaseName since caseName is unused

        // GET IOCs for case
        const iocRes = await axios.get<IOC[]>(`http://localhost:8080/api/v1/cases/${case_id}/iocs`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (Array.isArray(iocRes.data)) {
          setIocs(iocRes.data);
        } else {
          setIocs([]);
        }
      } catch (err) {
        toast.error("Failed to load case or IOCs");
        setIocs([]);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [case_id, token]);

  async function handleAddIOC() {
    if (!value.trim()) {
      toast.error("Please enter a value for the IOC.");
      return;
    }

    if (!case_id) {
      toast.error("Invalid case ID.");
      return;
    }
    if (!token) {
      toast.error("You are not authenticated. Please log in.");
      return;
    }

    try {
      setLoading(true);

      const res = await axios.post<IOC>(
        `http://localhost:8080/api/v1/cases/${case_id}/iocs`,
        { type, value },
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      setIocs((prev) => [...prev, res.data]);
      setValue("");
      toast.success("IOC added successfully!");
    } catch {
      toast.error("Failed to add IOC.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Theme compatible animated background */}
      <style>{`
        .ioc-gradient {
          background: linear-gradient(to right, var(--background) 1px, transparent 1px),
                      linear-gradient(to bottom, var(--background) 1px, transparent 1px);
          background-size: 4rem 4rem;
          mask-image: radial-gradient(ellipse 60% 50% at 50% 0%, #000 70%, transparent 110%);
          opacity: 0.15;
        }
      `}</style>
      <div className="absolute inset-0 ioc-gradient"></div>
      <div className="relative z-10 p-8 max-w-7xl mx-auto">
        {/* Header Section */}
        <div className="flex items-center gap-4 mb-8">
          <button
            onClick={() => navigate(-1)}
            className="group flex items-center gap-2 px-4 py-2 rounded-lg bg-card border border-border hover:bg-card/80 hover:border-border/80 transition-all duration-200 backdrop-blur-sm"
            aria-label="Go back"
          >
            <ArrowLeft className="w-4 h-4 text-primary group-hover:text-primary-foreground transition-colors" />
            <span className="text-muted-foreground group-hover:text-foreground font-medium">Back</span>
          </button>
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-accent border border shadow-lg">
              <Shield className="w-8 h-8 text-primary" />
            </div>
            <div>
              <h1 className="text-4xl font-bold text-foreground/80">
                Indicators of Compromise
              </h1>
              <p className="text-muted-foreground flex items-center gap-2 mt-1">
                <Database className="w-4 h-4" />
                CaseID: <span className="text-primary font-medium">{case_id}</span>
              </p>
            </div>
          </div>
        </div>

        {/* Statistics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="group bg-card border border-primary/20 rounded-xl p-6 hover:border-primary transition-all duration-300 hover:shadow-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-muted-foreground text-sm font-medium">Total IOCs</p>
                <p className="text-3xl font-bold text-foreground mt-2">{iocs.length}</p>
              </div>
              <div className="p-3 rounded-lg bg-accent border border">
                <Database className="w-8 h-8 text-primary" />
              </div>
            </div>
          </div>
          <div className="group bg-destructive/10 border border-destructive rounded-xl p-6 hover:border-destructive-foreground transition-all duration-300 hover:shadow-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-muted-foreground text-sm font-medium">High Threat</p>
                <p className="text-3xl font-bold text-destructive mt-2">
                  {iocs.filter(ioc => getThreatLevel(ioc.type).level === 'HIGH' || getThreatLevel(ioc.type).level === 'CRITICAL').length}
                </p>
              </div>
              <div className="p-3 rounded-lg bg-destructive/20 border border">
                <AlertTriangle className="w-8 h-8 text-destructive" />
              </div>
            </div>
          </div>
          <div className="group bg-secondary/10 border border-secondary rounded-xl p-6 hover:border-secondary-foreground transition-all duration-300 hover:shadow-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-muted-foreground text-sm font-medium">Medium Threat</p>
                <p className="text-3xl font-bold text-secondary mt-2">
                  {iocs.filter(ioc => getThreatLevel(ioc.type).level === 'MEDIUM').length}
                </p>
              </div>
              <div className="p-3 rounded-lg bg-secondary/20 border border">
                <Eye className="w-8 h-8 text-secondary" />
              </div>
            </div>
          </div>
          <div className="group bg-card border border-primary/20 rounded-xl p-6 hover:border-primary transition-all duration-300 hover:shadow-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-muted-foreground text-sm font-medium">Last Added</p>
                <p className="text-lg font-bold text-foreground mt-2">
                  {iocs.length > 0 ? new Date(Math.max(...iocs.map(ioc => new Date(ioc.created_at).getTime()))).toLocaleDateString() : 'None'}
                </p>
              </div>
              <div className="p-3 rounded-lg bg-accent border border">
                <Clock className="w-8 h-8 text-muted-foreground" />
              </div>
            </div>
          </div>
        </div>

        {/* Add IOC Section */}
  <div className="mb-8 bg-card border border-border rounded-xl p-6 shadow-lg">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-2 rounded-lg bg-accent border border-border shadow-lg">
              <Plus className="w-5 h-5 text-primary" />
            </div>
            <h2 className="text-2xl font-semibold text-foreground/80">Add New Indicator</h2>
            <div className="ml-auto flex items-center gap-2 px-3 py-1 rounded-full bg-accent/20 border border-border animate-pulse">
              <Zap className="w-3 h-3 text-green-400" />
              <span className="text-green-400 text-xs font-medium">ACTIVE</span>
            </div>
          </div>
          
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 items-end">
            <div className="lg:col-span-3">
              <label className="block text-sm font-medium text-muted-foreground mb-3">IOC Type</label>
              <select
                value={type}
                onChange={(e) => setType(e.target.value)}
                className="w-full bg-input border border-border rounded-lg px-4 py-3 text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all duration-200 hover:bg-input/80"
              >
                {iocTypes.map((t) => (
                  <option key={t} value={t} className="bg-card text-foreground">
                    {t}
                  </option>
                ))}
              </select>
            </div>
            
            <div className="lg:col-span-6">
              <label className="block text-sm font-medium text-muted-foreground mb-3">IOC Value</label>
              <div className="relative">
                <input
                  type="text"
                  placeholder="Enter IOC value..."
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  className="w-full bg-input border border-border rounded-lg px-4 py-3 pr-12 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all duration-200 hover:bg-input/80"
                />
                <div className="absolute right-4 top-1/2 transform -translate-y-1/2">
                  {getIOCIcon(type)}
                </div>
              </div>
            </div>
            
            <div className="lg:col-span-3">
              <Button
                onClick={handleAddIOC} 
                disabled={loading}
                className="w-full bg-primary text-primary-foreground font-medium py-3 px-6 rounded-lg transition-all duration-200 hover:bg-primary/80 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? (
                  <div className="flex items-center justify-center gap-2">
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                    Adding...
                  </div>
                ) : (
                  <div className="flex items-center justify-center gap-2">
                    <Plus className="w-4 h-4" />
                    Add IOC
                  </div>
                )}
              </Button>
            </div>
          </div>
        </div>

        {/* IOCs List Section */}
  <div className="bg-card border border-border rounded-xl p-6 shadow-lg">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-2 rounded-lg bg-accent border border-border shadow-lg">
              <Database className="w-5 h-5 text-primary" />
            </div>
            <h2 className="text-2xl font-semibold text-foreground/80">Threat Intelligence Database</h2>
            <div className="ml-auto flex items-center gap-2 px-3 py-1 rounded-full bg-accent/20 border border-border animate-pulse">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-muted-foreground/80 text-sm font-medium">Live</span>
            </div>
          </div>

          {loading && iocs.length === 0 ? (
            <div className="flex items-center justify-center py-16">
              <div className="flex flex-col items-center gap-4">
                <div className="w-8 h-8 border-2 border-primary/30 border-t-primary rounded-full animate-spin"></div>
                <p className="text-muted-foreground text-lg">Loading IOCs...</p>
              </div>
            </div>
          ) : !Array.isArray(iocs) || iocs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16">
              <div className="p-4 rounded-full bg-accent border border-border mb-4">
                <Shield className="w-16 h-16 text-primary" />
              </div>
              <p className="text-muted-foreground text-xl font-medium mb-2">No IOCs found for this case</p>
              <p className="text-muted-foreground text-sm">Add your first IOC above to begin threat analysis</p>
            </div>
          ) : (
            <div className="space-y-4">
              {iocs.map(({ id, type, value, created_at }) => {
                const threat = getThreatLevel(type);
                // Map threat colorClass to theme variables for border/text
                let borderClass = "border-border";
                let textClass = "text-foreground";
                let bgClass = "bg-card";
                if (threat.level === "HIGH" || threat.level === "CRITICAL") {
                  borderClass = "border-destructive";
                  textClass = "text-destructive";
                  bgClass = "bg-destructive/10";
                } else if (threat.level === "MEDIUM") {
                  borderClass = "border-secondary";
                  textClass = "text-secondary";
                  bgClass = "bg-secondary/10";
                }
                return (
                  <div
                    key={id}
                    className={`group ${bgClass} border ${borderClass} rounded-xl p-5 transition-all duration-300 hover:border-primary hover:shadow-lg`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-5 flex-1">
                        <div className="flex items-center gap-4">
                          <div className="p-3 rounded-xl bg-card border border-border">
                            {getIOCIcon(type)}
                          </div>
                          <div>
                            <div className="flex items-center gap-3 mb-2">
                              <span className="text-sm font-semibold text-muted-foreground bg-background px-3 py-1 rounded-lg border border-border">
                                {type}
                              </span>
                              <div className={`px-3 py-1 rounded-lg text-xs font-bold border ${borderClass} ${textClass}`}>
                                {threat.level}
                              </div>
                            </div>
                            <p className="text-foreground font-mono text-base break-all group-hover:text-primary transition-colors">
                              {value}
                            </p>
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-6">
                        <div className="text-right">
                          <p className="text-xs text-muted-foreground font-medium mb-1">ADDED</p>
                          <time
                            className="text-sm text-muted-foreground font-medium"
                            dateTime={created_at}
                            title={new Date(created_at).toLocaleString()}
                          >
                            {new Date(created_at).toLocaleDateString()}
                          </time>
                        </div>
                        <button className="opacity-0 group-hover:opacity-100 p-2 rounded-lg bg-accent border border-accent-foreground hover:bg-primary/10 hover:border-primary transition-all duration-200 hover:scale-110">
                          <Eye className="w-4 h-4 text-muted-foreground hover:text-primary transition-colors" />
                        </button>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
        
        {/* Footer */}
        <div className="mt-8 text-center">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-card border border-border backdrop-blur-sm">
            <Shield className="w-4 h-4 text-primary" />
            <span className="text-muted-foreground text-sm">AEGIS Platform</span>
          </div>
        </div>
      </div>
    </div>
  );
};
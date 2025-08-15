import { useEffect, useState } from "react";
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
  const { case_id } = useParams<{ case_id: string }>(); // match route param naming
  const navigate = useNavigate();

  const [caseName, setCaseName] = useState("");
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
        const caseRes = await axios.get<Case>(`http://localhost:8080/api/v1/cases/${case_id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        setCaseName(caseRes.data.name);

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
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      {/* Animated background effects */}
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#1e293b_1px,transparent_1px),linear-gradient(to_bottom,#1e293b_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)] opacity-20"></div>
      
      <div className="relative z-10 p-8 max-w-7xl mx-auto">
        {/* Header Section */}
        <div className="flex items-center gap-4 mb-8">
          <button
            onClick={() => navigate(-1)}
            className="group flex items-center gap-2 px-4 py-2 rounded-lg bg-slate-800/50 border border-slate-700 hover:bg-slate-700/50 hover:border-slate-600 transition-all duration-200 backdrop-blur-sm"
            aria-label="Go back"
          >
            <ArrowLeft className="w-4 h-4 text-cyan-400 group-hover:text-cyan-300 transition-colors" />
            <span className="text-slate-300 group-hover:text-white font-medium">Back</span>
          </button>
          
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-gradient-to-br from-cyan-500/20 to-blue-600/20 border border-cyan-500/30 shadow-lg shadow-cyan-500/10">
              <Shield className="w-8 h-8 text-cyan-400" />
            </div>
            <div>
              <h1 className="text-4xl font-bold bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
                Indicators of Compromise
              </h1>
              <p className="text-slate-400 flex items-center gap-2 mt-1">
                <Database className="w-4 h-4" />
                CaseID: <span className="text-cyan-400 font-medium">{case_id}</span>
              </p>
            </div>
          </div>
        </div>

        {/* Statistics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="group bg-gradient-to-br from-slate-800/80 to-slate-700/80 backdrop-blur-sm border border-slate-600/50 rounded-xl p-6 hover:border-slate-500/70 transition-all duration-300 hover:shadow-lg hover:shadow-slate-900/20">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm font-medium">Total IOCs</p>
                <p className="text-3xl font-bold text-white mt-2">{iocs.length}</p>
              </div>
              <div className="p-3 rounded-lg bg-slate-700/50 border border-slate-600/50 group-hover:border-slate-500/70 transition-all duration-300">
                <Database className="w-8 h-8 text-cyan-400" />
              </div>
            </div>
          </div>
          
          <div className="group bg-gradient-to-br from-red-900/30 to-red-800/30 backdrop-blur-sm border border-red-600/50 rounded-xl p-6 hover:border-red-500/70 transition-all duration-300 hover:shadow-lg hover:shadow-red-900/20">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm font-medium">High Threat</p>
                <p className="text-3xl font-bold text-red-400 mt-2">
                  {iocs.filter(ioc => getThreatLevel(ioc.type).level === 'HIGH' || getThreatLevel(ioc.type).level === 'CRITICAL').length}
                </p>
              </div>
              <div className="p-3 rounded-lg bg-red-800/30 border border-red-600/50 group-hover:border-red-500/70 transition-all duration-300">
                <AlertTriangle className="w-8 h-8 text-red-400" />
              </div>
            </div>
          </div>
          
          <div className="group bg-gradient-to-br from-yellow-900/30 to-yellow-800/30 backdrop-blur-sm border border-yellow-600/50 rounded-xl p-6 hover:border-yellow-500/70 transition-all duration-300 hover:shadow-lg hover:shadow-yellow-900/20">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm font-medium">Medium Threat</p>
                <p className="text-3xl font-bold text-yellow-400 mt-2">
                  {iocs.filter(ioc => getThreatLevel(ioc.type).level === 'MEDIUM').length}
                </p>
              </div>
              <div className="p-3 rounded-lg bg-yellow-800/30 border border-yellow-600/50 group-hover:border-yellow-500/70 transition-all duration-300">
                <Eye className="w-8 h-8 text-yellow-400" />
              </div>
            </div>
          </div>
          
          <div className="group bg-gradient-to-br from-slate-800/80 to-slate-700/80 backdrop-blur-sm border border-slate-600/50 rounded-xl p-6 hover:border-slate-500/70 transition-all duration-300 hover:shadow-lg hover:shadow-slate-900/20">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm font-medium">Last Added</p>
                <p className="text-lg font-bold text-slate-300 mt-2">
                  {iocs.length > 0 ? new Date(Math.max(...iocs.map(ioc => new Date(ioc.created_at).getTime()))).toLocaleDateString() : 'None'}
                </p>
              </div>
              <div className="p-3 rounded-lg bg-slate-700/50 border border-slate-600/50 group-hover:border-slate-500/70 transition-all duration-300">
                <Clock className="w-8 h-8 text-slate-400" />
              </div>
            </div>
          </div>
        </div>

        {/* Add IOC Section */}
        <div className="mb-8 bg-gradient-to-br from-slate-800/80 to-slate-700/80 backdrop-blur-sm border border-slate-600/50 rounded-xl p-6 shadow-lg">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-2 rounded-lg bg-gradient-to-r from-green-500/20 to-emerald-500/20 border border-green-500/30 shadow-lg shadow-green-500/10">
              <Plus className="w-5 h-5 text-green-400" />
            </div>
            <h2 className="text-2xl font-semibold text-white">Add New Indicator</h2>
            <div className="ml-auto flex items-center gap-2 px-3 py-1 rounded-full bg-green-500/10 border border-green-500/20 animate-pulse">
              <Zap className="w-3 h-3 text-green-400" />
              <span className="text-green-400 text-xs font-medium">ACTIVE</span>
            </div>
          </div>
          
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 items-end">
            <div className="lg:col-span-3">
              <label className="block text-sm font-medium text-slate-300 mb-3">IOC Type</label>
              <select
                value={type}
                onChange={(e) => setType(e.target.value)}
                className="w-full bg-slate-700/50 border border-slate-600 rounded-lg px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:border-transparent transition-all duration-200 hover:bg-slate-700/70"
              >
                {iocTypes.map((t) => (
                  <option key={t} value={t} className="bg-slate-800">
                    {t}
                  </option>
                ))}
              </select>
            </div>
            
            <div className="lg:col-span-6">
              <label className="block text-sm font-medium text-slate-300 mb-3">IOC Value</label>
              <div className="relative">
                <input
                  type="text"
                  placeholder="Enter IOC value..."
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  className="w-full bg-slate-700/50 border border-slate-600 rounded-lg px-4 py-3 pr-12 text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:border-transparent transition-all duration-200 hover:bg-slate-700/70"
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
                className="w-full bg-gradient-to-r from-cyan-500 to-blue-500 hover:from-cyan-600 hover:to-blue-600 disabled:from-slate-600 disabled:to-slate-700 border-0 text-white font-medium py-3 px-6 rounded-lg transition-all duration-200 transform hover:scale-105 hover:shadow-lg hover:shadow-cyan-500/25 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
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
        <div className="bg-gradient-to-br from-slate-800/80 to-slate-700/80 backdrop-blur-sm border border-slate-600/50 rounded-xl p-6 shadow-lg">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-2 rounded-lg bg-gradient-to-r from-purple-500/20 to-pink-500/20 border border-purple-500/30 shadow-lg shadow-purple-500/10">
              <Database className="w-5 h-5 text-purple-400" />
            </div>
            <h2 className="text-2xl font-semibold text-white">Threat Intelligence Database</h2>
            <div className="ml-auto flex items-center gap-2 px-3 py-1 rounded-full bg-slate-700/50 border border-slate-600/50">
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              <span className="text-slate-300 text-sm font-medium">Live</span>
            </div>
          </div>

          {loading && iocs.length === 0 ? (
            <div className="flex items-center justify-center py-16">
              <div className="flex flex-col items-center gap-4">
                <div className="w-8 h-8 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin"></div>
                <p className="text-slate-400 text-lg">Loading IOCs...</p>
              </div>
            </div>
          ) : !Array.isArray(iocs) || iocs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16">
              <div className="p-4 rounded-full bg-slate-700/50 border border-slate-600/50 mb-4">
                <Shield className="w-16 h-16 text-slate-500" />
              </div>
              <p className="text-slate-400 text-xl font-medium mb-2">No IOCs found for this case</p>
              <p className="text-slate-500 text-sm">Add your first IOC above to begin threat analysis</p>
            </div>
          ) : (
            <div className="space-y-4">
              {iocs.map(({ id, type, value, created_at }) => {
                const threat = getThreatLevel(type);
                return (
                  <div
                    key={id}
                    className="group bg-gradient-to-r from-slate-800/40 to-slate-700/40 border border-slate-700/50 hover:border-slate-600/70 rounded-xl p-5 transition-all duration-300 hover:bg-gradient-to-r hover:from-slate-800/60 hover:to-slate-700/60 hover:shadow-lg hover:shadow-slate-900/20"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-5 flex-1">
                        <div className="flex items-center gap-4">
                          <div className="p-3 rounded-xl bg-slate-700/50 border border-slate-600/50 group-hover:border-slate-500/70 transition-all duration-300">
                            {getIOCIcon(type)}
                          </div>
                          <div>
                            <div className="flex items-center gap-3 mb-2">
                              <span className="text-sm font-semibold text-slate-300 bg-slate-700/50 px-3 py-1 rounded-lg border border-slate-600/50">
                                {type}
                              </span>
                              <div className={`px-3 py-1 rounded-lg text-xs font-bold border ${threat.colorClass}`}>
                                {threat.level}
                              </div>
                            </div>
                            <p className="text-white font-mono text-base break-all group-hover:text-cyan-100 transition-colors">
                              {value}
                            </p>
                          </div>
                        </div>
                      </div>
                      
                      <div className="flex items-center gap-6">
                        <div className="text-right">
                          <p className="text-xs text-slate-500 font-medium mb-1">ADDED</p>
                          <time
                            className="text-sm text-slate-300 font-medium"
                            dateTime={created_at}
                            title={new Date(created_at).toLocaleString()}
                          >
                            {new Date(created_at).toLocaleDateString()}
                          </time>
                        </div>
                        <button className="opacity-0 group-hover:opacity-100 p-2 rounded-lg bg-slate-700/50 hover:bg-slate-600/50 border border-slate-600/50 hover:border-slate-500/70 transition-all duration-200 hover:scale-110">
                          <Eye className="w-4 h-4 text-slate-400 hover:text-cyan-400 transition-colors" />
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
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-slate-800/50 border border-slate-700/50 backdrop-blur-sm">
            <Shield className="w-4 h-4 text-cyan-400" />
            <span className="text-slate-400 text-sm">AEGIS Platform</span>
          </div>
        </div>
      </div>
    </div>
  );
};
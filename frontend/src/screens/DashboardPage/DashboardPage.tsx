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
  File,
} from "lucide-react";
import { Link } from "react-router-dom";
import { useState } from "react";
import { Progress } from "../../components/ui/progress";
import { cn } from "../../lib/utils";
import { useEffect } from "react";

const metricCards = [
  {
    value: "45",
    label: "Cases ongoing",
    increase: "8%",
    color: "text-[#636ae8]",
    icon: <Briefcase className="w-[75px] h-[52px] text-[#636ae8] flex-shrink-0" />,
  },
  {
    value: "120",
    label: "Cases Closed",
    increase: "15%",
    color: "text-green-500",
    icon: <CheckCircle className="w-[75px] h-[52px] text-green-500 flex-shrink-0" />,
  },
  {
    value: "875",
    label: "Evidence Collected",
    increase: "12%",
    color: "text-sky-500",
    icon: <Database className="w-[75px] h-[52px] text-sky-500 flex-shrink-0" />,
  },
];

const recentActivities = [
  {
    icon: File,
    text: "Team Alpha assigned to Case #AEG-1234",
    time: "yesterday",
  },
  {
    icon: AlertTriangle,
    text: "High severity alert triggered in Case #AEG-9012",
    time: "5 hours ago",
  },
  {
    icon: Briefcase,
    text: "Case #AEG-5678 status updated to 'Analysis' by Team Delta",
    time: "3 hours ago",
  },
];

// Default fallback cases (only used if localStorage is empty)
const defaultCaseCards = [
  {
    id: 1,
    creator: "System",
    team: "Team Gamma",
    priority: "critical",
    attackType: "Malware infection analysis",
    description: "System malware detected",
    lastActivity: "Yesterday",
    progress: 45,
    image: "https://th.bing.com/th/id/OIP.kq_Qib5c_49zZENmpMnuLQHaDt?w=331&h=180&c=7&r=0&o=5&dpr=1.3&pid=1.7",
  },
  {
    id: 2,
    creator: "System",
    team: "Team Alpha",
    priority: "high",
    attackType: "Data breach investigation",
    description: "Unauthorized data access detected",
    lastActivity: "2 hours ago",
    progress: 72,
    image: "https://th.bing.com/th/id/OIP.kq_Qib5c_49zZENmpMnuLQHaDt?w=331&h=180&c=7&r=0&o=5&dpr=1.3&pid=1.7",
  },
  {
    id: 3,
    creator: "System",
    team: "Team Beta",
    priority: "mid",
    attackType: "Phishing campaign analysis",
    description: "Suspicious email campaign identified",
    lastActivity: "Today",
    progress: 88,
    image: "https://th.bing.com/th/id/OIP.kq_Qib5c_49zZENmpMnuLQHaDt?w=331&h=180&c=7&r=0&o=5&dpr=1.3&pid=1.7",
  },
];

interface CaseCard {
  id: number;
  creator: string;
  team: string;
  priority: string;
  attackType: string;
  description: string;
  lastActivity: string;
  progress: number;
  image: string;
}

export default function Dashboard() {
  const [caseCards, setCaseCards] = useState<CaseCard[]>([]);

  useEffect(() => {
    // Load cases from localStorage
    const stored = localStorage.getItem("cases");
    console.log("Loaded from localStorage:", stored);

    if (stored) {
      try {
        const parsedCases: CaseCard[] = JSON.parse(stored);
        console.log("Parsed Cases:", parsedCases);
        setCaseCards(parsedCases.reverse()); // Show newest first
      } catch (error) {
        console.error("Error parsing stored cases:", error);
        setCaseCards(defaultCaseCards);
      }
    } else {
      // Use default cases if nothing in localStorage
      console.log("No cases in localStorage, using defaults");
      setCaseCards(defaultCaseCards);
    }
  }, []);

  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold text-white mb-6">Dashboard</h1>

      {caseCards.length === 0 ? (
        <div className="text-center text-gray-400 py-8">
          <p>No cases found. Create your first case to get started!</p>
        </div>
      ) : (
        <div className="flex flex-wrap gap-6">
          {caseCards.map((card) => (
            <div
              key={card.id}
              className="flex flex-col justify-between items-center w-[440px] h-[370px] p-4 bg-[#19191F] border border-[#393D47] rounded-[8px]"
            >
              <img
                src={card.image}
                alt={card.description}
                width={331}
                height={180}
                className="rounded-md mb-3"
              />

              <h3 className="text-white text-lg font-bold text-center mb-1">
                {card.attackType || "Untitled Case"}
              </h3>

              <div className="text-sm text-gray-400 text-center mb-2">
                Team: {card.team} | Last Activity: {card.lastActivity}
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
                  <span className="text-gray-300 capitalize">{card.priority}</span>
                </div>
                <div className="flex items-center gap-1">
                  <span className="w-2 h-2 rounded-full bg-blue-400"></span>
                  <span className="text-gray-300">Ongoing</span>
                </div>
              </div>

              <Progress
                value={card.progress}
                className="w-full h-3 bg-gray-800 mb-3 [&>div]:bg-green-500"
              />

              <button className="bg-[#633ae8] text-white text-sm px-14 py-2 rounded hover:bg-gray-800">
                View Details
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export const DashBoardPage = () => {
  const [activeTab, setActiveTab] = useState("active");
  const [caseCards, setCaseCards] = useState<CaseCard[]>([]);
  useEffect(() => {
    // Load cases from localStorage for the main dashboard page too
    const stored = localStorage.getItem("cases");
    console.log("DashboardPage - Loaded from localStorage:", stored);

    if (stored) {
      try {
        const parsedCases: CaseCard[] = JSON.parse(stored);
        console.log("DashboardPage - Parsed Cases:", parsedCases);
        setCaseCards(parsedCases.reverse()); // Show newest first
      } catch (error) {
        console.error("Error parsing stored cases:", error);
        setCaseCards(defaultCaseCards);
      }
    } else {
      setCaseCards(defaultCaseCards);
    }
  }, []);

  useEffect(() => {
    // Load cases from localStorage for the main dashboard page too
    const stored = localStorage.getItem("cases");
    console.log("DashboardPage - Loaded from localStorage:", stored);

    if (stored) {
      try {
        const parsedCases: CaseCard[] = JSON.parse(stored);
        console.log("DashboardPage - Parsed Cases:", parsedCases);
        setCaseCards(parsedCases.reverse()); // Show newest first
      } catch (error) {
        console.error("Error parsing stored cases:", error);
        setCaseCards(defaultCaseCards);
      }
    } else {
      setCaseCards(defaultCaseCards);
    }
  }, []);

  
  return (
    <div className="min-h-screen bg-black text-white">
      {/* Sidebar */}
      <div className="fixed left-0 top-0 h-full w-80 bg-black border-r border-gray-800 p-6 flex flex-col z-10">
        {/* Logo */}
        <div className=" flex items-center gap-3 mb-8">
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

          <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-3 rounded-lg transition-colors cursor-pointer">
            <FileText className="w-6 h-6" />
            <Link to="/case-management"><span className="text-lg">Case Management</span></Link>
          </div>

          <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-3 rounded-lg transition-colors cursor-pointer">
            <Folder className="w-6 h-6" />
            <Link to="/evidence-viewer"><span className="text-lg">Evidence Viewer</span></Link>
          </div>

          <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-3 rounded-lg transition-colors cursor-pointer">
            <MessageSquare className="w-6 h-6" />
            <span className="text-lg">
              <Link to="/secure-chat">Secure Chat</Link>
            </span>
          </div>
        </nav>

        {/* User Profile */}
        <div className="border-t border-gray-700 pt-4">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-gray-600 rounded-full flex items-center justify-center">
              <Link to="/profile" ><span className="text-white font-medium">AU</span></Link>
            </div>
            <div>
              <p className="font-semibold text-white">Agent User</p>
              <p className="text-gray-400 text-sm">user@dfir.com</p>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="ml-80 min-h-screen bg-black">
        {/* Topbar */}
        <div className="sticky top-0 bg-black border-b border-gray-800 p-4 z-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-6">
              <button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">
                Dashboard
              </button>
              <Link to="/evidence-viewer"><button className="text-gray-400 hover:text-white px-4 py-2 rounded-lg transition-colors">
                Evidence Viewer
              </button></Link>
              <Link to="/case-management"><button className="text-gray-400 hover:text-white px-4 py-2 rounded-lg transition-colors">
                Case Management
              </button></Link>
              <button className="text-gray-400 hover:text-white px-4 py-2 rounded-lg transition-colors">
                <Link to="/secure-chat">Secure Chat</Link>
              </button>
            </div>

            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  className="w-80 h-12 bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 text-white placeholder-gray-400 focus:outline-none focus:border-blue-500"
                  placeholder="Search cases, evidence, users"
                />
              </div>
              <button className="p-2 text-gray-400 hover:text-white transition-colors">
                <Bell className="w-6 h-6" />
              </button>
              <Link to="/settings" ><button className="p-2 text-gray-400 hover:text-white transition-colors">
                <Settings className="w-6 h-6" />
              </button></Link>
              <div className="w-10 h-10 bg-gray-600 rounded-full flex items-center justify-center">
                <Link to="/profile" ><span className="text-white font-medium text-sm">AU</span></Link>
              </div>
            </div>
          </div>
        </div>

        {/* Page Content */}
        <main className="p-8">
          <h1 className="text-3xl font-semibold mb-6">Dashboard Overview</h1>

          {/* Metric Cards */}
          <div className="flex gap-6 flex-wrap">
            {metricCards.map((card, index) => (
              <div
                key={index}
                className="w-[266px] h-[126px] flex-shrink-0 bg-[#19191F] border-[5px] border-[#30333C] rounded-[8px] p-4 flex items-center justify-between"
              >
                <div>
                  <p className={`text-3xl font-bold ${card.color}`}>{card.value}</p>
                  <p className="text-white/70 text-sm">{card.label}</p>
                  <p className="text-white/40 text-xs mt-1">↑ {card.increase} from last week</p>
                </div>
                {card.icon}
              </div>
            ))}
          </div>
            {/* Extra spacing before next row */}
                <div className="mt-[100px] flex gap-6">
                {/* Threat Landscape Card */}
                <div className="w-[550px] h-[366px] flex-shrink-0 rounded-lg border-[3px] border-[#30333C] bg-[#19191F] p-6">
                    <h2 className="font-bold text-white text-lg mb-2">Threat Landscape</h2>
                    <p className="text-gray-400 text-sm mb-4">Global incident trends, and a picture:</p>
                    <img
                    src="https://c.animaapp.com/maycc5gah5c0ar/img/graph.png"
                    alt="Threat Landscape Graph"
                    className="w-[503px] h-[265px] flex-shrink-0"
                    />
                </div>

                 {/* Recent Activities Card */}
            <div className="w-[529px] h-[366px] flex-shrink-0 rounded-lg border-[3px] border-[#30333C] bg-[#19191F] p-6 overflow-auto">
              <h2 className="font-bold text-white text-lg mb-4">Recent Activities</h2>
              <ul className="space-y-4">
                {recentActivities.map((activity, index) => {
                  const Icon = activity.icon;
                  const isAlert = Icon === AlertTriangle;
                  return (
                    <li key={index}>
                      <div className="flex items-start gap-3 mb-2">
                        <Icon className={`w-5 h-5 mt-1 ${isAlert ? 'text-red-500' : 'text-white'}`} />
                        <div>
                          <p className="text-white text-sm">{activity.text}</p>
                          <p className="text-gray-400 text-xs">{activity.time}</p>
                        </div>
                      </div>
                      {index < recentActivities.length - 1 && (
                        <hr className="w-[500px] border-t-[2px] border-[#8C8D8B] transform rotate-[0.053deg]" />
                      )}
                    </li>
                  );
                })}
              </ul>
            </div>
          </div>
         
          <div className="w-[1105px] h-[1700px] bg-[#19191F] border-[3px] border-[#393D47] rounded-[8px] mt-[30px] p-6">
      <div className="flex justify-between items-center mb-4">
        <div className="flex gap-2">
          <button
            onClick={() => setActiveTab("active")}
            className={cn(
              "text-sm rounded-lg h-8 px-4",
              activeTab === "active"
                ? "bg-gray-700 text-white"
                : "bg-gray-900 text-gray-400 border border-gray-700"
            )}
          >
            Active Cases ({caseCards.length})
          </button>
          <button
            onClick={() => setActiveTab("archived")}
            className={cn(
              "text-sm rounded-lg h-8 px-4",
              activeTab === "archived"
                ? "bg-gray-700 text-white"
                : "bg-gray-900 text-gray-400 border border-gray-700"
            )}
          >
            Archived Cases (0)
          </button>
        </div>
        <Link to="/create-case"><button className="bg-blue-600 text-white text-sm px-4 py-2 rounded-md hover:bg-blue-700">
          Create Case
        </button></Link>
      </div>

      {caseCards.length === 0 ? (
        <div className="text-center text-gray-400 py-8">
          <p>No cases found. Create your first case to get started!</p>
        </div>
      ) : (
        <div className="flex flex-wrap gap-6">
          {caseCards.map((card) => (
            <div
              key={card.id}
              className="flex flex-col justify-between items-center w-[440px] h-[370px] p-4 bg-[#19191F] border border-[#393D47] rounded-[8px]"
            >
              <img
              src={card.image}
              alt={card.attackType}
              width={331}
              height={180}
              className="rounded-md mb-3"
              />

              <h3 className="text-white text-lg font-bold text-center mb-1">
                {card.attackType || "Untitled Case"}
              </h3>
              <div className="text-sm text-gray-400 text-center mb-2">
                Team: {card.team} | Last Activity: {card.lastActivity}
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
                  <span className="text-gray-300 capitalize">{card.priority}</span>
                </div>
                <div className="flex items-center gap-1">
                  <span className="w-2 h-2 rounded-full bg-blue-400"></span>
                  <span className="text-gray-300">Ongoing</span>
                </div>
              </div>
              <Progress
              value={card.progress}
              className="w-full h-3 bg-gray-800 mb-3 [&>div]:bg-green-500"
              />

              <Link to={`/evidence-viewer/${card.id}`}>
                <button className="bg-[#633ae8] text-white text-sm px-14 py-2 rounded hover:bg-gray-800">
                  View Evidence Details
                </button>
              </Link>
            </div>
          ))}
        </div>
      )}
    </div>
        </main>
      </div>
    </div>
  );
};
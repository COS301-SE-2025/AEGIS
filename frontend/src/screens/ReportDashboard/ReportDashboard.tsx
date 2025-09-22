import React, { useState,useEffect } from 'react';
import { 
  Search, 
  //ChevronDown, 
  //Plus,
  Grid,
  List,
  FileText,
  Users,
  Download,
  //Lock,
  //Eye,
  //Calendar,
  //Clock,
  AlertTriangle,
  Shield,
  Bug
} from 'lucide-react';
import axios from 'axios';
//import { useParams } from 'react-router-dom';
import { useNavigate } from 'react-router-dom';  // React Router v6
// Types
interface ReportWithDetails {
  id: string;
  reportID: string;
  case_id: string;
  case_name?: string;
  team_id?: string;
  team_name?: string;
  name: string;             // corresponds to report name
  type: string;
  status: 'draft' | 'review' | 'published';
  version: number;
  last_modified: string;
  file_path: string;
  author: string;           // examiner full name
  collaborators: number;    // count from case_user_roles
}


// interface ReportTemplate {
//   id: string;
//   title: string;
//   description: string;
//   icon: React.ReactNode;
//   color: string;
// }
// interface Report {
//   name: string;
//   content: { title: string; content: string }[];
// }


export const ReportDashboard = () => {
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchTerm, setSearchTerm] = useState('');
  // const [selectedAuthor, setSelectedAuthor] = useState('All Authors');
  // const [selectedType, setSelectedType] = useState('All Types');
  // const [selectedTimeframe, setSelectedTimeframe] = useState('Last 30 days');
  
const [reports, setReports] = useState<ReportWithDetails[]>([]);
  //const [selectedCaseId, setSelectedCaseId] = useState<string>('923f5f04-0641-4e10-b9f8-ef6fcfbecbc2');
  const [_, setError] = useState<string | null>(null);
//  const { reportId } = useParams<Record<string, string | undefined>>(); 


  // Change from Axios.AxiosResponse to axios.AxiosResponse (lowercase)

  // API URL - make sure to update with the correct URL
  const API_URL = 'http://localhost:8080/api/v1';
const getTeamIdFromSession = (): string => {
  try {
    const raw = sessionStorage.getItem("user");
    if (!raw) return "";
    const u = JSON.parse(raw);
    return u?.team_id || u?.teamId || "";
  } catch {
    return "";
  }
};

const [teamId, setTeamId] = useState<string>(getTeamIdFromSession());

useEffect(() => {
  setTeamId(getTeamIdFromSession());
}, []);

   // Fetch reports by case - move the axios call here
// Fetch reports by case
useEffect(() => {
  const fetchReportsForTeam = async () => {
    try {
      const token = sessionStorage.getItem("authToken");
      if (!token) {
        console.error("No auth token found");
        return;
      }
      if (!teamId) {
        console.warn("No teamId found in session; cannot load team reports.");
        setReports([]);
        return;
      }

      // New endpoint shape: GET /api/v1/reports/teams/:teamID  -> { reports: [...] }
      const res = await axios.get<{ reports: ReportWithDetails[] }>(
        `${API_URL}/reports/teams/${teamId}`,
        { headers: { Authorization: `Bearer ${token}` } }
      );

      setReports(Array.isArray(res.data?.reports) ? res.data.reports : []);
    } catch (err) {
      console.error("Error fetching team reports:", err);
      setReports([]);
    }
  };

  fetchReportsForTeam();
}, [teamId]);


const formatTimestamp = (timestamp: string) => {
  if (!timestamp) return "";

  // Parse the timestamp string (assume "2025-08-13 15:58:46")
  const date = new Date(timestamp.replace(" ", "T")); // make it ISO compatible

  // Format options
  const options: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  };

  return date.toLocaleString("en-US", options); // e.g., "Dec 12, 2025 at 9:15 AM"
};

const shortId = (id?: string) => (id ? `${id.slice(0, 8)}…${id.slice(-4)}` : '');

const Badge = ({ children }: { children: React.ReactNode }) => (
  <span className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs
                   bg-gray-800 border border-gray-600 text-gray-200">
    {children}
  </span>
);



async function downloadReport(id: string) {
  try {
    const token = sessionStorage.getItem('authToken');
    if (!token) throw new Error('No auth token found');

    // Tell Axios we expect a Blob
 const res = await axios.get(`${API_URL}/reports/${id}/download/pdf`, {
  headers: { Authorization: `Bearer ${token}` },
  responseType: "blob",
});
const blob = new Blob([res.data as BlobPart], { type: "application/pdf" });
    const url = window.URL.createObjectURL(blob);

    const link = document.createElement("a");
    link.href = url;
    link.download = `report-${id}.pdf`;
    link.click();

    window.URL.revokeObjectURL(url);

  } catch (err) {
    console.error('Failed to download report', err);
    setError("Failed to download report");
  }
}

const navigate = useNavigate();
const handleOpenReport = (reportId: string) => {
  navigate(`/report-editor/${reportId}`);
};


 
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'draft': return 'bg-gray-600';
      case 'review': return 'bg-yellow-600';
      case 'published': return 'bg-green-600';
      default: return 'bg-gray-600';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'incident': return <AlertTriangle className="w-4 h-4" />;
      case 'forensic': return <Shield className="w-4 h-4" />;
      case 'malware': return <Bug className="w-4 h-4" />;
      default: return <FileText className="w-4 h-4" />;
    }
  };

  // const ReportTemplateCard = ({ template }: { template: ReportTemplate }) => (
  //   <div className="bg-gray-800 rounded-lg p-6 border border-gray-700 hover:border-gray-600 transition-colors">
  //     <div className="flex items-center justify-between mb-4">
  //       <div className={`${template.color} p-3 rounded-lg`}>
  //         {template.icon}
  //       </div>
  //       <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
  //         Create
  //       </button>
  //     </div>
  //     <h3 className="text-white font-semibold mb-2">{template.title}</h3>
  //     <p className="text-gray-400 text-sm">{template.description}</p>
  //   </div>
  // );

  const ReportCard = ({ report }: { report: ReportWithDetails }) => (
    <div className="bg-card rounded-lg p-6 border border hover:border-primary transition-colors shadow-xl">
      <div className="flex items-start justify-between mb-4">

        <div>
          <h3 className="text-foreground font-semibold mb-1">{report.name}</h3>
          <p className="text-foreground text-sm">Last Modified: {formatTimestamp(report.last_modified)}</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1 text-foreground">
            <Users className="w-4 h-4" />
            <span className="text-sm">{report.collaborators}</span>
            <div className="hidden md:flex items-center gap-2 mx-4">
          <Badge>
            Team: {report.team_name ?? (report.team_id ? `#${shortId(report.team_id)}` : `#${shortId(teamId)}`)}
          </Badge>
          <Badge>
            Case: {report.case_name ?? (report.case_id ? shortId(report.case_id) : 'Unknown')}
          </Badge>
        </div>

          </div>
          <span className={`w-2 h-2 rounded-full ${getStatusColor(report.status)}`}></span>
        </div>
      </div>
      
      <div className="flex items-center justify-between">
  <div className="flex items-center gap-3">
    <button 
     onClick={() => handleOpenReport(report.id)}  
    className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/60 transition-colors text-sm font-medium">
      Open
    </button>
    <button className="p-2 text-foreground hover:text-foreground/60 transition-colors">
      <Users className="w-4 h-4" />
    </button>
    <button
      onClick={() => downloadReport(report.id)}
      className="p-2 text-foreground hover:text-foreground/60 transition-colors"
    >
      <Download className="w-4 h-4" />
    </button>
  </div>
  {getTypeIcon(report.type)}
</div>
    </div>
  );

  const ReportListItem = ({ report }: { report: ReportWithDetails }) => (
    <div className="bg-card rounded-lg p-4 border border hover:border-primary/60 transition-colors">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4 flex-1">
          <div className="flex items-center gap-2">
            {getTypeIcon(report.type)}
            <span className={`w-2 h-2 rounded-full ${getStatusColor(report.status)}`}></span>
          </div>
          
          <div className="flex-1 min-w-0">
            <h3 className="text-foreground font-medium truncate">{report.name}</h3>
            <p className="text-foreground text-sm">{report.author} • {formatTimestamp(report.last_modified)}</p>
          </div>
          
          <div className="flex items-center gap-1 text-foreground">
            <Users className="w-4 h-4" />
            <span className="text-sm">{report.collaborators}</span>
          </div>
        </div>
        
        <div className="flex items-center gap-2 ml-4">
          <button 
            onClick={() => handleOpenReport(report.id)}  
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/60 transition-colors text-sm font-medium">
              Open
            </button>
                    <button
              onClick={() => downloadReport(report.id)}
              className="p-2 text-foreground hover:text-foreground/60 transition-colors"
            >
              <Download className="w-4 h-4" />
            </button>
        </div>
      </div>
    </div>
  );

//   const filteredReports = reports.filter(report => {
//   const matchesSearch = (report.name ?? '').toLowerCase().includes(searchTerm.toLowerCase());
//   const matchesAuthor = selectedAuthor === 'All Authors' || (report.author ?? '') === selectedAuthor;
//   const matchesType = selectedType === 'All Types' || (report.type ?? '') === selectedType;
//   return matchesSearch && matchesAuthor && matchesType;
// });



  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="bg-background border border px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-2xl font-bold text-foreground">Report Dashboard</h1>
            <span className="text-foreground">Team #{shortId(teamId)}</span>


          </div>
          <div className="flex items-center gap-3">
            {/* <Link to="/case-management">
            <button className="px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 transition-colors">
              Case management
            </button>
            </Link> */}
      
          </div>
        </div>
      </div>

      <div className="p-6">
        {/* Filters */}
        <div className="flex items-center gap-4 mb-8">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search reports..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full bg-background border rounded-lg pl-10 pr-4 py-2 text-foreground placeholder-muted-foreground focus:outline-none focus:border-primary"
            />
          </div>

          <div className="flex items-center gap-2">
            {/* <select 
              value={selectedAuthor}
              onChange={(e) => setSelectedAuthor(e.target.value)}
              className="bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
            >
              <option>All Authors</option>
              <option>J. Doe</option>
              <option>M. Smith</option>
              <option>A. Johnson</option>
              <option>S. Williams</option>
            </select> */}

            {/* <select 
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value)}
              className="bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
            >
              <option>All Types</option>
              <option value="incident">Incident</option>
              <option value="forensic">Forensic</option>
              <option value="malware">Malware</option>
            </select>

            <select 
              value={selectedTimeframe}
              onChange={(e) => setSelectedTimeframe(e.target.value)}
              className="bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
            >
              <option>Last 30 days</option>
              <option>Last 7 days</option>
              <option>Last 90 days</option>
              <option>All time</option>
            </select> */}
          </div>
        </div>

        {/* Report Templates */}
    

        {/* Existing Reports */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold text-foreground">Existing Reports</h2>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setViewMode('grid')}
                className={`p-2 rounded-lg transition-colors ${
                  viewMode === 'grid' 
                    ? 'bg-blue-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <Grid className="w-4 h-4" />
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={`p-2 rounded-lg transition-colors ${
                  viewMode === 'list' 
                    ? 'bg-primary text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <List className="w-4 h-4" />
              </button>
            </div>
          </div>

         {viewMode === 'grid' ? (
  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
    {reports.map((report, index) => (
      <ReportCard
        key={report.id ?? `report-${index}`}
        report={report}
      />
    ))}
  </div>
) : (
  <div className="space-y-3">
    {reports.map((report, index) => (
      <ReportListItem
        key={report.id ?? `report-${index}`}
        report={report}
      />
    ))}
  </div>
)}

        </div>
      </div>
    </div>
  );
};


import React, { useState } from 'react';
import { 
  Search, 
  //ChevronDown, 
  Plus,
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
import { Link } from "react-router-dom";

// Types
interface Report {
  id: string;
  title: string;
  type: 'incident' | 'forensic' | 'malware';
  author: string;
  collaborators: number;
  lastModified: string;
  status: 'draft' | 'review' | 'published';
  caseId?: string;
}

interface ReportTemplate {
  id: string;
  title: string;
  description: string;
  icon: React.ReactNode;
  color: string;
}

export const ReportDashboard = () => {
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedAuthor, setSelectedAuthor] = useState('All Authors');
  const [selectedType, setSelectedType] = useState('All Types');
  const [selectedTimeframe, setSelectedTimeframe] = useState('Last 30 days');

  const reportTemplates: ReportTemplate[] = [
    {
      id: 'incident-standard',
      title: 'Incident Response - Standard',
      description: 'Comprehensive incident response report template with timeline, impact assessment, and remediation steps.',
      icon: <AlertTriangle className="w-6 h-6" />,
      color: 'bg-blue-600'
    },
    {
      id: 'forensic-analysis',
      title: 'Forensic Analysis Report',
      description: 'Detailed digital forensics examination template including evidence chain of custody and findings.',
      icon: <Shield className="w-6 h-6" />,
      color: 'bg-emerald-600'
    },
    {
      id: 'malware-analysis',
      title: 'Malware Analysis',
      description: 'Structured template for malware reverse engineering and behavioral analysis documentation.',
      icon: <Bug className="w-6 h-6" />,
      color: 'bg-red-600'
    }
  ];

  const existingReports: Report[] = [
    {
      id: 'report-1',
      title: 'Phishing Campaign Analysis - Q4 2024',
      type: 'incident',
      author: 'J. Doe',
      collaborators: 3,
      lastModified: 'Dec 15, 2024 at 4:30 PM',
      status: 'draft',
      caseId: '2024-001'
    },
    {
      id: 'report-2',
      title: 'Network Intrusion Investigation',
      type: 'forensic',
      author: 'M. Smith',
      collaborators: 1,
      lastModified: 'Dec 12, 2024 at 9:15 AM',
      status: 'review',
      caseId: '2024-002'
    },
    {
      id: 'report-3',
      title: 'Ransomware Incident Response',
      type: 'incident',
      author: 'A. Johnson',
      collaborators: 2,
      lastModified: 'Dec 10, 2024 at 6:45 PM',
      status: 'published',
      caseId: '2024-003'
    },
    {
      id: 'report-4',
      title: 'Data Breach Assessment',
      type: 'forensic',
      author: 'S. Williams',
      collaborators: 4,
      lastModified: 'Dec 8, 2024 at 11:20 AM',
      status: 'published',
      caseId: '2024-004'
    }
  ];

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

  const ReportTemplateCard = ({ template }: { template: ReportTemplate }) => (
    <div className="bg-gray-800 rounded-lg p-6 border border-gray-700 hover:border-gray-600 transition-colors">
      <div className="flex items-center justify-between mb-4">
        <div className={`${template.color} p-3 rounded-lg`}>
          {template.icon}
        </div>
        <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
          Create
        </button>
      </div>
      <h3 className="text-white font-semibold mb-2">{template.title}</h3>
      <p className="text-gray-400 text-sm">{template.description}</p>
    </div>
  );

  const ReportCard = ({ report }: { report: Report }) => (
    <div className="bg-gray-800 rounded-lg p-6 border border-gray-700 hover:border-gray-600 transition-colors">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="text-white font-semibold mb-1">{report.title}</h3>
          <p className="text-gray-400 text-sm">Last modified: {report.lastModified}</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1 text-gray-400">
            <Users className="w-4 h-4" />
            <span className="text-sm">{report.collaborators}</span>
          </div>
          <span className={`w-2 h-2 rounded-full ${getStatusColor(report.status)}`}></span>
        </div>
      </div>
      
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
            Open
          </button>
          <button className="p-2 text-gray-400 hover:text-white transition-colors">
            <Users className="w-4 h-4" />
          </button>
          <button className="p-2 text-gray-400 hover:text-white transition-colors">
            <Download className="w-4 h-4" />
          </button>
        </div>
        {getTypeIcon(report.type)}
      </div>
    </div>
  );

  const ReportListItem = ({ report }: { report: Report }) => (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700 hover:border-gray-600 transition-colors">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4 flex-1">
          <div className="flex items-center gap-2">
            {getTypeIcon(report.type)}
            <span className={`w-2 h-2 rounded-full ${getStatusColor(report.status)}`}></span>
          </div>
          
          <div className="flex-1 min-w-0">
            <h3 className="text-white font-medium truncate">{report.title}</h3>
            <p className="text-gray-400 text-sm">{report.author} • {report.lastModified}</p>
          </div>
          
          <div className="flex items-center gap-1 text-gray-400">
            <Users className="w-4 h-4" />
            <span className="text-sm">{report.collaborators}</span>
          </div>
        </div>
        
        <div className="flex items-center gap-2 ml-4">
          <button className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 transition-colors">
            Open
          </button>
          <button className="p-1.5 text-gray-400 hover:text-white transition-colors">
            <Download className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );

  const filteredReports = existingReports.filter(report => {
    const matchesSearch = report.title.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesAuthor = selectedAuthor === 'All Authors' || report.author === selectedAuthor;
    const matchesType = selectedType === 'All Types' || report.type === selectedType;
    return matchesSearch && matchesAuthor && matchesType;
  });

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Header */}
      <div className="bg-gray-800 border-b border-gray-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-2xl font-bold text-white">Report Dashboard</h1>
            <span className="text-gray-400">Case #2024-001</span>
          </div>
          <div className="flex items-center gap-3">
            {/* <Link to="/case-management">
            <button className="px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 transition-colors">
              Case management
            </button>
            </Link> */}
            <div className="flex items-center gap-3 text-muted-foreground hover:text-white hover:bg-muted p-3 rounded-lg transition-colors cursor-pointer">
                        <FileText className="w-6 h-6" />
                        <Link to="/case-management"><span className="text-lg">Case Management</span></Link>
                      </div>
            
            <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
              <Plus className="w-4 h-4" />
              New Report
            </button>
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
              className="w-full bg-gray-800 border border-gray-600 rounded-lg pl-10 pr-4 py-2 text-white placeholder-gray-400 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div className="flex items-center gap-2">
            <select 
              value={selectedAuthor}
              onChange={(e) => setSelectedAuthor(e.target.value)}
              className="bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
            >
              <option>All Authors</option>
              <option>J. Doe</option>
              <option>M. Smith</option>
              <option>A. Johnson</option>
              <option>S. Williams</option>
            </select>

            <select 
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
            </select>
          </div>
        </div>

        {/* Report Templates */}
        <div className="mb-8">
          <h2 className="text-xl font-semibold text-white mb-4">Report Templates</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {reportTemplates.map(template => (
              <ReportTemplateCard key={template.id} template={template} />
            ))}
          </div>
        </div>

        {/* Existing Reports */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold text-white">Existing Reports</h2>
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
                    ? 'bg-blue-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                <List className="w-4 h-4" />
              </button>
            </div>
          </div>

          {viewMode === 'grid' ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {filteredReports.map(report => (
                <ReportCard key={report.id} report={report} />
              ))}
            </div>
          ) : (
            <div className="space-y-3">
              {filteredReports.map(report => (
                <ReportListItem key={report.id} report={report} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};


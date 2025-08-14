import React, { useState } from 'react';
import ReactQuill from 'react-quill';
import 'react-quill/dist/quill.snow.css';
import {
  FileText,
  Download,
  Save,
  Plus,
  Clock,
  Users,
  Calendar,
  Shield,
  Eye,
} from 'lucide-react';

interface ReportSection {
  id: string;
  title: string;
  content: string;
  completed: boolean;
}

interface RecentReport {
  id: string;
  title: string;
  status: 'draft' | 'review' | 'published';
  lastModified: string;
}

export const ReportEditor = () => {
  const [activeSection, setActiveSection] = useState(0);
  const [reportTitle, setReportTitle] = useState('Digital Forensics and Incident Response Report');
  const [incidentId, setIncidentId] = useState('2024-001');
  const [dateCreated, setDateCreated] = useState('July 15, 2024');
  const [analyst, setAnalyst] = useState('John Doe');
  const [reportType, setReportType] = useState('Security Incident');

  const [sections, setSections] = useState<ReportSection[]>([
    {
      id: 'executive-summary',
      title: 'Executive Summary',
      content: `<p>On January 14, 2024, the Security Operations Center (SOC) detected suspicious network activity indicating a potential security breach. This report documents the comprehensive digital forensics investigation conducted to determine the scope, attack vector, and root cause of the incident.</p><p>Initial analysis revealed unauthorized access to the corporate network through a compromised employee workstation. The investigation timeline, findings, and recommended remediation actions are detailed in this report.</p>`,
      completed: true
    },
    {
      id: 'incident-scope',
      title: 'Incident Scope & Objectives',
      content: `<p><strong>Investigation Objectives:</strong></p><ul><li>Identify the attack vector and timeline</li><li>Determine the extent of system compromise</li><li>Assess data exfiltration risks</li><li>Document evidence for potential legal proceedings</li></ul>`,
      completed: true
    },
    {
      id: 'evidence-findings',
      title: 'Evidence & Findings',
      content: '<p>Content for Evidence &amp; Findings section...</p>',
      completed: false
    },
    {
      id: 'compromised-assets',
      title: 'Compromised Assets',
      content: '',
      completed: false
    },
    {
      id: 'malware-identified',
      title: 'Malware Identified',
      content: '',
      completed: false
    }
  ]);

  const recentReports: RecentReport[] = [
    {
      id: 'security-incident-2024-001',
      title: 'Security Incident 2024-001',
      status: 'draft',
      lastModified: '2 hours ago'
    },
    {
      id: 'malware-analysis-report',
      title: 'Malware Analysis Report',
      status: 'review',
      lastModified: '1 day ago'
    },
    {
      id: 'network-forensics',
      title: 'Network Forensics',
      status: 'review',
      lastModified: '1 day ago'
    },
    {
      id: 'endpoint-investigation',
      title: 'Endpoint Investigation',
      status: 'published',
      lastModified: '3 days ago'
    }
  ];

  // Custom Quill modules with dark theme styling
  const modules = {
    toolbar: [
      [{ 'header': [1, 2, 3, false] }],
      ['bold', 'italic', 'underline', 'strike'],
      [{ 'color': [] }, { 'background': [] }],
      [{ 'list': 'ordered'}, { 'list': 'bullet' }],
      [{ 'indent': '-1'}, { 'indent': '+1' }],
      ['link', 'image', 'code-block'],
      [{ 'align': [] }],
      ['clean']
    ],
  };

  const formats = [
    'header', 'bold', 'italic', 'underline', 'strike', 
    'color', 'background', 'list', 'bullet', 'indent',
    'link', 'image', 'code-block', 'align'
  ];

  const updateSectionContent = (content: string) => {
    const updatedSections = sections.map((section, index) =>
      index === activeSection ? { ...section, content } : section
    );
    setSections(updatedSections);
  };

  const toggleSectionCompletion = (index: number) => {
    const updatedSections = sections.map((section, i) =>
      i === index ? { ...section, completed: !section.completed } : section
    );
    setSections(updatedSections);
  };

  const getStatusDot = (status: string) => {
    switch (status) {
      case 'draft': return 'bg-gray-400';
      case 'review': return 'bg-yellow-400';
      case 'published': return 'bg-green-400';
      default: return 'bg-gray-400';
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 flex">
      {/* Custom styles for Quill in dark mode */}
      <style jsx global>{`
        .ql-snow {
          border: 1px solid #374151 !important;
          background-color: #1f2937 !important;
        }
        
        .ql-snow .ql-toolbar {
          border-bottom: 1px solid #374151 !important;
          background-color: #1f2937 !important;
        }
        
        .ql-snow .ql-container {
          border-top: none !important;
          background-color: #1f2937 !important;
        }
        
        .ql-editor {
          color: #e5e7eb !important;
          background-color: #1f2937 !important;
          min-height: 300px !important;
          font-size: 16px !important;
          line-height: 1.6 !important;
        }
        
        .ql-editor.ql-blank::before {
          color: #6b7280 !important;
          font-style: italic;
        }
        
        .ql-snow .ql-tooltip {
          background-color: #374151 !important;
          border: 1px solid #4b5563 !important;
          color: #e5e7eb !important;
        }
        
        .ql-snow .ql-tooltip input {
          background-color: #1f2937 !important;
          color: #e5e7eb !important;
          border: 1px solid #4b5563 !important;
        }
        
        .ql-snow .ql-picker-options {
          background-color: #374151 !important;
          border: 1px solid #4b5563 !important;
        }
        
        .ql-snow .ql-picker-item:hover {
          background-color: #4b5563 !important;
          color: #e5e7eb !important;
        }
        
        .ql-snow .ql-stroke {
          stroke: #9ca3af !important;
        }
        
        .ql-snow .ql-fill {
          fill: #9ca3af !important;
        }
        
        .ql-snow .ql-picker-label:hover .ql-stroke,
        .ql-snow .ql-picker-label.ql-active .ql-stroke {
          stroke: #e5e7eb !important;
        }
        
        .ql-snow .ql-picker-label:hover .ql-fill,
        .ql-snow .ql-picker-label.ql-active .ql-fill {
          fill: #e5e7eb !important;
        }
      `}</style>

      {/* Left Sidebar */}
      <div className="w-80 bg-gray-800 border-r border-gray-700 flex flex-col">
        {/* Logo & Header */}
        <div className="p-4 border-b border-gray-700">
          <div className="flex items-center gap-2 mb-4">
            <div className="w-8 h-8 rounded flex items-center justify-center">
              <img
                src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
                alt="AEGIS Logo"
                className="w-full h-full object-cover"
              />
            </div>
            <span className="text-white font-bold text-xl">AEGIS</span>
          </div>
        </div>

        {/* Recent Reports */}
        <div className="p-4 border-b border-gray-700">
          <h3 className="text-gray-300 font-medium mb-3">Recent Reports</h3>
          <div className="space-y-2">
            {recentReports.map((report) => (
              <div 
                key={report.id}
                className="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-700 cursor-pointer transition-colors"
              >
                <FileText className="w-4 h-4 text-gray-400 flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className={`w-2 h-2 rounded-full ${getStatusDot(report.status)}`}></span>
                    <span className="text-white text-sm truncate">{report.title}</span>
                  </div>
                  <p className="text-gray-400 text-xs">{report.lastModified}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* New Report Button */}
        <div className="p-4">
          <button className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
            <Plus className="w-4 h-4" />
            New Report
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex">
        {/* Report Sections Navigation */}
        <div className="w-80 bg-gray-850 border-r border-gray-700 p-4">
          <div className="mb-6">
            <input
              type="text"
              value={reportTitle}
              onChange={(e) => setReportTitle(e.target.value)}
              className="w-full bg-transparent text-white font-semibold text-lg border-none outline-none"
            />
            <div className="flex items-center gap-2 mt-2 text-sm text-gray-400">
              <Calendar className="w-4 h-4" />
              <span>Incident 2024-001</span>
            </div>
            <div className="flex items-center gap-2 mt-1 text-sm text-gray-400">
              <Clock className="w-4 h-4" />
              <span>Date Created: {dateCreated}</span>
            </div>
            <div className="flex items-center gap-2 mt-1 text-sm text-gray-400">
              <Users className="w-4 h-4" />
              <span>Analyst: {analyst}</span>
            </div>
          </div>

          <div className="space-y-1">
            {sections.map((section, index) => (
              <button
                key={section.id}
                onClick={() => setActiveSection(index)}
                className={`w-full flex items-center justify-between p-3 rounded-lg text-left transition-colors ${
                  activeSection === index 
                    ? 'bg-blue-600 text-white' 
                    : 'hover:bg-gray-700 text-gray-300'
                }`}
              >
                <span className="font-medium">{section.title}</span>
                <div 
                  className={`w-3 h-3 rounded-full border-2 ${
                    section.completed 
                      ? 'bg-green-500 border-green-500' 
                      : 'border-gray-400'
                  }`}
                  onClick={(e) => {
                    e.stopPropagation();
                    toggleSectionCompletion(index);
                  }}
                >
                  {section.completed && (
                    <div className="w-full h-full flex items-center justify-center">
                      <div className="w-1 h-1 bg-white rounded-full"></div>
                    </div>
                  )}
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Editor */}
        <div className="flex-1 flex flex-col">
          {/* Editor Header */}
          <div className="bg-gray-800 border-b border-gray-700 p-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-white font-semibold text-lg">
                  Security Incident 2024-001
                </h2>
                <div className="flex items-center gap-4 text-sm text-gray-400 mt-1">
                  <span className="flex items-center gap-1">
                    <Eye className="w-4 h-4" />
                    Export
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                  <Download className="w-4 h-4" />
                  Export
                </button>
              </div>
            </div>
          </div>

          {/* Editor Content */}
          <div className="flex-1 p-8 overflow-y-auto">
            <div className="max-w-4xl mx-auto">
              {/* Report Header */}
              <div className="mb-8">
                <h1 className="text-3xl font-bold text-white mb-4">
                  {reportTitle}
                </h1>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-400">Incident ID:</span>
                    <span className="text-white ml-2">{incidentId}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">Date Created:</span>
                    <span className="text-white ml-2">{dateCreated}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">Analyst:</span>
                    <span className="text-white ml-2">{analyst}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">Report Type:</span>
                    <span className="text-white ml-2">{reportType}</span>
                  </div>
                </div>
              </div>

              {/* Current Section */}
              <div className="mb-6">
                <h2 className="text-2xl font-semibold text-white mb-4">
                  {sections[activeSection]?.title}
                </h2>
              </div>

              {/* React Quill Editor */}
              <div className="mb-8">
                <ReactQuill
                  theme="snow"
                  value={sections[activeSection]?.content || ''}
                  onChange={updateSectionContent}
                  modules={modules}
                  formats={formats}
                  placeholder="Start writing your report content here..."
                />
              </div>

              {/* Evidence Tables for specific sections */}
              {sections[activeSection]?.title === 'Evidence & Findings' && (
                <div className="mt-8 space-y-6">
                  <div className="bg-gray-800 rounded-lg border border-gray-700 p-6">
                    <h3 className="text-white font-semibold mb-4">Investigation Objectives</h3>
                    <div className="bg-gray-700 p-4 rounded">
                      <ul className="space-y-2 text-gray-200">
                        <li>• Identify the attack vector and timeline</li>
                        <li>• Determine the extent of system compromise</li>
                        <li>• Assess data exfiltration risks</li>
                        <li>• Document evidence for potential legal proceedings</li>
                      </ul>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-6">
                    <div className="bg-gray-800 rounded-lg border border-gray-700 p-6">
                      <h4 className="text-white font-medium mb-3">Compromised Assets</h4>
                      <div className="space-y-2 text-sm text-gray-300">
                        <div>• Server: web-prod-01</div>
                        <div>• Workstation: WS-001-JD</div>
                        <div>• Database: customer-db-01</div>
                      </div>
                    </div>
                    
                    <div className="bg-gray-800 rounded-lg border border-gray-700 p-6">
                      <h4 className="text-white font-medium mb-3">Malware Identified</h4>
                      <div className="space-y-2 text-sm text-gray-300">
                        <div>• Backdoor: Win32.Agent</div>
                        <div>• Keylogger: Win32.KeyCapture</div>
                        <div>• Payload: Win32.Adwindows</div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex items-center justify-between mt-8 pt-6 border-t border-gray-700">
                <div className="flex items-center gap-3">
                  <button className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors">
                    <Save className="w-4 h-4 inline mr-2" />
                    Save Changes
                  </button>
                  <button className="px-4 py-2 bg-gray-700 text-gray-300 rounded-lg hover:bg-gray-600 transition-colors">
                    <Eye className="w-4 h-4 inline mr-2" />
                    Preview
                  </button>
                </div>
                
                <div className="flex items-center gap-2 text-sm text-gray-400">
                  <Clock className="w-4 h-4" />
                  <span>Auto-saved 30 seconds ago</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
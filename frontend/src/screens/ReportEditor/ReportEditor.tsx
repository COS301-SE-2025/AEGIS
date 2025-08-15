import React, { useState, useEffect, useRef, useCallback } from "react";
import ReactQuill from "react-quill";
import "react-quill/dist/quill.snow.css";
import {
  FileText,
  Download,
  Save,
  Plus,
  Clock,
  Users,
  Calendar,
  Eye,
} from "lucide-react";
import { useParams } from "react-router-dom";
import axios from "axios";
import { CheckCircle, XCircle, Loader2 } from "lucide-react";

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

interface Report {
  id: string;
  name: string;
  type: string;
  content: ReportSection[];
  incidentId?: string;
  dateCreated?: string;
  analyst?: string;
}



//type Section = { id: string; title: string; content: string; order: number; updated_at?: string }
const API_URL = "http://localhost:8080/api/v1";

// keep endpoint the same; just return a simple result
async function putSectionContent(reportId: string, sectionId: string, content: string) {
  const token = sessionStorage.getItem("authToken");
  const res = await axios.put(
    `${API_URL}/reports/${reportId}/sections/${sectionId}/content`,
    { content },
    { headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } }
  );
  return { ok: res.status >= 200 && res.status < 300, status: res.status };
}


async function putSectionTitle(reportId: string, sectionId: string, title: string) {
  const token = sessionStorage.getItem("authToken");
  await axios.put(
    `${API_URL}/reports/${reportId}/sections/${sectionId}/title`,
    { title },
    { headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } }
  );
}

export const ReportEditor = () => {
 const { reportId } = useParams<{ reportId: string }>();

  // state
  const [report, setReport] = useState<Report | null>(null);
  const [sections, setSections] = useState<ReportSection[]>([]);
  const [activeSection, setActiveSection] = useState(0);

  const [reportTitle, setReportTitle] = useState("");
  const [incidentId, setIncidentId] = useState("");
  const [dateCreated, setDateCreated] = useState("");
  const [analyst, setAnalyst] = useState("");
  const [reportType, setReportType] = useState("");
  const lastSavedRef  = useRef<string>("");
  const lastQueuedRef = useRef<string>("");
  const sectionIdRef  = useRef<string>("");
  const timerRef      = useRef<number | null>(null);
  const [saveState, setSaveState] = useState<"idle"|"saving"|"saved"|"error">("idle");
const [lastSavedAt, setLastSavedAt] = useState<number|null>(null);
const [dirty, setDirty] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // debounce timer for autosave
  //const saveTimer = useRef<number | null>(null);

  // fetch report
  useEffect(() => {
    (async () => {
      if (!reportId) return;
      const token = sessionStorage.getItem("authToken");
      if (!token) return;

      // If your backend returns { metadata, content }, map accordingly.
      const { data } = await axios.get<Report>(`${API_URL}/reports/${reportId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      setReport(data);
      setSections(Array.isArray(data.content) ? data.content : []);
      setReportTitle(data.name || "");
      setIncidentId(data.incidentId || "");
      setDateCreated(data.dateCreated || "");
      setAnalyst(data.analyst || "");
      setReportType(data.type || "");
    })().catch(err => console.error("Error fetching report:", err));
  }, [reportId]);

  // keep activeSection in range if sections change
  useEffect(() => {
    if (activeSection >= sections.length && sections.length > 0) {
      setActiveSection(0);
    }
  }, [sections, activeSection]);

  // toggle completion (purely client-side unless you add an API)
  const toggleSectionCompletion = (index: number) => {
    setSections(prev =>
      prev.map((s, i) => (i === index ? { ...s, completed: !s.completed } : s))
    );
  };

const EMPTY_PATTERNS = ["<p><br></p>", "<p></p>"];
const normalizeHtml = (html: string) => {
  const t = html.trim();
  if (!t) return "";
  const compact = t.replace(/\s+/g, "").toLowerCase();
  return EMPTY_PATTERNS.includes(compact) ? "" : html;
};


  
// const toggleSectionCompletion = (index: number) => {
//   if (!sections) return; // Guard: do nothing if sections is null

//   const updatedSections = sections.map((section, i) =>
//     i === index ? { ...section, completed: !section.completed } : section
//   );
//   setSections(updatedSections);
// };


  // ReactQuill change handler: optimistic update + debounce save
useEffect(() => {
  const current = sections[activeSection];
  sectionIdRef.current  = current?.id || "";
  lastQueuedRef.current = "";
  lastSavedRef.current  = current ? normalizeHtml(current.content) : "";
  setDirty(false);
  if (timerRef.current) { window.clearTimeout(timerRef.current); timerRef.current = null; }
}, [sections, activeSection]);

useEffect(() => () => { if (timerRef.current) window.clearTimeout(timerRef.current); }, []);


const scheduleSave = useCallback((contentNorm: string) => {
  if (contentNorm === lastSavedRef.current || contentNorm === lastQueuedRef.current) return;
  setDirty(true);
  lastQueuedRef.current = contentNorm;
  if (timerRef.current) window.clearTimeout(timerRef.current);
  timerRef.current = window.setTimeout(async () => {
    const secId   = sectionIdRef.current;
    const payload = lastQueuedRef.current;
    lastQueuedRef.current = "";
    if (!secId || !reportId) return;

    try {
      setSaveState("saving");
      await putSectionContent(reportId, secId, payload);
      lastSavedRef.current = payload;
      setDirty(false);
      setSaveState("saved");
      setLastSavedAt(Date.now());
    } catch (e:any) {
      console.error("Autosave failed", e?.response?.data || e);
      setSaveState("error");
    }
  }, 600);
}, [reportId]);

const handleEditorChange = useCallback(
  (nextHtml: string, _delta: any, source: "user" | "api" | "silent") => {
    if (source !== "user") return; // 🚫 ignore programmatic toggles that emit empty

    // reflect what the user typed
    setSections(prev => {
      const copy = [...prev];
      if (copy[activeSection]) copy[activeSection] = { ...copy[activeSection], content: nextHtml };
      return copy;
    });

    // persist normalized content
    scheduleSave(normalizeHtml(nextHtml));
  },
  [activeSection, scheduleSave]
);


const flushSaveNow = useCallback(
  async (content?: string, opts?: { force?: boolean }) => {
    const force = !!opts?.force;

    // prefer ref, but fall back to current section if ref isn't ready yet
    const sectionId = sectionIdRef.current || sections[activeSection]?.id || "";
    if (!reportId || !sectionId) {
      console.warn("Save aborted: missing reportId or sectionId", { reportId, sectionId });
      return;
    }

    if (timerRef.current) { window.clearTimeout(timerRef.current); timerRef.current = null; }

    // build + normalize payload
    const pending = lastQueuedRef.current;                 // may be ""
    const current = sections[activeSection]?.content;      // string | undefined
    const raw = content !== undefined
      ? content
      : (pending !== "" ? pending : (current ?? lastSavedRef.current));

    const payload = normalizeHtml(raw);

    // Only skip when identical AND not forced
    if (!force && payload === lastSavedRef.current) {
      setSaveState("saved");
      return;
    }

    try {
      setSaveState("saving");
      await putSectionContent(reportId, sectionId, payload);
      lastSavedRef.current = payload;
      lastQueuedRef.current = "";
      setDirty(false);
      setSaveState("saved");
      setLastSavedAt(Date.now());
    } catch (e: any) {
      console.error("Manual save failed", e?.response?.data || e);
      setSaveState("error");
    }
  },
  [reportId, sections, activeSection]
);

// Keyboard shortcut: Ctrl/⌘+S
useEffect(() => {
  const onKey = (e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "s") {
      e.preventDefault();
      flushSaveNow(sections[activeSection]?.content ?? "");
    }
  };
  window.addEventListener("keydown", onKey);
  return () => window.removeEventListener("keydown", onKey);
}, [flushSaveNow, sections, activeSection]);


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
// wrapper that saves first then downloads
const flushAndDownload = async () => {
  const id = reportId ?? report?.id; // fall back to loaded report if you have it
  if (!id) {
    setError("No report ID available");
    return;
  }
  await flushSaveNow(sections[activeSection]?.content ?? "", { force: true });
  await downloadReport(id);
};

async function updateSectionTitle(reportId: string, sectionId: string, title: string) {
  const token = sessionStorage.getItem("authToken");
  const res = await axios.put(
    `${API_URL}/reports/${reportId}/sections/${sectionId}/title`,
    { title },
    { headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } }
  );
  return res.status;
}


  // const [sections, setSections] = useState<ReportSection[]>([
  //   {
  //     id: 'executive-summary',
  //     title: 'Executive Summary',
  //     content: `<p>On January 14, 2024, the Security Operations Center (SOC) detected suspicious network activity indicating a potential security breach. This report documents the comprehensive digital forensics investigation conducted to determine the scope, attack vector, and root cause of the incident.</p><p>Initial analysis revealed unauthorized access to the corporate network through a compromised employee workstation. The investigation timeline, findings, and recommended remediation actions are detailed in this report.</p>`,
  //     completed: true
  //   },
  //   {
  //     id: 'incident-scope',
  //     title: 'Incident Scope & Objectives',
  //     content: `<p><strong>Investigation Objectives:</strong></p><ul><li>Identify the attack vector and timeline</li><li>Determine the extent of system compromise</li><li>Assess data exfiltration risks</li><li>Document evidence for potential legal proceedings</li></ul>`,
  //     completed: true
  //   },
  //   {
  //     id: 'evidence-findings',
  //     title: 'Evidence & Findings',
  //     content: '<p>Content for Evidence &amp; Findings section...</p>',
  //     completed: false
  //   },
  //   {
  //     id: 'compromised-assets',
  //     title: 'Compromised Assets',
  //     content: '',
  //     completed: false
  //   },
  //   {
  //     id: 'malware-identified',
  //     title: 'Malware Identified',
  //     content: '',
  //     completed: false
  //   }
  // ]);

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

   const isLoading = !report;



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
      <style>{`
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
      <span>{incidentId}</span> {/* Dynamically show incident ID */}
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

  {/* Section List */}
  <div className="space-y-1">
  {sections.map((section, index) => (
    <button
      key={section.id}
      onClick={async () => {
        if (index === activeSection) return;
        await flushSaveNow(sections[activeSection]?.content ?? "", { force: true });
        setActiveSection(index);
      }}
      disabled={saveState === "saving"} // optional: prevent switching during save
      className={`w-full flex items-center justify-between p-3 rounded-lg text-left transition-colors ${
        activeSection === index
          ? "bg-blue-600 text-white"
          : "hover:bg-gray-700 text-gray-300"
      }`}
    >
      <span className="font-medium">{section.title}</span>

      {/* keep your completion toggle */}
      <div
        className={`w-3 h-3 rounded-full border-2 ${
          section.completed ? "bg-green-500 border-green-500" : "border-gray-400"
        }`}
        onClick={(e) => {
          e.stopPropagation(); // don’t trigger section switch
          toggleSectionCompletion(index);
        }}
      />
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
                <button
              type="button"
              onClick={flushAndDownload}
              disabled={!reportId && !report}  // disable until we know an id
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
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
            {sections && sections[activeSection] && (
          <div className="mb-6">
            <h2 className="text-2xl font-semibold text-white mb-4">
              {sections[activeSection].title}
            </h2>
          </div>
        )}


              {/* React Quill Editor */}
              {sections && sections[activeSection] && (
          <div className="mb-8">
            <ReactQuill
              theme="snow"
              value={sections[activeSection]?.content?? ""}
              onChange={handleEditorChange}
              modules={modules}
              formats={formats}
              placeholder="Start writing your report content here..."
            />
          </div>
        )}


              {/* Evidence Tables for specific sections */}
              {sections && sections[activeSection] && sections[activeSection].title === 'Evidence & Findings' && (
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
            <button
              type="button"
              onClick={() => flushSaveNow(sections[activeSection]?.content ?? "", { force: true })}
              disabled={saveState === "saving" || !sections[activeSection]}
              aria-busy={saveState === "saving"}
              className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
            >
              <Save className="w-4 h-4 inline mr-2" />
              Save Changes
            </button>
<div className="flex items-center gap-2 text-sm text-gray-300">
  {saveState === "saving" && (
    <>
      <Loader2 className="w-4 h-4 animate-spin" />
      <span>Saving…</span>
    </>
  )}
  {saveState === "saved" && (
    <>
      <CheckCircle className="w-4 h-4 text-emerald-500" />
      <span>
        Saved{lastSavedAt ? ` at ${new Date(lastSavedAt).toLocaleTimeString()}` : ""}
      </span>
    </>
  )}
  {saveState === "error" && (
    <>
      <XCircle className="w-4 h-4 text-red-500" />
      <span>Save failed. Try again.</span>
    </>
  )}
  {saveState === "idle" && dirty && <span>Unsaved changes</span>}

  {/* Accessible live region (screen readers) */}
  <span className="sr-only" role="status" aria-live="polite" aria-atomic="true">
    {saveState === "saving" && "Saving"}
    {saveState === "saved" && "Saved successfully"}
    {saveState === "error" && "Save failed"}
  </span>
</div>



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
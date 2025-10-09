
import "react-quill/dist/quill.snow.css";
import axios from "axios";
import { toast } from "react-hot-toast";
import { createRoot } from "react-dom/client";
import  { useState, useEffect, useRef, useCallback } from "react";
import ReactQuill from "react-quill";
import { useParams } from "react-router-dom";
interface ContextAutofillResponse {
  case_info: any;
  iocs: any[];
  evidence: any[];
  timeline: any[];
}
import { Plus, Download, Check, X, Pencil, Sparkles, ChevronUp, ChevronDown, Shield, AlertCircle, FileText, Clock, Calendar, Eye, CheckCircle, XCircle, Loader2, Trash2, GripVertical, AlertTriangle, Save, User } from "lucide-react";
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  useSortable,
  arrayMove,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";

interface ReportSection {
  id: string;
  title: string;
  content: string;
  completed?: boolean; // <- optional
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
  case_id?: string;
}

type ConfirmOpts = {
  title: string;
  description?: string;
  confirmText?: string;
  cancelText?: string;
  danger?: boolean;
};

//type Section = { id: string; title: string; content: string; order: number; updated_at?: string }
const API_URL = "https://localhost/api/v1";

async function putReportStatus(reportId: string, status: "draft" | "review" | "published") {
  const token = sessionStorage.getItem("authToken");
  console.log("report id:",reportId)
  const res = await axios.put(
    `${API_URL}/reports/${reportId}/status`,
    { status },
    { headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } }
  );
  return res.data;
}


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
type SortableSectionItemProps = {
  section: ReportSection;
  index: number;
  activeIndex: number;
  onSelect: (idx: number) => Promise<void> | void;
  onDelete: (id: string) => void;
  deletingId: string | null;
  saveState: "idle" | "saving" | "saved" | "error";
};
// utils (same file or a helpers file)
export function formatIsoDateTime(iso?: string) {
  if (!iso) return "";
  const d = new Date(iso);
  return d.toLocaleString("en-GB", {
    day: "2-digit",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",   // drop if you don't want seconds
    hour12: false,
    // timeZoneName: "short", // uncomment to show e.g. GMT
  });
}
export function confirm(opts: ConfirmOpts): Promise<boolean> {
  const container = document.createElement("div");
  document.body.appendChild(container);
  const root = createRoot(container);

  const cleanup = () => {
    setTimeout(() => {
      root.unmount();
      container.remove();
    }, 0);
  };

  return new Promise<boolean>((resolve) => {
    function Modal() {
      // Close on ESC
      useEffect(() => {
        const onKey = (e: KeyboardEvent) => {
          if (e.key === "Escape") {
            resolve(false);
            cleanup();
          }
        };
        document.addEventListener("keydown", onKey);
        return () => document.removeEventListener("keydown", onKey);
      }, []);

      return (
        <div
          className="fixed inset-0 z-[9999] flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
        >
          {/* Backdrop */}
          <div
            className="absolute inset-0 bg-black/60"
            onClick={() => {
              resolve(false);
              cleanup();
            }}
          />
          {/* Dialog */}
          <div className="relative w-[min(420px,92vw)] max-h-[85vh] overflow-y-auto rounded-xl border border-gray-700 bg-gray-800 p-5 shadow-2xl">
            <div className="flex items-start gap-3">
              <AlertTriangle className="mt-0.5 h-5 w-5 text-amber-400" />
              <div className="flex-1">
                <h3 className="text-white font-semibold">{opts.title}</h3>
                {opts.description && (
                  <p className="mt-1 text-sm text-gray-300">{opts.description}</p>
                )}
                <div className="mt-4 flex justify-end gap-2">
                  <button
                    className="rounded bg-gray-700 px-3 py-1.5 text-gray-200 hover:bg-gray-600"
                    onClick={() => {
                      resolve(false);
                      cleanup();
                    }}
                  >
                    {opts.cancelText ?? "Cancel"}
                  </button>
                  <button
                    className={
                      "rounded px-3 py-1.5 text-white " +
                      (opts.danger
                        ? "bg-red-600 hover:bg-red-700"
                        : "bg-blue-600 hover:bg-blue-700")
                    }
                    onClick={() => {
                      resolve(true);
                      cleanup();
                    }}
                  >
                    {opts.confirmText ?? "Confirm"}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      );
    }

    root.render(<Modal />);
  });
}

function SortableSectionItem({
  section,
  index,
  activeIndex,
  onSelect,
  onDelete,
  deletingId,
  saveState,
}: SortableSectionItemProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: section.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`flex items-center justify-between p-3 rounded-lg text-left transition-colors ${
        activeIndex === index
          ? "bg-primary text-white"
          : "hover:bg-primary/60 text-foreground/60"
      } ${isDragging ? "ring-2 ring-blue-500/70" : ""}`}
      onClick={() => onSelect(index)}
      role="button"
    >
      <div className="flex items-center gap-2 min-w-0">
        {/* Drag handle */}
        <button
          type="button"
          {...attributes}
          {...listeners}
          className="p-1 -ml-1 mr-1 rounded cursor-grab active:cursor-grabbing hover:bg-primary/60"
          onClick={(e) => e.stopPropagation()}
          aria-label="Drag to reorder"
          title="Drag to reorder"
        >
          <GripVertical className="w-4 h-4" />
        </button>

        <span className="font-medium truncate">{section.title}</span>
      </div>

      <div className="flex items-center gap-2">
      

        {/* delete */}
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onDelete(section.id);
          }}
          disabled={deletingId === section.id || saveState === "saving"}
          className={`p-1 rounded hover:bg-foreground/60 ${
            activeIndex === index ? "text-white" : "text-foreground/60"
          } disabled:opacity-50`}
          title="Delete section"
          aria-label="Delete section"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  );
}

export const ReportEditor = () => {
  // Dummy key to force ghost text unmount
  const [] = useState(0);
  // Enhancement button state for last dropped summary
  const [showEnhanceButton, setShowEnhanceButton] = useState(false);
  const [lastDroppedSummary, setLastDroppedSummary] = useState("");
  const [enhancing, setEnhancing] = useState(false);
  // state
  const [report, setReport] = useState<Report & { mongoId?: string } | null>(null);
  const [sections, setSections] = useState<ReportSection[]>([]);
  const [activeSection, setActiveSection] = useState(0);
  const [contextOpen, setContextOpen] = useState(true);
  const [sectionContext, setSectionContext] = useState<ContextAutofillResponse | null>(null);
  const { reportId } = useParams<{ reportId: string }>();
  // Ref to block suggestion re-setting during insertion
  const insertingSuggestionRef = useRef(false);
   const [isDFIRAdmin, setIsDFIRAdmin] = useState(false);
  const [, setTenantId] = useState<string | null>(null);

  useEffect(() => {
    // Check role and tenantId after mount (when sessionStorage is available)
    try {
      const token = sessionStorage.getItem('authToken');
      if (token) {
        const payload = JSON.parse(atob(token.split('.')[1]));
        setIsDFIRAdmin(payload.role === 'DFIR Admin' || payload.role === 'admin');
        setTenantId(payload.tenant_id || payload.tenantId || null);
      }
    } catch {}
  }, []);
  useEffect(() => {
    const fetchContext = async () => {
      const token = sessionStorage.getItem("authToken");
      if (!reportId || !sections[activeSection]?.id) return;
      try {
        const { data } = await axios.get(
          `${API_URL}/reports/${reportId}/sections/${sections[activeSection].id}/context`,
          { headers: { Authorization: `Bearer ${token}` } }
        );
        setSectionContext(data as ContextAutofillResponse);
      } catch (e) {
        setSectionContext(null);
        console.error("Failed to fetch section context", e);
      }
    };
    fetchContext();
  }, [reportId, sections, activeSection]);
  const quillRef = useRef<ReactQuill | null>(null);

    const [aiSuggestion, setAiSuggestion] = useState<string>("");
    const [aiLoading, setAiLoading] = useState(false);

  const [reportTitle, setReportTitle] = useState("");
  //const documentTitle = reportTitle; 
  const [incidentId, setIncidentId] = useState("");
  const [, setDateCreated] = useState("");
  const [, setAnalyst] = useState("");
  const [reportType, setReportType] = useState("");
  const lastSavedRef  = useRef<string>("");
  const lastQueuedRef = useRef<string>("");
  const sectionIdRef  = useRef<string>("");
  const timerRef      = useRef<number | null>(null);
  const [saveState, setSaveState] = useState<"idle"|"saving"|"saved"|"error">("idle");
const [lastSavedAt, setLastSavedAt] = useState<number|null>(null);
const [dirty, setDirty] = useState(false);
  const [, setError] = useState<string | null>(null);
// --- title edit state ---
const [editingTitleSectionId, setEditingTitleSectionId] = useState<string|null>(null);
const [tempSectionTitle, setTempSectionTitle] = useState("");
const [, setTitleDirty] = useState(false);
const [titleSaving, setTitleSaving] = useState<"idle"|"saving"|"saved"|"error">("idle");
const [addingBusy, setAddingBusy] = useState(false);
const [deletingId, setDeletingId] = useState<string | null>(null);
const [reportNameState, setReportNameState] = useState<"idle" | "saving" | "saved" | "error">("idle");
const [reportTitleDirty, setReportTitleDirty] = useState(false);
const reportNameTimerRef = useRef<number | null>(null);
const lastReportNameSavedRef = useRef<string>("");
const [caseId, setCaseId] = useState("");
const [isPreviewMode, setIsPreviewMode] = useState(false);
  // Remove duplicate ContextAutofillResponse type and sectionContext state
  // Sticky/floating context panel
  // Place this at the top of the main return

// local-only section helpers (so we can skip API calls until backend is wired)
const isLocalSection = (id: string) => id.startsWith("local-");

// add/delete UI state
const [adding, setAdding] = useState(false);
const [newSectionTitle, setNewSectionTitle] = useState("");



const [, setRecentReports] = useState<RecentReport[]>([]); // NEW
const [, setRecentLoading] = useState(true);               // NEW
const [, setRecentError] = useState<string | null>(null);    // NEW
const [lastModified, setLastModified] = useState<string>("");

  // fetch report
  // useEffect(() => {
  //   (async () => {
  //     if (!reportId) return;
  //     const token = sessionStorage.getItem("authToken");
  //     if (!token) return;

  //     // If your backend returns { metadata, content }, map accordingly.
  //     const { data } = await axios.get<Report>(`${API_URL}/reports/${reportId}`, {
  //       headers: { Authorization: `Bearer ${token}` },
  //     });
                // {(Array.isArray(sectionContext.iocs) ? sectionContext.iocs : []).map((ioc, i) => (
                //   <li key={i} className="flex items-center gap-2 text-sm text-red-300">
                //     {/* Render IOC details here, e.g. ioc.indicator or ioc.type */}
                //     {typeof ioc === 'string' ? ioc : JSON.stringify(ioc)}
                //   </li>
                // ))}
  //     setAnalyst(data.analyst || "");
  //     setReportType(data.type || "");
  //   })().catch(err => console.error("Error fetching report:", err));
  // }, [reportId]);
const loadReport = useCallback(async (id: string) => {
  const token = sessionStorage.getItem("authToken");
  if (!token) return;

  try {
    const res = await axios.get(`${API_URL}/reports/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data: any = res.data;
    // Defensive: ensure data is an object
    if (!data || typeof data !== "object") {
      throw new Error("Invalid report data");
    }
    const meta = data.metadata || data;
    const rawSections: any[] = Array.isArray(data.content)
      ? data.content
      : Array.isArray(data.sections)
      ? data.sections
      : [];
    const mappedSections: ReportSection[] = rawSections.map((s: any) => ({
      id: String(s.id),
      title: String(s.title ?? "Untitled Section"),
      content: String(s.content ?? ""),
      completed: !!s.completed,
    }));
    const uiReport: Report & { mongoId?: string } = {
      id: String(meta?.id ?? ""),
      name: String(meta?.name ?? ""),
      type: String(meta?.status ?? meta?.type ?? ""),
      content: mappedSections,
      incidentId: String(meta?.report_number ?? ""),
      dateCreated: String(
        meta?.created_at ?? meta?.createdAt ?? meta?.date_created ?? ""
      ),
      analyst: String(meta?.author ?? meta?.analyst ?? ""),
      case_id: String(meta?.case_id ?? ""),
      mongoId: String(meta?.mongo_id ?? meta?.MongoID ?? ""),
    };
    // Push into state
    setReport(uiReport);
    setSections(uiReport.content);
    // Title + debounce guard stay in sync
    setReportTitle(uiReport.name);
    lastReportNameSavedRef.current = uiReport.name;
    setReportNameState("idle");
    setReportTitleDirty(false);
    // Other fields
    setIncidentId(uiReport.incidentId ?? "");
    setDateCreated(uiReport.dateCreated ?? "");
    setAnalyst(uiReport.analyst ?? "");
    setReportType(uiReport.type ?? "");
    setCaseId(uiReport.case_id ?? "");
    // NEW: last modified / updated-at (support multiple keys)
    const lm =
      meta?.updated_at ??
      meta?.last_modified ??
      meta?.lastModified ??
      "";
    setLastModified(String(lm));
  } catch (err) {
    console.error("Error fetching report:", err);
  }
}, []);

// put this near the top of your component file (or in a utils file)

useEffect(() => {
  if (!reportId) return;
  loadReport(reportId).catch(err => console.error("Error fetching report:", err));
}, [reportId, loadReport]);

  // keep activeSection in range if sections change
  useEffect(() => {
    if (activeSection >= sections.length && sections.length > 0) {
      setActiveSection(0);
    }
  }, [sections, activeSection]);

  // toggle completion (purely client-side unless you add an API)
    // Automatically insert AI suggestion when it arrives
  useEffect(() => {
    if (aiSuggestion && !aiLoading) {
      setSections(prev => {
        const copy = [...prev];
        if (copy[activeSection]) {
          copy[activeSection] = {
            ...copy[activeSection],
            content: (copy[activeSection].content || "") + aiSuggestion,
          };
        }
        return copy;
      });
      setAiSuggestion("");
    }
  }, [aiSuggestion, aiLoading, activeSection, setSections]);

const EMPTY_PATTERNS = ["<p><br></p>", "<p></p>"];
const normalizeHtml = (html: string) => {
  const t = html.trim();
  if (!t) return "";
  const compact = t.replace(/\s+/g, "").toLowerCase();
  return EMPTY_PATTERNS.includes(compact) ? "" : html;
};




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
      if (isLocalSection(secId)) {
  // pretend-save locally (no network)
  lastSavedRef.current = payload;
  setDirty(false);
  setSaveState("saved");
  setLastSavedAt(Date.now());
  return;
}
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

    if (payload === "" && lastSavedRef.current === "") {
  setSaveState("saved");
  return;
}
    try {
      setSaveState("saving");
      if (isLocalSection(sectionId)) {
  lastSavedRef.current = payload;
  lastQueuedRef.current = "";
  setDirty(false);
  setSaveState("saved");
  setLastSavedAt(Date.now());
  return;
}

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

//toggle preview
//toggle preview
const togglePreview = useCallback(async () => {
  if (!isPreviewMode && (dirty || saveState === "saving")) {
    await flushSaveNow(sections[activeSection]?.content ?? "", { force: true });
  }
  setIsPreviewMode(prev => !prev);
}, [isPreviewMode, dirty, saveState, flushSaveNow, sections, activeSection]);

useEffect(() => { // NEW
  let cancelled = false;
  (async () => {
    try {
      setRecentLoading(true);
      setRecentError(null);
      const token = sessionStorage.getItem("authToken");
      if (!token) {
        setRecentReports([]);
        setRecentLoading(false);
        return;
      }
      const res = await axios.get(`${API_URL}/reports/recent?limit=6&mine=true`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      const items = (res.data as any[]).map(x => ({
        id: x.id,
        title: x.title,
        status: x.status as RecentReport["status"],
        lastModified: x.lastModified as string,
      })) as RecentReport[];

      if (!cancelled) setRecentReports(items);
    } catch (e: any) {
      if (!cancelled) setRecentError("Failed to load recent reports");
      console.error("Recent reports error:", e?.response?.data ?? e);
    } finally {
      if (!cancelled) setRecentLoading(false);
    }
  })();
  return () => { cancelled = true; };
}, []);



const handleAddSection = useCallback(async () => {
  const rid = reportId ?? report?.id;
  if (!rid) return;

  const title = newSectionTitle.trim() || "New Section";
  // Insert AFTER the active section; backend uses 1-based order.
  const order = activeSection + 2;

  try {
    setAddingBusy(true);
    await postAddSection(String(rid), title, "", order);
    await loadReport(String(rid));        // refresh with server IDs/order
    setActiveSection(order - 1);          // focus new section
    setAdding(false);
    setNewSectionTitle("");
  } catch (e) {
    console.error("Add section failed", e);
    setError("Failed to add section");
  } finally {
    setAddingBusy(false);
  }
}, [reportId, report?.id, newSectionTitle, activeSection, loadReport]);

// Remove a section locally and keep the UI stable
const handleDeleteSection = useCallback(
  async (sectionId: string) => {
    const rid = reportId ?? report?.id;
    if (!rid) return;

    const sec = sections.find((s) => s.id === sectionId);

    const confirmed = await confirm({
      title: "Delete this section?",
      description: sec?.title ? `“${sec.title}” will be removed.` : undefined,
      confirmText: "Delete",
      cancelText: "Cancel",
      danger: true,
    });
    if (!confirmed) return;

    const idx = sections.findIndex((s) => s.id === sectionId);
    if (idx === -1) return;

    // snapshot for rollback
    const prevSections = sections;
    const prevActive = activeSection;

    // optimistic UI
    const next = sections.filter((s) => s.id !== sectionId);
    const newActive =
      idx < activeSection
        ? activeSection - 1
        : idx === activeSection
        ? Math.max(0, activeSection - (idx === sections.length - 1 ? 1 : 0))
        : activeSection;

    setDeletingId(sectionId);
    setSections(next);
    setActiveSection(newActive);

    try {
      // If it never existed server-side, we're done
      if (sectionId.startsWith("local-")) {
        toast.success("Section deleted");
        return;
      }

      await deleteSection(String(rid), sectionId);
      // Optional: re-fetch canonical state
      await loadReport(String(rid));
      toast.success("Section deleted");
    } catch (e) {
      console.error("Delete section failed", e);
      // rollback
      setSections(prevSections);
      setActiveSection(prevActive);
      toast.error("Failed to delete section");
    } finally {
      setDeletingId(null);
    }
  },
  [reportId, report?.id, sections, activeSection, loadReport]
);


async function postAddSection(reportId: string, title: string, content = "", order?: number) {
  const token = sessionStorage.getItem("authToken");
  const res = await axios.post(
    `${API_URL}/reports/${reportId}/sections`,
    { title, content, order },
    { headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } }
  );
  return res.data; // handler returns {status: "..."} (no id), we’ll refetch below
}

async function deleteSection(reportId: string, sectionId: string) {
  const token = sessionStorage.getItem("authToken");
  const res = await axios.delete(
    `${API_URL}/reports/${reportId}/sections/${sectionId}`,
    { headers: { Authorization: `Bearer ${token}` } }
  );
  return res.data; // {status: "..."}
}



// API helper
async function putSectionOrder(
  reportId: string,
  sectionId: string,
  order: number
) {
  const token = sessionStorage.getItem("authToken");
  await axios.put(
    `${API_URL}/reports/${reportId}/sections/${sectionId}/reorder`,
    { order }, // 1-based
    {
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
    }
  );
}


// sensors (tiny drag threshold avoids accidental drags)
const sensors = useSensors(
  useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
);

// drag end => reorder UI + persist
const handleDragEnd = useCallback(
  async (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;

    const oldIndex = sections.findIndex((s) => s.id === String(active.id));
    const newIndex = sections.findIndex((s) => s.id === String(over.id));
    if (oldIndex === -1 || newIndex === -1) return;

    const newList = arrayMove(sections, oldIndex, newIndex);
    setSections(newList);

    // keep the active selection sensible
    setActiveSection((prev) => {
      if (prev === oldIndex) return newIndex;
      if (prev > oldIndex && prev <= newIndex) return prev - 1;
      if (prev < oldIndex && prev >= newIndex) return prev + 1;
      return prev;
    });

    // persist to server (choose bulk OR per-item)
   try {
  const rid = reportId ?? report?.id;
  if (!rid) return;

  // Optional: flush pending edits before persisting order
  if (dirty || saveState === "saving") {
    await flushSaveNow(sections[activeSection]?.content ?? "");
  }

  // Persist each item’s order (sequential to avoid racey reads)
  for (let i = 0; i < newList.length; i++) {
    await putSectionOrder(String(rid), newList[i].id, i + 1); // 1-based
  }

  // Optional but nice: confirm with server's canonical state
  // await loadReport(String(rid));
} catch (e) {
  console.error("Reorder persist failed", e);
  const rid = reportId ?? report?.id;
  if (rid) await loadReport(String(rid)); // revert to server order on error
}

  },
  [sections, reportId, report?.id, dirty, saveState, flushSaveNow, activeSection, loadReport]
);

async function putReportName(reportId: string, name: string) {
  const token = sessionStorage.getItem("authToken");
  await axios.put(
    `${API_URL}/reports/${reportId}/name`,
    { name },
    { headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } }
  );
}

// Debounce like your content saver
const scheduleReportNameSave = useCallback((nextName: string) => {
  if (nextName === lastReportNameSavedRef.current) {
    setReportTitleDirty(false);
    setReportNameState("saved");
    return;
  }

  setReportTitleDirty(true);
  if (reportNameTimerRef.current) window.clearTimeout(reportNameTimerRef.current);

  reportNameTimerRef.current = window.setTimeout(async () => {
    const rid = reportId ?? report?.id;
    if (!rid) return;
    try {
      setReportNameState("saving");
      await putReportName(String(rid), nextName.trim());
      lastReportNameSavedRef.current = nextName.trim();
      setReportTitleDirty(false);
      setReportNameState("saved");
    } catch (e) {
      console.error("Report name save failed", e);
      setReportNameState("error");
    }
  }, 600);
}, [reportId, report?.id]);

const flushReportNameNow = useCallback(
  async (opts?: { force?: boolean }) => {
    const rid = reportId ?? report?.id;
    if (!rid) return;
    if (reportNameTimerRef.current) {
      window.clearTimeout(reportNameTimerRef.current);
      reportNameTimerRef.current = null;
    }
    const current = (reportTitle ?? "").trim();
    if (!opts?.force && current === lastReportNameSavedRef.current) {
      setReportNameState("saved");
      setReportTitleDirty(false);
      return;
    }
    if (current === "") {
      // optional: prevent empty names
      return;
    }
    try {
      setReportNameState("saving");
      await putReportName(String(rid), current);
      lastReportNameSavedRef.current = current;
      setReportTitleDirty(false);
      setReportNameState("saved");
    } catch (e) {
      console.error("Report name save failed", e);
      setReportNameState("error");
    }
  },
  [reportId, report?.id, reportTitle]
);

useEffect(() => () => {
  if (reportNameTimerRef.current) window.clearTimeout(reportNameTimerRef.current);
}, []);


  //   {
  //     id: 'security-incident-2024-001',
  //     title: 'Security Incident 2024-001',
  //     status: 'draft',
  //     lastModified: '2 hours ago'
  //   },
  //   {
  //     id: 'malware-analysis-report',
  //     title: 'Malware Analysis Report',
  //     status: 'review',
  //     lastModified: '1 day ago'
  //   },
  //   {
  //     id: 'network-forensics',
  //     title: 'Network Forensics',
  //     status: 'review',
  //     lastModified: '1 day ago'
  //   },
  //   {
  //     id: 'endpoint-investigation',
  //     title: 'Endpoint Investigation',
  //     status: 'published',
  //     lastModified: '3 days ago'
  //   }
  // ];

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





const startEditingTitle = useCallback((section: ReportSection) => {
  setEditingTitleSectionId(section.id);
  setTempSectionTitle(section.title || "");
  setTitleDirty(false);
  setTitleSaving("idle");
}, []);

const cancelEditingTitle = useCallback(() => {
  setEditingTitleSectionId(null);
  setTempSectionTitle("");
  setTitleDirty(false);
  setTitleSaving("idle");
}, []);

const commitEditingTitle = useCallback(async () => {
  if (!editingTitleSectionId) return;
  const rid = reportId ?? report?.id;
  const newTitle = tempSectionTitle.trim();
  const currentTitle = sections.find(s => s.id === editingTitleSectionId)?.title ?? "";

  // no-op / empty
  if (newTitle === "" || newTitle === currentTitle) {
    setEditingTitleSectionId(null);
    setTempSectionTitle("");
    setTitleDirty(false);
    return;
  }

  // optimistic
  setSections(prev => prev.map(s => s.id === editingTitleSectionId ? { ...s, title: newTitle } : s));
  setTitleSaving("saving");

  try {
    // skip API while it's a local section
    if (!isLocalSection(editingTitleSectionId) && rid) {
      await putSectionTitle(String(rid), editingTitleSectionId, newTitle);
    }
    setTitleSaving("saved");
  } catch (e) {
    console.error("Title save failed", e);
    // rollback
    setSections(prev => prev.map(s => s.id === editingTitleSectionId ? { ...s, title: currentTitle } : s));
    setTitleSaving("error");
  } finally {
    setEditingTitleSectionId(null);
    setTempSectionTitle("");
    setTitleDirty(false);
    setTimeout(() => setTitleSaving("idle"), 1000);
  }
}, [editingTitleSectionId, tempSectionTitle, reportId, report, sections]);



  return (
    <div className="min-h-screen bg-background flex">
      
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
          min-height: 500px !important;
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
        
        /* Fix dropdown text colors */
        .ql-snow .ql-picker-label {
          color: #9ca3af !important;
        }

        .ql-snow .ql-picker-label:hover,
        .ql-snow .ql-picker-label.ql-active {
          color: #e5e7eb !important;
        }

        .ql-snow .ql-picker-item {
          color: #e5e7eb !important;
        }

        .ql-snow .ql-picker-item:hover {
          color: #ffffff !important;
          background-color: #4b5563 !important;
        }

        /* Fix header dropdown specifically */
        .ql-snow .ql-picker.ql-header .ql-picker-label {
          color: #9ca3af !important;
        }

        .ql-snow .ql-picker.ql-header .ql-picker-label:hover,
        .ql-snow .ql-picker.ql-header .ql-picker-label.ql-active {
          color: #e5e7eb !important;
        }

        .ql-snow .ql-picker.ql-header .ql-picker-item {
          color: #e5e7eb !important;
        }

        /* Color picker fixes - UPDATED */
        .ql-snow .ql-color-picker .ql-picker-options,
        .ql-snow .ql-background-picker .ql-picker-options {
          background-color: #374151 !important; /* Changed from white to dark gray */
          border: 1px solid #4b5563 !important;
          padding: 3px 5px !important;
          width: 152px !important;
        }

        .ql-snow .ql-color-picker .ql-picker-item,
        .ql-snow .ql-background-picker .ql-picker-item {
          border: 1px solid #4b5563 !important; /* Darker border for dark mode */
          float: left !important;
          height: 16px !important;
          margin: 2px !important;
          width: 16px !important;
          padding: 0 !important;
        }

        .ql-snow .ql-color-picker .ql-picker-item:hover,
        .ql-snow .ql-background-picker .ql-picker-item:hover {
          border-color: #9ca3af !important; /* Lighter border on hover for visibility */
        }

        /* Keep all pickers dark consistently */
        .ql-snow .ql-picker .ql-picker-options {
          background-color: #374151 !important;
          border: 1px solid #4b5563 !important;
        }

        .ql-snow .ql-picker .ql-picker-item {
          color: #e5e7eb !important;
        }

        /* Specific fix for font and size pickers */
        .ql-snow .ql-font-picker .ql-picker-options,
        .ql-snow .ql-size-picker .ql-picker-options {
          background-color: #374151 !important;
          border: 1px solid #4b5563 !important;
        }
      `}</style>

   
      {/* Main Content */}
      <div className="w-full flex">
        {/* Report Sections Navigation */}
        <div className="w-80 bg-background border-r border p-4">
        <div className="mb-6">
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={reportTitle}
              onChange={(e) => {
                setReportTitle(e.target.value);
                scheduleReportNameSave(e.target.value);
              }}
              onBlur={() => flushReportNameNow()}
              className="w-full bg-background text-foreground font-semibold text-lg border-none outline-none"
              placeholder="Untitled report"
            />
            {/* status dot */}
            {reportNameState === "saving" && <Loader2 className="w-4 h-4 animate-spin text-gray-300" />}
            {reportNameState === "saved" && !reportTitleDirty && <CheckCircle className="w-4 h-4 text-emerald-500" />}
            {reportNameState === "error" && <XCircle className="w-4 h-4 text-red-500" />}
          </div>

          <div className="flex items-center gap-2 mt-2 text-sm text-foreground/60">
            <Calendar className="w-4 h-4" />
            <span>{incidentId}</span>
          </div>
          <div className="flex items-center gap-2 mt-1 text-sm text-foreground/60">
            <Clock className="w-4 h-4" />
            <span>Last modified:  {lastModified ? formatIsoDateTime(lastModified) : "—"} </span>
          </div>
          <div className="flex items-center gap-2 mt-1 text-sm text-foreground/60">
          </div>
        </div>


  {/* Section List */}
  <div className="mb-3">
  {adding ? (
    <div className="flex items-center gap-2">
      <input
        autoFocus
        value={newSectionTitle}
        onChange={e => setNewSectionTitle(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") { e.preventDefault(); handleAddSection(); }
          if (e.key === "Escape") { e.preventDefault(); setAdding(false); setNewSectionTitle(""); }
        }}
        className="w-full bg-background text-foreground border rounded px-3 py-2 focus:outline-none focus:border-primary"
        placeholder="New section title"
      />
    <button
      type="button"
      onClick={handleAddSection}
      disabled={addingBusy}
      className="px-3 py-2 bg-primary text-white rounded hover:bg-primary/60 disabled:opacity-50"
    >
      Add
    </button>

      <button
        type="button"
        onClick={() => { setAdding(false); setNewSectionTitle(""); }}
        className="px-3 py-2 bg-gray-700 text-gray-200 rounded hover:bg-gray-600"
      >
        Cancel
      </button>
    </div>
  ) : (
    <button
      type="button"
      onClick={() => setAdding(true)}
      className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-primary text-white rounded hover:bg-primary/60"
    >
      <Plus className="w-4 h-4" />
      Add Section
    </button>
  )}
</div>

 <DndContext
  sensors={sensors}
  collisionDetection={closestCenter}
  modifiers={[restrictToVerticalAxis]}
  onDragEnd={handleDragEnd}
>
  <SortableContext items={sections.map((s) => s.id)} strategy={verticalListSortingStrategy}>
    <div className="space-y-1">
      {sections.map((section, index) => (
        <SortableSectionItem
          key={section.id}
          section={section}
          index={index}
          activeIndex={activeSection}
          deletingId={deletingId}
          saveState={saveState}
          onSelect={async (idx) => {
            if (idx === activeSection) return;
            if (dirty || saveState === "saving") {
              await flushSaveNow(sections[activeSection]?.content ?? "", { force: true });
            }
            setActiveSection(idx);
          }}
          onDelete={(id) => handleDeleteSection(id)}
        />
      ))}
    </div>
  </SortableContext>
</DndContext>


</div>


        {/* Editor */}
        <div className="flex-1 flex flex-col">
          {/* Editor Header */}
          <div className="bg-background border-b border p-4">
            <div className="flex items-center justify-between">
              <div>

                <div className="flex items-center gap-4 text-sm text-foreground mt-1">
                  <span className="flex items-center gap-1">
                    <Eye className="w-4 h-4" />
                    Export
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-3">
                {isDFIRAdmin && (
                <button
                  type="button"
                  onClick={flushAndDownload}
                  disabled={!reportId && !report}
                  className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/60 transition-colors"
                >
                  <Download className="w-4 h-4" />
                  Export
                </button>
              )}

              </div>
            </div>
          </div>

          {/* Editor Content */}
          <div className="flex-1 p-6 overflow-y-auto">
            <div className="max-w-5xl mx-auto">
              {/* Report Header */}
              <div className="mb-8">
                <h1 className="text-3xl font-bold text-foreground/80 mb-4">
                  {reportTitle}
                </h1>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-foreground">Incident ID:</span>
                    <span className="text-foreground/60 ml-2">{caseId}</span>
                  </div>
                  <div>
                    <span className="text-foreground">Last modified:</span>
                    <span className="text-foreground/60 ml-2">{lastModified ? formatIsoDateTime(lastModified) : "—"} </span>
                  </div>
                  {/* <div>
                    <span className="text-foreground">Analyst:</span>
                    <span className="text-foreground/60 ml-2">{analyst}</span>
                  </div> */}
                  <div>
                    <span className="text-foreground">Report Type:</span>
                    <span className="text-foreground/60 ml-2">{reportType}</span>
                  </div>
                </div>
              </div>

              {/* Current Section */}
            {sections && sections[activeSection] && (
      <div className="mb-6">

        {editingTitleSectionId === sections[activeSection].id ? (
          <div className="flex items-center gap-2">
            <input
              autoFocus
              value={tempSectionTitle}
              onChange={(e) => { setTempSectionTitle(e.target.value); setTitleDirty(true); }}
              onKeyDown={(e) => {
                if (e.key === "Enter") { e.preventDefault(); commitEditingTitle(); }
                if (e.key === "Escape") { e.preventDefault(); cancelEditingTitle(); }
              }}
              className="bg-background border border-border rounded-lg px-3 py-2 text-foreground/60 w-full max-w-xl focus:outline-none focus:border-primary"
              placeholder="Section title"
            />
            <button
              type="button"
              onClick={commitEditingTitle}
              disabled={!tempSectionTitle.trim()}
              className="p-2 rounded-lg bg-green-600 hover:bg-green-700 text-white disabled:opacity-50"
              title="Save"
            >
              <Check className="w-4 h-4" />
            </button>
            <button
              type="button"
              onClick={cancelEditingTitle}
              className="p-2 rounded-lg bg-gray-700 hover:bg-gray-600 text-white"
              title="Cancel"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        ) : (
          <div className="flex items-center gap-3">
      <h2
        className="text-2xl font-semibold text-foreground/80 cursor-pointer select-none"
        onDoubleClick={() => startEditingTitle(sections[activeSection])} // optional: dbl-click to rename
        title="Double-click to rename"
      >
        {sections[activeSection].title}
      </h2>

      {/* rename pencil opens the header editor */}
      <button
        type="button"
        disabled={saveState === "saving" || editingTitleSectionId !== null}
        onClick={() => startEditingTitle(sections[activeSection])}
        className="p-2 rounded-lg hover:bg-foreground/10 text-foreground/60 disabled:opacity-50"
        aria-label="Rename section"
        title={saveState === "saving" ? "Saving… please wait" : "Rename section"}
      >
        <Pencil className="w-4 h-4" />
      </button>
      {/* AI Suggest Button */}
      <button
        type="button"
        onClick={async () => {
          if (!report?.mongoId) return;
          setAiLoading(true);
          try {
            const token = sessionStorage.getItem("authToken");
            // Send section context in POST payload for context-aware suggestions
            const contextPayload = {
              case_info: sectionContext?.case_info,
              iocs: sectionContext?.iocs,
              evidence: sectionContext?.evidence,
              timeline: sectionContext?.timeline,
              section_title: sections[activeSection]?.title,
              section_content: sections[activeSection]?.content,
            };
            const res = await axios.post(
              `${API_URL}/reports/ai/${report.mongoId}/sections/${sections[activeSection].id}/suggest`,
              contextPayload,
              { headers: { Authorization: `Bearer ${token}` } }
            );
            // Defensive: extract string, fallback to empty string
            const suggestionObj = res.data as any;
            let suggestionText = "";
            // Handle nested object structure: { suggestion: { SuggestionText: "..." } }
            if (suggestionObj && typeof suggestionObj === "object") {
              if (typeof suggestionObj.SuggestionText === "string") {
                suggestionText = suggestionObj.SuggestionText;
              } else if (
                suggestionObj.suggestion &&
                typeof suggestionObj.suggestion === "object" &&
                typeof suggestionObj.suggestion.SuggestionText === "string"
              ) {
                suggestionText = suggestionObj.suggestion.SuggestionText;
              } else if (typeof suggestionObj.suggestion === "string") {
                suggestionText = suggestionObj.suggestion;
              } else {
                console.warn("AI suggestion object missing expected string property:", suggestionObj);
                suggestionText = "";
              }
            } else if (typeof suggestionObj === "string") {
              suggestionText = suggestionObj;
            }
            console.log("AI suggestion received:", suggestionText); // Debug log
            if (!insertingSuggestionRef.current) {
              setAiSuggestion(suggestionText);
            }
            // Removed setShowFeedback(true) to prevent duplicate overlay and ghost text persistence
          } catch (e) {
            console.error("AI suggest failed", e);
            toast.error("Failed to fetch AI suggestion");
          } finally {
            setAiLoading(false);
          }
        }}
        className="flex items-center gap-2 px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white rounded"
      >
        <Sparkles className="w-4 h-4" />
        Suggest
      </button>
      
      {/* Inline context panel (not overlay) */}
      <div
        className="w-[340px] max-h-[80vh] overflow-y-auto bg-gray-900 border border-gray-700 rounded-xl shadow-xl ml-6 my-4"
        style={{
    position: "absolute", top: "210px", right: "80px", zIndex: 40,
  }}
      >
        <button
          className="w-full flex items-center justify-between px-4 py-2 bg-gray-800 border-b border-gray-700 rounded-t-xl hover:bg-gray-700 transition"
          onClick={() => setContextOpen((v) => !v)}
        >
          <span className="font-semibold text-white text-lg">Section Context</span>
          {contextOpen ? <ChevronUp className="w-5 h-5 text-gray-400" /> : <ChevronDown className="w-5 h-5 text-gray-400" />}
        </button>
        {contextOpen && sectionContext && (
          <div className="p-4 space-y-4">
            {/* Case Info (draggable) */}
            <div
              draggable
              onDragStart={e => {
                e.dataTransfer.setData("application/json", JSON.stringify(sectionContext.case_info));
                e.dataTransfer.effectAllowed = "copy";
              }}
              className="cursor-grab select-none"
              style={{ pointerEvents: 'auto', userSelect: 'none' }}
            >
              <div className="flex items-center gap-2 mb-1">
                <Shield className="w-5 h-5 text-blue-400" />
                <span className="font-bold text-white">Case Info</span>
              </div>
              {sectionContext.case_info && typeof sectionContext.case_info === 'object' ? (
                <ul className="text-xs text-gray-300 bg-gray-800 rounded p-2 overflow-x-auto">
                  {Object.entries(sectionContext.case_info).map(([key, value]) => (
                    <li key={key}><span className="font-bold text-gray-400">{key}:</span> {String(value)}</li>
                  ))}
                </ul>
              ) : (
                <pre className="text-xs text-gray-300 bg-gray-800 rounded p-2 overflow-x-auto">{JSON.stringify(sectionContext.case_info, null, 2)}</pre>
              )}
              <div className="text-xs text-blue-300 mt-2">Drag to section editor to auto-fill description</div>
            </div>
            {/* IOCs (draggable) */}
            <div
              draggable
              onDragStart={e => {
                e.dataTransfer.setData("application/json", JSON.stringify(sectionContext.iocs));
                e.dataTransfer.effectAllowed = "copy";
              }}
              className="cursor-grab select-none"
              style={{ pointerEvents: 'auto', userSelect: 'none' }}
            >
              <div className="flex items-center gap-2 mb-1">
                <AlertCircle className="w-5 h-5 text-red-400" />
                <span className="font-bold text-white">IOCs</span>
              </div>
              <ul className="space-y-1">
                {(Array.isArray(sectionContext.iocs) ? sectionContext.iocs : []).map((ioc, i) => (
                  <li key={i} className="flex items-center gap-2 text-sm text-red-300">
                    <span className="bg-red-900/60 px-2 py-0.5 rounded font-mono">{ioc?.type || "IOC"}</span>
                    <span>{ioc?.value || (typeof ioc === 'string' ? ioc : JSON.stringify(ioc))}</span>
                  </li>
                ))}
              </ul>
              <div className="text-xs text-red-300 mt-2">Drag to section editor to auto-fill description</div>
            </div>
            {/* Evidence (draggable) */}
            <div
              className="cursor-grab select-none"
              style={{ pointerEvents: 'auto', userSelect: 'none' }}
            >
              <div className="flex items-center gap-2 mb-1">
                <FileText className="w-5 h-5 text-green-400" />
                <span className="font-bold text-white">Evidence</span>
              </div>
              <ul className="space-y-1">
                {(Array.isArray(sectionContext.evidence) ? sectionContext.evidence : []).map((ev, i) => {
                  // If evidence is a string, try to parse it
                  let evidenceObj = ev;
                  if (typeof ev === 'string') {
                    try {
                      evidenceObj = JSON.parse(ev);
                    } catch {
                      // fallback: skip rendering this evidence item
                      return null;
                    }
                  }
                  let sha512 = evidenceObj.sha512, sha256 = evidenceObj.sha256, uploader = evidenceObj.uploader;
                  if (evidenceObj.metadata) {
                    try {
                      const meta = typeof evidenceObj.metadata === 'string' ? JSON.parse(evidenceObj.metadata) : evidenceObj.metadata;
                      sha512 = meta.sha512 || sha512;
                      sha256 = meta.sha256 || sha256;
                      uploader = meta.uploader || uploader;
                    } catch {}
                  }
                  return (
                    <li
                      key={i}
                      className="bg-green-950/80 border border-green-800 rounded-xl shadow flex flex-col gap-2 p-3 mb-2"
                      draggable
                      onDragStart={e => {
                        e.dataTransfer.setData("application/json", JSON.stringify([ev]));
                        e.dataTransfer.effectAllowed = "copy";
                      }}
                    >
                      <div className="flex items-center gap-2 mb-1">
                        <FileText className="w-5 h-5 text-green-400" />
                        <span className="font-bold text-green-200 text-base">{evidenceObj.filename || evidenceObj.fileName || "Unknown file"}</span>
                      </div>
                      <div className="flex flex-wrap items-center gap-2 mb-1 overflow-x-auto">
                        <span
                          className="bg-gray-800 px-2 py-0.5 rounded font-mono text-xs text-gray-300 break-all max-w-full"
                          title={sha256}
                        >
                          SHA256: {sha256 ? sha256 : <span className='italic text-gray-500'>N/A</span>}
                        </span>
                        <span
                          className="bg-gray-800 px-2 py-0.5 rounded font-mono text-xs text-gray-300 break-all max-w-full"
                          title={sha512}
                        >
                          SHA512: {sha512 ? sha512 : <span className='italic text-gray-500'>N/A</span>}
                        </span>
                      </div>
                      {uploader && (
                        <div className="flex items-center gap-2 mt-1">
                          <span className="bg-green-900/60 px-2 py-0.5 rounded font-mono text-xs text-green-300"><User className="inline w-4 h-4 mr-1 text-green-400" /> Uploaded by: {uploader}</span>
                        </div>
                      )}
                    </li>
                  );
                })}
              </ul>
              <div className="text-xs text-green-300 mt-2">Drag to section editor to auto-fill description</div>
            </div>
            {/* Timeline (draggable) */}
            <div
              draggable
              onDragStart={e => {
                e.dataTransfer.setData("application/json", JSON.stringify(sectionContext.timeline));
                e.dataTransfer.effectAllowed = "copy";
              }}
              className="cursor-grab select-none"
              style={{ pointerEvents: 'auto', userSelect: 'none' }}
            >
              <div className="flex items-center gap-2 mb-1">
                <Clock className="w-5 h-5 text-yellow-400" />
                <span className="font-bold text-white">Timeline</span>
              </div>
              <ul className="space-y-1">
                {(Array.isArray(sectionContext.timeline) ? sectionContext.timeline : []).map((ev, i) => (
                  <li key={i} className="flex items-center gap-2 text-sm text-yellow-300">
                    <span className="bg-yellow-900/60 px-2 py-0.5 rounded font-mono">{ev?.createdAt || "Event"}</span>
                    <span>{ev?.description || (typeof ev === 'string' ? ev : JSON.stringify(ev))}</span>
                  </li>
                ))}
              </ul>
              <div className="text-xs text-yellow-300 mt-2">Drag to section editor to auto-fill description</div>
            </div>
          </div>
        )}
      </div>


      {titleSaving === "saving" && (
        <Loader2 className="w-4 h-4 text-gray-300 animate-spin" />
      )}
      {titleSaving === "error" && (
        <span className="text-sm text-red-400">Save failed</span>
      )}
      {titleSaving === "saved" && (
        <span className="text-sm text-emerald-400">Saved</span>
      )}
    </div>

        )}
      </div>
    )}
   
                  {/* React Quill Editor with AI ghost text */}
                  {sections && sections[activeSection] && (
                    <div className="mb-8" style={{ position: "relative" }}>
                      {/* Enhance Button for last dropped summary */}
                      {showEnhanceButton && lastDroppedSummary && (
                        <div className="mb-4 flex gap-2">
                          <button
                            type="button"
                            className="px-3 py-1 bg-blue-500 hover:bg-blue-600 text-white rounded flex items-center gap-2"
                            disabled={enhancing}
                            onClick={async () => {
                              setEnhancing(true);
                              try {
                                const token = sessionStorage.getItem("authToken");
                                // Gather context for the active section
                                const contextPayload = {
                                  text: lastDroppedSummary,
                                  case_info: sectionContext?.case_info,
                                  iocs: sectionContext?.iocs,
                                  evidence: sectionContext?.evidence,
                                  timeline: sectionContext?.timeline,
                                  section_title: sections[activeSection]?.title,
                                  section_content: sections[activeSection]?.content,
                                };
                                const res = await axios.post(
                                  `${API_URL}/reports/ai/enhance-summary`,
                                  contextPayload,
                                  { headers: { Authorization: `Bearer ${token}` } }
                                );
                                let enhancedText = "";
                                console.log("Enhance response:", res.data); // Debug log

                              if (res.data && typeof res.data === "object" && typeof (res.data as { enhanced?: string }).enhanced === "string") {
                                enhancedText = (res.data as { enhanced: string }).enhanced;
                              } else if (typeof res.data === "string") {
                                enhancedText = res.data;
                              } else {
                                toast.error("No enhanced summary received");
                              }
                                // Replace only the last dropped summary in the section content
                                setSections(prev => {
                                  const copy = [...prev];
                                  if (copy[activeSection]) {
                                    const prevContent = copy[activeSection].content || "";
                                    // Replace lastDroppedSummary with enhancedText (only last occurrence)
                                    const idx = prevContent.lastIndexOf(lastDroppedSummary);
                                    if (idx !== -1) {
                                      const before = prevContent.substring(0, idx);
                                      const after = prevContent.substring(idx + lastDroppedSummary.length);
                                      copy[activeSection] = {
                                        ...copy[activeSection],
                                        content: `${before}${enhancedText}${after}`,
                                      };
                                    }
                                  }
                                  return copy;
                                });
                                toast.success("Summary enhanced");
                                setShowEnhanceButton(false);
                              } catch (err) {
                                toast.error("Failed to enhance summary");
                              } finally {
                                setEnhancing(false);
                              }
                            }}
                          >
                            <Sparkles className="w-4 h-4" />
                            {enhancing ? "Enhancing..." : "Enhance summary"}
                          </button>
                        </div>
                      )}

{isPreviewMode ? (
  <div className="bg-gray-800 p-6 rounded-lg">
    {/* Word Document Simulation */}
    <div 
      className="mx-auto bg-gray-900 text-gray-100 shadow-2xl rounded overflow-hidden"
      style={{
        width: '210mm', // A4 width
        minHeight: '297mm', // A4 height
        boxShadow: '0 0 25px rgba(0,0,0,0.7)'
      }}
    >
      {/* Document Content Pages */}
      <div 
        className="p-12 bg-gray-900"
        style={{
          background: 'linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%)',
          minHeight: '297mm'
        }}
      >
        {/* Combined Sections */}
        <div className="space-y-8">
          {sections.length > 0 ? (
            sections.map((section, index) => (
              <div 
                key={section.id || index}
                className="section-word-style"
              >
                {/* Section Header */}
                {section.title && (
                  <h2 
                    className="text-blue-400 mb-4"
                    style={{
                      fontSize: '16pt',
                      fontWeight: '600',
                      fontFamily: 'Calibri, sans-serif',
                      marginBottom: '16px',
                      paddingBottom: '8px',
                      borderBottom: '2px solid #374151'
                    }}
                  >
                    {section.title}
                  </h2>
                )}
                
                {/* Section Content */}
                <div 
                  className="word-content-section"
                  style={{
                    fontFamily: 'Calibri, sans-serif',
                    fontSize: '11pt',
                    lineHeight: '1.5',
                    color: '#e5e7eb'
                  }}
                  dangerouslySetInnerHTML={{ 
                    __html: section.content || `
                      <div style="color: #6b7280; font-style: italic; text-align: center; padding: 32px; border: 1px dashed #4b5563; border-radius: 4px;">
                        <p style="margin: 0; font-size: 11pt;">No content added for this section</p>
                      </div>
                    `
                  }}
                />
                
                {/* Section Separator */}
                {index < sections.length - 1 && (
                  <div className="my-8 opacity-30">
                    <div className="w-full h-px bg-gray-600"></div>
                  </div>
                )}
              </div>
            ))
          ) : (
            <div className="text-center py-20 text-gray-400">
              <div className="text-4xl mb-4">📄</div>
              <h3 className="text-xl font-semibold mb-2" style={{ fontFamily: 'Calibri, sans-serif' }}>
                No Sections Added
              </h3>
              <p style={{ fontFamily: 'Calibri, sans-serif' }}>Add sections to start building your document</p>
            </div>
          )}
        </div>

        {/* Page Footer */}
        <div className="mt-16 pt-8 border-t border-gray-700 text-center">
          <div className="text-xs text-gray-500">
            Page 1 | {sections.reduce((total, section) => {
              const text = section.content?.replace(/<[^>]*>/g, '') || '';
              return total + (text.split(/\s+/).filter(word => word.length > 0).length);
            }, 0)} words
          </div>
        </div>
      </div>
    </div>

    {/* Word Document Styling */}
    <style>{`
      .word-content-section h1 {
        color: #ffffff;
        font-size: 16pt;
        font-weight: bold;
        margin: 20px 0 12px 0;
        font-family: Calibri, sans-serif;
      }

      .word-content-section h2 {
        color: #e5e7eb;
        font-size: 14pt;
        font-weight: bold;
        margin: 18px 0 10px 0;
        font-family: Calibri, sans-serif;
      }

      .word-content-section h3 {
        color: #e5e7eb;
        font-size: 12pt;
        font-weight: bold;
        margin: 16px 0 8px 0;
        font-family: Calibri, sans-serif;
      }

      .word-content-section p {
        margin-bottom: 12px;
        text-align: left;
        line-height: 1.5;
      }

      .word-content-section ul, 
      .word-content-section ol {
        margin: 12px 0;
        padding-left: 36px;
      }

      .word-content-section li {
        margin-bottom: 6px;
        line-height: 1.5;
      }

      .word-content-section strong {
        font-weight: bold;
      }

      .word-content-section em {
        font-style: italic;
      }

      .word-content-section u {
        text-decoration: underline;
      }

      .word-content-section s {
        text-decoration: line-through;
      }

      .word-content-section blockquote {
        border-left: 3px solid #2563eb;
        background: rgba(37, 99, 235, 0.1);
        padding: 12px 16px;
        margin: 16px 0;
        border-radius: 0 2px 2px 0;
      }

      .word-content-section table {
        width: 100%;
        border-collapse: collapse;
        margin: 16px 0;
        background: rgba(55, 65, 81, 0.5);
        border: 1px solid #4b5563;
      }

      .word-content-section th {
        background: #374151;
        color: #ffffff;
        padding: 8px 12px;
        text-align: left;
        font-weight: 600;
        border: 1px solid #4b5563;
      }

      .word-content-section td {
        padding: 8px 12px;
        border: 1px solid #4b5563;
      }

      @media print {
        .section-word-style {
          page-break-inside: avoid;
        }
      }
    `}</style>
  </div>
) : (
                        <div style={{ position: "relative" }}>
                          <div
                            style={{ position: "relative" }}
                            onDragOver={e => {
                              if (e.dataTransfer.types.includes("application/json")) {
                                e.preventDefault();
                                e.dataTransfer.dropEffect = "copy";
                              }
                            }}
                            onDrop={e => {
                              const data = e.dataTransfer.getData("application/json");
                              if (data) {
                                try {
                                  const parsed = JSON.parse(data);
                                  let description = "";
                                  // Helper: humanize Case Info
                                  const humanizeCaseInfo = (info: any) => {
                                    if (!info || typeof info !== "object") return "";
                                    let lines = [];
                                    if (info.title) lines.push(`Case Title: ${info.title}`);
                                    if (info.description) lines.push(`Description: ${info.description}`);
                                    if (info.priority) lines.push(`Priority: ${info.priority}`);
                                    if (info.status) lines.push(`Status: ${info.status}`);
                                    if (info.investigation_stage) lines.push(`Stage: ${info.investigation_stage}`);
                                    if (info.team_name) lines.push(`Team: ${info.team_name}`);
                                    if (info.created_at) lines.push(`Created: ${new Date(info.created_at).toLocaleString()}`);
                                    if (info.updated_at) lines.push(`Last Updated: ${new Date(info.updated_at).toLocaleString()}`);
                                    if (info.report_name) lines.push(`Report Name: ${info.report_name}`);
                                    if (info.report_status) lines.push(`Report Status: ${info.report_status}`);
                                    if (info.report_created_at) lines.push(`Report Created: ${new Date(info.report_created_at).toLocaleString()}`);
                                    if (info.report_updated_at) lines.push(`Report Updated: ${new Date(info.report_updated_at).toLocaleString()}`);
                                    if (info.created_by) lines.push(`Created By: ${info.created_by}`);
                                    if (info.examiner_id) lines.push(`Examiner: ${info.examiner_id}`);
                                    if (info.team_id) lines.push(`Team ID: ${info.team_id}`);
                                    if (info.tenant_id) lines.push(`Tenant ID: ${info.tenant_id}`);
                                    if (info.id) lines.push(`Case ID: ${info.id}`);
                                    return lines.join("\n");
                                  };
                                  if (Array.isArray(parsed)) {
                                    // Evidence, IOCs, Timeline
                                    if (parsed.length && parsed[0] && typeof parsed[0] === "object") {
                                      // Evidence: look for filename, hashes
                                      if (parsed[0].filename || parsed[0].fileName) {
                                        description = parsed.map((ev) => {
                                          let evidenceObj = ev;
                                          if (typeof evidenceObj === 'string') {
                                            try { evidenceObj = JSON.parse(evidenceObj); } catch {}
                                          }
                                          let sha512 = evidenceObj.sha512, sha256 = evidenceObj.sha256, uploader = evidenceObj.uploader;
                                          if (evidenceObj.metadata) {
                                            try {
                                              const meta = typeof evidenceObj.metadata === 'string' ? JSON.parse(evidenceObj.metadata) : evidenceObj.metadata;
                                              sha512 = meta.sha512 || sha512;
                                              sha256 = meta.sha256 || sha256;
                                              uploader = meta.uploader || uploader;
                                            } catch {}
                                          }
                                          return `Evidence: ${evidenceObj.filename || evidenceObj.fileName || "Unknown file"}${uploader ? ` (uploaded by ${uploader})` : ""}\nHashes: SHA256 ${sha256 || "N/A"}, SHA512 ${sha512 || "N/A"}`;
                                        }).join("\n\n");
                                        toast.success("Evidence added to section");
                                      } else if (parsed[0].type || parsed[0].value) {
                                        // IOCs
                                        description = parsed.map((ioc, idx) => {
                                          return `Indicator ${idx + 1}: ${ioc.type || "Type"} - ${ioc.value || (typeof ioc === 'string' ? ioc : JSON.stringify(ioc))}`;
                                        }).join("\n");
                                        toast.success("IOCs added to section");
                                      } else if (parsed[0].createdAt || parsed[0].description) {
                                        // Timeline
                                        description = parsed.map((ev, idx) => {
                                          return `Timeline Event ${idx + 1}: ${ev.description || (typeof ev === 'string' ? ev : JSON.stringify(ev))}${ev.createdAt ? ` (at ${new Date(ev.createdAt).toLocaleString()})` : ""}`;
                                        }).join("\n");
                                        toast.success("Timeline added to section");
                                      } else {
                                        // Fallback for unknown array
                                        description = parsed.map((item, idx) => `Item ${idx + 1}: ${typeof item === 'object' ? JSON.stringify(item) : String(item)}`).join("\n");
                                        toast.success("Items added to section");
                                      }
                                    } else {
                                      // Array of primitives
                                      description = parsed.map((item, idx) => `Item ${idx + 1}: ${String(item)}`).join("\n");
                                      toast.success("Items added to section");
                                    }
                                  } else if (parsed && typeof parsed === "object") {
                                    // Case Info
                                    description = humanizeCaseInfo(parsed);
                                    toast.success("Section description auto-filled from Case Info");
                                  } else {
                                    // Fallback
                                    description = String(parsed);
                                    toast.success("Content added to section");
                                  }
                                  setSections(prev => {
                                    const copy = [...prev];
                                    if (copy[activeSection]) {
                                      // Append to existing content, not overwrite
                                      const prevContent = copy[activeSection].content || "";
                                      copy[activeSection] = {
                                        ...copy[activeSection],
                                        content: prevContent ? `${prevContent}\n\n${description}` : description,
                                      };
                                    }
                                    return copy;
                                  });
                                  setLastDroppedSummary(description);
                                  setShowEnhanceButton(true);
                                } catch (err) {
                                  toast.error("Failed to parse dropped card data");
                                }
                              }
                            }}
                          >
                          <div style={{ position: "relative" }}
                            onDragOver={e => {
                              if (e.dataTransfer.types.includes("application/json")) {
                                e.preventDefault();
                                e.dataTransfer.dropEffect = "copy";
                              }
                            }}
                            onDrop={e => {
                              const data = e.dataTransfer.getData("application/json");
                              if (data) {
                                try {
                                  // ...existing code...
                                } catch (err) {
                                  toast.error("Failed to parse dropped card data");
                                }
                              }
                            }}>
                            <ReactQuill
                              theme="snow"
                              value={sections[activeSection]?.content ?? ""}
                              onChange={handleEditorChange}
                              modules={modules}
                              formats={formats}
                              placeholder="Start writing your report content here..."
                              ref={quillRef}
                              onKeyDown={e => {
                                if (aiSuggestion && !aiLoading && e.key === "Tab") {
                                  e.preventDefault();
                                  if (!insertingSuggestionRef.current) {
                                    insertingSuggestionRef.current = true;
                                    setSections(prev => {
                                      const copy = [...prev];
                                      if (copy[activeSection]) {
                                        copy[activeSection] = {
                                          ...copy[activeSection],
                                          content: (copy[activeSection].content || "") + aiSuggestion,
                                        };
                                      }
                                      return copy;
                                    });
                                    setAiSuggestion("");
                                  }
                                }
                              }}
                            />
                          </div>
                          {/* AI suggestion box always visible below editor */}
                          {(aiSuggestion && !aiLoading) && (
                            <div className="mt-2 p-3 bg-blue-50 border border-blue-300 rounded text-blue-900 flex items-center justify-between">
                              <span className="font-mono text-sm">{aiSuggestion}</span>
                              <button
                                type="button"
                                className="ml-4 px-2 py-1 bg-blue-600 text-white rounded hover:bg-blue-700"
                                onClick={() => {
                                  setSections(prev => {
                                    const copy = [...prev];
                                    if (copy[activeSection]) {
                                      copy[activeSection] = {
                                        ...copy[activeSection],
                                        content: (copy[activeSection].content || "") + aiSuggestion,
                                      };
                                    }
                                    return copy;
                                  });
                                  setAiSuggestion("");
                                }}
                              >
                                Insert Suggestion
                              </button>
                              <span className="ml-2 text-xs text-blue-700">(or press Tab)</span>
                            </div>
                          )}
                          </div>
                        </div>
                      )}
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
      <span className="text-emerald-400">
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

<div className="flex items-center gap-2">
  <label className="text-foreground text-sm">Status:</label>
  <select
    value={reportType}  // you mapped `status` into `reportType`
    onChange={async (e) => {
      const newStatus = e.target.value as "draft" | "review" | "published";
      if (!reportId) return;
      try {
        await putReportStatus(reportId, newStatus);
        setReportType(newStatus); // update UI
        toast.success(`Report status updated to ${newStatus}`);
      } catch (err) {
        console.error("Status update failed", err);
        toast.error("Failed to update status");
      }
    }}
    className="bg-background text-foreground px-2 py-1 rounded border border-border focus:outline-none focus:border-primary"
  >
    <option value="draft">Draft</option>
    <option value="review">Review</option>
    <option value="published">Published</option>
  </select>
</div>


                  {/* <button className="px-4 py-2 bg-gray-700 text-gray-300 rounded-lg hover:bg-gray-600 transition-colors">
                    <Eye className="w-4 h-4 inline mr-2" />
                    Preview
                  </button> */}

                  <button 
                  onClick={togglePreview}
                  className={`px-4 py-2 rounded-lg transition-colors ${
                    isPreviewMode 
                      ? 'bg-primary text-white hover:bg-primary/80' 
                      : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                  }`}
                >
                  <Eye className="w-4 h-4 inline mr-2" />
                  {isPreviewMode ? 'Edit' : 'Preview'}
                </button>
                </div>
                
                {/* <div className="flex items-center gap-2 text-sm text-gray-400">
                  <Clock className="w-4 h-4" />
                  <span>Auto-saved 30 seconds ago</span>
                </div> */}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
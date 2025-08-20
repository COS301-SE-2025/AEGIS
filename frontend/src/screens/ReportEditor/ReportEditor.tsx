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
import { useNavigate } from "react-router-dom"; 
import { Pencil, Check, X,Trash2 } from "lucide-react"; // NEW
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
import { GripVertical } from "lucide-react";
import { toast } from "react-hot-toast";
import { AlertTriangle } from "lucide-react";
import { createRoot } from "react-dom/client";

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
const API_URL = "http://localhost:8080/api/v1";


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
          ? "bg-blue-600 text-white"
          : "hover:bg-gray-700 text-gray-300"
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
          className="p-1 -ml-1 mr-1 rounded cursor-grab active:cursor-grabbing hover:bg-gray-700"
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
          className={`p-1 rounded hover:bg-gray-700 ${
            activeIndex === index ? "text-white" : "text-gray-400"
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
// --- title edit state ---
const [editingTitleSectionId, setEditingTitleSectionId] = useState<string|null>(null);
const [tempSectionTitle, setTempSectionTitle] = useState("");
const [titleDirty, setTitleDirty] = useState(false);
const [titleSaving, setTitleSaving] = useState<"idle"|"saving"|"saved"|"error">("idle");
const [addingBusy, setAddingBusy] = useState(false);
const [deletingId, setDeletingId] = useState<string | null>(null);
const [reportNameState, setReportNameState] = useState<"idle" | "saving" | "saved" | "error">("idle");
const [reportTitleDirty, setReportTitleDirty] = useState(false);
const reportNameTimerRef = useRef<number | null>(null);
const lastReportNameSavedRef = useRef<string>("");
const [caseId, setCaseId] = useState("");
const [isPreviewMode, setIsPreviewMode] = useState(false);

// local-only section helpers (so we can skip API calls until backend is wired)
const makeLocalId = () =>
  `local-${(crypto as any)?.randomUUID?.() ?? Date.now().toString(36)}`;
const isLocalSection = (id: string) => id.startsWith("local-");

// add/delete UI state
const [adding, setAdding] = useState(false);
const [newSectionTitle, setNewSectionTitle] = useState("");


const navigate = useNavigate(); // NEW

const [recentReports, setRecentReports] = useState<RecentReport[]>([]); // NEW
const [recentLoading, setRecentLoading] = useState(true);               // NEW
const [recentError, setRecentError] = useState<string | null>(null);    // NEW
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

  //     setReport(data);
  //     setSections(Array.isArray(data.content) ? data.content : []);
  //     setReportTitle(data.name || "");
  //     setIncidentId(data.incidentId || "");
  //     setDateCreated(data.dateCreated || "");
  //     setAnalyst(data.analyst || "");
  //     setReportType(data.type || "");
  //   })().catch(err => console.error("Error fetching report:", err));
  // }, [reportId]);
const loadReport = useCallback(async (id: string) => {
  const token = sessionStorage.getItem("authToken");
  if (!token) return;

  // Read raw payload; backend may send either { metadata, content } or a flat object
  const { data } = await axios.get<any>(`${API_URL}/reports/${id}`, {
    headers: { Authorization: `Bearer ${token}` },
  });

  const meta = data?.metadata ?? data;

  // Sections from either data.content or data.sections
  const rawSections: any[] = Array.isArray(data?.content)
    ? data.content
    : Array.isArray(data?.sections)
    ? data.sections
    : [];

  const mappedSections: ReportSection[] = rawSections.map((s) => ({
    id: String(s.id),
    title: String(s.title ?? ""),
    content: String(s.content ?? ""),
    completed: !!s.completed,
  }));

  // Build a UI-friendly report (keeps your existing interface)
  const uiReport: Report = {
    id: String(meta?.id ?? ""),
    name: String(meta?.name ?? ""),                          // report name
    type: String(meta?.status ?? meta?.type ?? ""),         // often status
    content: mappedSections,
    incidentId: String(meta?.report_number ?? ""),          // what you show as “Incident ID”
    dateCreated: String(
      meta?.created_at ?? meta?.createdAt ?? meta?.date_created ?? ""
    ),
    analyst: String(meta?.author ?? meta?.analyst ?? ""),   // try author first
    case_id: String(meta?.case_id ?? ""),
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
}, []);

// put this near the top of your component file (or in a utils file)
const formatIsoDate = (
  iso?: string,
  opts: { locale?: string; utc?: boolean } = {}
) => {
  if (!iso) return "";
  // handle your sentinel "no date" value
  if (iso.startsWith("0001-01-01")) return "";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso; // fallback if server sends odd value
  const { locale = "en-GB", utc = false } = opts;
  return new Intl.DateTimeFormat(locale, {
    day: "numeric",
    month: "long",
    year: "numeric",
    ...(utc ? { timeZone: "UTC" } : {}),
  }).format(d);
};

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

const refreshRecent = async () => { // NEW (optional)
  try {
    setRecentLoading(true);
    setRecentError(null);
    const token = sessionStorage.getItem("authToken");
    if (!token) return;
    const res = await axios.get(`${API_URL}/reports/recent?limit=6&mine=true`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const items = (res.data as any[]).map(x => ({
      id: x.id, title: x.title, status: x.status as RecentReport["status"], lastModified: x.lastModified
    })) as RecentReport[];
    setRecentReports(items);
  } catch {
    setRecentError("Failed to refresh");
  } finally {
    setRecentLoading(false);
  }
};

const openReport = async (id: string) => {
  try {
    // only flush if there are pending edits
    if (dirty || saveState === "saving") {
      await flushSaveNow(sections[activeSection]?.content ?? ""); // no { force: true }
    }
    navigate(`/report-editor/${id}`);
  } catch (e) {
    console.error("Failed to save before navigation", e);
    setError("Couldn't save changes before opening another report.");
  }
};

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

async function putSectionsOrderBulk(
  reportId: string,
  order: { id: string; order: number }[]
) {
  const token = sessionStorage.getItem("authToken");
  await axios.put(
    `${API_URL}/reports/${reportId}/sections/reorder`,
    { order },
    { headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } }
  );
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

   const isLoading = !report;



  const getStatusDot = (status: string) => {
    switch (status) {
      case 'draft': return 'bg-gray-400';
      case 'review': return 'bg-yellow-400';
      case 'published': return 'bg-green-400';
      default: return 'bg-gray-400';
    }
  };

  function timeAgo(iso: string) { // NEW
  const d = new Date(iso);
  const s = Math.floor((Date.now() - d.getTime()) / 1000);
  if (s < 60) return "just now";
  const m = Math.floor(s / 60);
  if (m < 60) return `${m} min${m > 1 ? "s" : ""} ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h} hour${h > 1 ? "s" : ""} ago`;
  const days = Math.floor(h / 24);
  if (days < 7) return `${days} day${days > 1 ? "s" : ""} ago`;
  return d.toLocaleString(); // fallback for older items
}
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

        .ql-snow .ql-picker-options .ql-picker-item {
          color: #e5e7eb !important;
          background-color: transparent !important;
        }

        .ql-snow .ql-picker-options .ql-picker-item:hover {
          color: #ffffff !important;
          background-color: #4b5563 !important;
        }
      `}</style>

   
      {/* Main Content */}
      <div className="w-full flex">
        {/* Report Sections Navigation */}
        <div className="w-80 bg-gray-850 border-r border-gray-700 p-4">
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
              className="w-full bg-transparent text-white font-semibold text-lg border-none outline-none"
              placeholder="Untitled report"
            />
            {/* status dot */}
            {reportNameState === "saving" && <Loader2 className="w-4 h-4 animate-spin text-gray-300" />}
            {reportNameState === "saved" && !reportTitleDirty && <CheckCircle className="w-4 h-4 text-emerald-500" />}
            {reportNameState === "error" && <XCircle className="w-4 h-4 text-red-500" />}
          </div>

          <div className="flex items-center gap-2 mt-2 text-sm text-gray-400">
            <Calendar className="w-4 h-4" />
            <span>{incidentId}</span>
          </div>
          <div className="flex items-center gap-2 mt-1 text-sm text-gray-400">
            <Clock className="w-4 h-4" />
            <span>Last modified:  {lastModified ? formatIsoDateTime(lastModified) : "—"} </span>
          </div>
          <div className="flex items-center gap-2 mt-1 text-sm text-gray-400">
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
        className="w-full bg-gray-800 text-white border border-gray-700 rounded px-3 py-2"
        placeholder="New section title"
      />
    <button
      type="button"
      onClick={handleAddSection}
      disabled={addingBusy}
      className="px-3 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
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
      className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-gray-700 text-gray-200 rounded hover:bg-gray-600"
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
          <div className="bg-gray-800 border-b border-gray-700 p-4">
            <div className="flex items-center justify-between">
              <div>

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
          <div className="flex-1 p-6 overflow-y-auto">
            <div className="max-w-5xl mx-auto">
              {/* Report Header */}
              <div className="mb-8">
                <h1 className="text-3xl font-bold text-white mb-4">
                  {reportTitle}
                </h1>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-400">Incident ID:</span>
                    <span className="text-white ml-2">{caseId}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">Last modified:</span>
                    <span className="text-white ml-2">{lastModified ? formatIsoDateTime(lastModified) : "—"} </span>
                  </div>
                  {/* <div>
                    <span className="text-gray-400">Analyst:</span>
                    <span className="text-white ml-2">{analyst}</span>
                  </div> */}
                  <div>
                    <span className="text-gray-400">Report Type:</span>
                    <span className="text-white ml-2">{reportType}</span>
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
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white w-full max-w-xl"
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
        className="text-2xl font-semibold text-white"
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
        className="p-2 rounded-lg hover:bg-gray-700 text-gray-300 disabled:opacity-50"
        aria-label="Rename section"
        title={saveState === "saving" ? "Saving… please wait" : "Rename section"}
      >
        <Pencil className="w-4 h-4" />
      </button>

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



                  {/* React Quill Editor */}
                  {sections && sections[activeSection] && (
              <div className="mb-8">
                {isPreviewMode ? (
                <div 
                  className="prose prose-invert max-w-none bg-gray-800 p-6 rounded-lg border border-gray-700"
                  dangerouslySetInnerHTML={{ __html: sections[activeSection]?.content || '<p class="text-gray-400 italic">No content yet...</p>' }}
                  style={{ minHeight: '500px', color: '#e5e7eb', lineHeight: '1.6' }}
                />
              ) : (
                <ReactQuill
                  theme="snow"
                  value={sections[activeSection]?.content?? ""}
                  onChange={handleEditorChange}
                  modules={modules}
                  formats={formats}
                  placeholder="Start writing your report content here..."
                />
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

<div className="flex items-center gap-2">
  <label className="text-gray-300 text-sm">Status:</label>
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
    className="bg-gray-800 text-white px-2 py-1 rounded border border-gray-700"
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
                      ? 'bg-blue-600 text-white hover:bg-blue-700' 
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
};
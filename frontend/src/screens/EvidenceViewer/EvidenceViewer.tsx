import {useEffect, useState, ReactNode } from "react";
import {
  Bell,
  File,
  Folder,
  Home,
  MessageSquare,
  Search,
  Settings,
  // SlidersHorizontal,
  // ArrowUpDown,
  Download,
  Share,
  Send,
  Info,
  MessageCircle,
  Shield,
  Clock,
  Users,
  CheckCircle,
  XCircle,
  FileText,
  Hash,
  Calendar,
  MoreVertical} from "lucide-react";
import axios from "axios";
import { Link, useSearchParams } from "react-router-dom";
import { SidebarToggleButton, useSidebar } from '../../context/SidebarToggleContext';
//import { string } from "prop-types";
import { useParams } from "react-router-dom";
import { fetchEvidenceByCaseId } from "./api";
import { fetchThreadsByFile } from "./api"; 
import { fetchThreadMessages } from "./api";
import { sendThreadMessage } from "./api";
import { createAnnotationThread } from "./api";
import { addThreadParticipant } from "./api";
import { fetchThreadParticipants } from "./api";
import { addReaction } from "./api";
import { approveMessage } from "./api";
import{MessageCard} from "../../components/ui/MessageCard";
import { ClipboardList } from "lucide-react";


// Import Select components from your UI library
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem
} from "../../components/ui/select";


// Helper to get user name from session storage if userId matches current user
function getUserNameById(userId: string): string {
  const storedUser = sessionStorage.getItem("user");
  if (!storedUser) return "Unknown";
  const user = JSON.parse(storedUser);
  if (user.id === userId) return user.name || user.email || "Unknown";
  return userId; // fallback to userId if not found
}
interface FileItem {
  hash: ReactNode;
  id: string; // Corresponds to Go's `ID` (uuid.UUID)
  caseId: string; // Corresponds to Go's `CaseID` (uuid.UUID)
  uploaded_by: string; // Corresponds to Go's `UploadedBy` (uuid.UUID)
  filename: string; // Corresponds to Go's `Filename`
  file_type: string; // Corresponds to Go's `FileType`
  ipfs_cid: string; // Corresponds to Go's `IpfsCID`
  file_size: number; // Corresponds to Go's `FileSize` (int64)
  checksum: string; // Corresponds to Go's `Checksum`
  metadata: string; // Corresponds to Go's `Metadata` (JSON string)
  uploaded_at: string; // Corresponds to Go's `UploadedAt` (time.Time)
  description?: string; // These would likely be parsed from 'metadata' JSON
  status?: 'verified' | 'pending' | 'failed' | string; // Parsed from 'metadata'
  chainOfCustody?: string[]; // Parsed from 'metadata'
  acquisitionDate?: string; // Parsed from 'metadata'
  acquisitionTool?: string; // Parsed from 'metadata'
  integrityCheck?: 'passed' | 'failed' | 'pending' | string; // Parsed from 'metadata'
  threadCount?: number; // Parsed from 'metadata'
  priority?: 'high' | 'medium' | 'low' | string; // Parsed from 'metadata'
}


 interface ThreadTag {
  id: string;        // UUID
  thread_id: string; // UUID
  tag_name: string;
}

interface AnnotationThread {
  id: string;
  title: string;
  user: string;
  avatar: string;
  time: string;
  messageCount: number;
  participantCount: number;
  isActive?: boolean;
  status: 'open' | 'resolved' | 'pending_approval';
  priority: 'high' | 'medium' | 'low';
  tags: ThreadTag[];
  fileId: string;
  createdBy?: string; // UUID of the user who created the thread
}



interface ThreadMessage {
  id: string;
  threadID: string;
  parentMessageID?: string | null;
  userID: string;
  message: string;
  isApproved?: boolean;
  approvedBy?: string | null;
  approvedAt?: string | null;
  createdAt: string;
  updatedAt: string;
  mentions: { messageID: string; mentionedUserID: string; createdAt: string }[];
  reactions: { id: string; messageID: string; userID: string; reaction: string; createdAt: string }[];

  // Optional, if you still want to display replies in nested form:
  replies?: ThreadMessage[];
}


function buildNestedMessages(messages: ThreadMessage[]): ThreadMessage[] {
  const messageMap: { [id: string]: ThreadMessage & { replies: ThreadMessage[] } } = {};
  const topLevel: ThreadMessage[] = [];

  messages.forEach((msg) => {
    messageMap[msg.id] = { ...msg, replies: [] };
  });

  messages.forEach((msg) => {
    if (msg.parentMessageID && messageMap[msg.parentMessageID]) {
      messageMap[msg.parentMessageID].replies.push(messageMap[msg.id]);
    } else {
      topLevel.push(messageMap[msg.id]);
    }
  });

  return topLevel;
}



export const EvidenceViewer  =() =>{
// Helper to extract hashes from metadata JSON string
function getHashesFromMetadata(metadata: string) {
  try {
    const meta = JSON.parse(metadata);
    return {
      sha256: meta.sha256 || "N/A",
      md5: meta.md5 || "N/A"
    };
  } catch {
    return { sha256: "N/A", md5: "N/A" };
  }
}
// Helper to extract details from chainOfCustody
function getCustodyDetails(chainOfCustody: any[]) {
  if (!Array.isArray(chainOfCustody) || chainOfCustody.length === 0) return {};
  const first = chainOfCustody[0];
  const last = chainOfCustody[chainOfCustody.length - 1];
  // Add debug log to inspect entries
  console.log("getCustodyDetails: first entry:", first);
  console.log("getCustodyDetails: last entry:", last);
  return {
    examiner: first?.forensic_info?.examiner || first?.forensic_info?.Examiner || "",
    custodian: last?.custodian || last?.custodian_name || last?.user_name || "",
    acquisitionDate: last?.acquisition_date || last?.acquisitionDate || "",
    acquisitionTool: last?.acquisition_tool || last?.acquisitionTool || "",
    hash: last?.hash || "",
    osVersion: last?.system_info?.osVersion || last?.system_info?.os_version || "",
    architecture: last?.system_info?.architecture || "",
    computerName: last?.system_info?.computerName || last?.system_info?.computer_name || "",
    domain: last?.system_info?.domain || "",
    lastBoot: last?.system_info?.lastBoot || last?.system_info?.last_boot || "",
    method: first?.forensic_info?.method || first?.forensic_info?.Method || "",
    location: first?.forensic_info?.location || "",
    legalStatus: first?.forensic_info?.legalStatus || first?.forensic_info?.legal_status || "",
    notes: first?.forensic_info?.notes || "",
  };
}
const BASE_URL = "http://localhost:8080/api/v1";
  const storedUser = sessionStorage.getItem("user");
  const user = storedUser ? JSON.parse(storedUser) : null;
  const displayName = user?.name || user?.email?.split("@")[0] || "Agent User";
  const initials = displayName
    .split(" ")
    .map((part: string) => part[0])
    .join("")
    .toUpperCase();

const [role, setRole] = useState<string>(user?.role || "");
const isDFIRAdmin = role === "DFIR Admin";
//const {sidebarVisible, toggleSidebar} = useSidebar();
  
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


    
const { caseId } = useParams();

const [searchParams] = useSearchParams();
const evidenceIdFromQuery = searchParams.get("evidenceId");

const [files, setFiles] = useState<FileItem[]>([]);
const [loading, setLoading] = useState(true);
const [error, setError] = useState<string | null>(null);
const [showReactionPicker, setShowReactionPicker] = useState<string | null>(null);

const [annotationThreads, setAnnotationThreads] = useState<AnnotationThread[]>([]);   
const [newThreadTitle, setNewThreadTitle] = useState('');
const [selectedFile, setSelectedFile] = useState<FileItem | null>(null);
const [selectedThread, setSelectedThread] = useState<AnnotationThread | null>(null);
const [replyingToMessageId, setReplyingToMessageId] = useState<string | null>(null);
const [replyText, setReplyText] = useState('');
const [newMessage, setNewMessage] = useState('');
const [searchTerm, setSearchTerm] = useState('');
const [activeTab, setActiveTab] = useState<'overview' | 'threads' | 'metadata' | 'chain'>('overview');

const [threadMessages, setThreadMessages] = useState<ThreadMessage[]>([]);

// Chain of Custody state
const [chainOfCustody, setChainOfCustody] = useState<any[]>([]);
const custodyDetails = getCustodyDetails(chainOfCustody);
console.debug('Parsed custodyDetails:', custodyDetails);

const [chainLoading] = useState(false);
const [chainError] = useState<string | null>(null);

useEffect(() => {
  async function loadEvidence() {
    if (!caseId || caseId === "undefined") {
      setFiles([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const evidenceFiles = await fetchEvidenceByCaseId(caseId);
      // For each file, fetch its threads and set threadCount
      const filesWithThreadCount = await Promise.all(
        evidenceFiles.map(async (file: any) => {
          const threads = await fetchThreadsByFile(file.id);
          return {
            ...file,
            threadCount: Array.isArray(threads) ? threads.length : 0,
          };
        })
      );
      setFiles(filesWithThreadCount);
      setError(null);
      // Auto-select evidence if evidenceId is present in query
      function ensureCaseId(file: FileItem | null): FileItem | null {
        if (!file) return null;
        // If file.caseId is missing or undefined, set it from caseId param
        if (!file.caseId || file.caseId === "undefined") {
          return { ...file, caseId: caseId ?? "" };
        }
        return file;
      }
      if (evidenceIdFromQuery) {
        const found = filesWithThreadCount.find(f => f.id === evidenceIdFromQuery);
        if (found) setSelectedFile(ensureCaseId(found));
        else setSelectedFile(ensureCaseId(filesWithThreadCount[0] || null));
      } else {
        setSelectedFile(ensureCaseId(filesWithThreadCount[0] || null));
      }
    } catch (err: any) {
      setError(err.message || "Failed to load evidence files");
    } finally {
      setLoading(false);
    }
  }

  loadEvidence();
}, [caseId, evidenceIdFromQuery]);

useEffect(() => {
  // Clear previous chain of custody data when switching files
  setChainOfCustody([]);

  if (!selectedFile) {
    return;
  }
  const selectedCaseId = selectedFile.caseId;
  if (!selectedCaseId || selectedCaseId === "undefined") {
    return;
  }

  // Fetch chain of custody for selected evidence
  const fetchChainOfCustody = async () => {
    let res, parsedEntries;
    const evidenceId = selectedFile.id;

    const token = sessionStorage.getItem("authToken");
    let axiosConfig = {};
    if (token) {
      axiosConfig = { headers: { Authorization: `Bearer ${token}` } };
    }

    try {
      res = await axios.get(`${BASE_URL}/cases/${selectedCaseId}/chain_of_custody?evidence_id=${evidenceId}`, axiosConfig);
      console.log("Chain of Custody API raw response:", res.data);
      parsedEntries = Array.isArray(res.data)
        ? res.data.map(entry => {
            // Parse JSON fields
            const forensic_info = typeof entry.forensic_info === 'string' ? safeParse(entry.forensic_info) : entry.forensic_info || {};
            const system_info = typeof entry.system_info === 'string' ? safeParse(entry.system_info) : entry.system_info || {};
            // Add debug logs for parsed fields
            console.log("Parsed forensic_info:", forensic_info);
            console.log("Parsed system_info:", system_info);
            // Map snake_case to camelCase
            return {
              ...entry,
              forensic_info: {
                notes: forensic_info.notes || forensic_info.Notes || "",
                method: forensic_info.method || forensic_info.Method || "",
                examiner: forensic_info.examiner || forensic_info.Examiner || "",
                location: forensic_info.location || "",
                legalStatus: forensic_info.legalStatus || forensic_info.legal_status || forensic_info.LegalStatus || "",
              },
              system_info: {
                osVersion: system_info.osVersion || system_info.os_version || "",
                architecture: system_info.architecture || "",
                computerName: system_info.computerName || system_info.computer_name || "",
                domain: system_info.domain || "",
                lastBoot: system_info.lastBoot || system_info.last_boot || "",
              }
            };
          })
        : [];
      console.log("Chain of Custody parsed entries:", parsedEntries);
      setChainOfCustody(parsedEntries);
    } catch (err) {
      setChainOfCustody([]);
    }
  };
  // Safely parse JSON, fallback to empty object
  function safeParse(json: string) {
    try {
      return JSON.parse(json);
    } catch {
      return {};
    }
  }

  // Fetch threads as before
  const loadThreads = async () => {
    try {
      const threads = await fetchThreadsByFile(selectedFile.id);
      const threadsWithCounts = await Promise.all(
        threads.map(async (t: any) => {
          const rawMessages = await fetchThreadMessages(t.id);
          const userName = getUserNameById(t.created_by);
          return {
            ...t,
            fileId: t.file_id,
            caseId: t.case_id,
            createdBy: t.created_by,
            tags: t.Tags || [],
            participantCount: t.Participants?.length || 0,
            messageCount: rawMessages.length,
            user: userName,
            avatar: userName.split(" ").map((n: string) => n[0]).join("").toUpperCase(),
            time: new Date(t.created_at).toLocaleString(),
          };
        })
      );
      setAnnotationThreads(threadsWithCounts);
    } catch (err) {
      console.error("Failed to load threads", err);
    }
  };

  fetchChainOfCustody();
  loadThreads();
}, [selectedFile]);
// Fetch evidence metadata for selected file
useEffect(() => {
  // Metadata will be derived from chainOfCustody entries
}, [selectedFile]);


function formatMessages(rawMessages: any[]): ThreadMessage[] {
  return rawMessages.map((m) => ({
    id: m.ID,
    threadID: m.ThreadID,
    parentMessageID: m.ParentMessageID ?? m.parentMessageID ?? null, // ensure field is always present
    userID: m.UserID,
    message: m.Message,
    isApproved: m.IsApproved,
    approvedBy: m.ApprovedBy,
    approvedAt: m.ApprovedAt,
    createdAt: m.CreatedAt ? new Date(m.CreatedAt).toLocaleString() : "",
    updatedAt: m.UpdatedAt ? new Date(m.UpdatedAt).toLocaleString() : "",
    mentions: m.Mentions || [],
    reactions: (m.Reactions || []).map((r: any) => ({
      id: r.ID,
      messageID: r.MessageID,
      userID: r.UserID,
      reaction: r.Reaction,
      createdAt: r.CreatedAt ? new Date(r.CreatedAt).toLocaleString() : "",
    })),
    replies: [],
  }));
}



useEffect(() => {
  if (!selectedThread) return;

  const loadMessages = async () => {
    try {
    const rawMessages = await fetchThreadMessages(selectedThread.id); // from `api.ts`
    console.log("Fetched messages for thread:", rawMessages);
    const formattedMessages = formatMessages(rawMessages);
    console.log("Formatted messages before nesting:", formattedMessages);
    setThreadMessages(buildNestedMessages(formattedMessages));
    } catch (err) {
      console.error("Failed to load thread messages", err);
    }
  };

  loadMessages();
}, [selectedThread]);

function formatFileSize(bytes: number): string {
  if (bytes === undefined || bytes === null) return "0 MB";
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}
const handleSendMessage = async (overrideText?: string) => {
  const text = overrideText ?? newMessage;
  if (!text.trim() || !selectedThread) return;

  try {
    if (user.id !== selectedThread.createdBy) {
      try {
        await addThreadParticipant(selectedThread.id, user.id);
        const participants = await fetchThreadParticipants(selectedThread.id);
        setAnnotationThreads(prev =>
          prev.map(thread =>
            thread.id === selectedThread.id
              ? { ...thread, participantCount: participants.length }
              : thread
          )
        );
      } catch (err) {
        console.warn("Participant already exists or failed:", err);
      }
    }

    await sendThreadMessage(selectedThread.id, {
      user_id: user.id,
      message: text,
      parent_message_id: replyingToMessageId || null,
      mentions: []
    });

    const updatedMessages = await fetchThreadMessages(selectedThread.id);
    const formatted = formatMessages(updatedMessages);
    setThreadMessages(buildNestedMessages(formatted));

    // Update messageCount in annotationThreads
    setAnnotationThreads(prev =>
      prev.map(thread =>
        thread.id === selectedThread.id
          ? { ...thread, messageCount: formatted.length }
          : thread
      )
    );

    setNewMessage('');
    setReplyText('');
    setReplyingToMessageId(null);

    

    setAnnotationThreads(prev =>
      prev.map(thread =>
        thread.id === selectedThread.id
          ? { ...thread, messageCount: formatted.length }
          : thread
      )
    );


    const messageCount = formatted.length;
    setAnnotationThreads(prev =>
      prev.map(thread =>
        thread.id === selectedThread.id
          ? { ...thread, messageCount }
          : thread
      )
    );

    setSelectedThread(prev => prev ? { ...prev, messageCount } : null);
  } catch (err) {
    console.error("Failed to send message:", err);
  }
};

const [profile] = useState({
    name: user?.name || "User",
    email: user?.email || "user@aegis.com",
    role: user?.role || "Admin",
    image: user?.image_url || null, // assuming you might store this too
  });

  const recentThread = [...annotationThreads]
  .sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime())[0]; // most recent

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'verified': case 'passed': case 'resolved': return 'text-green-400';
      case 'pending': case 'open': return 'text-yellow-400';
      case 'failed': return 'text-red-400';
      default: return 'text-muted-foreground';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'verified': case 'passed': case 'resolved': return <CheckCircle className="w-4 h-4" />;
      case 'pending': case 'open': return <Clock className="w-4 h-4" />;
      case 'failed': return <XCircle className="w-4 h-4" />;
      default: return <Info className="w-4 h-4" />;
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'text-red-400 bg-red-400/10';
      case 'medium': return 'text-yellow-400 bg-yellow-400/10';
      case 'low': return 'text-green-400 bg-green-400/10';
      default: return 'text-muted-foreground bg-gray-400/10';
    }
  };

  // const filteredThreads = annotationThreads.filter(thread => 
  //   selectedFile ? thread.fileId === selectedFile.id : true
  // );

  const filteredThreads = annotationThreads;
  const [typeFilter, setTypeFilter] = useState<string | null>(null);
  const [timeFilter, setTimeFilter] = useState<'recent' | 'oldest' | null>(null);

  let filteredFiles = [...files];

  // Filter by type
  if (typeFilter && typeFilter !== "all") {
    filteredFiles = filteredFiles.filter(file => file.file_type === typeFilter);
  }

  // Filter by search term
  if (searchTerm && searchTerm.trim() !== "") {
    const term = searchTerm.trim().toLowerCase();
    filteredFiles = filteredFiles.filter(file =>
      file.filename.toLowerCase().includes(term) ||
      (file.description && file.description.toLowerCase().includes(term))
    );
  }

  // Sort by time
  if (timeFilter === 'recent') {
    filteredFiles.sort((a, b) => new Date(b.uploaded_at || '').getTime() - new Date(a.uploaded_at || '').getTime());
  } else if (timeFilter === 'oldest') {
    filteredFiles.sort((a, b) => new Date(a.uploaded_at || '').getTime() - new Date(b.uploaded_at || '').getTime());
  }



// Add these helper functions inside the EvidenceViewer component (after existing functions)
// const handleAddReaction = async () => {
//   try {
//     // Call the backend and get the updated message with reactions

//     // Update the threadMessages state: replace the old message with updatedMessage
//     if (!selectedThread) return;
//     const updatedMessages = await fetchThreadMessages(selectedThread.id);
//     const formattedMessages = formatMessages(updatedMessages);
//     setThreadMessages(buildNestedMessages(formattedMessages));

//     setShowReactionPicker(null); // Close reaction picker
//   } catch (err) {
//     console.error("Failed to add reaction:", err);
//   }
// };

const handleAddReaction = async (messageId: string, reaction: string) => {
  try {
    await addReaction(messageId, user.id, reaction);

    // Refresh the messages to get the updated reactions
    if (!selectedThread) return;
    const updatedMessages = await fetchThreadMessages(selectedThread.id);
    const formattedMessages = formatMessages(updatedMessages);
    setThreadMessages(buildNestedMessages(formattedMessages));

    setShowReactionPicker(null); // Close reaction picker
  } catch (err) {
    console.error("Failed to add reaction:", err);
  }
};

const handleApproveMessage = async (messageId: string) => {
  try {
    await approveMessage(messageId);
    const updatedMessages = await fetchThreadMessages(selectedThread!.id);
    const formatted = formatMessages(updatedMessages);
    setThreadMessages(buildNestedMessages(formatted));
  } catch (err) {
    console.error("Failed to approve message:", err);
  }
};

const handleSendMessageWithParent = async (text: string, parentId?: string) => {
  if (!text.trim() || !selectedThread) return;

  try {
    if (user.id !== selectedThread.createdBy) {
      try {
        await addThreadParticipant(selectedThread.id, user.id);
        const participants = await fetchThreadParticipants(selectedThread.id);
        setAnnotationThreads(prev =>
          prev.map(thread =>
            thread.id === selectedThread.id
              ? { ...thread, participantCount: participants.length }
              : thread
          )
        );
      } catch (err) {
        console.warn("Participant already exists or failed:", err);
      }
    }

    await sendThreadMessage(selectedThread.id, {
      user_id: user.id,
      message: text,
      parent_message_id: parentId || null,
      mentions: []
    });

    const updatedMessages = await fetchThreadMessages(selectedThread.id);
    const formatted = formatMessages(updatedMessages);
    const nestedMessages = buildNestedMessages(formatted);
    setThreadMessages(nestedMessages);

    // Update message count in both thread list and selected thread
    const messageCount = formatted.length;
    setAnnotationThreads(prev =>
      prev.map(thread =>
        thread.id === selectedThread.id
          ? { ...thread, messageCount }
          : thread
      )
    );
    
    setSelectedThread(prev => prev ? { ...prev, messageCount } : null);

  } catch (err) {
    console.error("Failed to send message:", err);
  }
};


//TO BE DELETED
// function updateReplyApproval(replies: ThreadMessage[], replyId: string): ThreadMessage[] {
//   return replies.map(reply => {
//     if (reply.id === replyId) {
//       return { ...reply, isApproved: true };
//     } else if (reply.replies) {
//       return {
//         ...reply,
//         replies: updateReplyApproval(reply.replies, replyId)
//       };
//     }
//     return reply;
//   });
// }

function timeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  const intervals: [number, string][] = [
    [60, "seconds"],
    [3600, "minutes"],
    [86400, "hours"],
    [604800, "days"]
  ];

  for (let [limit, label] of intervals.reverse()) {
    const value = Math.floor(seconds / limit);
    if (value >= 1) return `${value} ${label} ago`;
  }

  return "just now";
}

// Show loading only when we actually have a caseId and are loading
if (loading && caseId && caseId !== "undefined") {
  return <div className="p-4">Loading evidence files...</div>;
}

// Show error only for actual errors (not missing caseId)  
if (error) {
  return <div className="p-4 text-red-500">Error: {error}</div>;
}

// Handle no case scenario early
if (!caseId || caseId === "undefined") {
  return (
    <div className="min-h-screen bg-background text-foreground flex">
      {/* Sidebar */}
      <aside className="fixed left-0 top-0 h-screen w-64 bg-background border-r border-border p-6 flex flex-col sidebar-toggle-target">        <div>
          {/* Logo */}
          <div className="flex items-center gap-3 mb-8">
            <div className="w-10 h-10 rounded-lg overflow-hidden">
              <img
                src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
                alt="AEGIS Logo"
                className="w-full h-full object-cover"
              />
            </div>
            <span className="font-bold text-foreground text-xl">AEGIS</span>
          </div>

          {/* Navigation */}
          <nav className="space-y-1">
            <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-2 rounded-lg transition-colors cursor-pointer">
              <Home className="w-5 h-5" />
              <Link to="/dashboard"><span className="text-sm">Dashboard</span></Link>
            </div>
            <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-2 rounded-lg transition-colors cursor-pointer">
              <Folder className="w-5 h-5" />
              <Link to="/case-management"><span className="text-sm">Case management</span></Link>
            </div>
            <div className="flex items-center gap-3 bg-blue-600 text-white p-3 rounded-lg">
              <File className="w-5 h-5" />
              <span className="text-sm font-medium">Evidence Viewer</span>
            </div>
            <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-2 rounded-lg transition-colors cursor-pointer">
              <MessageSquare className="w-5 h-5" />
              <Link to="/secure-chat"><span className="text-sm">Secure chat</span></Link>
            </div>
                 {isDFIRAdmin && (
               <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-2 rounded-lg transition-colors cursor-pointer">
              <ClipboardList className="w-5 h-5" />
              <Link to="/report-dashboard"><span className="text-sm">Case Reports</span></Link>
            </div>
          )}
          </nav>
        </div>

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
              <p className="text-muted-foreground text-xs">{user?.email || "user@dfir.com"}</p>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="ml-64 flex-grow bg-background flex">
        {/* Header */}
        <div className="fixed top-0 left-64 right-0 z-20 bg-background border-b border-border p-4 header-toggle-target">            <div className="flex items-center justify-between">
            {/* Case Number and Tabs */}
            <div className="flex items-center gap-4">
              <div className="bg-gray-600 text-white px-3 py-1 rounded text-sm font-medium">
                No Case Selected
              </div>
              <div className="flex items-center gap-6">
                
                <Link to="/dashboard">
                  <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                    Dashboard
                  </button>
                </Link>
                <Link to="/case-management">
                  <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                    Case Management
                  </button>
                </Link>
                <Link to="/evidence-viewer">
                  <button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">
                    Evidence Viewer
                  </button>
                </Link>
                <Link to="/secure-chat">
                  <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                    Secure Chat
                  </button>
                </Link>
              </div>
            </div>

            {/* Right actions */}
            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <input
                  className="w-64 h-10 bg-card border border-muted rounded-lg pl-10 pr-4 text-foreground placeholder-gray-400 text-sm focus:outline-none focus:border-blue-500"
                  placeholder="Search cases, evidence, users"
                />
              </div>
              <Link to="/notifications">
                <Bell className="text-muted-foreground hover:text-foreground w-5 h-5 cursor-pointer" />
              </Link>
              <Link to="/settings">
                <Settings className="text-muted-foreground hover:text-foreground w-5 h-5 cursor-pointer" />
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
        
        <div className="flex flex-col items-center justify-center w-full h-[60vh] text-center text-muted-foreground pt-20">
          <h2 className="text-2xl font-semibold mb-4">No case, no load</h2>
          <p>Go to case management and pick a case to view details.</p>
        </div>
      </main>
    </div>
  );
}

  return (
    <div className="min-h-screen bg-background text-foreground flex">
      {/* Sidebar */}
      <aside className="fixed left-0 top-0 h-full w-64 bg-background border-r border-border p-4 flex flex-col justify-between z-10">
        <div>
          {/* Logo */}
          <div className="flex items-center gap-3 mb-8">
            <div className="w-10 h-10 rounded-lg overflow-hidden">
              <img
                src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
                alt="AEGIS Logo"
                className="w-full h-full object-cover"
              />
            </div>
            <span className="font-bold text-foreground text-xl">AEGIS</span>
          </div>

          {/* Navigation */}
          <nav className="space-y-1">
            <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-2 rounded-lg transition-colors cursor-pointer">
              <Home className="w-5 h-5" />
              <Link to="/dashboard"><span className="text-sm">Dashboard</span></Link>
            </div>
            <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-2 rounded-lg transition-colors cursor-pointer">
              <Folder className="w-5 h-5" />
              <Link to="/case-management"><span className="text-sm">Case management</span></Link>
            </div>
          <div className="flex items-center gap-3 bg-blue-600 text-white p-3 rounded-lg">
              <File className="w-5 h-5" />
              <span className="text-sm font-medium">Evidence Viewer</span>
            </div>
            <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-2 rounded-lg transition-colors cursor-pointer">
              <MessageSquare className="w-5 h-5" />
              <Link to="/secure-chat"><span className="text-sm">Secure chat</span></Link>
            </div>
                  <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-2 rounded-lg transition-colors cursor-pointer">
              <ClipboardList className="w-5 h-5" />
              <Link to="/report-dashboard"><span className="text-sm">Case Reports</span></Link>
            </div>
          </nav>
        </div>

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
              <p className="text-muted-foreground text-xs">{user?.email || "user@dfir.com"}</p>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="ml-80 min-h-screen bg-background">
        
        {/* Header */}
        <div className="fixed top-0 left-64 right-0 z-20 bg-background border-b border-border p-4">
          
          <div className="flex items-center justify-between">
            {/* Case Number and Tabs */}
            <div className="flex items-center gap-4">
              <div className="bg-blue-600 text-white px-3 py-1 rounded text-sm font-medium">
                 #{`CS-${caseId.slice(0, 7)}...`}
              </div>
            <div className="flex items-center gap-6">
              <SidebarToggleButton/>
              <Link to="/dashboard">
              <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Dashboard
              </button></Link>
              <Link to="/case-management"><button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Case Management
              </button></Link>
              <Link to="/evidence-viewer"><button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">

                Evidence Viewer
              </button></Link>
              <Link to="/secure-chat"><button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Secure Chat
              </button></Link>
            </div>
            </div>

            {/* Right actions */}
            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <input
                  className="w-64 h-10 bg-card border border-muted rounded-lg pl-10 pr-4 text-foreground placeholder-gray-400 text-sm focus:outline-none focus:border-blue-500"
                  placeholder="Search cases, evidence, users"
                />
              </div>
              <Link to="/notifications">
              <Bell className="text-muted-foreground hover:text-foreground w-5 h-5 cursor-pointer" />
              </Link>
              <Link to="/settings"><Settings className="text-muted-foreground hover:text-foreground w-5 h-5 cursor-pointer" /></Link>
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

        {/* Content Area */}
        <div className="flex-1 flex pt-20">
          
            <>
          {/* Evidence Files Panel */}
          <div className="w-80 border-r border-border p-4">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold">Evidence Files</h2>
              <span className="text-sm text-muted-foreground">{files.length} items</span>
            </div>
            
            {/* Search */}
            <div className="relative mb-4">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <input
                className="w-full h-9 bg-card border border-muted rounded-lg pl-10 pr-4 text-foreground placeholder-gray-400 text-sm focus:outline-none focus:border-blue-500"
                placeholder="Search evidence files"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            {/* Filter and Sort */}
            <div className="flex gap-2 mb-4">
              {/* Filter by type */}
              <Select onValueChange={(value) => setTypeFilter(value)}>
                <SelectTrigger className="w-40 bg-muted border-border text-foreground text-xs hover:bg-muted/50">
                  <SelectValue placeholder="Filter by type" />
                </SelectTrigger>
                <SelectContent className="bg-gray-800/80 text-muted-foreground text-sm border-border">                  <SelectItem value="all">All files</SelectItem>
                  <SelectItem value="memory_dump">Memory Dump</SelectItem>
                  <SelectItem value="executable">Executable</SelectItem>
                  <SelectItem value="network_capture">Network Capture</SelectItem>
                  <SelectItem value="log">Log</SelectItem>
                  <SelectItem value="image">Image</SelectItem>
                  <SelectItem value="document">Document</SelectItem>

                </SelectContent>
              </Select>

              {/* Sort by time */}
              <Select onValueChange={(value) => setTimeFilter(value as 'recent' | 'oldest')}>
                <SelectTrigger className="w-36 bg-muted border-border text-foreground text-xs hover:bg-muted/50">
                  <SelectValue placeholder="Sort by time" />
                </SelectTrigger>
                <SelectContent className="bg-gray-800/80 text-muted-foreground text-sm border-border">                  <SelectItem value="recent">Most Recent</SelectItem>
                  <SelectItem value="oldest">Oldest First</SelectItem>
                </SelectContent>
              </Select>
            </div>


            {/* File List */}
            <div className="space-y-2">
              {filteredFiles.map((file) => (
                <button
                  key={file.id}
                  onClick={() => setSelectedFile(file)}
                  className={`w-full p-3 rounded-lg border transition-all ${
                    selectedFile?.id === file.id
                      ? 'bg-blue-600/20 border-blue-500'
                      : 'border-muted hover:bg-muted/50 hover:border-gray-600'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <File className="w-5 h-5 text-muted-foreground flex-shrink-0 mt-0.5" />
                    <div className="flex-1 min-w-0 text-left">
                      <div className="font-medium text-sm truncate mb-1">{file.filename}</div>
                      <div className="flex items-center gap-2 mb-2">
                        <span className={`inline-flex items-center gap-1 text-xs ${getStatusColor(file.status || "pending")}`}>
                          {getStatusIcon(file.status || "pending")}
                          {file.status || "pending"}
                        </span>
                        <span className={`px-2 py-0.5 rounded text-xs ${getPriorityColor(file.priority || "low")}`}>
                          {file.priority || "low"}
                        </span>
                      </div>
                      <div className="flex items-center justify-between text-xs text-muted-foreground">
                        <span>{formatFileSize(file.file_size)}</span>
                        <div className="flex items-center gap-1">
                          <MessageCircle className="w-3 h-3" />
                          <span>{file.threadCount}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Main Viewer Area */}
          <div className="flex-1 flex flex-col">
            {selectedFile && (
              <>
                {/* File Header */}
                <div className="border-b border-border p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                      <h1 className="text-2xl font-semibold">{selectedFile.filename}</h1>
                      <div className={`inline-flex items-center gap-1 px-2 py-1 rounded text-sm ${getStatusColor(selectedFile.status || "pending")}`}>
                        {getStatusIcon(selectedFile.status || "pending")}
                        {selectedFile.status || "pending"}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg">
                        <Download className="w-5 h-5" />
                      </button>
                      <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg">
                        <FileText className="w-5 h-5" />
                      </button>
                      <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg">
                        <Share className="w-5 h-5" />
                      </button>
                      <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg">
                        <MoreVertical className="w-5 h-5" />
                      </button>
                    </div>
                  </div>

                  {/* Tabs */}
                  <div className="flex items-center gap-6 border-b border-muted">
                    <button
                    onClick={() => setActiveTab('chain')}
                    className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                      activeTab === 'chain'
                        ? 'text-blue-400 border-blue-400'
                        : 'text-muted-foreground border-transparent hover:text-foreground hover:border-gray-600'
                    }`}
                  >
                    Chain of Custody
                  </button>

                    <button
                      onClick={() => setActiveTab('overview')}
                      className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'overview'
                          ? 'text-blue-400 border-blue-400'
                          : 'text-muted-foreground border-transparent hover:text-foreground hover:border-gray-600'
                      }`}
                    >
                      Overview
                    </button>
                    <button
                      onClick={() => setActiveTab('threads')}
                      className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'threads'
                          ? 'text-blue-400 border-blue-400'
                          : 'text-muted-foreground border-transparent hover:text-foreground hover:border-gray-600'
                      }`}
                    >
                      Discussions ({filteredThreads.length})
                    </button>
                    <button
                      onClick={() => setActiveTab('metadata')}
                      className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'metadata'
                          ? 'text-blue-400 border-blue-400'
                          : 'text-muted-foreground border-transparent hover:text-foreground hover:border-gray-600'
                      }`}
                    >
                      Metadata
                    </button>
                  </div>
                </div>

                {/* Tab Content */}
                <div className="flex-1 overflow-y-auto p-6">
                  {activeTab === 'chain' && (
                    <div className="space-y-4">
                      <div className="flex justify-between items-center">
                        <h3 className="text-lg font-semibold">Chain of Custody Records</h3>
                        <button
                          className="px-3 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 flex items-center gap-2"
                          onClick={() => {
                            if (selectedFile && selectedFile.id) {
                              window.location.assign(`/chain-of-custody/${caseId}?evidenceId=${selectedFile.id}`);
                            }
                          }}
                        >
                          <span className="text-sm font-medium">+ Add Entry</span>
                        </button>
                      </div>
                      <div className="bg-card p-4 rounded-lg">
                        {chainLoading ? (
                          <p className="text-muted-foreground text-sm">Loading chain of custody...</p>
                        ) : chainError ? (
                          <p className="text-red-500 text-sm">{chainError}</p>
                        ) : chainOfCustody.length > 0 ? (
                          chainOfCustody.map((entry, idx) => (
                            <div key={entry.id || idx} className="flex items-center gap-3 mb-2">
                              <div className="w-2 h-2 bg-green-400 rounded-full"></div>
                              <div className="flex-1">
                                <p className="text-sm font-medium">{entry.custodian_name || entry.custodian || entry.user_name || "Unknown"}</p>
                                <p className="text-xs text-muted-foreground">
                                  {idx === 0
                                    ? "Original Collector"
                                    : (idx === chainOfCustody.length - 1
                                        ? "Current Custodian"
                                        : "Transferred")}
                                </p>
                                <p className="text-xs text-muted-foreground">{entry.timestamp ? new Date(entry.timestamp).toLocaleString() : ""}</p>
                              </div>
                            </div>
                          ))
                        ) : (
                          <p className="text-muted-foreground text-sm">No chain of custody entries yet.</p>
                        )}
                      </div>
                    </div>
                  )}

                  {activeTab === 'overview' && (
                    <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
                      {/* Evidence Information */}
                      <div className="bg-card p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Shield className="w-5 h-5 text-blue-400" />
                          Evidence Information
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div>
                            <span className="text-muted-foreground">Description:</span>
                            <p className="text-muted-foreground mt-1">{custodyDetails.notes || "No description available"}</p>
                          </div>
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <span className="text-muted-foreground">Size:</span>
                              <p className="text-muted-foreground">{formatFileSize(selectedFile.file_size)}</p>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Type:</span>
                              <p className="text-muted-foreground capitalize">{selectedFile.file_type.replace('_', ' ')}</p>
                            </div>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Integrity Check:</span>
                            <div className={`inline-flex items-center gap-1 ml-2 ${getStatusColor(selectedFile.integrityCheck || "pending")}`}>
                              {getStatusIcon(selectedFile.integrityCheck  || "pending")}
                              <span className="capitalize">{selectedFile.integrityCheck}</span>
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Chain of Custody */}
                      <div className="bg-card p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Users className="w-5 h-5 text-green-400" />
                          Chain of Custody
                        </h3>
                        <div className="space-y-3">
                          {chainLoading ? (
                            <p className="text-muted-foreground text-sm">Loading chain of custody...</p>
                          ) : chainError ? (
                            <p className="text-red-500 text-sm">{chainError}</p>
                          ) : chainOfCustody.length > 0 ? (
                            chainOfCustody.map((entry, index) => (
                              <div key={entry.id || index} className="flex items-center gap-3">
                                <div className="w-2 h-2 bg-green-400 rounded-full"></div>
                                <div className="flex-1">
                                  <div className="text-sm font-medium">{entry.custodian_name || entry.custodian || entry.user_name || "Unknown"}</div>
                                  <div className="text-xs text-muted-foreground">
                                    {index === 0 ? 'Original Collector' : 
                                     (index === chainOfCustody.length - 1) ? 'Current Custodian' : 'Transferred'}
                                  </div>
                                  <div className="text-xs text-muted-foreground">{entry.timestamp ? new Date(entry.timestamp).toLocaleString() : ""}</div>
                                </div>
                                <CheckCircle className="w-4 h-4 text-green-400" />
                              </div>
                            ))
                          ) : (
                            <p className="text-muted-foreground text-sm">No chain of custody entries yet.</p>
                          )}
                        </div>
                      </div>

                      {/* Acquisition Details */}
                      <div className="bg-card p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Calendar className="w-5 h-5 text-purple-400" />
                          Acquisition Details
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div>
                            <span className="text-muted-foreground">Acquisition Date:</span>
                            <p className="text-muted-foreground">{custodyDetails.acquisitionDate ? new Date(custodyDetails.acquisitionDate).toLocaleString() : "N/A"}</p>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Tool Used:</span>
                            <p className="text-muted-foreground">{custodyDetails.acquisitionTool || "N/A"}</p>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Hash:</span>
                            <p className="text-muted-foreground font-mono text-xs break-all">{custodyDetails.hash || "N/A"}</p>
                          </div>
                        </div>
                      </div>

                      {/* Recent Activity */}
                      <div className="bg-card p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Clock className="w-5 h-5 text-yellow-400" />
                          Recent Activity
                        </h3>
                        <div className="space-y-3">
                          <div className="flex items-center gap-3 text-sm">
                              {recentThread && (
                                <div className="flex items-center gap-3 text-sm">
                                  <MessageCircle className="w-4 h-4 text-blue-400" />
                                  <div className="flex-1">
                                    <span className="text-muted-foreground">New discussion thread created</span>
                                    <div className="text-xs text-muted-foreground">
                                      by {recentThread.user} • {timeAgo(recentThread.time)}
                                    </div>
                                  </div>
                                </div>
                                )}
                          </div>
                          <div className="flex items-center gap-3 text-sm">
                            <CheckCircle className="w-4 h-4 text-green-400" />
                            <div className="flex-1">
                              <span className="text-muted-foreground">Integrity verification in progress </span>
                              <div className="text-xs text-muted-foreground">System • just now</div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  {activeTab === 'threads' && (
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <h3 className="text-lg font-semibold">Discussion Threads</h3>
                          <div className="flex items-center gap-2">
                            <div className="bg-card p-4 rounded-lg space-y-2">
                              <input
                                type="text"
                                placeholder="Thread title"
                                className="w-full px-3 py-2 bg-card border border-border rounded text-foreground text-sm focus:outline-none focus:border-primary/60"
                                value={newThreadTitle}
                                onChange={(e) => setNewThreadTitle(e.target.value)}
                              />
                              <button
                                className="w-full px-4 py-2 bg-primary text-white rounded hover:bg-primary/60 text-sm"
                                onClick={async () => {
                                  if (!newThreadTitle.trim()) return;
                                  try {
                                    const createdThread = await createAnnotationThread({
                                      case_id: caseId,
                                      file_id: selectedFile?.id || '',
                                      user_id: user.id,
                                      title: newThreadTitle,
                                      tags: [],
                                      priority: "medium"
                                    });

                                    const adaptedThread = {
                                    ...createdThread,
                                    user: profile.name,
                                    avatar: profile.name.split(" ").map((n: string) => n[0]).join("").toUpperCase(),
                                    time: "Just now",
                                    messageCount: 0,
                                    participantCount: createdThread.Participants?.length || 1,
                                   tags: createdThread.Tags || [],
                                  };

                                    console.log("Created thread:", createdThread);
                                    setAnnotationThreads(prev => [...prev, adaptedThread]);
                                    setSelectedThread(adaptedThread);                                    
                                    setNewThreadTitle('');
                                  } catch (err) {
                                    console.error("Failed to create thread:", err);
                                  }

                                }}
                              >
                                Create Thread
                              </button>
                            </div>
                          </div>
                      </div> 
                      
                      {filteredThreads.map((thread) => (
                        <div
                          key={thread.id}
                          className={`border rounded-lg p-4 cursor-pointer transition-all ${
                            selectedThread?.id === thread.id
                              ? 'border-blue-500 bg-blue-600/10'
                              : 'border-muted hover:border-gray-600 hover:bg-muted/50'
                          }`}
                          onClick={() => setSelectedThread(thread)}
                        >
                          <div className="flex items-start justify-between mb-3">
                            <div className="flex items-center gap-3">
                              <div className="w-8 h-8 bg-muted rounded-full flex items-center justify-center text-xs font-medium">
                                {thread.avatar}
                              </div>
                              <div>
                                <h4 className="font-medium text-sm">{thread.title}</h4>
                                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                  <span>{thread.user}</span>
                                  <span>•</span>
                                  <span>{thread.time}</span>
                                </div>
                              </div>
                            </div>
                            <div className="flex items-center gap-2">
                              <span className={`px-2 py-1 rounded text-xs ${getPriorityColor(thread.priority)}`}>
                                {thread.priority}
                              </span>
                              <span className={`px-2 py-1 rounded text-xs ${getStatusColor(thread.status)}`}>
                                {thread.status.replace('_', ' ')}
                              </span>
                            </div>
                          </div>
                          
                          <div className="flex items-center gap-4 text-xs text-muted-foreground">
                            <div className="flex items-center gap-1">
                              <MessageSquare className="w-3 h-3" />
                              <span>{thread.messageCount} messages</span>
                            </div>
                            <div className="flex items-center gap-1">
                              <Users className="w-3 h-3" />
                              <span>{thread.participantCount} participants</span>
                            </div>
                          </div>
                          
                          {Array.isArray(thread.tags) && thread.tags.length > 0 && (
                            <div className="flex items-center gap-2 mt-2">
                              {thread.tags.map((tag, index) => (
                                <span key={index} className="px-2 py-1 b-muted text-muted-foreground rounded text-xs">
                                  {tag.tag_name}
                                </span>
                              ))}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  )}

                  {activeTab === 'metadata' && (
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                      {/* File Metadata */}
                      <div className="bg-card p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Hash className="w-5 h-5 text-cyan-400" />
                          File Metadata
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <span className="text-muted-foreground">File Name:</span>
                              <p className="text-muted-foreground font-mono">{selectedFile.filename}</p>
                            </div>
                            <div>
                              <span className="text-muted-foreground">File Size:</span>
                              <p className="text-muted-foreground">{formatFileSize(selectedFile.file_size)}</p>
                            </div>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Hash Values:</span>
                            <div className="mt-2 space-y-2">
                              {(() => {
                                const hashes = getHashesFromMetadata(selectedFile.metadata);
                                return (
                                  <>
                                    <div className="bg-muted p-2 rounded">
                                      <div className="text-xs text-muted-foreground mb-1">SHA256:</div>
                                      <div className="text-muted-foreground font-mono text-xs break-all">{hashes.sha256}</div>
                                    </div>
                                    <div className="bg-muted p-2 rounded">
                                      <div className="text-xs text-muted-foreground mb-1">MD5:</div>
                                      <div className="text-muted-foreground font-mono text-xs">{hashes.md5}</div>
                                    </div>
                                  </>
                                );
                              })()}
                            </div>
                          </div>

                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <span className="text-muted-foreground">Created:</span>
                              <p className="text-muted-foreground">{new Date(selectedFile.uploaded_at || '').toLocaleString()}</p>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Modified:</span>
                              <p className="text-muted-foreground">{custodyDetails.acquisitionDate ? new Date(custodyDetails.acquisitionDate).toLocaleString() : "N/A"}</p>
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Forensic Metadata */}
                      <div className="bg-card p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Shield className="w-5 h-5 text-amber-400" />
                          Forensic Metadata
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div>
                            <span className="text-muted-foreground">Evidence ID:</span>
                            <p className="text-muted-foreground font-mono">EVD-{selectedFile.id.padStart(6, '0')}</p>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Acquisition Method:</span>
                            <p className="text-muted-foreground">{custodyDetails.method || "N/A"}</p>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Source Device:</span>
                            <p className="text-muted-foreground">{custodyDetails.computerName || "N/A"}</p>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Examiner:</span>
                            <p className="text-muted-foreground">{custodyDetails.examiner || "N/A"}</p>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Case Reference:</span>
                            <p className="text-muted-foreground"> #{`CS-${caseId.slice(0, 7)}...`}</p>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Legal Status:</span>
                            <div className="flex items-center gap-2 mt-1">
                              <CheckCircle className="w-4 h-4 text-green-400" />
                              <span className="text-green-400">{custodyDetails.legalStatus || "N/A"}</span>
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* System Information */}
                      <div className="bg-card p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Info className="w-5 h-5 text-indigo-400" />
                          System Information
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <span className="text-muted-foreground">OS Version:</span>
                              <p className="text-muted-foreground">{custodyDetails.osVersion || "N/A"}</p>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Architecture:</span>
                              <p className="text-muted-foreground">{custodyDetails.architecture || "N/A"}</p>
                            </div>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Computer Name:</span>
                            <p className="text-muted-foreground">{custodyDetails.computerName || "N/A"}</p>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Domain:</span>
                            <p className="text-muted-foreground">{custodyDetails.domain || "N/A"}</p>
                          </div>
                          
                          <div>
                            <span className="text-muted-foreground">Last Boot:</span>
                            <p className="text-muted-foreground">{custodyDetails.lastBoot || "N/A"}</p>
                          </div>
                        </div>
                      </div>

                      {/* Analysis Tools */}
                      <div className="bg-card p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Settings className="w-5 h-5 text-purple-400" />
                          Analysis History
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div className="border-l-2 border-blue-400 pl-3">
                            <div className="font-medium text-muted-foreground">Volatility Analysis</div>
                            <div className="text-muted-foreground text-xs">Completed • 3 hours ago</div>
                            <div className="text-muted-foreground text-xs">Tool: Volatility 3.2.0</div>
                          </div>
                          
                          <div className="border-l-2 border-green-400 pl-3">
                            <div className="font-medium text-muted-foreground">String Extraction</div>
                            <div className="text-muted-foreground text-xs">Completed • 4 hours ago</div>
                            <div className="text-muted-foreground text-xs">Tool: strings (GNU binutils)</div>
                          </div>
                          
                          <div className="border-l-2 border-yellow-400 pl-3">
                            <div className="font-medium text-muted-foreground">Malware Scan</div>
                            <div className="text-muted-foreground text-xs">In Progress • Started 1 hour ago</div>
                            <div className="text-muted-foreground text-xs">Tool: YARA Rules v4.3.2</div>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </>
            )}
          </div>

          {/* Right Sidebar - Thread Messages */}
          {selectedThread && (
            <div className="w-96 border-l border-border bg-background flex flex-col">
              {/* Thread Header */}
              <div className="p-4 border-b border-border">
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-semibold text-sm leading-tight">{selectedThread.title}</h3>
                  <button className="p-1 text-muted-foreground hover:text-foreground">
                    <MoreVertical className="w-4 h-4" />
                  </button>
                </div>
                
                <div className="flex items-center gap-2 mb-3">
                  <div className="w-6 h-6 bg-muted rounded-full flex items-center justify-center text-xs">
                    {selectedThread.avatar}
                  </div>
                  <div className="text-sm">
                    <span className="text-muted-foreground">Created by </span>
                    <span className="font-medium">{selectedThread.user}</span>
                  </div>
                </div>
                
                <div className="flex items-center gap-3 text-xs text-muted-foreground">
                  <span>{selectedThread.time}</span>
                  <span>•</span>
                  <span>{selectedThread.messageCount} messages</span>
                  <span>•</span>
                  <span>{selectedThread.participantCount} participants</span>
                </div>
                
                {selectedThread.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-2">
                    {selectedThread.tags.map((tag, index) => (
                      <span key={index} className="px-2 py-1 b-muted text-muted-foreground rounded text-xs">
                        {tag.tag_name}
                      </span>
                    ))}
                  </div>
                )}
              </div>

             {/* Messages */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {threadMessages.map(message => (
                  <MessageCard
                    key={message.id}
                    message={message}
                    user={user}
                    replyingToMessageId={replyingToMessageId}
                    setReplyingToMessageId={setReplyingToMessageId}
                    replyText={replyText}
                    setReplyText={setReplyText}
                    showReactionPicker={showReactionPicker}
                    setShowReactionPicker={setShowReactionPicker}
                    selectedThread={selectedThread}
                    onSendMessage={handleSendMessageWithParent}
                    onAddReaction={handleAddReaction}
                    onApproveMessage={handleApproveMessage}
                    profile={profile}
                  />
                ))}
              </div>

              {/* Message Input */}
              <div className="p-4 border-t border-border">
                <div className="bg-card rounded-lg p-3">
                  <input
                    type="text"
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    placeholder="Type your message... (@mention users)"
                    className="w-full bg-transparent text-foreground placeholder-gray-400 text-sm focus:outline-none mb-2"
                    onKeyPress={(e) => { 
                      if(e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault();
                        handleSendMessage();
                        }
                      }}
                  />
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span>@ to mention</span>
                      <span>•</span>
                      <span>Shift+Enter for new line</span>
                    </div>
                    <button
                      onClick={() => handleSendMessage()}
                      className="p-1.5 bg-primary text-white rounded hover:bg-primary/60"
                    >
                      <Send className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
          </>
        
        </div>
      </main>
    </div>
  );
}
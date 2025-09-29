import {
  Send,
  FileText,
  Folder,
  Home,
  MessageSquare,
  Menu,
  Plus,
  Search,
  MoreVertical,
  Users,
  Check,
  CheckCheck,
  Paperclip,
  LogOut,
  X,
  Reply,
  Download,
  Eye,
  Trash
} from "lucide-react";
import { Link } from "react-router-dom";
import { useState, useEffect, useRef } from "react";
import { toast } from 'react-hot-toast';
import { MutableRefObject } from "react";
import { ClipboardList } from "lucide-react";


/*----------------------------------------
End-to-End Encryption (E2EE) with X3DH + AES-GCM
-------------------------------------------
*/
import {
  deriveSharedSecretInitiator as deriveSharedSecretInitiatorRaw,
  deriveSharedSecretResponder, fetchBundle
} from '../../lib/crypto/x3dh';

import { encryptMessage, decryptMessage,encryptBytes,decryptBytes,unb64url,toStdB64} from '../../lib/crypto/aes_gcm'; // or your xchacha/secretbox helpers
 // or wherever your helpers live
import { useUserKeys } from '../../store/userKeys';
import { useSharedSecrets } from '../../store/keys';
import sodium from 'libsodium-wrappers';

import { initializeE2EE } from "../../lib/crypto/init-e2ee";

import { Buffer } from 'buffer';
(window as any).Buffer = Buffer;
import { hkdf } from "../../lib/crypto/hkdf";
import { generateGroupSharedSecretForChat } from "../../lib/crypto/groupSecret";
import { ratchetEncrypt, initRatchetState,ratchetDecrypt,ratchetEncryptBytes,ratchetDecryptBytes } from '../../lib/crypto/double_rachet';
import { ChatMember } from "../../lib/crypto/types";

interface Message {
  id: number;
  user: string;
  color: string;
  content: string;
  time: string;
  status: string;
  self?: boolean;
  attachments?: {
    file_name: string;
    file_type: string;
    file_size: number;
    url?: string;
    hash?: string;
    isImage?: boolean;
  }[];
  replyTo?: {
    id: number;
    user: string;
    content: string;
    attachment?: {
      name: string;
      type: string;
    };
  };
}

interface JwtPayload {
  user_id: string;
  email: string;
  role?: string;
  exp: number;
  iat: number;
}
interface Group {
  id: string;
  caseID?: string;
  case_id?: string;
  name: string;
  lastMessage: string;
  lastMessageTime: string;
  unreadCount: number;
  members: ChatMember[]; // Changed from string[]
  group_url: string;
  hasStarted?: boolean;
  caseId?: string;
}
interface Thread {
  thread_id: string;
  title: string;
  case_id: string;
  file_id: string;
  created_by: string;
  created_at: string;
  priority: string;
  new_status?: string;
}

// --- WebSocket Types ---

type WebSocketMessage = {
  type: WebSocketMessageType;
  payload: any;
  groupId?: string;
  userEmail?: string;
  timestamp?: string;
};


type WebSocketMessageType =
  | "new_message"
  | "typing_start"
  | "read_receipt"
  | "message_reaction"
  | "message_reply"
  | "user_joined"
  | "user_left"
  | "file_attachment"
  | "system_alert"
 | "typing_stop"
  | "group_secret_updated";
// Shared type everywhere
export type CryptoEnvelopeV1 = {
  v: 1;
  algo: "aes-gcm";
  ephemeral_pub: string; // base64 x25519
  opk_id?: string; // optional
  nonce: string; // base64
  ct: string; // base64 ciphertext
  header?: { dhPub: string; msgNum: number }; // Double Ratchet header
};

export function generateEphemeralPublicKey(): string {
  sodium.ready;
  const keyPair = sodium.crypto_box_keypair();
  return sodium.to_base64(keyPair.publicKey);
}

const te = new TextEncoder();

/** Hash a UTF-8 string to 32 bytes (SHA-256) for fixed-length salt. */
async function sha256Utf8(s: string): Promise<Uint8Array> {
  const digest = await crypto.subtle.digest("SHA-256", te.encode(s));
  return new Uint8Array(digest);
}

/**
 * Derive a 32-byte "shared secret" from a group of public keys.
 * NOTE: This is NOT secret unless `ikm` includes a private contribution.
 */
export async function deriveSharedSecret(
  groupId: string,
  publicKeys: Uint8Array[]
): Promise<Uint8Array> {
  if (!Array.isArray(publicKeys) || publicKeys.length === 0) {
    throw new Error("No public keys provided for secret derivation.");
  }
  if (!publicKeys.every((k) => k instanceof Uint8Array)) {
    throw new Error("All public keys must be instances of Uint8Array.");
  }
  if (!publicKeys.every((k) => k.length === publicKeys[0].length)) {
    throw new Error("All public keys must have the same length.");
  }

  // Concatenate keys into IKM (input keying material)
  const keyLen = publicKeys[0].length;
  const ikm = new Uint8Array(publicKeys.length * keyLen);
  publicKeys.forEach((k, i) => ikm.set(k, i * keyLen));

  // Salt = H(groupId) to ensure fixed, well-distributed salt bytes
  const salt = await sha256Utf8(groupId);

  // Context string to bind this derivation’s purpose
  const info = te.encode("shared-secret-info");

  // Derive 32-byte key with HKDF(SHA-256)
  const okm = await hkdf(ikm, salt, info, 32);
  return okm;
}

export const connectWebSocket = (
  caseId: string,
  token: string,
  socketRef: MutableRefObject<WebSocket | null>,
  reconnectTimeoutRef: MutableRefObject<ReturnType<typeof setTimeout> | null>,
  onMessage: (msg: WebSocketMessage) => void,
  onOpen?: () => void,
  onClose?: () => void,
  onTypingStatus?: (msg: WebSocketMessage) => void
) => {
  if (!caseId || !token) return;

  // Prevent duplicate connections on the same ref
  const state = socketRef.current?.readyState;
  if (state === WebSocket.OPEN || state === WebSocket.CONNECTING) {
    console.log("📡 WebSocket already connected/connecting.");
    return;
  }

  // Clean up any stale socket before creating a new one
  try { socketRef.current?.close(); } catch { /* noop */ }
  socketRef.current = null;

  // Scheme aware (use wss on https)
  const scheme = (typeof window !== "undefined" && window.location?.protocol === "https:") ? "wss" : "ws";
  const url = `${scheme}://localhost:8080/ws/cases/${encodeURIComponent(caseId)}?token=${encodeURIComponent(token)}`;

  // Backoff state stored on the module-scope-ish ref so retries persist
  const anyRef = socketRef as unknown as MutableRefObject<WebSocket & {
    __retryAttempt?: number;
    __hbTimer?: ReturnType<typeof setInterval> | null;
    __hbTimeout?: ReturnType<typeof setTimeout> | null;
  } | null>;

  const attempt = (anyRef.current?.__retryAttempt ?? 0);

  // Create socket
  const ws = new WebSocket(url);

  // Attach backoff/heartbeat fields
  (ws as any).__retryAttempt = attempt;
  (ws as any).__hbTimer = null;
  (ws as any).__hbTimeout = null;

  // --- Heartbeat helpers (app-level ping/pong) ---
  const clearHeartbeat = () => {
    const cur = socketRef.current as any;
    if (!cur) return;
    if (cur.__hbTimer) { clearInterval(cur.__hbTimer); cur.__hbTimer = null; }
    if (cur.__hbTimeout) { clearTimeout(cur.__hbTimeout); cur.__hbTimeout = null; }
  };

  const startHeartbeat = () => {
    const cur = socketRef.current as any;
    if (!cur) return;

    // send ping every 25s; expect a pong within 10s
    cur.__hbTimer = setInterval(() => {
      if (cur.readyState !== WebSocket.OPEN) return;
      try {
        const ts = Date.now();
        cur.send(JSON.stringify({ type: "ping", payload: { ts } }));
      } catch {/* ignore */}
      // fail the connection if no pong arrives in time
      cur.__hbTimeout = setTimeout(() => {
        console.warn("💔 WS heartbeat timeout; closing socket to trigger reconnect.");
        try { cur.close(); } catch {}
      }, 10000);
    }, 25000);
  };

  // --- Event handlers ---
  ws.onopen = () => {
    console.log("✅ WebSocket connected");
    // reset backoff on a good connection
    (ws as any).__retryAttempt = 0;
    startHeartbeat();
    onOpen?.();
  };

  ws.onmessage = (event) => {
    try {
      const parsed = JSON.parse(event.data);

      // Heartbeat pong
      if (parsed?.type === "pong") {
        // clear the pong timeout
        const cur = socketRef.current as any;
        if (cur?.__hbTimeout) { clearTimeout(cur.__hbTimeout); cur.__hbTimeout = null; }
        // Optional RTT log
        if (parsed?.payload?.ts) {
          const rtt = Date.now() - Number(parsed.payload.ts);
          if (Number.isFinite(rtt)) console.log("🏓 WS RTT:", rtt, "ms");
        }
        return;
      }

      const msg = parsed as WebSocketMessage;
      if (!msg || typeof msg !== "object") throw new Error("Malformed WS message: not an object");
      if (!msg.type || !msg.payload) throw new Error("Malformed WS message: missing type or payload");

      // Latency tracking (best-effort)
      const receivedAt = Date.now();
      const sentAt = Date.parse((msg.payload as any).timestamp || "");
      if (!Number.isNaN(sentAt)) {
        console.log("📥 WS message latency:", receivedAt - sentAt, "ms");
      }

      // Typing events
      if (msg.type === "typing_start" || msg.type === "typing_stop") {
        onTypingStatus?.(msg);
        return;
      }

      onMessage(msg);
    } catch (err) {
      console.error("❌ Error handling WebSocket message:", err);
    }
  };

  ws.onclose = (ev) => {
    clearHeartbeat();
    // Some servers use specific close codes for auth problems
    const authish = [4001, 4003, 4401, 4403];
    if (authish.includes(ev.code)) {
      console.error(`🔒 WS closed with auth-like code ${ev.code}; not reconnecting.`);
      onClose?.();
      return;
    }

    // If a reconnect is already scheduled, skip
    if (reconnectTimeoutRef.current) return;

    // Exponential backoff with jitter
    const prevAttempt = (ws as any).__retryAttempt ?? 0;
    const nextAttempt = prevAttempt + 1;
    const base = 1000;          // 1s
    const max = 30000;          // 30s
    const backoff = Math.min(max, base * Math.pow(2, prevAttempt));
    const jitter = Math.floor(Math.random() * 400); // +0-400ms
    const delay = backoff + jitter;

    console.warn(`⚠️ WebSocket closed (code=${ev.code}, reason="${ev.reason || ""}"). Reconnecting in ${delay}ms...`);
    onClose?.();

    reconnectTimeoutRef.current = setTimeout(() => {
      reconnectTimeoutRef.current = null;
      // bump attempt count for next socket
      (anyRef.current as any)?.__retryAttempt;
      // Re-call connect; carry attempt via ref if available
      // We store attempt on the new ws below after creating it.
      connectWebSocket(
        caseId,
        token,
        socketRef,
        reconnectTimeoutRef,
        onMessage,
        onOpen,
        onClose,
        onTypingStatus
      );
      // Store attempt on the new socket after creation:
      if (socketRef.current) {
        (socketRef.current as any).__retryAttempt = nextAttempt;
      }
    }, delay);
  };

  ws.onerror = (err) => {
    console.error("❌ WebSocket error:", err);
    // Let onclose drive the reconnect/backoff path
    try { ws.close(); } catch {/* noop */}
  };

  // Replace the ref with this socket
  socketRef.current = ws;

  // Optional: listen to online/offline to kick reconnect early
  const handleOnline = () => {
    const st = socketRef.current?.readyState;
    if (st !== WebSocket.OPEN && !reconnectTimeoutRef.current) {
      console.log("🌐 Network online — attempting early reconnect.");
      try { socketRef.current?.close(); } catch {}
    }
  };
  window.addEventListener("online", handleOnline);

  // Clean this listener if/when the socket is replaced
  const cleanupPrev = socketRef.current;
  const cleanup = () => {
    window.removeEventListener("online", handleOnline);
    clearHeartbeat();
  };
  // Attach a small hook so whoever closes this socket triggers cleanup
  (cleanupPrev as any).__cleanup = cleanup;
};


export const SecureChatPage = (): JSX.Element => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [activeChat, setActiveChat] = useState<Group | null>(null);
  const [message, setMessage] = useState("");
  const [showNewGroupModal, setShowNewGroupModal] = useState(false);
  const [newGroupName, setNewGroupName] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [showMoreMenu, setShowMoreMenu] = useState(false);
  const [showChatSearch, setShowChatSearch] = useState(false);
  const [chatSearchQuery, setChatSearchQuery] = useState("");
  const [replyingTo, setReplyingTo] = useState<Message | null>(null);
  const [showAttachmentPreview, setShowAttachmentPreview] = useState(false);
  const [previewFile, setPreviewFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string>("");
  const [attachmentMessage, setAttachmentMessage] = useState("");
  const [showImageModal, setShowImageModal] = useState(false);
  const [modalImageUrl, setModalImageUrl] = useState("");
  const [, setPreviewFileData] = useState<string>("");
  const [typingUsers, setTypingUsers] = useState<Record<string, string[]>>({});
  const [hasMounted, setHasMounted] = useState(false);
  const [] = useState<Thread[]>([]);
  const socketRef = useRef<WebSocket | null>(null);
  const [, setSocketConnected] = useState(false);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const storedUser = sessionStorage.getItem("user");
  const user = storedUser ? JSON.parse(storedUser) : null;
  const [role, setRole] = useState<string>(user?.role || "");
  const isDFIRAdmin = role === "DFIR Admin";
  const [userId] = useState(() => {
    try {
      const token = sessionStorage.getItem("authToken");
      if (!token) return null;

      const base64Payload = token.split(".")[1];
      const decodedPayload = JSON.parse(atob(base64Payload)) as JwtPayload;
      return decodedPayload?.user_id || null;
    } catch (err) {
      console.error("❌ Failed to decode token:", err);
      return null;
    }
  });
    const [token] = useState(sessionStorage.getItem("authToken"));
  interface Case {
    id: string;
    title?: string;
    // Add other properties if needed
  }
  const [activeCases, setActiveCases] = useState<Case[]>([]);
  const [selectedCaseId, setSelectedCaseId] = useState("");
  // Removed unused setRetryCount state

  const [editGroupName, setEditGroupName] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editIsPublic, setEditIsPublic] = useState(false);
  const [showEditGroupModal, setShowEditGroupModal] = useState(false);

  // Mock data for groups and messages
  const [groups, setGroups] = useState<Group[]>([]);

  type ChatMessages = Record<string, Message[]>; // <-- string keys
 const [chatMessages, setChatMessages] = useState<ChatMessages>(() => {
  const initial: ChatMessages = {};
  const savedGroups = localStorage.getItem('chatGroups');
  if (savedGroups) {
    const groups: Group[] = JSON.parse(savedGroups);
    groups.forEach(group => {
      initial[group.id] = [];
    });
  }
  return initial;
});
  const B64 = sodium.base64_variants.ORIGINAL;

  // Track any blob URLs we create so we can revoke them later safely
  const blobUrlsRef = useRef<string[]>([]);
  const trackBlobUrl = (u: string) => { blobUrlsRef.current.push(u); };

  // Revoke tracked URLs on unmount
  useEffect(() => {
    return () => {
      blobUrlsRef.current.forEach(u => URL.revokeObjectURL(u));
      blobUrlsRef.current = [];
    };
  }, []);

  useEffect(() => {
  return () => {
    if (activeChat?.id) {
      const groupTimeouts = typingTimeoutsRef.current[activeChat.id] || {};
      Object.values(groupTimeouts).forEach(clearTimeout);
      delete typingTimeoutsRef.current[activeChat.id];
    }
  };
}, [activeChat?.id]);

  // somewhere in app bootstrap (e.g., in a useEffect in SecureChatPage)
  useEffect(() => {
    const token = sessionStorage.getItem("authToken");
    if (!token) return;

    // decode your JWT the way you already do:
    const userId = String(user.id);
    const tenantId = /* if you have multi-tenant */ undefined;

    initializeE2EE({ userId, tenantId, minOPKs: 10, targetOPKs: 40 })
      .catch((e) => {
        console.error("E2EE init failed:", e);
      });
  }, []);

  useEffect(() => {
  const initKeys = async () => {
    if (userId) {
      await useUserKeys.getState().initializeKeys(userId);
      await useUserKeys.getState().loadPersistedKeys();
    }
  };
  initKeys();
}, [userId]);



 function verifySpkSignature(ikPubEd_b64: string, spkPubX_b64: string, sig_b64: string) {
    const ikPubEd = sodium.from_base64(ikPubEd_b64, B64);
    const spkPubX = sodium.from_base64(spkPubX_b64, B64);
    const sig = sodium.from_base64(sig_b64, B64);
    return sodium.crypto_sign_verify_detached(sig, spkPubX, ikPubEd);
  }


async function e2eeDecryptText(
  ctB64u: string,
  nB64u: string,
  secret: Uint8Array
) {
  // decryptMessage(key, nonceB64u, ciphertextB64u)
  return await decryptMessage(secret, nB64u, ctB64u);
}

async function e2eeDecryptBytes(
  ctB64u: string,
  nB64u: string,
  secret: Uint8Array
) {
  // decryptBytes(key, nonceB64u, ciphertextB64u)
  return await decryptBytes(secret, nB64u, ctB64u);
}




  // Replace your current fetchBundle with this UUID-only version.
  type ServerBundleWire = {
    identity_key?: string;
    signed_prekey?: string;
    spk_signature?: string;
    one_time_prekey?: string;
    opk_id?: string;
    opk?: { id?: string; publicKey?: string } | null;

    // also accept camelCase from newer backend
    identityKey?: string;
    signedPreKey?: string;
    spkSignature?: string;
    oneTimePreKey?: string;
  };


  /** UUID-only bundle fetch. */
  // put near fetchBundle()
  async function resolvePeerIdentifier(identifier: string): Promise<string> {
    // If already UUID-ish, return as-is
    if (/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(identifier)) {
      return identifier;
    }
    // If it's an email and your API supports resolving:
    // (Skip this block if your backend already accepts email on /bundle/:id)
    // const r = await fetch(`${API}/api/v1/users/resolve?email=${encodeURIComponent(identifier)}`, { headers:{ Authorization: `Bearer ${sessionStorage.getItem("authToken")||""}` } });
    // if (r.ok) { const j = await r.json(); return j.user_id; }
    return identifier;
  }

let bundleFetchInFlight: { [key: string]: Promise<ServerBundleWire> } = {};








useEffect(() => {
  const loadGroupDetails = async () => {
    if (!activeChat?.caseId || !token) return;

    try {
      const res = await fetch(`http://localhost:8080/api/v1/cases/${activeChat.caseId}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (!res.ok) {
        console.error("Failed to fetch group details:", await res.text());
        toast.error("Failed to load group details.");
        return;
      }
      const groupData: {
        id: string;
        case_id: string;
        name: string;
        members: { id: string; email: string }[];
        group_url?: string;
        lastMessage?: string;
        lastMessageTime?: string;
        unreadCount?: number;
        hasStarted?: boolean;
      } = await res.json();
      setActiveChat(prev => ({
        ...prev,
        id: groupData.id,
        caseId: groupData.case_id,
        case_id: groupData.case_id,
        name: groupData.name,
        members: groupData.members.map((m: any) => ({
          id: m.id, // UUID
          email: m.email // Optional email for display
        })),
        group_url: groupData.group_url || "/default-group-avatar.png",
        lastMessage: groupData.lastMessage || "",
        lastMessageTime: groupData.lastMessageTime || "now",
        unreadCount: groupData.unreadCount || 0,
        hasStarted: groupData.hasStarted || true
      }));
    } catch (err) {
      console.error("Error loading group details:", err);
      toast.error("Failed to load group details.");
    }
  };
  loadGroupDetails();
}, [activeChat?.caseId, token]);

// ... Inside SecureChatPage
const handleSendMessage = async (e?: React.MouseEvent | React.KeyboardEvent) => {
  e?.preventDefault();
  if (!activeChat || !message.trim() || !userEmail) return;

  const chatId = String(activeChat.id);
  const token = sessionStorage.getItem("authToken") || "";

  try {
    let ratchetState = useSharedSecrets.getState().getSharedSecret(chatId);
    let ephemeralPubB64u: string | undefined;
    let opkIdUsed: string | undefined;

    if (!ratchetState) {
      const chatForSecret: Chat = {
        id: chatId,
        members: activeChat.members.map(m => ({ id: m.id, email: m.email })), // Correct mapping
        name: activeChat.name
      };
      const res = await generateGroupSharedSecretForChat(chatForSecret, token, userId || "");
      if (!res || !res.sharedSecret) {
        toast.error("Failed to establish secure session. Ensure all members have registered key bundles.");
        return;
      }
      ratchetState = await initRatchetState(res.sharedSecret);
      ephemeralPubB64u = res.ephemeralPubKey || "";
      opkIdUsed = res.opkIdUsed || "";
      useSharedSecrets.getState().setSharedSecret(chatId, ratchetState);
    }

    const { header, ciphertext, nonce } = await ratchetEncrypt(ratchetState, message);
    const envelope: CryptoEnvelopeV1 = {
      v: 1,
      algo: "aes-gcm",
      ephemeral_pub: ephemeralPubB64u || header.dhPub,
      ...(opkIdUsed ? { opk_id: opkIdUsed } : {}),
      nonce,
      ct: ciphertext,
      header
    };

    const res = await fetch(`http://localhost:8080/api/v1/chat/groups/${chatId}/messages`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({
        sender_email: userEmail,
        sender_name: "You",
        message_type: "text",
        is_encrypted: true,
        envelope,
        content: "",
        ...(replyingTo && {
          reply_to: {
            id: replyingTo.id,
            user: replyingTo.user
          }
        })
      })
    });

    if (!res.ok) {
      toast.error("Message failed to send.");
      return;
    }

    const saved = await res.json();

    const socket = socketRef.current;
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        type: "new_message",
        payload: {
          messageId: saved.id,
          groupId: chatId,
          senderEmail: userEmail,
          senderName: "You",
          message_type: "text",
          is_encrypted: true,
          envelope,
          timestamp: new Date(saved.created_at || Date.now()).toISOString()
        }
      }));
    }

    const uiMsg: Message = {
      id: saved.id || Date.now(),
      user: "You",
      self: true,
      color: "text-blue-400",
      content: message,
      time: new Date(saved.created_at || Date.now()).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
      status: "sent",
      ...(replyingTo && {
        replyTo: {
          id: replyingTo.id,
          user: replyingTo.user,
          content: replyingTo.content,
          ...(replyingTo.attachments?.[0] && {
            attachment: {
              name: replyingTo.attachments[0].file_name,
              type: replyingTo.attachments[0].file_type
            }
          })
        }
      })
    };

    setChatMessages(prev => ({
      ...prev,
      [chatId]: [...(prev[chatId] || []), uiMsg]
    }));

    setGroups(prev => prev.map(g =>
      g.id === activeChat.id
        ? { ...g, lastMessage: uiMsg.content, lastMessageTime: "now" }
        : g
    ));

    setMessage("");
    setReplyingTo(null);
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        type: "typing_stop",
        payload: { userEmail },
        groupId: chatId,
        userEmail
      }));
    }
  } catch (err: unknown) {
    console.error("handleSendMessage error:", err);
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    toast.error(`Message failed: ${errorMessage}`);
  }
};

// Generate shared secret for group
// Define a Chat member type
interface ChatMember {
  id: string;
  email?: string;
  name?: string;
  publicKey?: Uint8Array; 
}

// Define a Chat type
interface Chat {
  id: string;              // group ID
  members: ChatMember[];   // array of members
  name?: string;           // optional group name
}






  // --- Receiving a message ---



  // replace your current ensureOPKs
let ensureOPKsInFlight = false;









  const handleTypingStatus = (msg: WebSocketMessage) => {
    const { type, payload } = msg;
    const groupId = activeChat?.id;
    if (!groupId || payload.userEmail === userEmail) return;

    if (type === "typing_start") {
      // Add user to typing list (without duplicates)
      setTypingUsers(prev => {
        const users = new Set([...(prev[groupId] || []), payload.userEmail]);
        return { ...prev, [groupId]: Array.from(users) };
      });

      // Clear existing timeout
      if (typingTimeoutsRef.current[groupId]?.[payload.userEmail]) {
        clearTimeout(typingTimeoutsRef.current[groupId][payload.userEmail]);
      }

      // Set new timeout to remove user
      const timeout = setTimeout(() => {
        setTypingUsers(prev => ({
          ...prev,
          [groupId]: (prev[groupId] || []).filter(u => u !== payload.userEmail),
        }));
      }, 3000);

      // Store the timeout
      typingTimeoutsRef.current[groupId] = {
        ...(typingTimeoutsRef.current[groupId] || {}),
        [payload.userEmail]: timeout,
      };
    }
  };
const handleSelectGroup = (group: any) => {
  const id: string = group.id || group._id || "";
  setActiveChat({
    id,
    caseId: group.case_id || group.caseId || "",
    case_id: group.case_id || group.caseId || "",
    name: group.name || "Unnamed Group",
    members: (group.members || []).map((m: any) => ({
      id: m.id || m.user_id || (typeof m === 'string' ? m : ""),
      email: m.email || m.user_email || (typeof m === 'string' ? m : "")
    })),
    group_url: group.group_url || getAvatar(String(id)),
    lastMessage: group.lastMessage || "",
    lastMessageTime: group.lastMessageTime || "now",
    unreadCount: group.unreadCount || 0,
    hasStarted: group.hasStarted || true
  });
  localStorage.setItem("activeChat", JSON.stringify({
    id,
    caseId: group.case_id || group.caseId || "",
    case_id: group.case_id || group.caseId || "",
    name: group.name || "Unnamed Group",
    members: (group.members || []).map((m: any) => ({
      id: m.id || m.user_id || (typeof m === 'string' ? m : ""),
      email: m.email || m.user_email || (typeof m === 'string' ? m : "")
    })),
    group_url: group.group_url || getAvatar(String(id)),
    lastMessage: group.lastMessage || "",
    lastMessageTime: group.lastMessageTime || "now",
    unreadCount: group.unreadCount || 0,
    hasStarted: group.hasStarted || true
  }));
};
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


  useEffect(() => {
    const fetchActiveCases = async () => {
      try {
        const res = await fetch("http://localhost:8080/api/v1/cases/filter?status=open", {
          headers: {
            Authorization: `Bearer ${sessionStorage.getItem("authToken") || ""}`,
          },
        });
        const data = await res.json();
        setActiveCases(data.cases || []);
      } catch (err) {
        console.error("Error fetching cases:", err);
      }
    };

    fetchActiveCases();
  }, []);


  const chatEndRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const moreMenuRef = useRef<HTMLDivElement>(null);

  const [showAddMembersModal, setShowAddMembersModal] = useState(false);
  const [newMemberEmail, setNewMemberEmail] = useState("");

  const [availableUsers, setAvailableUsers] = useState<{ user_email: string, role: string }[]>([]);



  const [userEmail] = useState(() => {
    try {
      const token = sessionStorage.getItem("authToken");
      if (!token) return null;

      const base64Payload = token.split(".")[1];
      const decodedPayload = JSON.parse(atob(base64Payload)) as JwtPayload;
      return decodedPayload?.email || null;
    } catch {
      return null;
    }
  });




  const filteredGroups = groups.filter(group =>
    group.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Filter messages based on chat search
  const filteredMessages = activeChat && chatMessages[activeChat.id]
    ? chatMessages[activeChat.id].filter(msg =>
      msg.content.toLowerCase().includes(chatSearchQuery.toLowerCase())
    )
    : [];

  const displayMessages = (showChatSearch && chatSearchQuery 
  ? filteredMessages 
  : (activeChat && chatMessages[activeChat.id] ? chatMessages[activeChat.id] : [])) || [];

  const scrollToBottom = () => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [chatMessages, activeChat]);

  // Close more menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (moreMenuRef.current && !moreMenuRef.current.contains(event.target as Node)) {
        setShowMoreMenu(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  useEffect(() => {
    const savedSidebar = localStorage.getItem("sidebarOpen");
    if (savedSidebar) {
      setSidebarOpen(savedSidebar === "true");
    }
  }, []);

  useEffect(() => {
    localStorage.setItem("sidebarOpen", sidebarOpen.toString());
  }, [sidebarOpen]);

  useEffect(() => {
    setHasMounted(true);
  }, []);




  const fileInputGroupRef = useRef<HTMLInputElement>(null);

  const handleGroupImageClick = () => {
    if (activeChat?.id) {
      fileInputGroupRef.current?.click();
    }
  };

  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const typingTimeoutsRef = useRef<{
    [groupId: string]: { [email: string]: ReturnType<typeof setTimeout> };
  }>({});

  const sendTypingNotification = (type: "typing_start" | "typing_stop") => {
    if (!activeChat?.id || !socketRef.current) return;

    const message = {
      type,
      payload: { userEmail },
      groupId: String(activeChat.id),
      userEmail,
    };

    socketRef.current.send(JSON.stringify(message));

    if (type === "typing_start") {
      // Debounce sending "typing_stop"
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);

      typingTimeoutRef.current = setTimeout(() => {
        const stopMessage = {
          type: "typing_stop",
          payload: { userEmail },
          groupId: String(activeChat.id),
          userEmail,
        };
        socketRef.current?.send(JSON.stringify(stopMessage));
      }, 3000); // 3s of inactivity
    }
  };





  useEffect(() => {
    return () => {
      Object.values(typingTimeoutsRef.current).forEach(group =>
        Object.values(group).forEach(timeout => clearTimeout(timeout))
      );
    };
  }, []);



  const handleGroupImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !activeChat) return;

    const formData = new FormData();
    formData.append('group_url', file);

    try {
      const res = await fetch(`http://localhost:8080/api/v1/chat/groups/${activeChat.id}/image`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData
      });

      if (!res.ok) throw new Error("Upload failed");

      const data = await res.json();
      const newImageUrl = data.group_url;

      setActiveChat(prev =>
        prev ? { ...prev, group_url: newImageUrl } : prev
      );

      setGroups(prev =>
        prev.map(group =>
          group.id === activeChat.id
            ? { ...group, group_url: newImageUrl }
            : group
        )
      );

    } catch (err) {
      console.error("Failed to upload group image", err);
    }
  };


  // Save to localStorage
  useEffect(() => {
    const hasRealMessages = Object.values(chatMessages).some(msgs => (msgs || []).length > 0);

    const activeGroups = groups.filter(g => g.hasStarted);

    if (activeGroups.length > 0 || hasRealMessages) {
      localStorage.setItem('chatGroups', JSON.stringify(activeGroups));
      //  localStorage.setItem('chatMessages', JSON.stringify(chatMessages));
    }
  }, [groups, chatMessages]);


  useEffect(() => {
    const meaningfulGroups = groups.filter(g => g.hasStarted);
    if (meaningfulGroups.length > 0) {
      localStorage.setItem('chatGroups', JSON.stringify(meaningfulGroups));
    }
  }, [groups]);

  useEffect(() => {
    const savedGroups = localStorage.getItem('chatGroups');
    // const savedMessages = localStorage.getItem('chatMessages');

    if (savedGroups) setGroups(JSON.parse(savedGroups));
    // if (savedMessages) setChatMessages(JSON.parse(savedMessages));
  }, []);
const fetchGroups = async () => {
  console.log("Calling fetchGroups with:", userEmail, token);
  if (!token || !userEmail) {
    console.warn("No token or userEmail, cannot fetch groups.");
    setGroups([]);
    setChatMessages({});
    return;
  }

  try {
    const res = await fetch(`http://localhost:8080/api/v1/chat/groups/user/${encodeURIComponent(userEmail)}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    console.log("Response status:", res.status);
    if (!res.ok) {
      console.error("Fetch failed:", await res.text());
      setGroups([]);
      setChatMessages({});
      return;
    }

    const data = await res.json();
    console.log("Fetched groups data:", data);

    const groupsWithAvatars: Group[] = (Array.isArray(data) ? data : data.groups || []).map((g: any) => ({
      id: g.id || g._id || "",
      caseId: g.case_id || g.caseId || "",
      case_id: g.case_id || g.caseId || "",
      name: g.name || "Unnamed Group",
      members: (g.members || []).map((m: any) => ({
        id: m.id || m.user_id || "",
        email: m.email || m.user_email || ""
      })),
      group_url: g.group_url || getAvatar(String(g.id || g._id || "")),
      lastMessage: g.lastMessage || "",
      lastMessageTime: g.lastMessageTime || "now",
      unreadCount: g.unreadCount || 0,
      hasStarted: g.hasStarted || true
    }));

    setGroups(groupsWithAvatars);
    // Initialize chatMessages for each group
    setChatMessages(prev => {
      const updated: ChatMessages = { ...prev };
      groupsWithAvatars.forEach(group => {
        if (!updated[group.id]) {
          updated[group.id] = [];
        }
      });
      return updated;
    });
  } catch (err: unknown) {
    console.error("Failed to fetch groups:", err);
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    toast.error(`Failed to fetch groups: ${errorMessage}`);
    setGroups([]);
    setChatMessages({});
  }
};
  useEffect(() => {
    if (!userEmail || !token) return;

    const normalizedEmail = userEmail.trim().toLowerCase();
    if (normalizedEmail) {
      fetchGroups();
    }
  }, [userEmail, token]);



useEffect(() => {
  const interval = setInterval(async () => {
    if (!activeChat || !userId) return;
    const ratchetState = useSharedSecrets.getState().getSharedSecret(activeChat.id);
    if (ratchetState && ratchetState.sendCount > 100) {
      try {
        const token = sessionStorage.getItem("authToken") || "";
        const chatForSecret: Chat = {
          id: activeChat.id,
          members: activeChat.members.map(m => ({ id: m.id, email: m.email })), // Correct mapping
          name: activeChat.name
        };

        const res = await generateGroupSharedSecretForChat(chatForSecret, token, userId);
        if (!res || !res.sharedSecret) {
          toast.error("Failed to rotate group shared secret.");
          return;
        }

        const newRatchetState = await initRatchetState(res.sharedSecret);
        useSharedSecrets.getState().setSharedSecret(activeChat.id, newRatchetState);
        console.log("Periodic key rotation performed");
      } catch (err: unknown) {
        console.error("Key rotation error:", err);
        const errorMessage = err instanceof Error ? err.message : "Unknown error";
        toast.error(`Failed to rotate keys: ${errorMessage}`);
      }
    }
  }, 24 * 60 * 60 * 1000); // Check daily
  return () => clearInterval(interval);
}, [activeChat, userId]);
  // Robust normalizer (no unsafe type assertions)
  function normalizeIncomingBundle(raw: ServerBundleWire) {
    const identityKeyEd =
      raw.identityKey ?? raw.identity_key ?? raw.identityKey; // ed25519 pub

    const spkPublicX =
      raw.signedPreKey ?? raw.signed_prekey;

    const spkSignature =
      raw.spkSignature ?? raw.spk_signature;

    const opk =
      raw.opk ?? (raw.one_time_prekey ? { publicKey: raw.one_time_prekey } : null);

    if (!identityKeyEd) throw new Error("Bundle missing identity key");
    return { identityKeyEd, spkPublicX, spkSignature, opk };
  }


useEffect(() => {
  if (!activeChat?.caseId || !token) return;

  connectWebSocket(
    activeChat.caseId,
    token,
    socketRef,
    reconnectTimeoutRef,
    async (msg: WebSocketMessage) => {
      if (msg.type === "group_secret_updated" && String(msg.payload.groupId) === String(activeChat.id)) {
        const chatForSecret: Chat = {
          id: String(activeChat.id),
          members: activeChat.members,
          name: activeChat.name
        };
        const res = await generateGroupSharedSecretForChat(chatForSecret, token, userId || "");
        if (res && res.sharedSecret) {
          const newRatchetState = await initRatchetState(res.sharedSecret);
          useSharedSecrets.getState().setSharedSecret(String(activeChat.id), newRatchetState);
          console.log("Group shared secret updated for new member");
        } else {
          toast.error("Failed to update group shared secret.");
        }
        return;
      }

      if (msg.type !== "new_message" || String(msg.payload.groupId) !== String(activeChat.id)) return;

      const incoming = msg.payload;
      let displayedText = incoming.content || "";
      let attachmentsForUI: Message["attachments"] = [];

      try {
        if (incoming.is_encrypted && incoming.envelope) {
          const envelope = incoming.envelope as CryptoEnvelopeV1;
          let ratchetState = useSharedSecrets.getState().getSharedSecret(String(incoming.groupId));

          if (!ratchetState) {
            const { ikPriv: ourIKPrivEd, spkPriv: ourSPKPrivX, opks: ourOPKs } = useUserKeys.getState();
            if (!ourIKPrivEd || !ourSPKPrivX) throw new Error("Missing private keys.");

            const ephPubX = base64ToU8((envelope.ephemeral_pub || "").trim());
            const opkId: string | undefined = envelope.opk_id;
            const senderId = incoming.senderEmail || incoming.from || "";

            if (!senderId) throw new Error("Missing sender identity");
            const raw = await fetchBundle(senderId,userId || "", token);
            const nb = normalizeIncomingBundle(raw);

            const ok = nb.spkSignature && nb.spkPublicX && verifySpkSignature(nb.identityKeyEd, nb.spkPublicX, nb.spkSignature);
            if (!ok) throw new Error("Sender SPK signature invalid");

            const senderIKPubEd = sodium.from_base64(nb.identityKeyEd);
            const senderIKPubX = sodium.crypto_sign_ed25519_pk_to_curve25519(senderIKPubEd);
            const ourIKPrivX = sodium.crypto_sign_ed25519_sk_to_curve25519(ourIKPrivEd);

            let ourOPKPrivX: Uint8Array | undefined;
            if (opkId) {
              const found = (useUserKeys.getState().opks || []).find(o => o.id === opkId);
              ourOPKPrivX = found?.priv;
            }

            const shared = await deriveSharedSecretResponder(
              u8Fresh(ourIKPrivX),
              u8Fresh(ourSPKPrivX),
              u8Fresh(ephPubX),
              u8Fresh(senderIKPubX),
              ourOPKPrivX ? u8Fresh(ourOPKPrivX) : undefined
            );

            if (!shared) throw new Error("Failed to derive shared secret");
            ratchetState = await initRatchetState(shared);
            useSharedSecrets.getState().setSharedSecret(String(incoming.groupId), ratchetState);
          }

          if (incoming.message_type === "attachment") {
            const bytes = await ratchetDecryptBytes(ratchetState, envelope.header || { dhPub: "", msgNum: 0 }, envelope.ct, envelope.nonce);
            const mime = incoming.file_mime || "application/octet-stream";
            const blob = new Blob([u8ToArrayBuffer(bytes)], { type: mime });
            const blobUrl = URL.createObjectURL(blob);
            trackBlobUrl(blobUrl);
            displayedText = incoming.content || "";
            attachmentsForUI = [{
              file_name: incoming.file_name || "file",
              file_type: mime,
              file_size: incoming.file_size || bytes.byteLength,
              url: blobUrl,
              isImage: mime.startsWith("image/")
            }];
          } else {
            displayedText = await ratchetDecrypt(ratchetState, envelope.header || { dhPub: "", msgNum: 0 }, envelope.ct, envelope.nonce);
          }

          if (envelope.opk_id) {
            try { await useUserKeys.getState().markOPKUsed(envelope.opk_id, true); } catch {}
          }
        } else {
          if (incoming.message_type === "attachment" && incoming.file_url) {
            const mime = incoming.file_mime || "application/octet-stream";
            attachmentsForUI = [{
              file_name: incoming.file_name || "file",
              file_type: mime,
              file_size: incoming.file_size || 0,
              url: incoming.file_url,
              isImage: mime.startsWith("image/")
            }];
          }
          displayedText = incoming.content || "";
        }
      } catch (err: unknown) {
        console.warn("Failed to decrypt message:", err);
        const errorMessage = err instanceof Error ? err.message : "Unknown error";
        displayedText = `[Decryption failed: ${errorMessage}]`;
        try {
          const chatForSecret: Chat = {
            id: String(incoming.groupId),
            members: activeChat.members,
            name: activeChat.name
          };
          const res = await generateGroupSharedSecretForChat(chatForSecret, token, userId || "");
          if (!res || !res.sharedSecret) throw new Error("Failed to rotate group shared secret.");
          const newRatchetState = await initRatchetState(res.sharedSecret);
          useSharedSecrets.getState().setSharedSecret(String(incoming.groupId), newRatchetState);
          toast.error("Session refreshed due to key issue. Please retry.");
        } catch (rotationErr: unknown) {
          console.error("Key rotation failed:", rotationErr);
          const rotationErrorMessage = rotationErr instanceof Error ? rotationErr.message : "Unknown error";
          toast.error(`Key rotation failed: ${rotationErrorMessage}`);
        }
      }

      const mappedMessage: Message = {
        id: incoming.messageId,
        user: incoming.senderName,
        color: incoming.senderEmail === userEmail ? "text-green-400" : "text-blue-400",
        content: displayedText,
        time: new Date(incoming.timestamp).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
        status: "read",
        self: incoming.senderEmail === userEmail,
        attachments: attachmentsForUI
      };

      setChatMessages(prev => {
        const existing = prev[activeChat.id] || [];
        if (existing.some(m => m.id === mappedMessage.id)) return prev;
        return { ...prev, [activeChat.id]: [...existing, mappedMessage] };
      });
    },
    () => setSocketConnected(true),
    () => setSocketConnected(false),
    handleTypingStatus
  );

  return () => {
    if (reconnectTimeoutRef.current) { clearTimeout(reconnectTimeoutRef.current); reconnectTimeoutRef.current = null; }
    try { socketRef.current?.close(); } catch {}
  };
}, [activeChat?.caseId, activeChat?.id, token, userId]);
  const handleFileSelection = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files || files.length === 0) return;

    const file = files[0];

    const url = URL.createObjectURL(file);
    setPreviewFile(file);
    setPreviewUrl(url);
    setShowAttachmentPreview(true);
    setAttachmentMessage("");

    // Store base64 
    const fileData = await new Promise<string>((resolve) => {
      const reader = new FileReader();
      reader.onload = (e) => resolve(e.target?.result as string);
      reader.readAsDataURL(file);
    });

    setPreviewFileData(fileData);
  };
  const meaningfulGroups = groups.filter(g => g.hasStarted);
  if (meaningfulGroups.length > 0) {
    localStorage.setItem('chatGroups', JSON.stringify(meaningfulGroups));
  }



  const handleCancelAttachment = () => {
    setShowAttachmentPreview(false);
    setPreviewFile(null);
    setAttachmentMessage("");

    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
      setPreviewUrl(""); // <-- only after revoking
    }
  };


// make sure this import (or equivalent) exists somewhere near your other crypto imports:
// import { encryptBytes } from "src/lib/crypto/aesgcm"; // adjust path to your aesgcm.ts
// base64url -> Uint8Array
function unb64url(s: string): Uint8Array {
  const pad = s.length % 4 === 2 ? "==" : s.length % 4 === 3 ? "=" : "";
  const base64 = s.replace(/-/g, "+").replace(/_/g, "/") + pad;
  const bin = atob(base64);
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
  return out;
}





// ... Other imports remain unchanged

const sendAttachment = async () => {
  if (!activeChat || !previewFile) return;

  const chatId = String(activeChat.id);
  const groupId = chatId;
  const token = sessionStorage.getItem("authToken") || "";
  const userId = user.id;

  let envelope: CryptoEnvelopeV1;
  try {
    let ratchetState = useSharedSecrets.getState().getSharedSecret(groupId);
    let ephPubB64u: string | undefined;
    let opkIdUsed: string | undefined;

    if (!ratchetState) {
      const chatForSecret: Chat = {
        id: chatId,
        members: activeChat.members.map(m => ({ id: m.id, email: m.email })), // Correct mapping
        name: activeChat.name
      };
      const res = await generateGroupSharedSecretForChat(chatForSecret, token, userId || "");
      if (!res || !res.sharedSecret) {
        console.error("Failed to derive group shared secret");
        toast.error("Failed to establish secure session. Ensure all members have registered key bundles.");
        return;
      }
      ratchetState = await initRatchetState(res.sharedSecret);
      ephPubB64u = res.ephemeralPubKey || "";
      opkIdUsed = res.opkIdUsed || "";
      useSharedSecrets.getState().setSharedSecret(groupId, ratchetState);
    }

    const rawBytes = new Uint8Array(await previewFile.arrayBuffer());
    const { header, nonce, ciphertext } = await ratchetEncryptBytes(ratchetState, rawBytes);
    const ctU8 = unb64url(ciphertext);
    const fileStdB64 = toStdB64(ctU8);

    envelope = {
      v: 1,
      algo: "aes-gcm",
      ephemeral_pub: ephPubB64u || header.dhPub,
      ...(opkIdUsed ? { opk_id: opkIdUsed } : {}),
      nonce,
      ct: ciphertext,
      header
    };

    console.log("Sending attachment payload:", {
      file: fileStdB64,
      fileName: previewFile.name,
      file_mime: previewFile.type,
      file_size: previewFile.size,
      content: attachmentMessage || "",
      envelope
    });

    const res = await fetch(`http://localhost:8080/api/v1/chat/groups/${activeChat.id}/messages`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({
        sender_email: userEmail,
        sender_name: "You",
        message_type: "attachment",
        is_encrypted: true,
        file: fileStdB64,
        fileName: previewFile.name,
        file_mime: previewFile.type || "application/octet-stream",
        file_size: previewFile.size,
        envelope,
        content: attachmentMessage || ""
      })
    });

    if (!res.ok) {
      const txt = await res.text().catch(() => "");
      console.error("Failed to send encrypted attachment:", res.status, txt);
      toast.error("Failed to send attachment.");
      return;
    }

    const saved = await res.json();
    console.log("Backend response:", saved);

    const socket = socketRef.current;
    if (socket?.readyState === WebSocket.OPEN) {
      const payload = {
        type: "new_message",
        payload: {
          messageId: saved.id,
          groupId: chatId,
          senderEmail: userEmail,
          senderName: "You",
          message_type: "attachment",
          is_encrypted: true,
          envelope,
          file_name: previewFile.name,
          file_mime: previewFile.type,
          file_size: previewFile.size,
          content: saved.content || attachmentMessage || "",
          timestamp: new Date(saved.created_at || Date.now()).toISOString()
        }
      };
      console.log("WebSocket broadcast:", payload);
      socket.send(JSON.stringify(payload));
    }

    const messageBlobUrl = URL.createObjectURL(previewFile);
    trackBlobUrl(messageBlobUrl);

    const optimistic: Message = {
      id: saved.id || Date.now(),
      user: "You",
      self: true,
      color: "text-blue-400",
      content: saved.content || attachmentMessage || `Shared file: ${previewFile.name}`,
      time: new Date(saved.created_at || Date.now()).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
      status: "sent",
      attachments: [{
        file_name: previewFile.name,
        file_type: previewFile.type || "application/octet-stream",
        file_size: previewFile.size,
        url: messageBlobUrl,
        isImage: (previewFile.type || "").startsWith("image/")
      }],
      ...(replyingTo && {
        replyTo: {
          id: replyingTo.id,
          user: replyingTo.user,
          content: replyingTo.content,
          ...(replyingTo.attachments?.[0] && {
            attachment: {
              name: replyingTo.attachments[0].file_name,
              type: replyingTo.attachments[0].file_type
            }
          })
        }
      })
    };

    setChatMessages(prev => {
      const cur = prev[activeChat.id] || [];
      return { ...prev, [activeChat.id]: [...cur, optimistic] };
    });

    setGroups(prev => prev.map(group =>
      group.id === activeChat.id
        ? { ...group, lastMessage: optimistic.content, lastMessageTime: "now" }
        : group
    ));

  } catch (err: unknown) {
    console.error("Failed to send attachment:", err);
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    toast.error(`Attachment failed: ${errorMessage}`);
  } finally {
    setShowAttachmentPreview(false);
    setPreviewFile(null);
    setAttachmentMessage("");
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
      setPreviewUrl("");
    }
    setReplyingTo(null);
  }
};


  function getAvatar(_groupId: string): string {
    return ""; // Let the fallback image handle it
  }

  const handleCreateGroup = async (e?: React.MouseEvent | React.KeyboardEvent) => {
    e?.preventDefault();

    if (!newGroupName.trim() || !selectedCaseId) return;

    try {
      const res = await fetch('http://localhost:8080/api/v1/chat/groups', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          name: newGroupName,
          description: "Group created from frontend",
          type: "group",
          case_id: selectedCaseId,
          created_by: userEmail,
          members: [{ user_email: userEmail, role: "admin" }],
          settings: { is_public: false, allow_invites: true },
          group_url: ""
        })
      });

      if (!res.ok) {
        const errorData = await res.json();
        console.error("Failed to create group:", errorData);
        return;
      }

      const createdGroup = await res.json();

      // Assign group_url using getAvatar (based on UUID string)
      const groupWithAvatar = {
        ...createdGroup,
        group_url: createdGroup.group_url || "/default-group-avatar.png",
        unreadCount: 0,
        lastMessage: "Group created",
        lastMessageTime: "now",
      };

      setGroups(prev => {
        const updated = [...prev, groupWithAvatar];
        return updated.sort((a, b) => a.name.localeCompare(b.name));
      });

      setChatMessages(prev => ({
        ...prev,
        [createdGroup.id]: []
      }));

      setNewGroupName("");
      setSelectedCaseId("");
      setShowNewGroupModal(false);

    } catch (error) {
      console.error("Error creating group:", error);
    }
  };


  const handleReplyToMessage = (message: Message) => {
    setReplyingTo(message);
  };

  const handleImageClick = (url: string) => {
    setModalImageUrl(url);
    setShowImageModal(true);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "sent":
        return <Check className="w-4 h-4 text-muted-foreground" />;
      case "delivered":
        return <CheckCheck className="w-4 h-4 text-muted-foreground" />;
      case "read":
        return <CheckCheck className="w-4 h-4 text-blue-400" />;
      default:
        return null;
    }
  };


const handleAddMember = async (e?: React.MouseEvent | React.KeyboardEvent) => {
  e?.preventDefault();
  if (!newMemberEmail.trim() || !activeChat) return;

  if (activeChat.members.some(m => m.email === newMemberEmail)) {
    toast.error("User is already a group member");
    return;
  }

  try {
    const token = sessionStorage.getItem("authToken") || "";
    const userStr = sessionStorage.getItem("user");
    let currentUserId: string | null = null;
    try {
      const userObj = userStr ? JSON.parse(userStr) : null;
      currentUserId = userObj?.id ?? null;
    } catch {
      toast.error("Could not determine current user. Please sign in again.");
      return;
    }
    if (!currentUserId) {
      toast.error("Could not determine current user. Please sign in again.");
      return;
    }

    // Fetch the new member's UUID
    const userRes = await fetch(`http://localhost:8080/api/v1/users?email=${encodeURIComponent(newMemberEmail)}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (!userRes.ok) {
      const err = await userRes.text().catch(() => "Unknown error");
      toast.error(`Failed to find user ${newMemberEmail}: ${err}`);
      return;
    }
    const userData = await userRes.json();
    const newMemberId = userData.id || userData.user_id;
    if (!newMemberId) {
      toast.error(`No user ID found for ${newMemberEmail}`);
      return;
    }

    // Add member to group
    const res = await fetch(`http://localhost:8080/api/v1/chat/groups/${activeChat.id}/members`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        user_id: newMemberId,
        user_email: newMemberEmail,
        role: "member"
      })
    });

    if (!res.ok) {
      let message = "Failed to add member";
      try {
        const err = await res.json();
        message = err?.message || message;
      } catch {}
      toast.error(message);
      return;
    }

    const updatedGroup = await res.json();

    setGroups(prev =>
      prev.map(group =>
        group.id === updatedGroup.id
          ? {
              ...group,
              members: updatedGroup.members.map((m: any) => ({
                id: m.id || m.user_id,
                email: m.email || m.user_email
              }))
            }
          : group
      )
    );

    setActiveChat(prev => prev ? {
      ...prev,
      members: [
        ...(Array.isArray(prev.members) ? prev.members : []),
        { id: newMemberId, email: newMemberEmail } // Add ChatMember
      ]
    } : null);

    setAvailableUsers(prev =>
      prev.filter(u => u.user_email !== newMemberEmail)
    );

    const systemMessage: Message = {
      id: Date.now(),
      user: "System",
      color: "text-gray-400",
      content: `${newMemberEmail} was added to the group`,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      status: "read"
    };

    setChatMessages(prev => ({
      ...prev,
      [activeChat.id]: [...(prev[activeChat.id] || []), systemMessage]
    }));

    const chatForSecret: Chat = {
      id: String(activeChat.id),
      members: updatedGroup.members.map((m: any) => ({
        id: m.id || m.user_id,
        email: m.email || m.user_email
      })),
      name: activeChat.name
    };
    const secretResult = await generateGroupSharedSecretForChat(chatForSecret, token, currentUserId);
    if (secretResult && secretResult.sharedSecret) {
      const newRatchetState = await initRatchetState(secretResult.sharedSecret);
      useSharedSecrets.getState().setSharedSecret(String(activeChat.id), newRatchetState);

      const socket = socketRef.current;
      if (socket?.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
          type: "group_secret_updated",
          payload: {
            groupId: activeChat.id,
            message: `Group key updated due to new member: ${newMemberEmail}`
          }
        }));
      }
    } else {
      toast.error("Failed to update group shared secret.");
      return;
    }

    toast.success(`${newMemberEmail} successfully added to the group`);
    setNewMemberEmail("");
  } catch (err: unknown) {
    console.error("Failed to add member:", err);
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    toast.error(`Failed to add member: ${errorMessage}`);
  }
};



function asEmailList(members: unknown): string[] {
  if (!Array.isArray(members)) return [];
  if (members.length === 0) return [];
  if (typeof members[0] === "string") return members as string[];
  return (members as Array<{ email?: string; id?: string; name?: string }>)
    .map(m => m?.email || m?.id || m?.name || "")
    .filter(Boolean);
}

function base64ToU8(s: string): Uint8Array {
  // url-safe -> std
  const pad = s.length % 4 === 2 ? "==" : s.length % 4 === 3 ? "=" : "";
  const b64 = s.replace(/-/g, "+").replace(/_/g, "/") + pad;
  const bin = atob(b64);
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
  return out;
}

function u8Fresh(u8: Uint8Array): Uint8Array {
  const ab = new ArrayBuffer(u8.byteLength);
  const copy = new Uint8Array(ab);
  copy.set(u8);
  return copy;
}

function u8ToArrayBuffer(u8: Uint8Array): ArrayBuffer {
  const ab = new ArrayBuffer(u8.byteLength);
  new Uint8Array(ab).set(u8);
  return ab;
}

// Keep this somewhere in your module (e.g., under loadMessages)
async function resolvePreviewUrl(url: string, mime: string): Promise<string> {
  if (!url.startsWith('https:')) throw new Error('Non-HTTPS URL'); // Enforce HTTPS
  const resp = await fetch(url, {
    mode: 'cors',
    credentials: 'omit',
    referrerPolicy: 'no-referrer',
    cache: 'no-store', // Avoid caching encrypted blobs
  });
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
  const blob = await resp.blob();
  const typedBlob = new Blob([await blob.arrayBuffer()], { type: mime });
  const blobUrl = URL.createObjectURL(typedBlob);
  // Optional: If backend provides hash, verify: await crypto.subtle.digest('SHA-256', await blob.arrayBuffer()) == expected
  return blobUrl;
}



// ... Other imports (e.g., React, socketRef, setChatMessages, etc.)
async function loadMessages(groupId: string) {
  if (!activeChat?.id || !token) {
    console.warn("❌ No activeChat ID or token, skipping message load.");
    return;
  }

  try {
    const res = await fetch(`http://localhost:8080/api/v1/chat/groups/${groupId}/messages`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (!res.ok) {
      console.error("Failed to load messages:", await res.text());
      setChatMessages(prev => ({ ...prev, [groupId]: [] }));
      return;
    }

    const data = await res.json();
    console.log("Fetched messages:", data);
    if (!Array.isArray(data)) {
      setChatMessages(prev => ({ ...prev, [groupId]: [] }));
      return;
    }

    await sodium.ready;

    const groupKey = String(groupId);
    const chatForSecret: Chat = {
      id: groupKey,
      name: activeChat.name,
      members: activeChat.members // ChatMember[] with UUIDs
    };

    const messages: Message[] = [];

    for (const msg of data) {
      console.log("Processing message:", { id: msg.id, content: msg.content, message_type: msg.message_type });
      let text = msg.content ?? "";
      let attachmentsForUI: Message["attachments"] = [];

      try {
        if (msg.is_encrypted && msg.envelope) {
          const envelope = msg.envelope as CryptoEnvelopeV1;
          const opkId: string | undefined = envelope.opk_id;

          let ratchetState = useSharedSecrets.getState().getSharedSecret(groupKey);

          if (!ratchetState) {
            const ephB64u = (envelope.ephemeral_pub || "").trim();

            if (ephB64u) {
              const { ikPriv: ourIKPrivEd, spkPriv: ourSPKPrivX, opks: ourOPKs } = useUserKeys.getState();
              if (!ourIKPrivEd || !ourSPKPrivX) throw new Error("Missing private keys (IK/SPK) for decryption.");

              const senderIdentity = msg.sender_email ?? msg.from ?? "";
              if (!senderIdentity) throw new Error("Cannot determine sender identity to fetch bundle.");

              const raw = await fetchBundle(senderIdentity,userId || "", token); // Fixed: pass token
              const nb = normalizeIncomingBundle(raw);
              const ok =
                nb.spkSignature &&
                nb.spkPublicX &&
                verifySpkSignature(nb.identityKeyEd, nb.spkPublicX, nb.spkSignature);
              if (!ok) throw new Error("Sender SPK signature invalid");

              const senderIKPubEd = sodium.from_base64(nb.identityKeyEd);
              const senderIKPubX = sodium.crypto_sign_ed25519_pk_to_curve25519(senderIKPubEd);
              const ourIKPrivX = sodium.crypto_sign_ed25519_sk_to_curve25519(ourIKPrivEd);

              let ourOPKPrivX: Uint8Array | undefined;
              if (opkId) {
                const found = ourOPKs?.find((o: { id: string; priv: Uint8Array }) => o.id === opkId);
                ourOPKPrivX = found?.priv;
              }

              const ephPubX = base64ToU8(ephB64u);

              const shared = await deriveSharedSecretResponder(
                u8Fresh(ourIKPrivX),
                u8Fresh(ourSPKPrivX),
                u8Fresh(ephPubX),
                u8Fresh(senderIKPubX),
                ourOPKPrivX ? u8Fresh(ourOPKPrivX) : undefined
              );

              if (!shared) throw new Error("Failed to derive shared secret");
              ratchetState = await initRatchetState(shared);
            } else {
              const currentUserId =
                userId ||
                (() => {
                  try {
                    const [, p] = (token || "").split(".");
                    if (!p) return "";
                    const j = JSON.parse(atob(p.replace(/-/g, "+").replace(/_/g, "/")));
                    return j?.user_id || "";
                  } catch {
                    return "";
                  }
                })();

              const resSecret = await generateGroupSharedSecretForChat(chatForSecret, token, currentUserId || "");
              if (!resSecret?.sharedSecret) {
                toast.error(`Failed to derive group secret for ${groupId}. Ensure all members have key bundles.`);
                throw new Error("Group secret derivation failed.");
              }
              ratchetState = await initRatchetState(resSecret.sharedSecret);
            }

            useSharedSecrets.getState().setSharedSecret(groupKey, ratchetState);
          }

          if ((msg.message_type ?? "text") === "text") {
            text = await ratchetDecrypt(ratchetState, envelope.header || { dhPub: "", msgNum: 0 }, envelope.ct, envelope.nonce);
          } else if ((msg.message_type ?? "").toLowerCase() === "file" && Array.isArray(msg.attachments) && msg.attachments.length) {
            const att = msg.attachments[0];
            const mime = att.file_type || "application/octet-stream";
            let plainBytes: Uint8Array | undefined;

            if (msg.is_encrypted && msg.envelope?.ct && msg.envelope?.nonce) {
              plainBytes = await ratchetDecryptBytes(ratchetState, envelope.header || { dhPub: "", msgNum: 0 }, envelope.ct, envelope.nonce);
            } else if (att.url && !att.is_encrypted) {
              attachmentsForUI = [{
                file_name: att.file_name || "file",
                file_type: mime,
                file_size: Number(att.file_size || 0),
                url: att.url,
                isImage: mime.startsWith("image/")
              }];
              text = msg.content || "";
            }

            if (plainBytes) {
              const blobUrl = URL.createObjectURL(new Blob([u8ToArrayBuffer(plainBytes)], { type: mime }));
              trackBlobUrl?.(blobUrl);
              attachmentsForUI = [{
                file_name: att.file_name || "file",
                file_type: mime,
                file_size: Number(att.file_size || plainBytes.byteLength || 0),
                url: blobUrl,
                isImage: mime.startsWith("image/")
              }];
            }
          }

          if (envelope.opk_id) {
            try { await useUserKeys.getState().markOPKUsed(envelope.opk_id, true); } catch {}
          }
        } else {
          if ((msg.message_type ?? "text") === "text") {
            text = msg.content ?? "";
          } else if ((msg.message_type ?? "").toLowerCase() === "file" && Array.isArray(msg.attachments) && msg.attachments.length) {
            const att = msg.attachments[0];
            const mime = att.file_type || "application/octet-stream";
            attachmentsForUI = [{
              file_name: att.file_name || "file",
              file_type: mime,
              file_size: Number(att.file_size || 0),
              url: att.url,
              isImage: mime.startsWith("image/")
            }];
            text = msg.content || "";
          }
        }
      } catch (e: unknown) {
        console.warn("History decrypt failed:", e);
        const errorMessage = e instanceof Error ? e.message : "Unknown error";
        text = `[Decryption failed: ${errorMessage}]`;
      }

      const status: "sent" | "delivered" | "read" =
        msg?.status?.read ? "read" :
        msg?.status?.delivered ? "delivered" : "delivered";

      messages.push({
        id: msg.id,
        user: msg.sender_name || msg.sender_email,
        color: msg.sender_email === userEmail ? "text-green-400" : "text-blue-400",
        content: text,
        time: new Date(msg.created_at?.$date || msg.created_at)
          .toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
        status,
        self: msg.sender_email === userEmail,
        attachments: attachmentsForUI
      });
    }

    setChatMessages(prev => ({ ...prev, [groupId]: messages }));
  } catch (err: unknown) {
    console.error("Failed to load messages:", err);
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    setChatMessages(prev => ({ ...prev, [groupId]: [] }));
    toast.error(`Failed to load messages: ${errorMessage}`);
  }
}

useEffect(() => {
  if (!activeChat?.id || chatMessages[activeChat.id]?.length > 0) {
    console.warn("Skipping loadMessages as messages are already present");
    return;
  }
  console.log("🔄 Loading messages for group ID:", activeChat.id);
  loadMessages(String(activeChat.id));
}, [activeChat]);

useEffect(() => {
  if (activeChat) return; // Skip if activeChat is already set (prevents reset on groups update)
  const storedChat = localStorage.getItem("activeChat");
  if (storedChat) {
    const parsed = JSON.parse(storedChat);
    const match = groups.find(g => g.id === parsed.id);
    if (match) {
      setActiveChat(parsed);
    }
  }
}, [groups]);
  useEffect(() => {
    if (!previewFile) {
      setPreviewUrl(""); // ensure consistency
    }
  }, [previewFile]);




  const updateGroup = async () => {
    if (!activeChat) {
      console.error("No active chat selected for update.");
      return;
    }
    try {
      await fetch(`http://localhost:8080/api/v1/chat/groups/${activeChat.id}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          name: editGroupName,
          description: editDescription,
          settings: { is_public: editIsPublic }
        })
      });

      await fetchGroups(); // refresh your groups list
      setShowEditGroupModal(false);
    } catch (err) {
      console.error("Failed to update group:", err);
    }
  };


  const removeGroupLocally = (groupId: number) => {
    setGroups(prev => prev.filter(group => String(group.id) !== String(groupId)));
    setChatMessages(prev => {
      const updated = { ...prev };
      delete updated[groupId];
      return updated;
    });
    setActiveChat(null);
    setShowMoreMenu(false);
  };


  const handleLeaveGroup = async () => {
    if (!activeChat) return;
    try {
      await fetch(`http://localhost:8080/api/v1/chat/groups/${activeChat.id}/members/${userEmail}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      });

      removeGroupLocally(Number(activeChat.id));
      console.log("You left the group successfully!");
    } catch (err) {
      console.error("Failed to leave group:", err);
    }
  };
  useEffect(() => {
    if (!activeChat || !activeChat.caseId) return;

    const fetchCollaborators = async () => {
      try {
        const res = await fetch(`http://localhost:8080/api/v1/cases/${activeChat.caseId}/collaborators`, {
          headers: { Authorization: `Bearer ${token}` }
        });

        if (!res.ok) {
          console.error("Failed to fetch collaborators:", await res.text());
          return;
        }

        const data = await res.json();
        console.log("Fetched collaborators:", data);

        // Ensure collaborators match the structure: { user_email: string, role: string }
        const formatted = (data.data || []).map((collab: any) => ({
          user_email: collab.email,
          role: collab.role || "member"
        }));
        console.log("✅ availableUsers after fetch:", formatted);

        setAvailableUsers(formatted);
      } catch (err) {
        console.error("Failed to fetch collaborators:", err);
      }
    };

    fetchCollaborators();
  }, [activeChat, token]);



  const handleOpenAddMembersModal = async () => {
    if (!activeChat?.caseId) {
      console.warn("❌ No caseId available in activeChat");
      return;
    }

    try {
      const res = await fetch(`http://localhost:8080/api/v1/cases/${activeChat.caseId}/collaborators`, {
        headers: { Authorization: `Bearer ${token}` }
      });

      if (!res.ok) {
        console.error("❌ Failed to fetch collaborators:", await res.text());
        return;
      }

      const data = await res.json();
      console.log("✅ Fetched collaborators:", data);

      const formatted = (data.data || []).map((collab: any) => ({
        user_email: collab.email,
        role: collab.role || "member"
      }));

      setAvailableUsers(formatted);
      setShowAddMembersModal(true);
      setShowMoreMenu(false);
    } catch (err) {
      console.error("❌ Error fetching collaborators:", err);
    }
  };






  const handleDeleteGroup = async () => {
    if (!activeChat) return;
    try {
      await fetch(`http://localhost:8080/api/v1/chat/groups/${activeChat.id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      });

      removeGroupLocally(Number(activeChat.id));
      console.log("Group deleted successfully!");
    } catch (err) {
      console.error("Failed to delete group:", err);
    }
  };



const MessageComponent = ({ msg }: { msg: Message }) => (
  <div className={`flex ${msg.self ? "justify-end" : "justify-start"} group`}>
    <div
      className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg relative ${msg.self ? "bg-blue-600 text-white" : "bg-muted text-foreground"}`}
    >
      {/* Reply preview */}
      {msg.replyTo && (
        <div
          className={`mb-2 p-2 rounded border-l-4 ${msg.self ? "border-white/30 bg-white/10" : "border-blue-400 bg-blue-50 dark:bg-blue-900/20"}`}
        >
          <p className={`text-xs font-semibold ${msg.self ? "text-white/80" : "text-blue-600"}`}>
            {msg.replyTo.user}
          </p>
          <p className={`text-xs truncate ${msg.self ? "text-white/70" : "text-muted-foreground"}`}>
            {msg.replyTo.attachment ? `📎 ${msg.replyTo.attachment.name}` : msg.replyTo.content}
          </p>
        </div>
      )}

      {/* Sender name (for incoming only) */}
      {!msg.self && (
        <p className={`text-xs font-bold ${msg.color} mb-1`}>{msg.user}</p>
      )}

      {/* Attachments */}
      {Array.isArray(msg.attachments) &&
        msg.attachments.map((attachment, idx) => (
          <div key={idx} className="mb-2">
            {attachment.file_type?.startsWith?.("image/") ? (
              <div className="relative">
                <img
                  src={attachment.url}
                  alt={attachment.file_name}
                  className="max-w-full h-auto rounded cursor-pointer hover:opacity-90 transition-opacity"
                  onClick={() => attachment.url && handleImageClick(attachment.url)}
                />
                <button
                  onClick={() => attachment.url && handleImageClick(attachment.url)}
                  className="absolute top-2 right-2 bg-black/50 text-white p-1 rounded-full hover:bg-black/70 transition-all"
                  title="Preview"
                >
                  <Eye className="w-4 h-4" />
                </button>
              </div>
            ) : (
              <div className={`p-3 rounded border ${msg.self ? "bg-black/20 border-white/20" : "bg-accent border-border"}`}>
                <div className="flex items-center gap-2">
                  <FileText className="w-5 h-5" />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate text-sm">{attachment.file_name}</p>
                    <p className="text-xs opacity-70">
                      {Number.isFinite(attachment.file_size) ? `${(attachment.file_size / 1024).toFixed(1)} KB` : ""}
                    </p>
                  </div>
                  {attachment.url && (
                    <a href={attachment.url} download className="p-1 hover:bg-black/20 rounded" title="Download">
                      <Download className="w-4 h-4" />
                    </a>
                  )}
                </div>
              </div>
            )}
          </div>
        ))}

      {/* Message Content */}
      {msg.content && ( // Only render if content exists
        <div className="text-sm">
          <p>{msg.content}</p>
        </div>
      )}

      {/* Footer: time + actions */}
      <div className="flex items-center justify-between mt-1">
        <span className="text-xs opacity-70">{msg.time}</span>
        <div className="flex items-center gap-1">
          {msg.self && getStatusIcon(msg.status)}
          {!msg.self && (
            <button
              onClick={() => handleReplyToMessage(msg)}
              className="opacity-0 group-hover:opacity-100 p-1 hover:bg-black/20 rounded transition-all"
              title="Reply"
            >
              <Reply className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>
    </div>
  </div>
);



  return (
    <div className="bg-background flex w-full h-screen text-foreground relative">
      {/* Main Sidebar - Fixed positioning without overlay */}
      {hasMounted && (
        <div className={`fixed z-30 top-0 left-0 h-full w-72 bg-card border-r border-border transition-transform duration-300 ease-in-out ${sidebarOpen ? "translate-x-0" : "-translate-x-full"}`}>
          <div className="p-6">
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
            <nav className="space-y-2">
              <Link to="/dashboard"><button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-muted rounded-lg">
                <Home className="w-5 h-5" />
                Dashboard
              </button></Link>
              <Link to="/case-management"><button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-muted rounded-lg">
                <Folder className="w-5 h-5" />
                Case Management
              </button></Link>
              <Link to="/evidence-viewer"><button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-muted rounded-lg">
                <FileText className="w-5 h-5" />
                Evidence Viewer
              </button></Link>
              <button className="w-full flex items-center gap-3 text-left px-4 py-2 bg-muted hover:bg-accent rounded-lg">
                <MessageSquare className="w-5 h-5" />
                Secure Chat
              </button>
              {isDFIRAdmin && (
                <Link to="/report-dashboard"><button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-muted rounded-lg">
                  <ClipboardList className="w-5 h-5" />
                  Case Reports
                </button></Link>
              )}
            </nav>
          </div>
        </div>)}
      {/* Overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-20"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Chat Layout - Adjusted margin for sidebar */}
      <div className={`flex flex-1 h-screen transition-all duration-300 ${sidebarOpen ? 'ml-72' : 'ml-0'}`}>
        {/* Chat List Sidebar */}
        <div className="w-80 min-w-80 max-w-80 bg-card border-r border-border flex flex-col">          {/* Chat Header */}
          <div className="p-4 border-b border">
            <div className="flex items-center justify-between mb-4">
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="text-foreground hover:text-blue-400 mr-3"
              >
                <Menu className="w-6 h-6" />
              </button>
              <h2 className="text-xl font-bold flex-1">Chats</h2>
              <button
                onClick={() => setShowNewGroupModal(true)}
                className="text-foreground hover:text-blue-400"
                title="Create new group"
              >
                <Plus className="w-6 h-6" />
              </button>
            </div>

            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search chats..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-muted border border-border rounded-lg text-foreground placeholder-muted-foreground"
              />
            </div>
          </div>

          {/* Chat List */}
          <div className="flex-1 overflow-y-auto">
            {filteredGroups.map((group) => (
              <div
                key={group.id}
                onClick={() => {
                  handleSelectGroup(group);
                  setGroups(prev =>
                    prev.map(g => g.id === group.id ? { ...g, unreadCount: 0 } : g)
                  );
                  setShowChatSearch(false);
                  setChatSearchQuery("");
                  setReplyingTo(null);
                }}
                className={`p-4 border-b border-border cursor-pointer hover:bg-muted transition-colors ${activeChat?.id === group.id ? "bg-accent" : ""
                  }`}
              >
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-full overflow-hidden cursor-pointer hover:opacity-80 transition" onClick={handleGroupImageClick}>
                    <img
                      src={group.group_url || "/default-group-avatar.png"}  // fallback image
                      alt="Group Avatar"
                      className="w-full h-full object-cover"
                    />
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold text-foreground truncate">{group.name}</h3>
                      <span className="text-xs text-muted-foreground">{group.lastMessageTime}</span>
                    </div>
                    <div className="flex items-center justify-between mt-1">
                      <p className="text-sm text-muted-foreground truncate">{group.lastMessage}</p>
                      {group.unreadCount > 0 && (
                        <span className="bg-blue-500 text-white text-xs rounded-full px-2 py-1 min-w-5 text-center">
                          {group.unreadCount}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Active Chat Area */}
        <div className="flex-1 flex flex-col bg-background">
          {activeChat ? (
            <>
              {/* Chat Header */}
              <div className="p-4 border-b border-border bg-card">
                {showChatSearch ? (
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => {
                        setShowChatSearch(false);
                        setChatSearchQuery("");
                      }}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <X className="w-5 h-5" />
                    </button>
                    <div className="flex-1 relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <input
                        type="text"
                        placeholder="Search messages..."
                        value={chatSearchQuery}
                        onChange={(e) => setChatSearchQuery(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 bg-muted border border-border rounded-lg text-foreground placeholder-muted-foreground"
                        autoFocus
                      />
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="relative">
                        {/* Group Avatar (clickable) */}
                        <div
                          className="w-10 h-10 rounded-full overflow-hidden cursor-pointer hover:opacity-80 transition"
                          onClick={handleGroupImageClick}
                        >
                          <img
                            src={activeChat.group_url || "/default-group-avatar.png"}
                            alt="Group Avatar"
                            className="w-full h-full object-cover"
                          />
                        </div>

                        {/* Hidden File Input */}
                        <input
                          ref={fileInputGroupRef}
                          type="file"
                          accept="image/*"
                          className="hidden"
                          onChange={handleGroupImageUpload}
                        />
                      </div>

                      <div>
                        <h3 className="font-semibold text-foreground">{activeChat.name}</h3>
                        <p
                          className="text-sm text-muted-foreground flex items-center gap-1 cursor-pointer hover:underline"
                          onClick={() => setShowAddMembersModal(true)}
                        >
                          <Users className="w-4 h-4" />
                          {activeChat?.members?.length ?? 0} members
                        </p>

                        {activeChat.caseId && (
                          <p className="text-xs text-muted-foreground">
                            Linked Case:{" "}
                            <Link
                              to={`/case-management/${activeChat.caseId}`}
                              className="text-blue-500 hover:underline"
                            >
                              {activeChat.caseId.slice(0, 8)}...
                            </Link>
                          </p>
                        )}
                      </div>

                    </div>
                    <div className="relative" ref={moreMenuRef}>
                      <button
                        onClick={() => setShowMoreMenu(!showMoreMenu)}
                        className="text-muted-foreground hover:text-foreground"
                      >
                        <MoreVertical className="w-5 h-5" />
                      </button>

                      {/* More Menu Dropdown */}
                      {showMoreMenu && (
                        <div className="absolute right-0 top-8 bg-background border border-border rounded-lg shadow-lg py-2 w-48 z-50">
                          <button
                            onClick={() => {
                              setShowChatSearch(true);
                              setShowMoreMenu(false);
                            }}
                            className="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-muted"
                          >
                            <Search className="w-4 h-4" />
                            Search
                          </button>
                          <button
                            onClick={handleOpenAddMembersModal}
                            className="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-muted"
                          >
                            <Users className="w-4 h-4" />
                            Add Members
                          </button>

                          <button
                            onClick={() => {
                              if (window.confirm("Are you sure you want to leave this group?")) {
                                handleLeaveGroup();
                              }
                            }}
                            className="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-muted text-red-400"
                          >
                            <LogOut className="w-4 h-4" />
                            Exit Group
                          </button>
                          <button
                            onClick={() => {
                              if (window.confirm("Are you sure you want to delete this group for everyone?")) {
                                handleDeleteGroup();
                              }
                            }}
                            className="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-muted text-red-400"
                          >
                            <Trash className="w-4 h-4" />
                            Delete Group
                          </button>

                        </div>

                      )}
                    </div>
                  </div>
                )}
              </div>

              {/* Messages Area */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {displayMessages.map((msg: Message) => (
                  <MessageComponent key={msg.id} msg={msg} />
                ))}
                <div ref={chatEndRef} />
              </div>


              {/* Reply Preview */}
              {replyingTo && (
                <div className="px-4 py-2 bg-muted border-t border-border">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <p className="text-sm font-medium text-blue-600">
                        Replying to {replyingTo.user}
                      </p>
                      <p className="text-xs text-muted-foreground truncate">
                        {replyingTo.attachments?.[0]
                          ? `📎 ${replyingTo.attachments[0].file_name}`
                          : replyingTo.content
                        }
                      </p>
                    </div>
                    <button
                      onClick={() => setReplyingTo(null)}
                      className="p-1 hover:bg-accent rounded"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              )}

              {/* ✅ Message Input + Typing Indicator */}
              <div className="p-4 border-t border-border bg-card">
                {/* Typing Indicator for other users */}
                {typingUsers[activeChat?.id]?.filter((email) => email !== userEmail).length > 0 && (
                  <div className="text-sm text-muted-foreground mb-1 ml-2">
                    {typingUsers[activeChat.id]
                      .filter((email) => email !== userEmail)
                      .join(", ")}{" "}
                    {typingUsers[activeChat.id].length > 2 ? "are" : "is"} typing...
                  </div>
                )}

                {/* Input Row */}
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => fileInputRef.current?.click()}
                    className="p-3 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg transition-colors"
                    title="Attach file"
                  >
                    <Paperclip className="w-5 h-5" />
                  </button>
                  <input
                    ref={fileInputRef}
                    type="file"
                    onChange={handleFileSelection}
                    className="hidden"
                    accept="*/*"
                  />
                  <input
                    type="text"
                    value={message}
                    onChange={(e) => {
                      setMessage(e.target.value);
                      sendTypingNotification("typing_start");
                    }}
                    onBlur={() => sendTypingNotification("typing_stop")}
                    onKeyPress={(e) => e.key === 'Enter' && handleSendMessage(e)}
                    placeholder="Type a secure message..."
                    className="flex-1 p-3 rounded-lg bg-muted text-foreground border border-border placeholder-muted-foreground"
                  />
                  <button
                    onClick={handleSendMessage}
                    className="px-4 py-3 bg-blue-600 hover:bg-blue-500 rounded-lg flex items-center justify-center transition-colors"
                  >
                    <Send className="w-5 h-5" />
                  </button>
                </div>
              </div>
              {/* ✅ Old Message Input (Preserved) */}

            </>
          ) : (
            <div className="flex-1 flex items-center justify-center">
              <div className="text-center text-muted-foreground">
                <MessageSquare className="w-16 h-16 mx-auto mb-4 opacity-50" />
                <h3 className="text-xl font-semibold mb-2">Welcome to Secure Chat</h3>
                <p>Select a group to start secure communication</p>
              </div>
            </div>
          )}
        </div>
      </div>
      {/* Attachment Preview Modal */}
      {showAttachmentPreview && previewFile && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="rounded-lg p-6 w-full max-w-md max-h-[90vh] overflow-y-auto border-[3px] border-border bg-background shadow-xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-bold text-foreground text-lg mb-4">Send Attachment</h3>
              <button
                onClick={handleCancelAttachment}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* File Preview */}
            <div className="mb-4">
              {previewFile.type.startsWith('image/') ? (
                // Fixed size image preview container
                <div className="w-full h-64 overflow-hidden rounded-lg border border-border bg-muted flex items-center justify-center">
                  <img
                    src={previewUrl}
                    alt={previewFile.name}
                    className="max-w-full max-h-full object-contain"
                  />
                </div>
              ) : (
                // Fixed size file preview
                <div className="w-full h-32 p-4 bg-muted rounded-lg border border-border flex items-center justify-center">
                  <div className="flex items-center gap-3">
                    <FileText className="w-12 h-12 text-blue-500 flex-shrink-0" />
                    <div className="min-w-0">
                      <p className="font-medium truncate">{previewFile.name}</p>
                      <p className="text-sm text-muted-foreground">
                        {(previewFile.size / 1024).toFixed(1)} KB
                      </p>
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Reply Preview in Attachment Modal */}
            {replyingTo && (
              <div className="mb-4 p-3 bg-muted rounded-lg border-l-4 border-blue-400">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-blue-600">
                      Replying to {replyingTo.user}
                    </p>
                    <p className="text-xs text-muted-foreground truncate">
                      {replyingTo.attachments?.[0]
                        ? `📎 ${replyingTo.attachments[0].file_name}`
                        : replyingTo.content
                      }
                    </p>
                  </div>
                  <button
                    onClick={() => setReplyingTo(null)}
                    className="p-1 hover:bg-accent rounded"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              </div>
            )}

            {/* Message Input */}
            <div className="mb-4">
              <input
                type="text"
                value={attachmentMessage}
                onChange={(e) => setAttachmentMessage(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && sendAttachment()}
                placeholder="Add a message..."
                className="w-full p-3 rounded-lg bg-muted text-foreground border border-border placeholder-muted-foreground"
                autoFocus
              />
            </div>

            {/* Action Buttons */}
            <div className="flex justify-end gap-2">
              <button
                onClick={handleCancelAttachment}
                className="px-4 py-2 text-muted-foreground hover:text-foreground"
              >
                Cancel
              </button>
              <button
                onClick={sendAttachment}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-white flex items-center gap-2"
              >
                <Send className="w-4 h-4" />
                Send
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Image Viewer Modal */}
      {showImageModal && (
        <div className="fixed inset-0 bg-black bg-opacity-90 z-50 flex items-center justify-center p-4">
          <div className="relative max-w-4xl max-h-full">
            <button
              onClick={() => setShowImageModal(false)}
              className="absolute top-4 right-4 text-white hover:text-gray-300 bg-black bg-opacity-50 rounded-full p-2"
            >
              <X className="w-6 h-6" />
            </button>
            <img
              src={modalImageUrl}
              alt="Full size view"
              className="max-w-full max-h-full object-contain rounded-lg"
            />
          </div>
        </div>
      )}

      {/* New Group Modal */}
      {showNewGroupModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="rounded-lg p-6 w-full max-w-md max-h-[90vh] overflow-y-auto border-[3px] border-border bg-background shadow-xl">
            <h3 className="text-xl font-bold mb-4">Create New Group</h3>

            {/* Case Selector */}
            <label className="block text-sm text-muted-foreground mb-1">Case associated with group chat</label>
            <select
              className="w-full p-3 rounded-lg bg-muted text-foreground border border-border mb-4"
              value={selectedCaseId}
              onChange={(e) => setSelectedCaseId(e.target.value)}
            >
              <option value="">-- Select an active case --</option>
              {activeCases.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.title || "Untitled Case"} ({c.id})
                </option>
              ))}
            </select>

            {/* Group name input */}
            <input
              type="text"
              value={newGroupName}
              onChange={(e) => setNewGroupName(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleCreateGroup(e)}
              placeholder="Enter group name..."
              className="w-full p-3 rounded-lg bg-muted text-foreground border border-border placeholder-muted-foreground mb-4"
              autoFocus
            />

            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowNewGroupModal(false)}
                className="px-4 py-2 text-muted-foreground hover:text-foreground"
              >
                Cancel
              </button>
              <button
                onClick={(e) => handleCreateGroup(e)}
                disabled={!selectedCaseId}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-white disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Members Modal */}
      {showAddMembersModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
            <div className="rounded-lg p-6 w-full max-w-md max-h-[90vh] overflow-y-auto border-[3px] border-border bg-background shadow-xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold">Add Members</h3>
              <button
                onClick={() => setShowAddMembersModal(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Current Members */}
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-muted-foreground mb-2">Current Members</h4>
              <div className="space-y-1 max-h-32 overflow-y-auto">
             {activeChat?.members.map((member: string | { user_email: string; role: string }, index: number) => {
              const email = typeof member === "string" ? member : member.user_email;

              return (
                <div key={index} className="flex items-center gap-2 p-2 bg-muted rounded text-sm">
                  <Users className="w-4 h-4 text-muted-foreground" />
                  <span>{email}</span>
                  {email === userEmail && (
                    <span className="text-xs bg-blue-500 text-white px-2 py-1 rounded">You</span>
                  )}
                </div>
              );
            })}


              </div>
            </div>

           {/* Add New Member */}
          <div className="mb-4">
            <h4 className="text-sm font-semibold text-muted-foreground mb-2">Add New Member</h4>

            <div className="space-y-3">
              <select
                value={newMemberEmail}
                onChange={(e) => setNewMemberEmail(e.target.value)}
                className="w-full p-3 rounded-lg bg-muted text-foreground border border-border"
              >
                <option value="">-- Select a collaborator --</option>

                {availableUsers &&
                  activeChat?.members &&
                  availableUsers.filter(userObj =>
                    !activeChat.members.includes(userObj.user_email)
                  ).length === 0 && (
                    <option disabled value="">
                      No available collaborators
                    </option>
                  )}

                {availableUsers &&
                  activeChat?.members &&
                  availableUsers
                    .filter(userObj =>
                      !activeChat.members.includes(userObj.user_email)
                    )
                    .map(userObj => (
                      <option key={userObj.user_email} value={userObj.user_email}>
                        {userObj.user_email} ({userObj.role})
                      </option>
                    ))}
              </select>


        

                
                {/* Quick Add Suggestions */}
                <div>
                  <p className="text-xs text-muted-foreground mb-2">Quick Add:</p>
                  <div className="grid grid-cols-1 gap-1 max-h-32 overflow-y-auto">
                    {availableUsers
                      .filter(userObj => !activeChat?.members.includes(userObj.user_email))
                      .map(userObj => (
                        <button
                          key={userObj.user_email}
                          onClick={() => setNewMemberEmail(userObj.user_email)}
                          className="..."
                        >
                          {userObj.user_email} ({userObj.role})
                        </button>
                    ))}

                  </div>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowAddMembersModal(false)}
                className="px-4 py-2 text-muted-foreground hover:text-foreground"
              >
                Cancel
              </button>
              <button
                onClick={handleAddMember}
                disabled={!newMemberEmail.trim()}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:bg-gray-400 disabled:cursor-not-allowed rounded-lg text-white flex items-center gap-2"
              >
                <Users className="w-4 h-4" />
                Add Member
              </button>
            </div>
          </div>
        </div>
      )}
      {/* Edit Group Modal */}
      {showEditGroupModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="rounded-lg p-6 w-full max-w-md bg-background shadow-xl border-[3px] border-border">
            <h3 className="text-xl font-bold mb-4">Edit Group</h3>
            <input
              type="text"
              value={editGroupName}
              onChange={(e) => setEditGroupName(e.target.value)}
              placeholder="Group name"
              className="w-full mb-3 p-3 rounded bg-muted border border-border"
            />
            <textarea
              value={editDescription}
              onChange={(e) => setEditDescription(e.target.value)}
              placeholder="Description"
              className="w-full mb-3 p-3 rounded bg-muted border border-border"
            />
            <label className="flex items-center gap-2 mb-4">
              <input
                type="checkbox"
                checked={editIsPublic}
                onChange={() => setEditIsPublic(!editIsPublic)}
              />
              Public group
            </label>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowEditGroupModal(false)}
                className="px-4 py-2 text-muted-foreground hover:text-foreground"
              >
                Cancel
              </button>
              <button
                onClick={updateGroup}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-white"
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}


    </div>
  );
}
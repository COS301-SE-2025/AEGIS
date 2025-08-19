import {
  Trash2,
  Inbox,
  Archive,
  Bell,
} from "lucide-react";
import { useEffect, useState, useRef } from "react";
import { Link } from "react-router-dom";

import { cn } from "../../lib/utils"; // Utility for conditional styling

type Notification = {
  id: string;
  title: string;
  message: string;
  timestamp: string;
  read: boolean;
  archived: boolean;
};


export const NotificationsPage = () => {
 // const [notifications, setNotifications] = useState<Notification[]>(mockData);
  const [selected, setSelected] = useState<string[]>([]);
  const [filter, setFilter] = useState<"all" | "unread" | "archived">("all");
  const [searchQuery, setSearchQuery] = useState("");

const wsRef = useRef<WebSocket | null>(null);
const [notifications, setNotifications] = useState<Notification[]>([]);
const token = sessionStorage.getItem("authToken");
const tenantID = sessionStorage.getItem("tenantId"); // store on login

// Compute unread notifications
const unreadCount = notifications.filter((n) => !n.read && !n.archived).length;

// 1️⃣ Fetch initial notifications from backend REST API
useEffect(() => {
  async function fetchNotifications() {
    if (!token) return;
    try {
      const res = await fetch("http://localhost:8080/api/v1/notifications", {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error("Failed to fetch notifications");
      const data = await res.json();

      // Format timestamps to human-readable form
      const formatted = data.map((n: Notification) => ({
        ...n,
        timestamp: new Date(n.timestamp).toLocaleString(),
      }));
      setNotifications(formatted);
    } catch (err) {
      console.error("Error fetching notifications:", err);
    }
  }
  fetchNotifications();
}, [token]);

// 2️⃣ WebSocket Connection
useEffect(() => {
  if (!token) return;

  // 🔹 Scope the WS to tenant or team instead of global
  const ws = new WebSocket(
    `ws://localhost:8080/ws/cases/${tenantID }?token=${token}`
  );
  wsRef.current = ws;

  ws.onopen = () => console.log("🔗 Connected to notifications WebSocket");

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);

      switch (msg.type) {
        case "notification":
        case "EventNotification":
          setNotifications((prev) => [
            {
              ...msg.payload,
              timestamp: new Date(msg.payload.timestamp).toLocaleString(),
            },
            ...prev,
          ]);
          break;

        case "mark_notification_read":
          const ids: string[] = msg.payload.notificationIds || [];
          setNotifications((prev) =>
            prev.map((n) => (ids.includes(n.id) ? { ...n, read: true } : n))
          );
          break;

        default:
          console.warn("⚠️ Unhandled WS message type:", msg.type);
      }
    } catch (err) {
      console.error("Failed to parse WS message", err);
    }
  };

  ws.onclose = () => console.log("❌ Notifications WebSocket closed");

  return () => ws.close();
}, [token, tenantID]);

// 3️⃣ Mark as Read (local + WebSocket + optional REST call)
const markAsRead = async () => {
  if (selected.length === 0) return;

  // Update locally
  setNotifications((prev) =>
    prev.map((n) => (selected.includes(n.id) ? { ...n, read: true } : n))
  );

  // Clear selection
  const idsToMark = [...selected];
  setSelected([]);

  // Send WebSocket event
  const ws = wsRef.current;
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(
      JSON.stringify({
        type: "MARK_NOTIFICATION_READ",
        payload: { notificationIds: idsToMark },
      })
    );
  }

  // Optional: Persist via REST
  try {
    await fetch("http://localhost:8080/api/v1/notifications/read", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ notificationIds: idsToMark }),
    });
  } catch (err) {
    console.error("Failed to persist mark-as-read:", err);
  }
};


// Archive selected notifications
const archiveSelected = async () => {
  if (selected.length === 0) return;

  // 1. Update local state immediately
  setNotifications((prev) =>
    prev.map((n) =>
      selected.includes(n.id) ? { ...n, archived: true } : n
    )
  );

  const idsToArchive = [...selected];
  setSelected([]);

  // 2. Send WebSocket event
  const ws = wsRef.current;
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(
      JSON.stringify({
        type: "ARCHIVE_NOTIFICATION",
        payload: { notificationIds: idsToArchive },
      })
    );
  }

  // 3. Persist via REST
  try {
    await fetch("http://localhost:8080/api/v1/notifications/archive", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ notificationIds: idsToArchive }),
    });
  } catch (err) {
    console.error("Failed to archive notifications:", err);
  }
};

// Delete selected notifications
const deleteSelected = async () => {
  if (selected.length === 0) return;

  // 1. Update local state immediately
  const idsToDelete = [...selected];
  setNotifications((prev) =>
    prev.filter((n) => !idsToDelete.includes(n.id))
  );

  setSelected([]);

  // 2. Send WebSocket event
  const ws = wsRef.current;
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(
      JSON.stringify({
        type: "DELETE_NOTIFICATION",
        payload: { notificationIds: idsToDelete },
      })
    );
  }

  // 3. Persist via REST
  try {
    await fetch("http://localhost:8080/api/v1/notifications/delete", {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ notificationIds: idsToDelete }),
    });
  } catch (err) {
    console.error("Failed to delete notifications:", err);
  }
};



  const filteredNotifications = notifications
    .filter((n) => {
      if (filter === "unread") return !n.read && !n.archived;
      if (filter === "archived") return n.archived;
      return true;
    })
    .filter((n) =>
      `${n.title} ${n.message}`.toLowerCase().includes(searchQuery.toLowerCase())
    );

  const toggleSelect = (id: string) => {
    setSelected((prev) =>
      prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id]
    );
  };

  // const markAsRead = () => {
  //   setNotifications((prev) =>
  //     prev.map((n) =>
  //       selected.includes(n.id) ? { ...n, read: true } : n
  //     )
  //   );
  //   setSelected([]);
  // };

  // const archiveSelected = () => {
  //   setNotifications((prev) =>
  //     prev.map((n) =>
  //       selected.includes(n.id) ? { ...n, archived: true } : n
  //     )
  //   );
  //   setSelected([]);
  // };

  // const deleteSelected = () => {
  //   setNotifications((prev) =>
  //     prev.filter((n) => !selected.includes(n.id))
  //   );
  //   setSelected([]);
  // };

  // Real-time updates placeholder
  useEffect(() => {
    const interval = setInterval(() => {
      // Poll or listen to WebSocket in a real app
    }, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-background text-foreground px-8 py-10 transition-colors">
      <div className="flex items-center justify-between pb-6 border-b border-border mb-6">
        {/* <h1 className="text-2xl font-semibold flex items-center gap-2">
          <Bell className="w-6 h-6" /> Notifications
        </h1> */}
        <h1 className="text-2xl font-semibold flex items-center gap-2 relative">
          <Bell className="w-6 h-6" />
          Notifications
          {unreadCount > 0 && (
            <span className="absolute -top-2 -right-2 bg-red-600 text-white text-xs px-2 py-0.5 rounded-full">
              {unreadCount}
            </span>
          )}
        </h1>

        <div className="flex items-center gap-4">
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
            <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
              Evidence Viewer
            </button>
          </Link>
          <Link to="/secure-chat">
            <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
              Secure Chat
            </button>
          </Link>

        </div>
      
        <div className="flex items-center gap-3">
          <input
            type="text"
            placeholder="Search notifications..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="bg-input border border-border rounded-lg px-3 py-2 text-sm w-64"
          />
          <button
            onClick={markAsRead}
            className="bg-green-600 text-white rounded px-3 py-2 text-sm hover:bg-green-500"
          >
            Mark as Read
          </button>
          <button
            onClick={archiveSelected}
            className="bg-yellow-600 text-white rounded px-3 py-2 text-sm hover:bg-yellow-500"
          >
            Archive
          </button>
          <button
            onClick={deleteSelected}
            className="bg-red-600 text-white rounded px-3 py-2 text-sm hover:bg-red-500"
          >
            <Trash2 className="w-4 h-4 inline mr-1" />
            Delete
          </button>
        </div>
      </div>

      <div className="flex gap-4 mb-4">
        <button
          onClick={() => setFilter("all")}
          className={cn(
            "px-4 py-2 rounded-lg text-sm",
            filter === "all"
              ? "bg-blue-600 text-white"
              : "bg-muted text-muted-foreground"
          )}
        >
          <Inbox className="inline w-4 h-4 mr-1" />
          All
        </button>
        <button
          onClick={() => setFilter("unread")}
          className={cn(
            "px-4 py-2 rounded-lg text-sm",
            filter === "unread"
              ? "bg-blue-600 text-white"
              : "bg-muted text-muted-foreground"
          )}
        >
          Unread
        </button>
        <button
          onClick={() => setFilter("archived")}
          className={cn(
            "px-4 py-2 rounded-lg text-sm",
            filter === "archived"
              ? "bg-blue-600 text-white"
              : "bg-muted text-muted-foreground"
          )}
        >
          <Archive className="inline w-4 h-4 mr-1" />
          Archived
        </button>
      </div>

      {/* Notification List */}
      <div className="space-y-3">
        {filteredNotifications.map((n) => (
          <div
            key={n.id}
            className={cn(
              "flex items-start justify-between p-3 border border-border rounded-lg bg-card text-card-foreground",
              !n.read && "border-blue-500"
            )}
          >
            <div className="flex items-start gap-3">
              <input
                type="checkbox"
                checked={selected.includes(n.id)}
                onChange={() => toggleSelect(n.id)}
                className="mt-1"
              />
              <div className="text-sm">
                <p className="font-semibold">{n.title}</p>
                <p className="text-muted-foreground">{n.message}</p>
                <p className="text-xs text-muted-foreground mt-1">{n.timestamp}</p>
              </div>
            </div>
            {!n.read && (
              <span className="text-xs px-2 py-1 bg-blue-500 text-white rounded-lg">
                New
              </span>
            )}
          </div>
        ))}
      </div>

      {/* Pagination (placeholder) */}
      <div className="flex justify-end mt-6 gap-3">
        <button className="px-3 py-1 rounded bg-muted text-muted-foreground hover:bg-muted/70">
          Prev
        </button>
        <button className="px-3 py-1 rounded bg-muted text-muted-foreground hover:bg-muted/70">
          Next
        </button>
      </div>
    </div>
  );
};

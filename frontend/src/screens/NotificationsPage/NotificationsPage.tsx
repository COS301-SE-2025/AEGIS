import {
  Trash2,
  Inbox,
  Archive,
  Bell,
} from "lucide-react";
import { useEffect, useState } from "react";
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

const mockData: Notification[] = [
  {
    id: "1",
    title: "Case Updated",
    message: "Case #123 was updated by Responder Alpha.",
    timestamp: "2025-07-24 14:32",
    read: false,
    archived: false,
  },
  {
    id: "2",
    title: "Evidence Uploaded",
    message: "New evidence uploaded to Case #456.",
    timestamp: "2025-07-23 10:12",
    read: true,
    archived: false,
  },
  {
    id: "3",
    title: "New Assignment",
    message: "You were assigned to Case #789.",
    timestamp: "2025-07-22 09:00",
    read: false,
    archived: true,
  },
];

export const NotificationsPage = () => {
  const [notifications, setNotifications] = useState<Notification[]>(mockData);
  const [selected, setSelected] = useState<string[]>([]);
  const [filter, setFilter] = useState<"all" | "unread" | "archived">("all");
  const [searchQuery, setSearchQuery] = useState("");

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

  const markAsRead = () => {
    setNotifications((prev) =>
      prev.map((n) =>
        selected.includes(n.id) ? { ...n, read: true } : n
      )
    );
    setSelected([]);
  };

  const archiveSelected = () => {
    setNotifications((prev) =>
      prev.map((n) =>
        selected.includes(n.id) ? { ...n, archived: true } : n
      )
    );
    setSelected([]);
  };

  const deleteSelected = () => {
    setNotifications((prev) =>
      prev.filter((n) => !selected.includes(n.id))
    );
    setSelected([]);
  };

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
        <h1 className="text-2xl font-semibold flex items-center gap-2">
          <Bell className="w-6 h-6" /> Notifications
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

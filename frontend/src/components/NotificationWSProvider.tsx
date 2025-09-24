// src/components/NotificationsWSProvider.tsx
import { useEffect, useRef, useState } from "react";
import {  setUnreadCount,decUnread,incUnread } from "../lib/unreadCount";

const AUTH_EVENT = "auth:updated"; // we'll dispatch this on login

export default function NotificationsWSProvider() {
  const wsRef = useRef<WebSocket | null>(null);

  const [token, setToken] = useState<string | null>(() => sessionStorage.getItem("authToken"));
  const [tenantId, setTenantId] = useState<string | null>(() => sessionStorage.getItem("tenantId"));

  // Keep creds in sync when login writes sessionStorage
  useEffect(() => {
    const sync = () => {
      const t = sessionStorage.getItem("authToken");
      const ten = sessionStorage.getItem("tenantId");
      if (t !== token) setToken(t);
      if (ten !== tenantId) setTenantId(ten);
    };

    // listen for our custom login event (same tab)
    const onAuth = () => sync();
    window.addEventListener(AUTH_EVENT, onAuth);

    // small poll window right after app starts (helps if login just happened)
    const poll = setInterval(sync, 300);
    const stopPoll = setTimeout(() => clearInterval(poll), 5000);

    return () => {
      window.removeEventListener(AUTH_EVENT, onAuth);
      clearInterval(poll);
      clearTimeout(stopPoll);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // (Re)connect socket whenever creds ready/changed
  useEffect(() => {
    if (!token || !tenantId) return;

    // bootstrap unread so the badge shows right away
    (async () => {
      try {
        const res = await fetch("/api/v1/notifications", {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const data = await res.json();
          const unread = Array.isArray(data)
            ? data.filter((n: any) => !n.read && !n.archived).length
            : (data?.count ?? 0);
          setUnreadCount(unread);
        }
      } catch {}
    })();

    // close previous socket (if any)
    wsRef.current?.close();

    // IMPORTANT: confirm this path is correct.
    // Your logs say server logs: “WebSocket route hit … in case <id>”
    // If the server expects a CASE ID, do not pass a tenantId here.
    const ws = new WebSocket(`ws://localhost:8080/ws/cases/${tenantId}?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => console.log("[WS] notifications connected");
    ws.onclose = (e) => console.log("[WS] notifications closed", e.code, e.reason);
    ws.onerror = (e) => console.warn("[WS] error", e);

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);

        if (msg.type === "notification" || msg.type === "EventNotification") {
           incUnread(1);
        }

        if (msg.type === "mark_notification_read") {
          const ids: string[] = msg.payload?.notificationIds ?? [];
           if (ids.length) decUnread(ids.length);
        }
      } catch (err) {
        console.error("[WS] parse error", err);
      }
    };

    return () => ws.close();
  }, [token, tenantId]);

  return null;
}

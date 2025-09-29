const KEY = "aegis:unreadCount";
const EVT = "unread:changed";
const CHANNEL = "aegis-notifications";

const bc =
  typeof window !== "undefined" && "BroadcastChannel" in window
    ? new BroadcastChannel(CHANNEL)
    : null;

export function getUnreadCount(): number {
  if (typeof window === "undefined") return 0;
  const raw = window.localStorage.getItem(KEY);
  return raw ? Number(raw) || 0 : 0;
}

export function setUnreadCount(n: number) {
  if (typeof window === "undefined") return;
  const prev = getUnreadCount();
  if (prev === n) return; // avoid noisy writes
  window.localStorage.setItem(KEY, String(n));
  bc?.postMessage({ type: EVT, value: n }); // other tabs
  window.dispatchEvent(new CustomEvent(EVT, { detail: n })); // same tab
}

export function adjustUnreadCount(
  updater: number | ((prev: number) => number)
) {
  const prev = getUnreadCount();
  const next =
    typeof updater === "function" ? (updater as (p: number) => number)(prev) : updater;
  setUnreadCount(Math.max(0, next));
}

export function incUnread(delta = 1) {
  adjustUnreadCount((p) => p + delta);
}
export function decUnread(delta = 1) {
  adjustUnreadCount((p) => p - delta);
}

export function subscribe(cb: (n: number) => void): () => void {
  if (typeof window === "undefined") return () => {};
  const onCustom = (e: Event) => cb((e as CustomEvent).detail as number);
  const onStorage = (e: StorageEvent) => {
    if (e.key === KEY) cb(e.newValue ? Number(e.newValue) || 0 : 0);
  };
  const onBC = (e: MessageEvent) => {
    if (e.data?.type === EVT) cb(Number(e.data.value) || 0);
  };

  window.addEventListener(EVT, onCustom);
  window.addEventListener("storage", onStorage);
  bc?.addEventListener("message", onBC);

  return () => {
    window.removeEventListener(EVT, onCustom);
    window.removeEventListener("storage", onStorage);
    bc?.removeEventListener("message", onBC);
  };
}

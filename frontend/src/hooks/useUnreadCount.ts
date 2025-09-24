import { useEffect, useState } from "react";
import { getUnreadCount, subscribe } from "../lib/unreadCount";


export function useUnreadCount() {
  const [count, setCount] = useState<number>(() => getUnreadCount());
  useEffect(() => subscribe(setCount), []);
  return count;
}

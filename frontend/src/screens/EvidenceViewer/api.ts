
interface EvidenceMetadata {
  description?: string;
  status?: string;
  chainOfCustody?: string[];
  acquisitionDate?: string;
  acquisitionTool?: string;
  integrityCheck?: string;
  threadCount?: number;
  priority?: string;
}

const baseURL = "https://localhost/api/v1";

// EvidenceViewer/api.ts
export async function fetchEvidenceByCaseId(caseId: string) {
  const token = sessionStorage.getItem("authToken") || "";
  const res = await fetch(`https://localhost/api/v1/evidence-metadata/case/${caseId}`, {
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
  });

  if (!res.ok) throw new Error("Failed to fetch evidence");

 const data = await res.json();
 console.log("Fetched evidence data api:", data);

return data.map((item: any) => {
  let meta: EvidenceMetadata = {};
  try {
     const parsed = item.metadata ? JSON.parse(item.metadata) : null;
    meta = parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    meta = {};
  }

  return {
    ...item,
    description: meta.description || "",
    status: meta.status || "pending",
    chainOfCustody: meta.chainOfCustody || [],
    acquisitionDate: meta.acquisitionDate || "",
    acquisitionTool: meta.acquisitionTool || "",
    integrityCheck: meta.integrityCheck || "pending",
    threadCount: meta.threadCount || 0,
    priority: meta.priority || "low",
  };
});
}


export async function createThread(data: {
  case_id: string;
  file_id: string;
  user_id: string;
  title: string;
  tags: string[];
  priority: string;
}) {
  const token = sessionStorage.getItem("authToken") || "";

  const res = await fetch(`https://localhost/api/v1/threads`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });

  if (!res.ok) throw new Error("Failed to create thread");
  return await res.json();
}


export async function fetchThreadsByFile(fileID: string) {
  const token = sessionStorage.getItem("authToken");
  const res = await fetch(`https://localhost/api/v1/threads/file/${fileID}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!res.ok) throw new Error("Failed to fetch threads");
  return res.json();
}


// export async function sendThreadMessage(threadId: string, body: any) {
//   const token = sessionStorage.getItem("authToken");

//   const res = await fetch(`https://localhost/api/v1/threads/${threadId}/messages`, {
//     method: "POST",
//     headers: {
//       "Content-Type": "application/json",
//       Authorization: `Bearer ${token}`,
//     },
//     body: JSON.stringify(body),
//   });

//   if (!res.ok) throw new Error("Failed to send message");
//   return res.json();
// }

export async function sendThreadMessage(threadId: string, body: any) {
  const token = sessionStorage.getItem("authToken");

  const res = await fetch(`https://localhost/api/v1/threads/${threadId}/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
  });

  console.log("API response status:", res.status);
  console.log("API response headers:", Object.fromEntries(res.headers.entries()));

  if (!res.ok) {
    const errorText = await res.text();
    console.error("API error response:", errorText);
    throw new Error(`Failed to send message: ${res.status} - ${errorText}`);
  }
  
  return res.json();
}


export async function fetchThreadMessages(threadID: string) {
  const token = sessionStorage.getItem("authToken");
  const res = await fetch(`https://localhost/api/v1/threads/${threadID}/messages`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!res.ok) throw new Error("Failed to fetch messages");
  return res.json();
}

export async function createAnnotationThread(data: {
  case_id: string;
  file_id: string;
  user_id: string;
  title: string;
  tags: string[];
  priority: string; // should be 'high' | 'medium' | 'low'
}) {
  const token = sessionStorage.getItem("authToken");

  const res = await fetch("https://localhost/api/v1/threads", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`
    },
    body: JSON.stringify(data)
  });

  if (!res.ok) throw new Error("Failed to create annotation thread");

  return await res.json(); // the AnnotationThread
}


export async function addThreadParticipant(threadId: string, userId: string) {
  const token = sessionStorage.getItem("authToken");

  const res = await fetch(`${baseURL}/threads/${threadId}/participants`, {
    method: "POST",
    headers: { 
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`
    },
    body: JSON.stringify({ user_id: userId })
  });

  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.message || "Failed to add participant");
  }

  return await res.json();
}

export async function fetchThreadParticipants(threadId: string) {
  const token = sessionStorage.getItem("authToken");

  const res = await fetch(`${baseURL}/threads/${threadId}/participants`, {
    headers: { Authorization: `Bearer ${token}` }
  });

  if (!res.ok) throw new Error("Failed to fetch participants");
  return await res.json();
}

export const approveMessage = async (messageID: string) => {
  const token = sessionStorage.getItem("authToken");

  const res = await fetch(`${baseURL}/messages/${messageID}/approve`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`
    }
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.message || "Failed to approve message");
  }
};

export const addReaction = async (messageID: string, userID: string, type: string) => {
  const token = sessionStorage.getItem("authToken");

  const res = await fetch(`${baseURL}/messages/${messageID}/reactions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`
    },
    body: JSON.stringify({ user_id: userID, reaction: type }) // backend expects 'reaction' not 'type'
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.message || "Failed to add reaction");
  }
  return await res.json(); // return updated message object
};

export const removeReaction = async (messageID: string, userID: string) => {
  const token = sessionStorage.getItem("authToken");

  const res = await fetch(`${baseURL}/messages/${messageID}/reactions/${userID}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`
    }
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.message || "Failed to remove reaction");
  }
};

//TO BE DELETED
// export function nestMessages(flatMessages: ThreadMessage[]): ThreadMessage[] {
//   const messageMap = new Map<string, ThreadMessage>();
//   const roots: ThreadMessage[] = [];

//   // Prepare messages with empty replies
//   flatMessages.forEach(msg => {
//     messageMap.set(msg.id, { ...msg, replies: msg.replies ?? [] });
//   });

//   flatMessages.forEach(msg => {
//     if (msg.parent_message_id) {
//       const parent = messageMap.get(msg.parent_message_id);
//       if (parent) {
//         parent.replies!.push(messageMap.get(msg.id)!);
//       }
//     } else {
//       roots.push(messageMap.get(msg.id)!);
//     }
//   });

//   return roots;
// }


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



// EvidenceViewer/api.ts
export async function fetchEvidenceByCaseId(caseId: string) {
  const token = sessionStorage.getItem("authToken") || "";
  const res = await fetch(`http://localhost:8080/api/v1/evidence-metadata/case/${caseId}`, {
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

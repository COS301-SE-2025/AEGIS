import React, { useState } from "react";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Textarea } from "../../components/ui/TextArea";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "../../components/ui/select";
import { ShieldAlert } from "lucide-react";

export function CreateCaseForm(): JSX.Element {
  const [form, setForm] = useState({
    creator: "",
    team: "",
    priority: "",
    attackType: "",
    description: "",
  });

  type CreateCaseFormField = keyof typeof form;
  // Mock activity logging function
  const logActivity = (caseId: string, action: string, details: any = {}) => {
    const activities = JSON.parse(localStorage.getItem("caseActivities") || "[]");
    
    const newActivity = {
      id: `activity-${Date.now()}-${Math.random().toString(36).substring(2, 6)}`,
      caseId,
      action,
      details,
      timestamp: new Date().toISOString(),
      user: form.creator || "Unknown User",
      userRole: "Case Creator"
    };

    activities.push(newActivity);
    localStorage.setItem("caseActivities", JSON.stringify(activities));
    
    // Optional: Console log for debugging
    console.log("Activity logged:", newActivity);
  };

  const handleChange =
    (field: CreateCaseFormField) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setForm({ ...form, [field]: e.target.value });
    };

  // const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
  //   e.preventDefault();



  //   const stored = localStorage.getItem("cases");
  //   const cases = stored ? JSON.parse(stored) : [];

  //   const newId = cases.length > 0 ? Math.max(...cases.map((c: any) => c.id || 0)) + 1 : 1;

  //   const newCase = {
  //     id: newId,
  //     ...form,
  //     lastActivity: new Date().toISOString().split("T")[0],
  //     progress: 0,
  //     image:
  //       "https://th.bing.com/th/id/OIP.kq_Qib5c_49zZENmpMnuLQHaDt?w=331&h=180&c=7&r=0&o=5&dpr=1.3&pid=1.7",
  //   };

  //   cases.push(newCase);
  //   localStorage.setItem("cases", JSON.stringify(cases));

  //    // Log the case creation activity
  //   logActivity(newId.toString(), "Case Created", {
  //     priority: form.priority,
  //     attackType: form.attackType,
  //     team: form.team,
  //     description: form.description.substring(0, 100) + "..." // Truncate for logging
  //   });

    
  //   window.location.href = "/dashboard";
  // };

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
  e.preventDefault();

  const stored = localStorage.getItem("cases");
  const cases = stored ? JSON.parse(stored) : [];

  const pendingCaseId = Number(localStorage.getItem("pendingCaseId"));
  let updatedCases;

  const now = new Date().toISOString();

  const updatedCase = {
  id: pendingCaseId || (cases.length > 0 ? Math.max(...cases.map((c: any) => c.id || 0)) + 1 : 1),
  ...form,
  lastActivity: now.split("T")[0],
  createdAt: cases.find((c: any) => c.id === pendingCaseId)?.createdAt || now, // preserve original createdAt if exists
  updatedAt: now, // update on final submit
  progress: 0,
  image: "https://th.bing.com/th/id/OIP.kq_Qib5c_49zZENmpMnuLQHaDt?w=331&h=180&c=7&r=0&o=5&dpr=1.3&pid=1.7",
  };

  const existingIndex = cases.findIndex((c: any) => c.id === updatedCase.id);
  if (existingIndex >= 0) {
    cases[existingIndex] = updatedCase;
    updatedCases = cases;
  } else {
    updatedCases = [...cases, updatedCase];
  }

  localStorage.setItem("cases", JSON.stringify(updatedCases));
  localStorage.removeItem("pendingCaseId");

  logActivity(updatedCase.id.toString(), "Case Created", {
    priority: form.priority,
    attackType: form.attackType,
    team: form.team,
    description: form.description.substring(0, 100) + "..."
  });

  window.location.href = "/dashboard";
};


 const ensureCaseExists = () => {
  let pendingCaseId = localStorage.getItem("pendingCaseId");
  const stored = localStorage.getItem("cases");
  const cases = stored ? JSON.parse(stored) : [];

  if (pendingCaseId) {
    const existing = cases.find((c: any) => String(c.id) === pendingCaseId);
    if (existing) return Number(pendingCaseId);
  }

  const newId = cases.length > 0 ? Math.max(...cases.map((c: any) => c.id || 0)) + 1 : 1;

  const now = new Date().toISOString();

  const placeholderCase = {
  id: newId,
  creator: form.creator || "Unknown",
  team: form.team || "Unassigned",
  priority: form.priority || "low",
  attackType: form.attackType || "",
  description: form.description || "",
  lastActivity: now.split("T")[0],
  createdAt: now,
  updatedAt: now,
  progress: 0,
  image: "https://th.bing.com/th/id/OIP.kq_Qib5c_49zZENmpMnuLQHaDt?w=331&h=180&c=7&r=0&o=5&dpr=1.3&pid=1.7"
  };


  localStorage.setItem("pendingCaseId", String(newId));
  localStorage.setItem("cases", JSON.stringify([...cases, placeholderCase]));

  return newId;
};


  return (
    <div className="min-h-screen bg-background text-foreground flex items-center justify-center p-6">
      <div className="max-w-3xl mx-auto mt-10 bg-card text-foreground p-6 rounded-2xl shadow-xl border border-border font-mono w-full">
        <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">
          <ShieldAlert size={28} /> Create New Case
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Creator */}
          <div>
            <label className="block mb-1 text-sm">Name of Person Creating the Case</label>
            <Input
              className="bg-muted border-border text-foreground placeholder-muted-foreground"
              placeholder="e.g. Alice Johnson"
              value={form.creator}
              onChange={handleChange("creator")}
              required
            />
          </div>

          {/* Team */}
          <div>
            <label className="block mb-1 text-sm">Team Name</label>
            <Input
              className="bg-muted border-border text-foreground placeholder-muted-foreground"
              placeholder="e.g. AEGIS Forensics"
              value={form.team}
              onChange={handleChange("team")}
              required
            />
          </div>

          {/* Priority */}
          <div>
            <label className="block mb-1 text-sm">Case Priority</label>
            <Select onValueChange={(value: string) => setForm({ ...form, priority: value })}>
              <SelectTrigger className="bg-muted border-border text-foreground">
                <SelectValue placeholder="Select priority" />
              </SelectTrigger>
              <SelectContent className="bg-zinc-800 text-popover-foreground">
                <SelectItem value="low">Low</SelectItem>
                <SelectItem value="mid">Mid</SelectItem>
                <SelectItem value="high">High</SelectItem>
                <SelectItem value="critical">Critical</SelectItem>
                <SelectItem value="time-sensitive">Time Sensitive</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Attack Type */}
          <div>
            <label className="block mb-1 text-sm">Type of Attack</label>
            <Input
              className="bg-muted border-border text-foreground placeholder-muted-foreground"
              placeholder="e.g. Ransomware, Malware, Phishing"
              value={form.attackType}
              onChange={handleChange("attackType")}
            />
          </div>

          {/* Description */}
          <div>
            <label className="block mb-1 text-sm">Short Description</label>
            <Textarea
              className="bg-muted border-border text-foreground placeholder-muted-foreground"
              placeholder="Brief summary of the incident..."
              value={form.description}
              onChange={handleChange("description")}
              rows={4}
            />
          </div>

          {/* Actions */}
          <div className="flex flex-wrap gap-4 pt-4">
            <Button
              type="button"
              variant="outline"
              className="border-cyan-500 text-primary hover:bg-cyan-800/10"
              onClick={() => {
                //logActivity("temp-id", "Assigned Members");
                const id = ensureCaseExists();
                logActivity(id.toString(), "Assigned Members");
                window.location.href = "/assign-case-members";
              }}
            >
              Assign Case Members
            </Button>

            <Button
              type="button"
              variant="outline"
              className="border-purple-500 text-purple-500 hover:bg-purple-500/10"
              onClick={() => {
                //logActivity("temp-id", "Uploaded Evidence");
                const id = ensureCaseExists();
                logActivity(id.toString(), "Uploaded Evidence");
                window.location.href = "/upload-evidence";
              }}
            >
              Upload Evidence
            </Button>

            <Button
              type="submit"
              className="bg-cyan-600 hover:bg-cyan-700 text-white"
            >
              Create Case
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

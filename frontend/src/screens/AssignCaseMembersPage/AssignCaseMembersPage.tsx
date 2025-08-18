import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "../../components/ui/select";
import { UserPlus2 } from "lucide-react";

const dfirRoles = [
  "Incident Responder", "Forensic Analyst", "Malware Analyst",
  "Threat Intelligent Analyst", "DFIR Manager", "Detection Engineer",
  "Network Evidence Analyst", "Image Forensics Analyst", "Disk Forensics Analyst",
  "Log Analyst", "Mobile Device Analyst", "Memory Forensics Analyst",
  "Cloud Forensics Specialist", "Endpoint Forensics Analyst", "Reverse Engineer",
  "SIEM Analyst", "Vulnerability Analyst", "Digital Evidence Technician",
  "Packet Analyst", "Legal/Compliance Liaison", "Compliance Officer",
  "Legal Counsel", "Policy Analyst", "SOC Analyst", "Incident Commander",
  "Crisis Communications Officer", "IT Infrastructure Liaison", "Triage Analyst",
  "Evidence Archivist", "Training Coordinator", "Audit Reviewer", "Threat Hunter"
];

export function AssignCaseMembersForm(): JSX.Element {
  const navigate = useNavigate();

  const [members, setMembers] = useState<{ user: string; role: string }[]>([]);
  const [availableUsers, setAvailableUsers] = useState<{ id: string; name: string }[]>([]);

useEffect(() => {
  const fetchUsers = async () => {
    try {
      const res = await fetch("http://localhost:8080/api/v1/users", {
        headers: {
          "Authorization": `Bearer ${sessionStorage.getItem("authToken") || ""}`
        }
      });
      if (!res.ok) throw new Error("Failed to fetch users");
      const data = await res.json();

      const users = Array.isArray(data.data)
        ? data.data
            .filter((u: any) => u.FullName && u.ID)
            .map((u: any) => ({ id: u.ID, name: u.FullName }))
        : [];

      setAvailableUsers(users);
    } catch (err) {
      console.error("Fetch failed:", err);
      alert("Failed to load users");
    }
  };
  fetchUsers();
}, []);


  const handleMemberUserChange = (index: number, value: string) => {
    const updated = [...members];
    updated[index].user = value;
    setMembers(updated);
  };

  const handleMemberRoleChange = (index: number, value: string) => {
    const updated = [...members];
    updated[index].role = value;
    setMembers(updated);
  };

  const addMember = () => {
    if (members.length < 10) {
      setMembers([...members, { user: "", role: "" }]);
    }
  };

  const removeMember = (index: number) => {
    setMembers(members.filter((_, i) => i !== index));
  };

const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();

  if (members.length === 0) {
    alert("Please add at least one member.");
    return;
  }

  for (const m of members) {
    if (!m.user.trim() || !m.role.trim()) {
      alert("Please fill in user and role for all members.");
      return;
    }
  }

  const currentCase = JSON.parse(localStorage.getItem("currentCase") || "{}");
  const currentCaseId = currentCase.ID || currentCase.id;

  if (!currentCaseId) {
    alert("No active case found. Please create or select a case first.");
    return;
  }

  try {
    for (const member of members) {
      const user = availableUsers.find(u => u.name === member.user);
      if (!user) {
        throw new Error(`Could not find user ID for ${member.user}`);
      }

      const res = await fetch("http://localhost:8080/api/v1/cases/assign", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${sessionStorage.getItem("authToken") || ""}`
        },
        body: JSON.stringify({
          assignee_id: user.id,
          case_id: currentCaseId,
          role: member.role
        })
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Failed to assign user");
      }
    }

    alert("All members assigned successfully!");
    navigate(-1);
  } catch (err: any) {
    console.error(err);
    alert(`Error assigning members: ${err.message}`);
  }
};

  return (
    <div className="min-h-screen bg-background text-foreground flex items-center justify-center p-6">
      <div className="max-w-4xl w-full bg-card p-6 rounded-2xl shadow-xl border border-border font-mono">
        <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">
          <UserPlus2 size={28} /> Assign Case Members
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div className="space-y-4">
            {members.map((member, idx) => (
              <div
                key={idx}
                className="flex flex-wrap gap-4 items-center border border-border rounded p-3 bg-muted"
              >
                <div className="flex-1 min-w-[220px]">
                  <label className="block mb-1 text-sm text-muted-foreground">
                    Choose Member
                  </label>
                  <Select
                    value={member.user}
                    onValueChange={(val) => handleMemberUserChange(idx, val)}
                    disabled={availableUsers.length === 0}
                  >
                    <SelectTrigger className="bg-background border-border text-foreground w-full">
                      <SelectValue
                        placeholder={
                          availableUsers.length === 0
                            ? "Loading users..."
                            : "Select member"
                        }
                      />
                    </SelectTrigger>
                    <SelectContent className="bg-muted text-foreground max-h-40 overflow-y-auto">
                      {availableUsers.length === 0 ? (
                        <div className="p-2 text-muted-foreground">
                          No users loaded
                        </div>
                      ) : (
                        availableUsers
                          .filter(u => !members.some((m, i) => m.user === u.name && i !== idx))
                          .slice(0, 100)
                          .map((user) => (
                            <SelectItem key={user.id} value={user.name}>
                              {user.name}
                            </SelectItem>
                          ))
                      )}
                      {availableUsers.length > 100 && (
                        <div className="p-2 text-xs text-muted-foreground italic">
                          Showing first 100 users...
                        </div>
                      )}
                    </SelectContent>
                  </Select>
                </div>

                <div className="flex-1 min-w-[220px]">
                  <label className="block mb-1 text-sm text-muted-foreground">
                    Assign Role
                  </label>
                  <Select
                    value={dfirRoles.includes(member.role) ? member.role : ""}
                    onValueChange={(val: string) => handleMemberRoleChange(idx, val)}
                  >
                    <SelectTrigger className="bg-background border-border text-foreground w-full">
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent
                      side="bottom"
                      align="start"
                      avoidCollisions={false}
                      className="bg-muted text-foreground max-h-[30rem] overflow-y-auto"
                    >
                      {dfirRoles.slice(0, 100).map((role) => (
                        <SelectItem key={role} value={role}>
                          {role}
                        </SelectItem>
                      ))}
                      {dfirRoles.length > 100 && (
                        <div className="p-2 text-xs text-muted-foreground italic">
                          Showing first 100 roles...
                        </div>
                      )}
                    </SelectContent>
                  </Select>

                  <Input
                    type="text"
                    placeholder="Or type custom role"
                    className="mt-1 bg-background border-border text-foreground"
                    value={member.role}
                    onChange={(e) => handleMemberRoleChange(idx, e.target.value)}
                  />
                </div>

                <button
                  type="button"
                  onClick={() => removeMember(idx)}
                  className="self-start text-red-500 hover:text-red-700 font-bold text-lg"
                  aria-label="Remove member"
                >
                  &times;
                </button>
              </div>
            ))}
          </div>

          {members.length < 10 && (
            <Button
              type="button"
              variant="outline"
              className="border-muted-foreground border-cyan-500 text-cyan-400 hover:bg-cyan-800"
              onClick={addMember}
            >
              + Add Member
            </Button>
          )}

          <div className="flex gap-4 pt-4">
            <Button
              type="button"
              variant="outline"
              className="border-muted-foreground text-muted-foreground hover:bg-muted"
              onClick={() => navigate(-1)}
            >
              Back
            </Button>

            <Button
              type="submit"
              className="bg-cyan-600 hover:bg-cyan-700 text-white"
            >
              Done
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

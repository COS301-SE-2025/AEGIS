import React, { useState } from "react";
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

const exampleUsers = [
  "Alice Johnson",
  "Bob Smith",
  "Charlie Davis",
  "Diana Evans",
  "Ethan Ford",
  "Fiona Green",
  "George Hill",
  "Hannah Ingram",
  "Ian Jones",
  "Jessica King",
  "Kevin Lee",
  "Laura Miller",
];

const dfirRoles = [
  "Incident Responder",
  "Forensic Analyst",
  "Malware Analyst",
  "Threat Hunter",
  "Network Analyst",
  "Digital Forensics Investigator",
  "SOC Analyst",
  "Log Analyst",
  "Malware Reverse Engineer",
  "Case Manager",
];

export function AssignCaseMembersForm(): JSX.Element {
  const navigate = useNavigate();

  const [members, setMembers] = useState<{ user: string; role: string }[]>([]);
  const [userSearch, setUserSearch] = useState("");

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

  const filteredUsers = exampleUsers.filter(
    (u) =>
      u.toLowerCase().includes(userSearch.toLowerCase()) &&
      !members.some((m) => m.user === u)
  );

  const handleSubmit = (e: React.FormEvent) => {
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
    // alert("Members assigned!");
    // console.log("Assigned members:", members);

    
  const cases = JSON.parse(localStorage.getItem("cases") || "[]");
const currentCase = cases[cases.length - 1]; // last created

const currentCaseId = String(currentCase.id);

const allCaseMembers = JSON.parse(localStorage.getItem("caseMembers") || "[]");

// remove old entry for the same case, if any
const updatedCaseMembers = allCaseMembers.filter(
  (entry: any) => entry.caseId !== currentCaseId
);

// add new entry
updatedCaseMembers.push({
  caseId: currentCaseId,
  members: members.map(m => ({ name: m.user, role: m.role }))
});

// store updated list
localStorage.setItem("caseMembers", JSON.stringify(updatedCaseMembers));

alert("Members assigned!");
navigate("/create-case"); 
  };


  return (
    <div className="min-h-screen bg-background text-foreground flex items-center justify-center p-6">
      <div className="max-w-4xl w-full bg-card p-6 rounded-2xl shadow-xl border border-border font-mono">
        <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">          <UserPlus2 size={28} /> Assign Case Members
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Members */}
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
                  <Input
                    type="text"
                    list={`users-${idx}`}
                    placeholder="Type or select user"
                    className="bg-background border-border text-foreground"
                    value={member.user}
                    onChange={(e) => {
                      handleMemberUserChange(idx, e.target.value);
                      setUserSearch(e.target.value);
                    }}
                    required
                  />
                  <datalist id={`users-${idx}`}>
                    {filteredUsers.map((user) => (
                      <option key={user} value={user} />
                    ))}
                  </datalist>
                </div>

                <div className="flex-1 min-w-[220px]">
                  <label className="block mb-1 text-sm text-muted-foreground">
                    Assign Role
                  </label>
                  <Select
                    value={member.role}
                    onValueChange={(val: string) => handleMemberRoleChange(idx, val)}
                  >
                    <SelectTrigger className="bg-background border-border text-foreground w-full">
                      <SelectValue placeholder="Select or type role" />
                    </SelectTrigger>
                    <SelectContent className="bg-muted text-foreground max-h-40 overflow-y-auto">
                      {dfirRoles.map((role) => (
                        <SelectItem key={role} value={role}>
                          {role}
                        </SelectItem>
                      ))}
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

          {/* Add Member */}
          {members.length < 10 && (
            <Button
              type="button"
              variant="outline"
              className="border-muted-foreground border-cyan-500 text-cyan-400 hover:bg-cyan-800"              onClick={addMember}
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

            <Button type="submit" 
            className="bg-cyan-600 hover:bg-cyan-700 text-white"
            onClick={() => navigate(-1)}>
                   Done
           </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

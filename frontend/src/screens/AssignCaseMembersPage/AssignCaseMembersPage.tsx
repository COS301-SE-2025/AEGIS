import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "../../components/ui/select";
import { ShieldUser, UserPlus2 } from "lucide-react";

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

export function AssignCaseMembersForm():JSX.Element {
  const navigate = useNavigate();

  // State to store selected members, each with user & role
  const [members, setMembers] = useState<
    { user: string; role: string }[]
  >([]);

  // For user input search/filter for users dropdown
  const [userSearch, setUserSearch] = useState("");

  // For role free-text input per member
  const handleMemberUserChange = (index: number, value: string) => {
    const newMembers = [...members];
    newMembers[index].user = value;
    setMembers(newMembers);
  };

  const handleMemberRoleChange = (index: number, value: string) => {
    const newMembers = [...members];
    newMembers[index].role = value;
    setMembers(newMembers);
  };

  // Add new empty member slot (max 10)
  const addMember = () => {
    if (members.length < 10) {
      setMembers([...members, { user: "", role: "" }]);
    }
  };

  // Remove member by index
  const removeMember = (index: number) => {
    const newMembers = members.filter((_, i) => i !== index);
    setMembers(newMembers);
  };

  // Filter users for dropdown based on search input
  const filteredUsers = exampleUsers.filter(
    (u) => u.toLowerCase().includes(userSearch.toLowerCase()) && !members.some((m) => m.user === u)
  );

  // Submit handler
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (members.length === 0) {
      alert("Please add at least one member.");
      return;
    }
    // Validate members have user and role
    for (const m of members) {
      if (!m.user.trim() || !m.role.trim()) {
        alert("Please fill in user and role for all members.");
        return;
      }
    }
    alert("Members assigned!");
    console.log("Assigned members:", members);
  };

  return (
    <div className="min-h-screen bg-zinc-900 text-white flex items-center justify-center p-6">
      <div className="max-w-4xl w-full bg-zinc-900 p-6 rounded-2xl shadow-xl border border-zinc-700 font-mono">
        <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">
          <UserPlus2 size={28} /> Assign Case Members
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Members List */}
          <div className="space-y-4">
            {members.map((member, idx) => (
              <div
                key={idx}
                className="flex flex-wrap gap-4 items-center border border-zinc-700 rounded p-3 bg-zinc-800"
              >
                {/* User select with free text + dropdown */}
                <div className="flex-1 min-w-[220px]">
                  <label className="block mb-1 text-sm">Choose Member</label>
                  <Input
                    type="text"
                    list={`users-list-${idx}`}
                    placeholder="Type or select user"
                    className="bg-zinc-700 border-zinc-600 text-white"
                    value={member.user}
                    onChange={(e) => {
                      handleMemberUserChange(idx, e.target.value);
                      setUserSearch(e.target.value);
                    }}
                    required
                  />
                  <datalist id={`users-list-${idx}`}>
                    {filteredUsers.map((user) => (
                      <option key={user} value={user} />
                    ))}
                  </datalist>
                </div>

                {/* Role select + free text */}
                <div className="flex-1 min-w-[220px]">
                  <label className="block mb-1 text-sm">Assign Role</label>
                  <Select
                    value={member.role}
                    onValueChange={(value: string) => handleMemberRoleChange(idx, value)}
                  >
                    <SelectTrigger className="bg-zinc-700 border-zinc-600 text-white w-full">
                      <SelectValue placeholder="Select or type role" />
                    </SelectTrigger>
                    <SelectContent className="bg-zinc-800 text-white max-h-40 overflow-y-auto">
                      {dfirRoles.map((role: string) => (
                        <SelectItem key={role} value={role}>
                          {role}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {/* Allow free text role override */}
                  <Input
                    type="text"
                    placeholder="Or type custom role"
                    className="mt-1 bg-zinc-700 border-zinc-600 text-white"
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

          {/* Add member button */}
          {members.length < 10 && (
            <Button
              type="button"
              variant="outline"
              className="bg-zinc-800 border-cyan-500 text-cyan-400 hover:bg-cyan-800"
              onClick={addMember}
            >
              + Add Member
            </Button>
          )}

          <div className="flex gap-4 pt-4">
            <Button
              type="button"
              variant="outline"
              className="bg-zinc-800 border-gray-500 text-gray-300 hover:bg-gray-700"
              onClick={() => navigate("/create-case")}
            >
              Back
            </Button>

            <Button type="submit" className="bg-cyan-600 hover:bg-cyan-700 text-white">
              Done
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

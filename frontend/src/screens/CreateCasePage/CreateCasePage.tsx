import React, { useState } from "react";
import { Button } from "../../components/ui/button";
import { Input } from "../../components//ui/input";
import { Textarea } from "../../components/ui/TextArea";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "../../components/ui/select";
import { TagIcon, ShieldAlert } from "lucide-react";

export function CreateCaseForm(): JSX.Element {
  const [form, setForm] = useState({
    creator: "",
    team: "",
    priority: "",
    attackType: "",
    description: "",
  });

interface CreateCaseFormState {
    creator: string;
    team: string;
    priority: string;
    attackType: string;
    description: string;
}

type CreateCaseFormField = keyof CreateCaseFormState;

const handleChange =
    (field: CreateCaseFormField) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        setForm({ ...form, [field]: e.target.value });
    };

interface CreateCaseFormSubmitEvent extends React.FormEvent<HTMLFormElement> {}

const handleSubmit = (e: CreateCaseFormSubmitEvent) => {
    e.preventDefault();
    // TODO: Submit logic
    console.log("Creating case:", form);
};

  return (
 <div className="min-h-screen bg-zinc-900 text-white flex items-center justify-center p-6">
    <div className="max-w-3xl mx-auto mt-10 bg-zinc-900 text-white p-6 rounded-2xl shadow-xl border border-zinc-700 font-mono">
      <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">
        <ShieldAlert size={28} /> Create New Case
      </h1>
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label className="block mb-1 text-sm">Name of Person Creating the Case</label>
          <Input
            className="bg-zinc-800 border-zinc-600 text-white"
            placeholder="e.g. Alice Johnson"
            value={form.creator}
            onChange={handleChange("creator")}
            required
          />
        </div>

        <div>
          <label className="block mb-1 text-sm">Team Name</label>
          <Input
            className="bg-zinc-800 border-zinc-600 text-white"
            placeholder="e.g. AEGIS Forensics"
            value={form.team}
            onChange={handleChange("team")}
            required
          />
        </div>

        <div>
          <label className="block mb-1 text-sm">Case Priority</label>
          <Select
            onValueChange={(value: string) => setForm({ ...form, priority: value })}
          >
            <SelectTrigger className="bg-zinc-800 border-zinc-600 text-white">
              <SelectValue placeholder="Select priority" />
            </SelectTrigger>
            <SelectContent className="bg-zinc-800 text-white">
              <SelectItem value="low">Low</SelectItem>
              <SelectItem value="mid">Mid</SelectItem>
              <SelectItem value="high">High</SelectItem>
              <SelectItem value="critical">Critical</SelectItem>
              <SelectItem value="time-sensitive">Time Sensitive</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div>
          <label className="block mb-1 text-sm">Type of Attack</label>
          <Input
            className="bg-zinc-800 border-zinc-600 text-white"
            placeholder="e.g. Ransomware, Malware, Phishing"
            value={form.attackType}
            onChange={handleChange("attackType")}
          />
        </div>

        <div>
          <label className="block mb-1 text-sm">Short Description</label>
          <Textarea
            className="bg-zinc-800 border-zinc-600 text-white"
            placeholder="Brief summary of the incident..."
            value={form.description}
            onChange={handleChange("description")}
            rows={4}
          />
        </div>

        <div className="flex flex-wrap gap-4 pt-4">
          <Button
            type="button"
            variant="outline"
            className="bg-zinc-800 border-cyan-500 text-cyan-400 hover:bg-cyan-800"
            onClick={() => window.location.href = "/assign-case-members"}>
            Assign Case Members
          </Button>

          <Button
            type="button"
            variant="outline"
            className="bg-zinc-800 border-purple-500 text-purple-400 hover:bg-purple-800"
            onClick={() => window.location.href = "/upload-evidence"}>
            Upload Evidence
          </Button>

          <Button type="submit" className="bg-cyan-600 hover:bg-cyan-700 text-white">
            Create Case
          </Button>
        </div>
      </form>
    </div>
    </div>
  );
}

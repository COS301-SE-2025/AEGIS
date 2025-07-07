import React, { useState, useEffect } from "react";
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
import { useNavigate } from 'react-router-dom';
import axios from "axios";

export function CreateCaseForm(): JSX.Element {
  const navigate = useNavigate();

  const [form, setForm] = useState({
    creator: "",
    team: "",
    priority: "",
    attackType: "",
    description: "",
    creatorId: "", // from session storage
  });

  type CreateCaseFormField = keyof typeof form;

  useEffect(() => {
    const savedFormData = localStorage.getItem("tempCreateCaseForm");
    if (savedFormData) {
      try {
        setForm(JSON.parse(savedFormData));
      } catch (error) {
        console.error("Error loading saved form data:", error);
      }
    }

    // Load user from sessionStorage
    const userStr = sessionStorage.getItem("user");
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        setForm((prev) => ({ ...prev, creatorId: user.id }));
      } catch (err) {
        console.error("Failed to parse user from session storage:", err);
      }
    }
  }, []);

  useEffect(() => {
    localStorage.setItem("tempCreateCaseForm", JSON.stringify(form));
  }, [form]);

  const clearSavedFormData = () => {
    localStorage.removeItem("tempCreateCaseForm");
  };

  const handleChange =
    (field: CreateCaseFormField) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setForm({ ...form, [field]: e.target.value });
    };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!form.creatorId) {
      alert("Cannot create case: user ID not found in session. Please log in again.");
      return;
    }

    const payload = {
      title: form.attackType,
      description: form.description,
      status: "open",
      priority: form.priority || "low",
      investigation_stage: "analysis",
      created_by: form.creatorId, // direct from session storage
      team_name: form.team,
    };

    console.log("Submitting payload:", payload);

    try {
      const response = await axios.post("http://localhost:8080/api/v1/cases", payload, {
        headers: { "Content-Type": "application/json" },
      });

      if (response.status === 201) {
        alert("Case created successfully!");
        clearSavedFormData();
        const data = response.data as { case: { ID: string } };
        navigate(`/case/${data.case.ID}/next-steps`);
        // Save the case ID to localStorage so upload can find it
      localStorage.setItem("currentCase", JSON.stringify(data.case));

      }
    } catch (error: any) {
      console.error("Error creating case:", error.response?.data || error);
      alert("Failed to create case. Please check console for details.");
    }
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex items-center justify-center p-6">
      <div className="max-w-3xl mx-auto mt-10 bg-card text-foreground p-6 rounded-2xl shadow-xl border border-border font-mono w-full">
        <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">
          <ShieldAlert size={28} /> Create New Case
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5">
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

          <div>
            <label className="block mb-1 text-sm">Type of Attack</label>
            <Input
              className="bg-muted border-border text-foreground placeholder-muted-foreground"
              placeholder="e.g. Ransomware, Malware, Phishing"
              value={form.attackType}
              onChange={handleChange("attackType")}
              required
            />
          </div>

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

          <Button
            type="submit"
            className="bg-cyan-600 hover:bg-cyan-700 text-white"
          >
            Create Case
          </Button>
        </form>
      </div>
    </div>
  );
}

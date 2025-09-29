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
import { jwtDecode } from "jwt-decode";


export function CreateCaseForm(): JSX.Element {
  const navigate = useNavigate();
  const [teams, setTeams] = useState<{ id: string; name: string }[]>([]);


  const [form, setForm] = useState({
    creator: "",
    team: "",
    priority: "",
    attackType: "",
    description: "",
    creatorId: "", // from session storage
    tenantId: "", // from session storage
    
  });

    type CreateCaseFormField = keyof typeof form;
    type DecodedToken = {
    user_id: string;
    tenant_id: string;
    team_id: string;
    team_name: string;
    full_name: string;
    role: string;
    exp: number;
    email: string;
  };

  useEffect(() => {
  const savedFormData = localStorage.getItem("tempCreateCaseForm");
  if (savedFormData) {
    try {
      setForm(JSON.parse(savedFormData));
    } catch (error) {
      console.error("Error loading saved form data:", error);
    }
  }

  const userStr = sessionStorage.getItem("user");
  const token = sessionStorage.getItem("authToken");

  if (userStr) {
    try {
      const user = JSON.parse(userStr);
      setForm(prev => ({ ...prev, creatorId: user.id }));
    } catch (err) {
      console.error("Failed to parse user:", err);
    }
  }

  if (token) {
    try {
      const decoded = jwtDecode<DecodedToken>(token);
      setForm(prev => ({
      ...prev,
      tenantId: decoded.tenant_id,
      team: decoded.team_name, 
    }));
    } catch (err) {
      console.error("Failed to decode token:", err);
    }
  }
}, []);

  useEffect(() => {
  const token = sessionStorage.getItem("authToken");
  if (!token) return;

  const decoded = jwtDecode<DecodedToken>(token);
  setForm(prev => ({ ...prev, tenantId: decoded.tenant_id }));

  // Use team_id here, NOT tenant_id
axios.get<{ id: string; name: string }>(`https://localhost/api/v1/teams/${decoded.team_id}`)
  .then((res) => {
    setTeams([res.data]);  // wrap in array so you can map safely
  })
  .catch(err => {
    console.error("Failed to load teams:", err);
  });

}, []);


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
      //alert("Cannot create case: user ID not found in session. Please log in again.");
      return;
    }

const user = JSON.parse(sessionStorage.getItem("user") || "{}");
console.log("User from session:", user);

const payload = {
  title: form.attackType,
  description: form.description,
  status: "open",
  priority: form.priority || "low",
  investigation_stage: "Triage",
  created_by: user.id,
  team_name: form.team,
  tenant_id: form.tenantId, // 🛠️ ADD THIS LINE
};

    console.log("Submitting payload:", payload);
   const token = sessionStorage.getItem("authToken");
   console.log("Token:", token);
   if (token) {
  const base64Payload = token.split('.')[1];
  const decoded = JSON.parse(atob(base64Payload));
  console.log("Decoded token payload:", decoded);
}

    try {
      const response = await axios.post(
  "https://localhost/api/v1/cases",
  payload,
  {
    headers: {
      Authorization: `Bearer ${token}`
    }
  }
);

      if (response.status === 201) {
        //alert("Case created successfully!");
        clearSavedFormData();
        const data = response.data as { case: { id: string } };
        console.log("Case ID to navigate:", data.case.id);
        navigate(`/case/${data.case.id}/next-steps`);
        // Save the case ID to localStorage so upload can find it
      localStorage.setItem("currentCase", JSON.stringify(data.case));

      

      }
    } catch (error: any) {
      console.error("Error creating case:", error.response?.data || error);
      //alert("Failed to create case. Please check console for details.");
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
        <label className="block mb-1 text-sm">Team Name</label>
        <select
          className="w-full p-2 border rounded bg-muted text-foreground"
          value={form.team}
          onChange={(e) => setForm({ ...form, team: e.target.value })}
          required
        >
          <option value="">Select a team</option>
          {teams.map((team) => (
            <option key={team.id} value={team.name}>
              {team.name}
            </option>
          ))}
        </select>
      </div>

          <div>
            <label className="block mb-1 text-sm">Case Priority</label>
            <Select onValueChange={(value: string) => setForm({ ...form, priority: value })}>
              <SelectTrigger className="bg-muted border-border text-foreground">
                <SelectValue placeholder="Select priority" />
              </SelectTrigger>
              <SelectContent className="bg-zinc-800 text-popover-foreground">
                <SelectItem value="low">Low</SelectItem>
                <SelectItem value="medium">Mid</SelectItem>
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
          <div className="flex justify-between items-center pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => navigate(-1)}
              className="border-muted-foreground text-muted-foreground hover:bg-muted"
            >
              Cancel
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

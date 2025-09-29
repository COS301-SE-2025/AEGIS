import { useState, useEffect } from "react";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import { Shield, Cpu, Calendar, Hash } from "lucide-react";
import axios from "axios";
const BASE_URL = "https://localhost/api/v1";

export const ChainOfCustody = () => {
  const { caseId, entryId } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  // Get evidenceId from location state (passed from EvidenceViewer)
  const evidenceId = searchParams.get("evidenceId") || "";

  const [formData, setFormData] = useState({
    evidenceId: evidenceId,
    chainOfCustody: [],
    custodian: "",
    acquisitionDate: "",
    acquisitionTool: "",
    hash: "",
    systemInfo: {
      osVersion: "",
      architecture: "",
      computerName: "",
      domain: "",
      lastBoot: "",
    },
    forensic: {
      method: "",
      examiner: "",
      location: "",
      legalStatus: "Admissible",
      notes: "",
    }
  });
  const [showConfirmation, setShowConfirmation] = useState(false);

  // If editing, fetch existing entry
  useEffect(() => {
    if (entryId && caseId) {
      axios.get(`${BASE_URL}/cases/${caseId}/chain_of_custody/${entryId}`)
        .then(res => setFormData(mapBackendToForm(res.data)))
        .catch(err => console.error("Failed to load entry", err));
    }
  }, [entryId, caseId]);

  // Helper to map backend response to form shape
  function mapBackendToForm(data: any) {
    return {
      evidenceId: data.evidenceId || evidenceId,
      chainOfCustody: Array.isArray(data.chainOfCustody) ? data.chainOfCustody : [],
      custodian: data.custodian || "",
      acquisitionDate: data.acquisitionDate || "",
      acquisitionTool: data.acquisitionTool || "",
      hash: data.hash || "",
      systemInfo: {
        osVersion: data.systemInfo?.osVersion || "",
        architecture: data.systemInfo?.architecture || "",
        computerName: data.systemInfo?.computerName || "",
        domain: data.systemInfo?.domain || "",
        lastBoot: data.systemInfo?.lastBoot || "",
      },
      forensic: {
        method: data.forensic?.method || "",
        examiner: data.forensic?.examiner || "",
        location: data.forensic?.location || "",
        legalStatus: data.forensic?.legalStatus || "Admissible",
        notes: data.forensic?.notes || "",
      }
    };
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    if (name in formData.systemInfo) {
      setFormData(prev => ({ ...prev, systemInfo: { ...prev.systemInfo, [name]: value } }));
    } else if (name in formData.forensic) {
      setFormData(prev => ({ ...prev, forensic: { ...prev.forensic, [name]: value } }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const token = sessionStorage.getItem("authToken");
      const method = entryId ? "put" : "post";
      const url = entryId
        ? `${BASE_URL}/cases/${caseId}/chain_of_custody/${entryId}`
        : `${BASE_URL}/cases/${caseId}/chain_of_custody`;
      // Convert acquisitionDate to RFC3339 format if not empty
      let acquisitionDate = formData.acquisitionDate;
      if (acquisitionDate) {
        // If input is "2025-08-17T14:30", convert to "2025-08-17T14:30:00Z"
        acquisitionDate = new Date(acquisitionDate).toISOString();
      }
      const payload = {
        chainOfCustody: formData.chainOfCustody,
        custodian: formData.custodian,
        acquisition_date: acquisitionDate, // snake_case for backend
        acquisition_tool: formData.acquisitionTool,
        hash: formData.hash,
        system_info: formData.systemInfo, // match backend field name
        forensic_info: formData.forensic, // match backend field name
        case_id: caseId,
        evidence_id: evidenceId,
      };
      console.log("Submitting chain of custody payload:", payload);
      await axios({
        method,
        url,
        data: payload,
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {})
        },
      });
      setShowConfirmation(true);
      setTimeout(() => {
        navigate(`/evidence-viewer/${caseId}?tab=chain&evidenceId=${evidenceId}`);
      }, 1500); // Show confirmation for 1.5 seconds
    } catch (err) {
      console.error("Failed to save entry", err);
    }
  };

  return (
    <div className="min-h-screen p-8 bg-background text-foreground">
      {showConfirmation && (
        <div className="fixed top-4 left-1/2 transform -translate-x-1/2 bg-success text-success-foreground px-6 py-3 rounded shadow-lg z-50">
          Entry saved successfully!
        </div>
      )}
      <form onSubmit={handleSubmit} className="bg-card p-6 rounded-lg space-y-6 max-w-3xl mx-auto">
        <h2 className="text-xl font-semibold flex items-center gap-2">
          <Shield className="w-5 h-5 text-primary" />
          {entryId ? "Update Custody Information" : "New Custody Information"}
        </h2>

        {/* Custodian */}
        <div>
          <label className="text-sm text-muted-foreground">Custodian</label>
          <input
            name="custodian"
            placeholder="Name of custodian"
            className="w-full mt-1 p-2 bg-muted border border-border rounded text-sm"
            value={formData.custodian}
            onChange={handleChange}
          />
        </div>

        {/* Acquisition */}
        <div>
          <h3 className="font-semibold flex items-center gap-2">
            <Calendar className="w-5 h-5 text-secondary" /> Acquisition Details
          </h3>
          <input type="datetime-local" name="acquisitionDate" className="w-full mt-2 p-2 bg-muted border rounded text-sm" value={formData.acquisitionDate} onChange={handleChange} />
          <input name="acquisitionTool" placeholder="Tool Used" className="w-full mt-2 p-2 bg-muted border rounded text-sm" value={formData.acquisitionTool} onChange={handleChange} />
        </div>

        {/* System Info */}
        <div>
          <h3 className="font-semibold flex items-center gap-2">
            <Cpu className="w-5 h-5 text-accent" /> System Information
          </h3>
          <div className="grid grid-cols-2 gap-4 mt-2">
            <input name="osVersion" placeholder="OS Version" value={formData.systemInfo.osVersion} onChange={handleChange} className="p-2 bg-muted border rounded text-sm" />
            <input name="architecture" placeholder="Architecture" value={formData.systemInfo.architecture} onChange={handleChange} className="p-2 bg-muted border rounded text-sm" />
            <input name="computerName" placeholder="Computer Name" value={formData.systemInfo.computerName} onChange={handleChange} className="p-2 bg-muted border rounded text-sm col-span-2" />
            <input name="domain" placeholder="Domain" value={formData.systemInfo.domain} onChange={handleChange} className="p-2 bg-muted border rounded text-sm col-span-2" />
            <label className="text-xs text-muted-foreground col-span-2">Last Boot Time (mm/dd/yyyy)</label>
            <input
              type="datetime-local"
              name="lastBoot"
              value={formData.systemInfo.lastBoot || ""}
              onChange={handleChange}
              className="p-2 bg-muted border rounded text-sm col-span-2"
            />
          </div>
        </div>

        {/* Forensic Metadata */}
        <div>
          <h3 className="font-semibold flex items-center gap-2">
            <Hash className="w-5 h-5 text-warning" /> Forensic Metadata
          </h3>
          <input name="method" placeholder="Acquisition Method" value={formData.forensic.method} onChange={handleChange} className="w-full mt-2 p-2 bg-muted border rounded text-sm" />
          <input name="examiner" placeholder="Examiner Name" value={formData.forensic.examiner} onChange={handleChange} className="w-full mt-2 p-2 bg-muted border rounded text-sm" />
          <input name="location" placeholder="Evidence Location" value={formData.forensic.location} onChange={handleChange} className="w-full mt-2 p-2 bg-muted border rounded text-sm" />
          <textarea name="notes" placeholder="Notes / Comments" value={formData.forensic.notes} onChange={handleChange} className="w-full mt-2 p-2 bg-muted border rounded text-sm" />
        </div>

        {/* Save */}
        <button type="submit" className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90">
          {entryId ? "Update" : "Save"}
        </button>
      </form>
    </div>
  );
};

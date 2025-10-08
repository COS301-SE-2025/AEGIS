
import { useState, useEffect } from "react";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import { Shield, Cpu, Calendar, Hash, AlertCircle } from "lucide-react";
import axios from "axios";

const BASE_URL = "https://localhost/api/v1";

// Valid legal statuses for chain of custody
const LEGAL_STATUSES = [
  "Admissible",
  "Pending Review",
  "Inadmissible",
  "Challenged",
  "Sealed",
  "Destroyed"
] as const;

type LegalStatus = typeof LEGAL_STATUSES[number];

interface ValidationErrors {
  custodian?: string;
  acquisitionDate?: string;
  acquisitionTool?: string;
 // hash?: string;
  lastBoot?: string;
  examiner?: string;
  method?: string;
  location?: string;
}

export const ChainOfCustody = () => {
  const { caseId, entryId } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const evidenceId = searchParams.get("evidenceId") || "";

  const [formData, setFormData] = useState({
    evidenceId: evidenceId,
    chainOfCustody: [],
    custodian: "",
    acquisitionDate: "",
    acquisitionTool: "",
    //hash: "",
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
      legalStatus: "Admissible" as LegalStatus,
      notes: "",
    }
  });

  const [errors, setErrors] = useState<ValidationErrors>({});
  const [touched, setTouched] = useState<Set<string>>(new Set());
  const [showConfirmation, setShowConfirmation] = useState(false);

  useEffect(() => {
    if (entryId && caseId) {
      axios.get(`${BASE_URL}/cases/${caseId}/chain_of_custody/${entryId}`)
        .then(res => setFormData(mapBackendToForm(res.data)))
        .catch(err => console.error("Failed to load entry", err));
    }
  }, [entryId, caseId]);

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

  // Validation functions
  const validateDate = (dateStr: string, fieldName: string): string | undefined => {
    if (!dateStr) return `${fieldName} is required`;
    
    const date = new Date(dateStr);
    const now = new Date();
    
    if (isNaN(date.getTime())) {
      return `Invalid ${fieldName.toLowerCase()} format`;
    }
    
    if (date > now) {
      return `${fieldName} cannot be in the future`;
    }
    
    // Check if date is unreasonably old (before 1980)
    const minDate = new Date('1980-01-01');
    if (date < minDate) {
      return `${fieldName} seems unreasonably old`;
    }
    
    return undefined;
  };

  const validateHash = (hash: string): string | undefined => {
    if (!hash) return "Hash is required for evidence integrity";
    
    const trimmedHash = hash.trim();
    
    // MD5: 32 hex chars
    // SHA1: 40 hex chars
    // SHA256: 64 hex chars
    // SHA512: 128 hex chars
    const validLengths = [32, 40, 64, 128];
    const hexPattern = /^[a-fA-F0-9]+$/;
    
    if (!hexPattern.test(trimmedHash)) {
      return "Hash must contain only hexadecimal characters (0-9, a-f)";
    }
    
    if (!validLengths.includes(trimmedHash.length)) {
      return `Invalid hash length. Expected MD5 (32), SHA1 (40), SHA256 (64), or SHA512 (128) characters`;
    }
    
    return undefined;
  };

  const validateRequired = (value: string, fieldName: string): string | undefined => {
    if (!value || value.trim().length === 0) {
      return `${fieldName} is required`;
    }
    return undefined;
  };

  // Validate only touched fields
  const validateTouchedFields = (): ValidationErrors => {
    const newErrors: ValidationErrors = {};

    if (touched.has('custodian')) {
      const custodianError = validateRequired(formData.custodian, "Custodian");
      if (custodianError) newErrors.custodian = custodianError;
    }

    if (touched.has('acquisitionDate')) {
      const acquisitionDateError = validateDate(formData.acquisitionDate, "Acquisition date");
      if (acquisitionDateError) newErrors.acquisitionDate = acquisitionDateError;
    }

    if (touched.has('acquisitionTool')) {
      const acquisitionToolError = validateRequired(formData.acquisitionTool, "Acquisition tool");
      if (acquisitionToolError) newErrors.acquisitionTool = acquisitionToolError;
    }

    // if (touched.has('hash')) {
    //   const hashError = validateHash(formData.hash);
    //   if (hashError) newErrors.hash = hashError;
    // }

    if (touched.has('examiner')) {
      const examinerError = validateRequired(formData.forensic.examiner, "Examiner name");
      if (examinerError) newErrors.examiner = examinerError;
    }

    if (touched.has('method')) {
      const methodError = validateRequired(formData.forensic.method, "Acquisition method");
      if (methodError) newErrors.method = methodError;
    }

    if (touched.has('location')) {
      const locationError = validateRequired(formData.forensic.location, "Evidence location");
      if (locationError) newErrors.location = locationError;
    }

    if (touched.has('lastBoot') && formData.systemInfo.lastBoot) {
      const lastBootError = validateDate(formData.systemInfo.lastBoot, "Last boot time");
      if (lastBootError) newErrors.lastBoot = lastBootError;
    }

    return newErrors;
  };

  const getCurrentLocalDateTime = () => {
    const now = new Date();
    const offset = now.getTimezoneOffset();
    const localDate = new Date(now.getTime() - (offset * 60 * 1000));
    return localDate.toISOString().slice(0, 16);
  };

  // Validate all required fields (for submit)
  const validateAllFields = (): ValidationErrors => {
    const newErrors: ValidationErrors = {};

    const custodianError = validateRequired(formData.custodian, "Custodian");
    if (custodianError) newErrors.custodian = custodianError;

    const acquisitionDateError = validateDate(formData.acquisitionDate, "Acquisition date");
    if (acquisitionDateError) newErrors.acquisitionDate = acquisitionDateError;

    const acquisitionToolError = validateRequired(formData.acquisitionTool, "Acquisition tool");
    if (acquisitionToolError) newErrors.acquisitionTool = acquisitionToolError;

    // const hashError = validateHash(formData.hash);
    // if (hashError) newErrors.hash = hashError;

    const examinerError = validateRequired(formData.forensic.examiner, "Examiner name");
    if (examinerError) newErrors.examiner = examinerError;

    const methodError = validateRequired(formData.forensic.method, "Acquisition method");
    if (methodError) newErrors.method = methodError;

    const locationError = validateRequired(formData.forensic.location, "Evidence location");
    if (locationError) newErrors.location = locationError;

    if (formData.systemInfo.lastBoot) {
      const lastBootError = validateDate(formData.systemInfo.lastBoot, "Last boot time");
      if (lastBootError) newErrors.lastBoot = lastBootError;
    }

    return newErrors;
  };

  const handleBlur = (fieldName: string) => {
    setTouched(prev => new Set(prev).add(fieldName));
  };

  // Update validation whenever formData or touched changes
  useEffect(() => {
    if (touched.size > 0) {
      setErrors(validateTouchedFields());
    }
  }, [formData, touched]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
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
    
    // Mark all required fields as touched
    const allFields = new Set([
      'custodian', 'acquisitionDate', 'acquisitionTool',
      'examiner', 'method', 'location'
    ]);
    
    // Add lastBoot only if it has a value
    if (formData.systemInfo.lastBoot) {
      allFields.add('lastBoot');
    }
    
    setTouched(allFields);
    
    // Validate entire form
    const validationErrors = validateAllFields();
    setErrors(validationErrors);
    
    // Check if there are any errors
    if (Object.keys(validationErrors).length > 0) {
      // Scroll to first error
      const firstErrorField = Object.keys(validationErrors)[0];
      const element = document.querySelector(`[name="${firstErrorField}"]`);
      element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
      return;
    }

    try {
      const token = sessionStorage.getItem("authToken");
      const method = entryId ? "put" : "post";
      const url = entryId
        ? `${BASE_URL}/cases/${caseId}/chain_of_custody/${entryId}`
        : `${BASE_URL}/cases/${caseId}/chain_of_custody`;
      
      let acquisitionDate = formData.acquisitionDate;
      if (acquisitionDate) {
        acquisitionDate = new Date(acquisitionDate).toISOString();
      }

      let lastBoot = formData.systemInfo.lastBoot;
      if (lastBoot) {
        lastBoot = new Date(lastBoot).toISOString();
      }
      
      const payload = {
        chainOfCustody: formData.chainOfCustody,
        custodian: formData.custodian.trim(),
        acquisition_date: acquisitionDate,
        acquisition_tool: formData.acquisitionTool.trim(),
        //hash: formData.hash.trim().toLowerCase(),
        system_info: {
          ...formData.systemInfo,
          lastBoot: lastBoot || formData.systemInfo.lastBoot,
        },
        forensic_info: formData.forensic,
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
      }, 1500);
    } catch (err) {
      console.error("Failed to save entry", err);
      //alert("Failed to save entry. Please check the console for details.");
    }
  };

  const hasErrors = Object.keys(errors).length > 0;
  const showError = (fieldName: string) => touched.has(fieldName) && errors[fieldName as keyof ValidationErrors];

  return (
    <div className="min-h-screen p-8 bg-gray-900 text-gray-100">
      {showConfirmation && (
        <div className="fixed top-4 left-1/2 transform -translate-x-1/2 bg-green-600 text-white px-6 py-3 rounded shadow-lg z-50">
          Entry saved successfully!
        </div>
      )}
      
      <form onSubmit={handleSubmit} className="bg-gray-800 p-6 rounded-lg space-y-6 max-w-3xl mx-auto border border-gray-700">
        <h2 className="text-xl font-semibold flex items-center gap-2">
          <Shield className="w-5 h-5 text-blue-400" />
          {entryId ? "Update Custody Information" : "New Custody Information"}
        </h2>

        {/* Custodian */}
        <div>
          <label className="text-sm text-gray-400 flex items-center gap-1">
            Custodian <span className="text-red-400">*</span>
          </label>
          <input
            name="custodian"
            placeholder="Name of custodian responsible for evidence"
            className={`w-full mt-1 p-2 bg-gray-700 border rounded text-sm ${
              showError('custodian') ? 'border-red-500' : 'border-gray-600'
            }`}
            value={formData.custodian}
            onChange={handleChange}
            onBlur={() => handleBlur('custodian')}
          />
          {showError('custodian') && (
            <p className="text-red-400 text-xs mt-1 flex items-center gap-1">
              <AlertCircle size={12} /> {errors.custodian}
            </p>
          )}
        </div>

        {/* Acquisition */}
        <div>
          <h3 className="font-semibold flex items-center gap-2 mb-3">
            <Calendar className="w-5 h-5 text-cyan-400" /> Acquisition Details
          </h3>
          
          <div className="space-y-3">
            <div>
              <label className="text-sm text-gray-400 flex items-center gap-1">
                Acquisition Date & Time <span className="text-red-400">*</span>
              </label>
              <input
                type="datetime-local"
                name="acquisitionDate"
                max={getCurrentLocalDateTime()}
                className={`w-full mt-1 p-2 bg-gray-700 border rounded text-sm ${
                  showError('acquisitionDate') ? 'border-red-500' : 'border-gray-600'
                }`}
                value={formData.acquisitionDate}
                onChange={handleChange}
                onBlur={() => handleBlur('acquisitionDate')}
              />
              {showError('acquisitionDate') && (
                <p className="text-red-400 text-xs mt-1 flex items-center gap-1">
                  <AlertCircle size={12} /> {errors.acquisitionDate}
                </p>
              )}
            </div>

            <div>
              <label className="text-sm text-gray-400 flex items-center gap-1">
                Acquisition Tool <span className="text-red-400">*</span>
              </label>
              <input
                name="acquisitionTool"
                placeholder="e.g., FTK Imager, dd, EnCase"
                className={`w-full mt-1 p-2 bg-gray-700 border rounded text-sm ${
                  showError('acquisitionTool') ? 'border-red-500' : 'border-gray-600'
                }`}
                value={formData.acquisitionTool}
                onChange={handleChange}
                onBlur={() => handleBlur('acquisitionTool')}
              />
              {showError('acquisitionTool') && (
                <p className="text-red-400 text-xs mt-1 flex items-center gap-1">
                  <AlertCircle size={12} /> {errors.acquisitionTool}
                </p>
              )}
            </div>
          </div>
        </div>

        {/* System Info */}
        <div>
          <h3 className="font-semibold flex items-center gap-2 mb-3">
            <Cpu className="w-5 h-5 text-purple-400" /> System Information
          </h3>
          <div className="grid grid-cols-2 gap-4">
            <input
              name="osVersion"
              placeholder="OS Version (e.g., Windows 10 Pro)"
              value={formData.systemInfo.osVersion}
              onChange={handleChange}
              className="p-2 bg-gray-700 border border-gray-600 rounded text-sm"
            />
            <input
              name="architecture"
              placeholder="Architecture (e.g., x64)"
              value={formData.systemInfo.architecture}
              onChange={handleChange}
              className="p-2 bg-gray-700 border border-gray-600 rounded text-sm"
            />
            <input
              name="computerName"
              placeholder="Computer Name"
              value={formData.systemInfo.computerName}
              onChange={handleChange}
              className="p-2 bg-gray-700 border border-gray-600 rounded text-sm col-span-2"
            />
            <input
              name="domain"
              placeholder="Domain/Workgroup"
              value={formData.systemInfo.domain}
              onChange={handleChange}
              className="p-2 bg-gray-700 border border-gray-600 rounded text-sm col-span-2"
            />
            
            <div className="col-span-2">
              <label className="text-xs text-gray-400">Last Boot Time (Optional)</label>
              <input
                type="datetime-local"
                name="lastBoot"
                max={getCurrentLocalDateTime()}
                value={formData.systemInfo.lastBoot || ""}
                onChange={handleChange}
                onBlur={() => handleBlur('lastBoot')}
                className={`w-full mt-1 p-2 bg-gray-700 border rounded text-sm ${
                  showError('lastBoot') ? 'border-red-500' : 'border-gray-600'
                }`}
              />
              {showError('lastBoot') && (
                <p className="text-red-400 text-xs mt-1 flex items-center gap-1">
                  <AlertCircle size={12} /> {errors.lastBoot}
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Forensic Metadata */}
        <div>
          <h3 className="font-semibold flex items-center gap-2 mb-3">
            <Hash className="w-5 h-5 text-amber-400" /> Forensic Metadata
          </h3>
          
          <div className="space-y-3">
            <div>
              <label className="text-sm text-gray-400 flex items-center gap-1">
                Acquisition Method <span className="text-red-400">*</span>
              </label>
              <input
                name="method"
                placeholder="e.g., Live acquisition, Dead box imaging, Memory dump"
                value={formData.forensic.method}
                onChange={handleChange}
                onBlur={() => handleBlur('method')}
                className={`w-full mt-1 p-2 bg-gray-700 border rounded text-sm ${
                  showError('method') ? 'border-red-500' : 'border-gray-600'
                }`}
              />
              {showError('method') && (
                <p className="text-red-400 text-xs mt-1 flex items-center gap-1">
                  <AlertCircle size={12} /> {errors.method}
                </p>
              )}
            </div>

            <div>
              <label className="text-sm text-gray-400 flex items-center gap-1">
                Examiner Name <span className="text-red-400">*</span>
              </label>
              <input
                name="examiner"
                placeholder="Name of forensic examiner"
                value={formData.forensic.examiner}
                onChange={handleChange}
                onBlur={() => handleBlur('examiner')}
                className={`w-full mt-1 p-2 bg-gray-700 border rounded text-sm ${
                  showError('examiner') ? 'border-red-500' : 'border-gray-600'
                }`}
              />
              {showError('examiner') && (
                <p className="text-red-400 text-xs mt-1 flex items-center gap-1">
                  <AlertCircle size={12} /> {errors.examiner}
                </p>
              )}
            </div>

            <div>
              <label className="text-sm text-gray-400 flex items-center gap-1">
                Evidence Location <span className="text-red-400">*</span>
              </label>
              <input
                name="location"
                placeholder="Physical/digital storage location"
                value={formData.forensic.location}
                onChange={handleChange}
                onBlur={() => handleBlur('location')}
                className={`w-full mt-1 p-2 bg-gray-700 border rounded text-sm ${
                  showError('location') ? 'border-red-500' : 'border-gray-600'
                }`}
              />
              {showError('location') && (
                <p className="text-red-400 text-xs mt-1 flex items-center gap-1">
                  <AlertCircle size={12} /> {errors.location}
                </p>
              )}
            </div>

            <div>
              <label className="text-sm text-gray-400 flex items-center gap-1">
                Legal Status <span className="text-red-400">*</span>
              </label>
              <select
                name="legalStatus"
                value={formData.forensic.legalStatus}
                onChange={handleChange}
                className="w-full mt-1 p-2 bg-gray-700 border border-gray-600 rounded text-sm"
              >
                {LEGAL_STATUSES.map(status => (
                  <option key={status} value={status}>{status}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="text-sm text-gray-400">Notes / Comments</label>
              <textarea
                name="notes"
                placeholder="Additional notes about the evidence handling..."
                value={formData.forensic.notes}
                onChange={handleChange}
                className="w-full mt-1 p-2 bg-gray-700 border border-gray-600 rounded text-sm"
                rows={4}
              />
            </div>
          </div>
        </div>

        {/* Save Button */}
        <div className="flex items-center justify-between pt-4 border-t border-gray-700">
          <p className="text-xs text-gray-400">
            <span className="text-red-400">*</span> Required fields
          </p>
          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => navigate(`/evidence-viewer/${caseId}?tab=chain&evidenceId=${evidenceId}`)}
              className="px-4 py-2 bg-gray-700 text-gray-200 rounded hover:bg-gray-600 transition"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={hasErrors && touched.size > 0}
              className={`px-4 py-2 rounded transition flex items-center gap-2 ${
                hasErrors && touched.size > 0
                  ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                  : 'bg-blue-600 text-white hover:bg-blue-700'
              }`}
            >
              <Shield size={16} />
              {entryId ? "Update" : "Save"} Custody Record

            </button>
          </div>
        </div>
      </form>
    </div>
  );
};
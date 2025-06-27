import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../../components/ui/button";
import { ShieldPlus, UploadCloud } from "lucide-react";

export function UploadEvidenceForm(): JSX.Element {
  const [files, setFiles] = useState<File[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const navigate = useNavigate();

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const droppedFiles = Array.from(e.dataTransfer.files);
    setFiles((prev) => [...prev, ...droppedFiles]);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = e.target.files ? Array.from(e.target.files) : [];
    setFiles((prev) => [...prev, ...selectedFiles]);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    // Get the current case from localStorage
    const cases = JSON.parse(localStorage.getItem("cases") || "[]");
    const currentCase = cases[cases.length - 1]; // last created

    const allEvidence = JSON.parse(localStorage.getItem("evidenceFiles") || "[]");

    const newEvidence = files.map((file) => ({
      id: Date.now().toString() + Math.random().toString(36).substring(2, 6),
      name: file.name,
      type: "unknown", // optional: allow user to choose type
      size: `${(file.size / (1024 * 1024)).toFixed(1)} MB`,
      hash: "SHA256: fake-hash-placeholder", // optionally generate real hash later
      created: new Date().toISOString(),
      description: "Uploaded evidence", // allow custom description later
      status: "pending",
      chainOfCustody: [], // add logged-in user if needed
      acquisitionDate: new Date().toISOString(),
      acquisitionTool: "Manual Upload",
      integrityCheck: "pending",
      threadCount: 0,
      priority: "medium", // or let user choose
      caseId: String(currentCase.id)
    }));

    localStorage.setItem("evidenceFiles", JSON.stringify([...allEvidence, ...newEvidence]));
    // 1. Get current user info
      const user = JSON.parse(sessionStorage.getItem("user") || "{}");

      // 2. Create an audit log entry
      const evidenceUploadLog = {
        timestamp: new Date().toISOString(),
        user: user.email || "Unknown user",
        userId: user.id || "unknown",
        action: `Uploaded ${files.length} evidence file(s) to case ${currentCase.id}`
      };

      // 3. Append to caseActivities log
      const previousLogs = JSON.parse(localStorage.getItem("caseActivities") || "[]");
      const updatedLogs = [evidenceUploadLog, ...previousLogs];
      localStorage.setItem("caseActivities", JSON.stringify(updatedLogs));

    alert("Evidence uploaded successfully!");
    // navigate("/dashboard");


    // alert("Evidence Uploaded!");

    // try {
    //   const response = await axios.post("http://localhost:8080/api/v1", formData, {
    //     headers: { "Content-Type": "multipart/form-data" },
    //   });

    //   console.log("Uploading files:", files);
    //   console.log("Upload success:", response.data);
    // } catch (error) {
    //   console.error("Upload error:", error);
    // }
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex items-center justify-center p-6">
      <div className="max-w-3xl w-full bg-card border border-border p-6 rounded-2xl shadow-xl font-mono">
        <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">          <ShieldPlus size={28} /> Upload Evidence
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* File Input */}
          <div>
            <label className="block mb-1 text-sm">Select Evidence Files</label>
            <input
              type="file"
              multiple
              onChange={handleFileChange}
              className="block w-full text-sm bg-muted border border-border text-foreground rounded p-2 file:mr-4 file:py-1 file:px-2 file:rounded file:border-0 file:text-sm file:font-semibold file:bg-primary file:text-primary-foreground hover:file:bg-primary/90"
            />
            <p className="text-xs text-muted-foreground mt-1">Max upload: 200MB total</p>
          </div>

          {/* Drag & Drop Area */}
          <div
            onDrop={handleDrop}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
              isDragging ? "border-green-500 bg-muted" : "border-cyan-500 bg-muted" 
            }`}
          >
            <UploadCloud size={32} className="mx-auto mb-2 text-cyan-400" />
            Drag & drop files here
          </div>

          {/* File Preview */}
          {files.length > 0 && (
            <div className="bg-muted rounded p-3 border border-border">
              <p className="text-sm mb-2 text-primary font-semibold">Files to be uploaded:</p>
              <ul className="text-sm list-disc list-inside space-y-1">
                {files.map((file, index) => (
                  <li key={index}>
                    {file.name} ({(file.size / (1024 * 1024)).toFixed(2)} MB)
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Actions */}
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
              className="bg-green-600 hover:bg-green-700 text-white"
              onClick={() => navigate(-1)}
            >
              Done
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

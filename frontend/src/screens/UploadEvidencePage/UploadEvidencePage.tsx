import React, { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import axios from "axios";
import { Button } from "../../components/ui/button";
import { ShieldPlus, UploadCloud } from "lucide-react";
import { toast } from "react-toastify"; // Add this import

export function UploadEvidenceForm(): JSX.Element {
  const [files, setFiles] = useState<File[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const navigate = useNavigate();
  const { caseId } = useParams<{ caseId: string }>(); 
  
  const MAX_FILE_SIZE = 200 * 1024 * 1024; // 200MB in bytes

  // Helper function to validate file sizes
  const validateFiles = (filesToValidate: File[]): File[] => {
    const validFiles: File[] = [];
    
    filesToValidate.forEach(file => {
      if (file.size > MAX_FILE_SIZE) {
        toast.error(`Cannot upload because file size > 200MB: ${file.name} (${(file.size / (1024 * 1024)).toFixed(2)} MB)`);
      } else {
        validFiles.push(file);
      }
    });
    
    return validFiles;
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const droppedFiles = Array.from(e.dataTransfer.files);
    const validFiles = validateFiles(droppedFiles);
    setFiles((prev) => [...prev, ...validFiles]);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = e.target.files ? Array.from(e.target.files) : [];
    const validFiles = validateFiles(selectedFiles);
    setFiles((prev) => [...prev, ...validFiles]);
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

    // Additional validation before upload
    const totalSize = files.reduce((sum, file) => sum + file.size, 0);
    if (totalSize > MAX_FILE_SIZE) {
      toast.error(`Cannot upload because total file size > 200MB (${(totalSize / (1024 * 1024)).toFixed(2)} MB)`);
      return;
    }

    // Load current case from localStorage
    const currentCase = JSON.parse(localStorage.getItem("currentCase") || "{}");

    if (!caseId) {
      return;
    }

    // Load current user from sessionStorage
    const user = JSON.parse(sessionStorage.getItem("user") || "{}");
    if (!user.id) {
      return;
    }

    const formData = new FormData();
    files.forEach(file => {
      formData.append("files", file);
    });
    formData.append("caseId", caseId);
    formData.append("uploadedBy", user.id);
    formData.append("fileType", "generic");

    console.log("Uploading evidence for case:", currentCase.ID, "by user:", user.id);

    try {
      await axios.post("https://localhost/api/v1/evidence", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
          "Authorization": `Bearer ${sessionStorage.getItem("authToken") || ""}`
        }
      });

      toast.success("Evidence uploaded successfully!");
      navigate(-1);
    } catch (err: any) {
      console.error("Upload failed:", err.response?.data || err);
      toast.error("Failed to upload evidence. Check console for details.");
    }
  };

  // Add this helper function to check if upload is allowed
  const hasOversizedFiles = files.some(file => file.size > MAX_FILE_SIZE);
  const totalSize = files.reduce((sum, file) => sum + file.size, 0);
  const canUpload = files.length > 0 && !hasOversizedFiles;

  return (
    <div className="min-h-screen bg-background text-foreground flex items-center justify-center p-6">
      <div className="max-w-3xl w-full bg-card border border-border p-6 rounded-2xl shadow-xl font-mono">
        <h1 className="text-3xl font-bold text-primary mb-6 flex items-center gap-2">
          <ShieldPlus size={28} className="text-primary" /> Upload Evidence
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block mb-1 text-sm">Select Evidence Files</label>
            <input
              type="file"
              multiple
              onChange={handleFileChange}
              className="block w-full text-sm bg-muted border border-border text-foreground rounded p-2 file:mr-4 file:py-1 file:px-2 file:rounded file:border-0 file:text-sm file:font-semibold file:bg-primary file:text-primary-foreground hover:file:bg-primary/90"
            />
            <p className="text-xs text-muted-foreground mt-1">Max upload: 200MB per file</p>
          </div>

          <div
            onDrop={handleDrop}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
              isDragging ? "border-secondary bg-muted" : "border-primary bg-muted"
            }`}
          >
            <UploadCloud size={32} className="mx-auto mb-2 text-primary" />
            Drag & drop files here
          </div>

          {files.length > 0 && (
            <div className="bg-muted rounded p-3 border border-border">
              <p className="text-sm mb-2 text-primary font-semibold">Files to be uploaded:</p>
              <ul className="text-sm list-disc list-inside space-y-1">
                {files.map((file, index) => (
                  <li key={index} className={file.size > MAX_FILE_SIZE ? "text-red-500" : ""}>
                    {file.name} ({(file.size / (1024 * 1024)).toFixed(2)} MB)
                    {file.size > MAX_FILE_SIZE && " - Too large!"}
                  </li>
                ))}
              </ul>
              <p className="text-xs text-muted-foreground mt-2">
                Total size: {(totalSize / (1024 * 1024)).toFixed(2)} MB
              </p>
              
              {/* Add warning messages */}
              {hasOversizedFiles && (
                <div className="mt-2 p-2 bg-red-100 border border-red-300 rounded text-red-700 text-xs">
                  <strong>⚠️ Cannot upload:</strong> Some files exceed the 200MB limit. Please remove them to continue.
                </div>
              )}
              
              {!hasOversizedFiles && files.length === 0 && (
                <div className="mt-2 p-2 bg-yellow-100 border border-yellow-300 rounded text-yellow-700 text-xs">
                  <strong>📁 No files selected:</strong> Please select files to upload.
                </div>
              )}
            </div>
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
              className="bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={!canUpload}
              title={
                files.length === 0 
                  ? "Please select files to upload" 
                  : hasOversizedFiles 
                  ? "Remove files larger than 200MB to upload"
                  : "Upload evidence files"
              }
            >
              {hasOversizedFiles 
                ? "Cannot Upload - Files Too Large" 
                : files.length === 0 
                ? "Select Files to Upload" 
                : "Upload Evidence"
              }
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

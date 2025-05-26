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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // TODO: Upload logic
    console.log("Uploading files:", files);
  };

  return (
    <div className="min-h-screen bg-zinc-900 text-white flex items-center justify-center p-6">
      <div className="max-w-3xl w-full bg-zinc-900 p-6 rounded-2xl shadow-xl border border-zinc-700 font-mono">
        <h1 className="text-3xl font-bold text-cyan-400 mb-6 flex items-center gap-2">
          <ShieldPlus size={28} /> Upload Evidence
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block mb-1 text-sm">Select Evidence Files</label>
            <input
              type="file"
              multiple
              onChange={handleFileChange}
              className="block w-full text-sm text-white bg-zinc-800 border border-zinc-600 rounded p-2 file:mr-4 file:py-1 file:px-2 file:rounded file:border-0 file:text-sm file:font-semibold file:bg-cyan-600 file:text-white hover:file:bg-cyan-700"
            />
            <p className="text-xs text-zinc-400 mt-1">Max upload: 200MB total</p>
          </div>

          <div
            onDrop={handleDrop}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
              isDragging ? "border-green-400 bg-zinc-700" : "border-cyan-500 bg-zinc-800"
            }`}
          >
            <UploadCloud size={32} className="mx-auto mb-2 text-cyan-400" />
            Drag & drop files here
          </div>

          {files.length > 0 && (
            <div className="bg-zinc-800 rounded p-3 border border-zinc-600">
              <p className="text-sm mb-2 text-cyan-300 font-semibold">Files to be uploaded:</p>
              <ul className="text-sm list-disc list-inside space-y-1">
                {files.map((file, index) => (
                  <li key={index}>
                    {file.name} ({(file.size / (1024 * 1024)).toFixed(2)} MB)
                  </li>
                ))}
              </ul>
            </div>
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

            <Button type="submit" className="bg-green-600 hover:bg-green-700 text-white">
              Done
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

// src/pages/EvidenceViewer.tsx
import { useState } from "react";
import {
  Bell,
  File,
  Folder,
  Home,
  MessageSquare,
  Search,
  Settings,
  SlidersHorizontal,
  ArrowUpDown,
  Code,
  Image as ImageIcon,
  Video,
  MessageCircle
} from "lucide-react";
import { Link } from "react-router-dom";

// Define file structure
interface FileItem {
  id: string;
  name: string;
  type: 'executable' | 'log' | 'image' | 'document';
  content?: string;
  imageUrl?: string;
}

export const EvidenceViewer = () => {
  // Sample files data
  const files: FileItem[] = [
    {
      id: '1',
      name: 'system_logs.exe',
      type: 'executable',
      content: 'This is a system executable file. Binary content cannot be displayed in text format.'
    },
    {
      id: '2',
      name: 'malware_sample.exe',
      type: 'executable',
      content: 'This is a malware sample file. Handle with extreme caution. Binary content cannot be displayed in text format.'
    },
    {
      id: '3',
      name: 'screenshot_evidence.png',
      type: 'image',
      content: 'Screenshot taken from suspect\'s computer showing suspicious activity.',
      imageUrl: 'https://images.unsplash.com/photo-1516110833967-0b5716ca1387?w=800&h=600&fit=crop'
    }
  ];

  const [selectedFile, setSelectedFile] = useState<FileItem | null>(null);

  const handleFileClick = (file: FileItem) => {
    setSelectedFile(file);
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex">
      {/* Sidebar */}
      <aside className="fixed left-0 top-0 h-full w-80 bg-background border-r border p-6 flex flex-col justify-between z-10">
        <div>
          {/* Logo */}
          <div className="flex items-center gap-3 mb-8">
            <div className="w-14 h-14 rounded-lg overflow-hidden">
              <img
                src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
                alt="AEGIS Logo"
                className="w-full h-full object-cover"
              />
            </div>
            <span className="font-bold text-foreground text-2xl">AEGIS</span>
          </div>

          {/* Navigation */}
          <nav className="space-y-2">
            <Link to="/dashboard">
              <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors">
                <Home className="w-6 h-6" />
                <span className="text-lg">Dashboard</span>
              </div>
            </Link>
            <Link to="/case-management">
              <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors">
                <Folder className="w-6 h-6" />
                <span className="text-lg">Case Management</span>
              </div>
            </Link>
            <div className="flex items-center gap-3 bg-blue-600 text-white p-3 rounded-lg">
              <File className="w-6 h-6" />
              <span className="text-lg font-semibold">Evidence Viewer</span>
            </div>
            <Link to="/secure-chat">
              <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors">
                <MessageSquare className="w-6 h-6" />
                <span className="text-lg">Secure Chat</span>
              </div>
            </Link>
          </nav>
        </div>

        {/* User Profile */}
        <div className="border-t border pt-4">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-muted rounded-full flex items-center justify-center">
              <Link to="/profile">
                <span className="text-foreground font-medium">AU</span>
              </Link>
            </div>
            <div>
              <p className="font-semibold text-foreground">Agent User</p>
              <p className="text-muted-foreground text-sm">user@dfir.com</p>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="ml-80 flex-grow bg-background">
        {/* Topbar */}
        <div className="sticky top-0 z-10 bg-background border-b border p-4">
          <div className="flex items-center justify-between">
            {/* Tabs */}
        {/* Tabs */}
            <div className="flex items-center gap-6">
              <Link to="/dashboard">
                <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                  Dashboard
                </button>
              </Link>
              <button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">
                Evidence Viewer
              </button>
              <Link to="/case-management">
                <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                  Case Management
                </button>
              </Link>
              <Link to="/secure-chat">
                <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                  Secure Chat
                </button>
              </Link>
            </div>

            {/* Right actions */}
            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <input
                  className="w-80 h-12 bg-popover border rounded-lg pl-10 pr-4 text-foreground placeholder-muted-foreground focus:outline-none focus:border-[#636AE8]"
                  placeholder="Search cases, evidence, users"
                />
              </div>
              <Bell className="text-muted-foreground hover:text-foreground w-6 h-6 cursor-pointer" />
              <Link to="/settings"><Settings className="text-muted-foreground hover:text-foreground w-6 h-6 cursor-pointer" /></Link>
              <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center">
                <Link to="/profile">
                  <span className="text-foreground font-medium text-sm">AU</span>
                </Link>
              </div>
            </div>
          </div>
        </div>

        {/* Evidence Viewer Content */}
        <div className="p-8">
          <h1 className="text-3xl font-semibold mb-6">Evidence Viewer</h1>

          <div className="flex gap-8">
            {/* File list panel */}
            <div className="w-1/3 space-y-4">
              <div className="flex justify-between items-center">
                <h2 className="text-xl font-semibold">Case Files</h2>
                <div className="flex gap-2">
                  <button className="flex items-center gap-1 px-3 py-1 text-sm border rounded-lg text-foreground hover:bg-muted">
                    <SlidersHorizontal size={16} />
                    Filter
                  </button>
                  <button className="flex items-center gap-1 px-3 py-1 text-sm border rounded-lg text-foreground hover:bg-muted">
                    <ArrowUpDown size={16} />
                    Sort
                  </button>
                </div>
              </div>
              <div className="space-y-2">
                {files.map((file) => (
                  <button
                    key={file.id}
                    onClick={() => handleFileClick(file)}
                    className={`w-full flex items-center gap-2 p-2 rounded-md transition-colors cursor-pointer ${
                      selectedFile?.id === file.id
                        ? 'bg-[#636AE8] text-white'
                        : 'bg-muted hover:bg-muted/80'
                    }`}
                  >
                    {file.type === 'image' ? (
                      <ImageIcon className="w-5 h-5 text-green-500" />
                    ) : (
                      <File className="w-5 h-5 text-blue-500" />
                    )}
                    <span className="text-left">{file.name}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Viewer panel */}
            <div className="w-2/3 h-[400px] border rounded-lg bg-card">
              {selectedFile ? (
                <div className="p-4 h-full flex flex-col">
                  <div className="border-b border pb-2 mb-4">
                    <h3 className="text-lg font-semibold text-foreground">{selectedFile.name}</h3>
                    <p className="text-sm text-muted-foreground capitalize">{selectedFile.type} file</p>
                  </div>
                  <div className="flex-1 overflow-auto">
                    {selectedFile.type === 'image' && selectedFile.imageUrl ? (
                      <div className="space-y-4">
                        <div className="flex justify-center">
                          <img
                            src={selectedFile.imageUrl}
                            alt={selectedFile.name}
                            className="max-w-full max-h-64 object-contain rounded-lg border"
                          />
                        </div>
                        {selectedFile.content && (
                          <div className="text-foreground text-sm">
                            <strong>Description:</strong> {selectedFile.content}
                          </div>
                        )}
                      </div>
                    ) : (
                      <div className="text-foreground whitespace-pre-wrap">
                        {selectedFile.content}
                      </div>
                    )}
                  </div>
                </div>
              ) : (
                <div className="h-full flex items-center justify-center text-muted-foreground">
                  Select a file to view
                </div>
              )}
            </div>
          </div>

          {/* Annotation tools */}
          <div className="mt-10">
            <h2 className="text-xl font-semibold mb-2">Annotation Tools</h2>
            <div className="flex gap-4">
              <button className="w-10 h-10 flex items-center justify-center bg-muted rounded-full text-foreground hover:bg-[#636AE8] hover:text-white">
                <Code />
              </button>
              <button className="w-10 h-10 flex items-center justify-center bg-muted rounded-full text-foreground hover:bg-[#636AE8] hover:text-white">
                <ImageIcon />
              </button>
              <button className="w-10 h-10 flex items-center justify-center bg-muted rounded-full text-foreground hover:bg-[#636AE8] hover:text-white">
                <MessageCircle />
              </button>
              <button className="w-10 h-10 flex items-center justify-center bg-muted rounded-full text-foreground hover:bg-[#636AE8] hover:text-white">
                <Video />
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};
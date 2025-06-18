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
  Download,
  Share,
  Maximize2,
  Send,
  Info,
  MessageCircle
} from "lucide-react";

// Define file structure
interface FileItem {
  id: string;
  name: string;
  type: 'executable' | 'log' | 'image' | 'document';
  size?: string;
  hash?: string;
  created?: string;
  description?: string;
}

interface AnnotationThread {
  id: string;
  title: string;
  user: string;
  avatar: string;
  time: string;
  messageCount: number;
  isActive?: boolean;
}

interface ThreadMessage {
  id: string;
  user: string;
  avatar: string;
  time: string;
  message: string;
}

export default function EvidenceViewer() {
  // Sample files data
  const files: FileItem[] = [
    {
      id: '1',
      name: 'system_logs.exe',
      type: 'executable',
      size: '110MB',
      hash: 'a1b2c3d4e5f6789abc',
      created: '2024-03-15',
      description: 'Memory dump of workstation WS-0234 captured using FTK Imager following detection of unauthorized PowerShell activity'
    },
    {
      id: '2',
      name: 'malware_sample.exe',
      type: 'executable',
      size: '1.8 MB',
      hash: 'x1y2z3a4b5c6def',
      created: '2024-03-14'
    }
  ];

  const annotationThreads: AnnotationThread[] = [
    {
      id: '1',
      title: 'I noticed something in the image',
      user: 'Adm.1',
      avatar: 'A1',
      time: '18 June 2025',
      messageCount: 2,
      isActive: true
    }
  ];

  const threadMessages: ThreadMessage[] = [
    {
      id: '1',
      user: 'User.1',
      avatar: 'U1',
      time: '1 min ago',
      message: 'I noticed something in the image'
    },
    {
      id: '2',
      user: 'User.1',
      avatar: 'U1',
      time: '1 min ago',
      message: 'I agree it\'s something random'
    }
  ];

  const [selectedFile, setSelectedFile] = useState<FileItem | null>(files[0]);
  const [selectedThread, setSelectedThread] = useState<AnnotationThread | null>(annotationThreads[0]);
  const [newMessage, setNewMessage] = useState('');
  const [searchTerm, setSearchTerm] = useState('');

  const handleFileClick = (file: FileItem) => {
    setSelectedFile(file);
  };

  const handleThreadClick = (thread: AnnotationThread) => {
    setSelectedThread(thread);
  };

  const handleSendMessage = () => {
    if (newMessage.trim()) {
      // Handle sending message logic here
      setNewMessage('');
    }
  };

  return (
    <div className="min-h-screen bg-black text-white flex">
      {/* Sidebar */}
      <aside className="fixed left-0 top-0 h-full w-64 bg-black border-r border-gray-800 p-4 flex flex-col justify-between z-10">
        <div>
          {/* Logo */}
          <div className="flex items-center gap-3 mb-8">
            <div className="w-10 h-10 rounded-lg overflow-hidden">
              <img
                src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
                alt="AEGIS Logo"
                className="w-full h-full object-cover"
              />
            </div>
            <span className="font-bold text-white text-xl">AEGIS</span>
          </div>

          {/* Navigation */}
          <nav className="space-y-1">
            <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-2 rounded-lg transition-colors cursor-pointer">
              <Home className="w-5 h-5" />
              <span className="text-sm">Dashboard</span>
            </div>
            <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-2 rounded-lg transition-colors cursor-pointer">
              <Folder className="w-5 h-5" />
              <span className="text-sm">Case management</span>
            </div>
            <div className="flex items-center gap-3 bg-blue-600 text-white p-2 rounded-lg">
              <File className="w-5 h-5" />
              <span className="text-sm font-medium">Evidence Viewer</span>
            </div>
            <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-2 rounded-lg transition-colors cursor-pointer">
              <MessageSquare className="w-5 h-5" />
              <span className="text-sm">Secure chat</span>
            </div>
          </nav>
        </div>

        {/* User Profile */}
        <div className="border-t border-gray-700 pt-4">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center">
              <span className="text-white font-medium text-xs">AU</span>
            </div>
            <div>
              <p className="font-medium text-white text-sm">Agent User</p>
              <p className="text-gray-400 text-xs cursor-pointer hover:text-white">settings</p>
              <p className="text-gray-400 text-xs cursor-pointer hover:text-white">Logout</p>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="ml-64 flex-grow bg-black flex">
        {/* Header */}
        <div className="fixed top-0 left-64 right-0 z-20 bg-black border-b border-gray-800 p-4">
          <div className="flex items-center justify-between">
            {/* Case Number and Tabs */}
            <div className="flex items-center gap-4">
              {/* Tabs */}
              <div className="flex items-center gap-6">
                <button className="text-gray-400 hover:text-white text-sm transition-colors">
                  Dashboard
                </button>
                <button className="text-blue-400 font-medium text-sm border-b-2 border-blue-400 pb-2">
                  Evidence Viewer
                </button>
                <button className="text-gray-400 hover:text-white text-sm transition-colors">
                  case management
                </button>
                <button className="text-gray-400 hover:text-white text-sm transition-colors">
                  Secure chat
                </button>
              </div>
            </div>

            {/* Right actions */}
            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  className="w-64 h-10 bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 text-white placeholder-gray-400 text-sm focus:outline-none focus:border-blue-500"
                  placeholder="Search cases, evidence, users"
                />
              </div>
              <Bell className="text-gray-400 hover:text-white w-5 h-5 cursor-pointer" />
              <Settings className="text-gray-400 hover:text-white w-5 h-5 cursor-pointer" />
              <div className="w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center">
                <span className="text-white font-medium text-xs">AU</span>
              </div>
            </div>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 flex pt-20">
          {/* Annotation Threads Panel */}
          <div className="w-96 border-r border-gray-800 p-6">
            <div className="flex items-center gap-3 mb-4">
              <h2 className="text-xl font-semibold">Annotation threads</h2>
              <div className="bg-blue-600 text-white px-3 py-1 rounded text-sm font-medium">
                #CS-00579
              </div>
            </div>
            
            {/* Search */}
            <div className="relative mb-4">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                className="w-full h-10 bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 text-white placeholder-gray-400 text-sm focus:outline-none focus:border-blue-500"
                placeholder="Search Evidence"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            {/* Filter and Sort */}
            <div className="flex gap-2 mb-4">
              <button className="flex items-center gap-1 px-3 py-1 text-sm border border-gray-600 rounded-lg text-white hover:bg-gray-800">
                <SlidersHorizontal size={14} />
                filter
              </button>
              <button className="flex items-center gap-1 px-3 py-1 text-sm border border-gray-600 rounded-lg text-white hover:bg-gray-800">
                <ArrowUpDown size={14} />
                sort
              </button>
            </div>

            {/* File List */}
            <div className="space-y-2 mb-6">
              {files.map((file) => (
                <button
                  key={file.id}
                  onClick={() => handleFileClick(file)}
                  className={`w-full flex items-center gap-3 p-3 rounded-md transition-colors cursor-pointer ${
                    selectedFile?.id === file.id
                      ? 'bg-gray-800 border-l-2 border-blue-500'
                      : 'hover:bg-gray-800'
                  }`}
                >
                  <File className="w-5 h-5 text-gray-400 flex-shrink-0" />
                  <div className="text-left flex-1">
                    <div className="font-medium text-sm truncate">{file.name}</div>
                  </div>
                  {file.name === 'malware_sample.exe' && (
                    <Info className="w-4 h-4 text-yellow-500 flex-shrink-0" />
                  )}
                </button>
              ))}
            </div>

            {/* Thread List */}
            {/* Removed annotation threads section */}
          </div>

          {/* Main Viewer Area */}
          <div className="flex-1 flex flex-col">
            {/* File Header */}
            {selectedFile && (
              <div className="border-b border-gray-800 p-6">
                <div className="flex items-center justify-between mb-6">
                  <h1 className="text-2xl font-semibold">#{selectedFile.name}</h1>
                  <div className="flex items-center gap-2">
                    <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg">
                      <Download className="w-5 h-5" />
                    </button>
                    <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg">
                      <MessageCircle className="w-5 h-5" />
                    </button>
                    <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg">
                      <Share className="w-5 h-5" />
                    </button>
                    <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg">
                      <Maximize2 className="w-5 h-5" />
                    </button>
                  </div>
                </div>

                {/* Evidence Information */}
                <div className="grid grid-cols-2 gap-6 mb-6">
                  <div className="bg-gray-900 p-4 rounded-lg">
                    <h3 className="font-semibold mb-3">Evidence Information</h3>
                    <div className="space-y-2 text-sm">
                      <div><span className="text-gray-400">Description:</span></div>
                      <div className="text-gray-300">{selectedFile.description || 'No description available'}</div>
                      <div className="mt-3"><span className="text-gray-400">HDD-Image generic</span></div>
                      <div className="mt-2"><span className="text-gray-400">Size:</span></div>
                      <div className="text-gray-300">{selectedFile.size}</div>
                    </div>
                  </div>

                  <div className="bg-gray-900 p-4 rounded-lg">
                    <h3 className="font-semibold mb-3">Indicators of compromise (IOCs)</h3>
                    <div className="space-y-3 text-sm">
                      <div className="border border-gray-700 rounded p-3">
                        <div className="text-gray-400 mb-1">IP Address: 192.168.1.100</div>
                      </div>
                      <div className="border border-gray-700 rounded p-3">
                        <div className="text-gray-400 mb-1">Hash (MD5):</div>
                        <div className="text-gray-300 font-mono text-xs break-all">a1b2c3d4e5f67890</div>
                        <div className="text-red-400 text-xs mt-1">High</div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Thread Discussion */}
                {selectedThread && (
                  <div className="bg-gray-900 rounded-lg p-4">
                    <div className="flex items-center gap-2 mb-4">
                      <div className="w-6 h-6 bg-gray-600 rounded-full flex items-center justify-center text-xs">
                        A1
                      </div>
                      <span className="font-medium text-sm">Adm.1 started a thread:</span>
                      <span className="text-sm text-gray-400">{selectedThread.title}</span>
                    </div>
                    
                    <div className="bg-gray-800 rounded-lg p-3 mb-4">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className="w-6 h-6 bg-gray-600 rounded-full flex items-center justify-center text-xs">
                            U1
                          </div>
                          <span className="text-sm font-medium">User.1</span>
                          <span className="text-xs text-gray-400">I agree it's something random</span>
                          <span className="text-xs text-gray-500">1 min ago</span>
                        </div>
                        <button className="text-blue-400 text-sm hover:underline">
                          2 Messages →
                        </button>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Right Sidebar - Thread Messages */}
          {selectedThread && (
            <div className="w-80 border-l border-gray-800 bg-black flex flex-col">
              <div className="p-4 border-b border-gray-800">
                <h3 className="font-semibold">I noticed something in the image</h3>
                <p className="text-sm text-gray-400 mt-1">Created by Adm.1</p>
                <p className="text-xs text-gray-500">18 June 2025</p>
              </div>

              {/* Messages */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {threadMessages.map((message) => (
                  <div key={message.id} className="flex gap-3">
                    <div className="w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center text-xs font-medium flex-shrink-0">
                      {message.avatar}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-sm">{message.user}</span>
                        <span className="text-xs text-gray-400">{message.time}</span>
                      </div>
                      <div className="text-sm text-gray-300">{message.message}</div>
                    </div>
                  </div>
                ))}
              </div>

              {/* Message Input */}
              <div className="p-4 border-t border-gray-800">
                <div className="flex items-center gap-2 bg-gray-900 rounded-lg p-3">
                  <input
                    type="text"
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    placeholder="Type your message..."
                    className="flex-1 bg-transparent text-white placeholder-gray-400 text-sm focus:outline-none"
                    onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
                  />
                  <button
                    onClick={handleSendMessage}
                    className="p-1 text-blue-400 hover:bg-gray-800 rounded"
                  >
                    <Send className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
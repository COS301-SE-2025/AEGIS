// // src/pages/EvidenceViewer.tsx
// import { useState } from "react";
// import {
//   Bell,
//   File,
//   Folder,
//   Home,
//   MessageSquare,
//   Search,
//   Settings,
//   SlidersHorizontal,
//   ArrowUpDown,
//   Code,
//   Image as ImageIcon,
//   Video,
//   MessageCircle
// } from "lucide-react";
// import { Link } from "react-router-dom";

// // Define file structure
// interface FileItem {
//   id: string;
//   name: string;
//   type: 'executable' | 'log' | 'image' | 'document';
//   content?: string;
//   imageUrl?: string;
// }

// export const EvidenceViewer = () => {
//   // Sample files data
//   const files: FileItem[] = [
//     {
//       id: '1',
//       name: 'system_logs.exe',
//       type: 'executable',
//       content: 'This is a system executable file. Binary content cannot be displayed in text format.'
//     },
//     {
//       id: '2',
//       name: 'malware_sample.exe',
//       type: 'executable',
//       content: 'This is a malware sample file. Handle with extreme caution. Binary content cannot be displayed in text format.'
//     },
//     {
//       id: '3',
//       name: 'screenshot_evidence.png',
//       type: 'image',
//       content: 'Screenshot taken from suspect\'s computer showing suspicious activity.',
//       imageUrl: 'https://images.unsplash.com/photo-1516110833967-0b5716ca1387?w=800&h=600&fit=crop'
//     }
//   ];

//   const [selectedFile, setSelectedFile] = useState<FileItem | null>(null);

//   const handleFileClick = (file: FileItem) => {
//     setSelectedFile(file);
//   };

//   return (
//     <div className="min-h-screen bg-background text-foreground flex">
//       {/* Sidebar */}
//       <aside className="fixed left-0 top-0 h-full w-80 bg-background border-r border p-6 flex flex-col justify-between z-10">
//         <div>
//           {/* Logo */}
//           <div className="flex items-center gap-3 mb-8">
//             <div className="w-14 h-14 rounded-lg overflow-hidden">
//               <img
//                 src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
//                 alt="AEGIS Logo"
//                 className="w-full h-full object-cover"
//               />
//             </div>
//             <span className="font-bold text-foreground text-2xl">AEGIS</span>
//           </div>

//           {/* Navigation */}
//           <nav className="space-y-2">
//             <Link to="/dashboard">
//               <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors">
//                 <Home className="w-6 h-6" />
//                 <span className="text-lg">Dashboard</span>
//               </div>
//             </Link>
//             <Link to="/case-management">
//               <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors">
//                 <Folder className="w-6 h-6" />
//                 <span className="text-lg">Case Management</span>
//               </div>
//             </Link>
//             <div className="flex items-center gap-3 bg-blue-600 text-white p-3 rounded-lg">
//               <File className="w-6 h-6" />
//               <span className="text-lg font-semibold">Evidence Viewer</span>
//             </div>
//             <Link to="/secure-chat">
//               <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors">
//                 <MessageSquare className="w-6 h-6" />
//                 <span className="text-lg">Secure Chat</span>
//               </div>
//             </Link>
//           </nav>
//         </div>

//         {/* User Profile */}
//         <div className="border-t border pt-4">
//           <div className="flex items-center gap-3">
//             <div className="w-12 h-12 bg-muted rounded-full flex items-center justify-center">
//               <Link to="/profile">
//                 <span className="text-foreground font-medium">AU</span>
//               </Link>
//             </div>
//             <div>
//               <p className="font-semibold text-foreground">Agent User</p>
//               <p className="text-muted-foreground text-sm">user@dfir.com</p>
//             </div>
//           </div>
//         </div>
//       </aside>

//       {/* Main Content */}
//       <main className="ml-80 flex-grow bg-background">
//         {/* Topbar */}
//         <div className="sticky top-0 z-10 bg-background border-b border p-4">
//           <div className="flex items-center justify-between">
//             {/* Tabs */}
//         {/* Tabs */}
//             <div className="flex items-center gap-6">
//               <Link to="/dashboard">
//                 <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
//                   Dashboard
//                 </button>
//               </Link>
//               <button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">
//                 Evidence Viewer
//               </button>
//               <Link to="/case-management">
//                 <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
//                   Case Management
//                 </button>
//               </Link>
//               <Link to="/secure-chat">
//                 <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
//                   Secure Chat
//                 </button>
//               </Link>
//             </div>

//             {/* Right actions */}
//             <div className="flex items-center gap-4">
//               <div className="relative">
//                 <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
//                 <input
//                   className="w-80 h-12 bg-popover border rounded-lg pl-10 pr-4 text-foreground placeholder-muted-foreground focus:outline-none focus:border-[#636AE8]"
//                   placeholder="Search cases, evidence, users"
//                 />
//               </div>
//               <Bell className="text-muted-foreground hover:text-foreground w-6 h-6 cursor-pointer" />
//               <Link to="/settings"><Settings className="text-muted-foreground hover:text-foreground w-6 h-6 cursor-pointer" /></Link>
//               <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center">
//                 <Link to="/profile">
//                   <span className="text-foreground font-medium text-sm">AU</span>
//                 </Link>
//               </div>
//             </div>
//           </div>
//         </div>

//         {/* Evidence Viewer Content */}
//         <div className="p-8">
//           <h1 className="text-3xl font-semibold mb-6">Evidence Viewer</h1>

//           <div className="flex gap-8">
//             {/* File list panel */}
//             <div className="w-1/3 space-y-4">
//               <div className="flex justify-between items-center">
//                 <h2 className="text-xl font-semibold">Case Files</h2>
//                 <div className="flex gap-2">
//                   <button className="flex items-center gap-1 px-3 py-1 text-sm border rounded-lg text-foreground hover:bg-muted">
//                     <SlidersHorizontal size={16} />
//                     Filter
//                   </button>
//                   <button className="flex items-center gap-1 px-3 py-1 text-sm border rounded-lg text-foreground hover:bg-muted">
//                     <ArrowUpDown size={16} />
//                     Sort
//                   </button>
//                 </div>
//               </div>
//               <div className="space-y-2">
//                 {files.map((file) => (
//                   <button
//                     key={file.id}
//                     onClick={() => handleFileClick(file)}
//                     className={`w-full flex items-center gap-2 p-2 rounded-md transition-colors cursor-pointer ${
//                       selectedFile?.id === file.id
//                         ? 'bg-[#636AE8] text-white'
//                         : 'bg-muted hover:bg-muted/80'
//                     }`}
//                   >
//                     {file.type === 'image' ? (
//                       <ImageIcon className="w-5 h-5 text-green-500" />
//                     ) : (
//                       <File className="w-5 h-5 text-blue-500" />
//                     )}
//                     <span className="text-left">{file.name}</span>
//                   </button>
//                 ))}
//               </div>
//             </div>

//             {/* Viewer panel */}
//             <div className="w-2/3 h-[400px] border rounded-lg bg-card">
//               {selectedFile ? (
//                 <div className="p-4 h-full flex flex-col">
//                   <div className="border-b border pb-2 mb-4">
//                     <h3 className="text-lg font-semibold text-foreground">{selectedFile.name}</h3>
//                     <p className="text-sm text-muted-foreground capitalize">{selectedFile.type} file</p>
//                   </div>
//                   <div className="flex-1 overflow-auto">
//                     {selectedFile.type === 'image' && selectedFile.imageUrl ? (
//                       <div className="space-y-4">
//                         <div className="flex justify-center">
//                           <img
//                             src={selectedFile.imageUrl}
//                             alt={selectedFile.name}
//                             className="max-w-full max-h-64 object-contain rounded-lg border"
//                           />
//                         </div>
//                         {selectedFile.content && (
//                           <div className="text-foreground text-sm">
//                             <strong>Description:</strong> {selectedFile.content}
//                           </div>
//                         )}
//                       </div>
//                     ) : (
//                       <div className="text-foreground whitespace-pre-wrap">
//                         {selectedFile.content}
//                       </div>
//                     )}
//                   </div>
//                 </div>
//               ) : (
//                 <div className="h-full flex items-center justify-center text-muted-foreground">
//                   Select a file to view
//                 </div>
//               )}
//             </div>
//           </div>

//           {/* Annotation tools */}
//           <div className="mt-10">
//             <h2 className="text-xl font-semibold mb-2">Annotation Tools</h2>
//             <div className="flex gap-4">
//               <button className="w-10 h-10 flex items-center justify-center bg-muted rounded-full text-foreground hover:bg-[#636AE8] hover:text-white">
//                 <Code />
//               </button>
//               <button className="w-10 h-10 flex items-center justify-center bg-muted rounded-full text-foreground hover:bg-[#636AE8] hover:text-white">
//                 <ImageIcon />
//               </button>
//               <button className="w-10 h-10 flex items-center justify-center bg-muted rounded-full text-foreground hover:bg-[#636AE8] hover:text-white">
//                 <MessageCircle />
//               </button>
//               <button className="w-10 h-10 flex items-center justify-center bg-muted rounded-full text-foreground hover:bg-[#636AE8] hover:text-white">
//                 <Video />
//               </button>
//             </div>
//           </div>
//         </div>
//       </main>
//     </div>
//   );
// };
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
  MessageCircle,
  Shield,
  Clock,
  Users,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Eye,
  FileText,
  Hash,
  Calendar,
  User,
  Tag,
  MoreVertical,
  Reply,
  ThumbsUp
} from "lucide-react";

// Define file structure
interface FileItem {
  id: string;
  name: string;
  type: 'executable' | 'log' | 'image' | 'document' | 'memory_dump' | 'network_capture';
  size?: string;
  hash?: string;
  created?: string;
  description?: string;
  status: 'verified' | 'pending' | 'failed';
  chainOfCustody: string[];
  acquisitionDate: string;
  acquisitionTool: string;
  integrityCheck: 'passed' | 'failed' | 'pending';
  threadCount: number;
  priority: 'high' | 'medium' | 'low';
}

interface AnnotationThread {
  id: string;
  title: string;
  user: string;
  avatar: string;
  time: string;
  messageCount: number;
  participantCount: number;
  isActive?: boolean;
  status: 'open' | 'resolved' | 'pending_approval';
  priority: 'high' | 'medium' | 'low';
  tags: string[];
  fileId: string;
}

interface ThreadMessage {
  id: string;
  user: string;
  avatar: string;
  time: string;
  message: string;
  isApproved?: boolean;
  reactions: { type: string; count: number; users: string[] }[];
  replies?: ThreadMessage[];
}

export const EvidenceViewer  =() =>{
  // Enhanced sample data
  const files: FileItem[] = [
    {
      id: '1',
      name: 'system_memory.dmp',
      type: 'memory_dump',
      size: '8.2 GB',
      hash: 'SHA256: a1b2c3d4e5f6789abc...',
      created: '2024-03-15T14:30:00Z',
      description: 'Memory dump of workstation WS-0234 captured using FTK Imager following detection of unauthorized PowerShell activity',
      status: 'verified',
      chainOfCustody: ['Agent.Smith', 'Forensic.Analyst.1', 'Lead.Investigator'],
      acquisitionDate: '2024-03-15T14:30:00Z',
      acquisitionTool: 'FTK Imager v4.7.1',
      integrityCheck: 'passed',
      threadCount: 3,
      priority: 'high'
    },
    {
      id: '2',
      name: 'malware_sample.exe',
      type: 'executable',
      size: '1.8 MB',
      hash: 'MD5: x1y2z3a4b5c6def...',
      created: '2024-03-14T09:15:00Z',
      description: 'Suspected malware executable recovered from infected system',
      status: 'pending',
      chainOfCustody: ['Field.Agent.2'],
      acquisitionDate: '2024-03-14T09:15:00Z',
      acquisitionTool: 'Manual Collection',
      integrityCheck: 'pending',
      threadCount: 1,
      priority: 'high'
    },
    {
      id: '3',
      name: 'network_capture.pcap',
      type: 'network_capture',
      size: '245 MB',
      hash: 'SHA1: abc123def456...',
      created: '2024-03-13T16:45:00Z',
      description: 'Network traffic capture during incident timeframe',
      status: 'verified',
      chainOfCustody: ['Network.Analyst', 'Forensic.Analyst.1'],
      acquisitionDate: '2024-03-13T16:45:00Z',
      acquisitionTool: 'Wireshark v4.0.3',
      integrityCheck: 'passed',
      threadCount: 2,
      priority: 'medium'
    }
  ];

  const annotationThreads: AnnotationThread[] = [
    {
      id: '1',
      title: 'Suspicious PowerShell activity detected',
      user: 'Forensic.Analyst.1',
      avatar: 'FA',
      time: '2 hours ago',
      messageCount: 5,
      participantCount: 3,
      isActive: true,
      status: 'open',
      priority: 'high',
      tags: ['PowerShell', 'Malware', 'Initial Analysis'],
      fileId: '1'
    },
    {
      id: '2',
      title: 'Memory strings analysis findings',
      user: 'Senior.Analyst',
      avatar: 'SA',
      time: '4 hours ago',
      messageCount: 8,
      participantCount: 2,
      status: 'pending_approval',
      priority: 'medium',
      tags: ['Memory Analysis', 'Strings', 'IOCs'],
      fileId: '1'
    },
    {
      id: '3',
      title: 'Malware classification needed',
      user: 'Malware.Specialist',
      avatar: 'MS',
      time: '6 hours ago',
      messageCount: 3,
      participantCount: 4,
      status: 'open',
      priority: 'high',
      tags: ['Classification', 'Signature Analysis'],
      fileId: '2'
    }
  ];

  const threadMessages: ThreadMessage[] = [
    {
      id: '1',
      user: 'Forensic.Analyst.1',
      avatar: 'FA',
      time: '2 hours ago',
      message: 'Found suspicious PowerShell commands in memory dump. @Senior.Analyst please review the decoded base64 strings.',
      isApproved: true,
      reactions: [
        { type: '👍', count: 2, users: ['Senior.Analyst', 'Lead.Investigator'] }
      ],
      replies: [
        {
          id: '1-1',
          user: 'Senior.Analyst',
          avatar: 'SA',
          time: '1 hour ago',
          message: 'Confirmed. This appears to be a fileless attack. Initiating deeper memory analysis.',
          isApproved: true,
          reactions: []
        }
      ]
    },
    {
      id: '2',
      user: 'Junior.Analyst',
      avatar: 'JA',
      time: '1 hour ago',
      message: 'Should we also check for persistence mechanisms?',
      isApproved: false,
      reactions: []
    }
  ];

  const [selectedFile, setSelectedFile] = useState<FileItem | null>(files[0]);
  const [selectedThread, setSelectedThread] = useState<AnnotationThread | null>(annotationThreads[0]);
  const [newMessage, setNewMessage] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [activeTab, setActiveTab] = useState<'overview' | 'threads' | 'metadata'>('overview');

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'verified': case 'passed': case 'resolved': return 'text-green-400';
      case 'pending': case 'open': return 'text-yellow-400';
      case 'failed': return 'text-red-400';
      default: return 'text-gray-400';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'verified': case 'passed': case 'resolved': return <CheckCircle className="w-4 h-4" />;
      case 'pending': case 'open': return <Clock className="w-4 h-4" />;
      case 'failed': return <XCircle className="w-4 h-4" />;
      default: return <Info className="w-4 h-4" />;
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'text-red-400 bg-red-400/10';
      case 'medium': return 'text-yellow-400 bg-yellow-400/10';
      case 'low': return 'text-green-400 bg-green-400/10';
      default: return 'text-gray-400 bg-gray-400/10';
    }
  };

  const filteredThreads = annotationThreads.filter(thread => 
    selectedFile ? thread.fileId === selectedFile.id : true
  );

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
              <div className="bg-blue-600 text-white px-3 py-1 rounded text-sm font-medium">
                #CS-00579
              </div>
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
          {/* Evidence Files Panel */}
          <div className="w-80 border-r border-gray-800 p-4">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold">Evidence Files</h2>
              <span className="text-sm text-gray-400">{files.length} items</span>
            </div>
            
            {/* Search */}
            <div className="relative mb-4">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                className="w-full h-9 bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 text-white placeholder-gray-400 text-sm focus:outline-none focus:border-blue-500"
                placeholder="Search evidence files"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            {/* Filter and Sort */}
            <div className="flex gap-2 mb-4">
              <button className="flex items-center gap-1 px-3 py-1.5 text-xs border border-gray-600 rounded-lg text-white hover:bg-gray-800">
                <SlidersHorizontal size={12} />
                Filter
              </button>
              <button className="flex items-center gap-1 px-3 py-1.5 text-xs border border-gray-600 rounded-lg text-white hover:bg-gray-800">
                <ArrowUpDown size={12} />
                Sort
              </button>
            </div>

            {/* File List */}
            <div className="space-y-2">
              {files.map((file) => (
                <button
                  key={file.id}
                  onClick={() => setSelectedFile(file)}
                  className={`w-full p-3 rounded-lg border transition-all ${
                    selectedFile?.id === file.id
                      ? 'bg-blue-600/20 border-blue-500'
                      : 'border-gray-700 hover:bg-gray-800/50 hover:border-gray-600'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <File className="w-5 h-5 text-gray-400 flex-shrink-0 mt-0.5" />
                    <div className="flex-1 text-left">
                      <div className="font-medium text-sm truncate mb-1">{file.name}</div>
                      <div className="flex items-center gap-2 mb-2">
                        <span className={`inline-flex items-center gap-1 text-xs ${getStatusColor(file.status)}`}>
                          {getStatusIcon(file.status)}
                          {file.status}
                        </span>
                        <span className={`px-2 py-0.5 rounded text-xs ${getPriorityColor(file.priority)}`}>
                          {file.priority}
                        </span>
                      </div>
                      <div className="flex items-center justify-between text-xs text-gray-400">
                        <span>{file.size}</span>
                        <div className="flex items-center gap-1">
                          <MessageCircle className="w-3 h-3" />
                          <span>{file.threadCount}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Main Viewer Area */}
          <div className="flex-1 flex flex-col">
            {selectedFile && (
              <>
                {/* File Header */}
                <div className="border-b border-gray-800 p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                      <h1 className="text-2xl font-semibold">{selectedFile.name}</h1>
                      <div className={`inline-flex items-center gap-1 px-2 py-1 rounded text-sm ${getStatusColor(selectedFile.status)}`}>
                        {getStatusIcon(selectedFile.status)}
                        {selectedFile.status}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg">
                        <Download className="w-5 h-5" />
                      </button>
                      <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg">
                        <FileText className="w-5 h-5" />
                      </button>
                      <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg">
                        <Share className="w-5 h-5" />
                      </button>
                      <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg">
                        <MoreVertical className="w-5 h-5" />
                      </button>
                    </div>
                  </div>

                  {/* Tabs */}
                  <div className="flex items-center gap-6 border-b border-gray-700">
                    <button
                      onClick={() => setActiveTab('overview')}
                      className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'overview'
                          ? 'text-blue-400 border-blue-400'
                          : 'text-gray-400 border-transparent hover:text-white hover:border-gray-600'
                      }`}
                    >
                      Overview
                    </button>
                    <button
                      onClick={() => setActiveTab('threads')}
                      className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'threads'
                          ? 'text-blue-400 border-blue-400'
                          : 'text-gray-400 border-transparent hover:text-white hover:border-gray-600'
                      }`}
                    >
                      Discussions ({filteredThreads.length})
                    </button>
                    <button
                      onClick={() => setActiveTab('metadata')}
                      className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                        activeTab === 'metadata'
                          ? 'text-blue-400 border-blue-400'
                          : 'text-gray-400 border-transparent hover:text-white hover:border-gray-600'
                      }`}
                    >
                      Metadata
                    </button>
                  </div>
                </div>

                {/* Tab Content */}
                <div className="flex-1 overflow-y-auto p-6">
                  {activeTab === 'overview' && (
                    <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
                      {/* Evidence Information */}
                      <div className="bg-gray-900 p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Shield className="w-5 h-5 text-blue-400" />
                          Evidence Information
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div>
                            <span className="text-gray-400">Description:</span>
                            <p className="text-gray-300 mt-1">{selectedFile.description}</p>
                          </div>
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <span className="text-gray-400">Size:</span>
                              <p className="text-gray-300">{selectedFile.size}</p>
                            </div>
                            <div>
                              <span className="text-gray-400">Type:</span>
                              <p className="text-gray-300 capitalize">{selectedFile.type.replace('_', ' ')}</p>
                            </div>
                          </div>
                          <div>
                            <span className="text-gray-400">Integrity Check:</span>
                            <div className={`inline-flex items-center gap-1 ml-2 ${getStatusColor(selectedFile.integrityCheck)}`}>
                              {getStatusIcon(selectedFile.integrityCheck)}
                              <span className="capitalize">{selectedFile.integrityCheck}</span>
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Chain of Custody */}
                      <div className="bg-gray-900 p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Users className="w-5 h-5 text-green-400" />
                          Chain of Custody
                        </h3>
                        <div className="space-y-3">
                          {selectedFile.chainOfCustody.map((person, index) => (
                            <div key={index} className="flex items-center gap-3">
                              <div className="w-2 h-2 bg-green-400 rounded-full"></div>
                              <div className="flex-1">
                                <div className="text-sm font-medium">{person}</div>
                                <div className="text-xs text-gray-400">
                                  {index === 0 ? 'Original Collector' : 
                                   index === selectedFile.chainOfCustody.length - 1 ? 'Current Custodian' : 'Transferred'}
                                </div>
                              </div>
                              <CheckCircle className="w-4 h-4 text-green-400" />
                            </div>
                          ))}
                        </div>
                      </div>

                      {/* Acquisition Details */}
                      <div className="bg-gray-900 p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Calendar className="w-5 h-5 text-purple-400" />
                          Acquisition Details
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div>
                            <span className="text-gray-400">Acquisition Date:</span>
                            <p className="text-gray-300">{new Date(selectedFile.acquisitionDate).toLocaleString()}</p>
                          </div>
                          <div>
                            <span className="text-gray-400">Tool Used:</span>
                            <p className="text-gray-300">{selectedFile.acquisitionTool}</p>
                          </div>
                          <div>
                            <span className="text-gray-400">Hash:</span>
                            <p className="text-gray-300 font-mono text-xs break-all">{selectedFile.hash}</p>
                          </div>
                        </div>
                      </div>

                      {/* Recent Activity */}
                      <div className="bg-gray-900 p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Clock className="w-5 h-5 text-yellow-400" />
                          Recent Activity
                        </h3>
                        <div className="space-y-3">
                          <div className="flex items-center gap-3 text-sm">
                            <MessageCircle className="w-4 h-4 text-blue-400" />
                            <div className="flex-1">
                              <span className="text-gray-300">New discussion thread created</span>
                              <div className="text-xs text-gray-400">by Forensic.Analyst.1 • 2 hours ago</div>
                            </div>
                          </div>
                          <div className="flex items-center gap-3 text-sm">
                            <CheckCircle className="w-4 h-4 text-green-400" />
                            <div className="flex-1">
                              <span className="text-gray-300">Integrity verification completed</span>
                              <div className="text-xs text-gray-400">System • 4 hours ago</div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  {activeTab === 'threads' && (
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <h3 className="text-lg font-semibold">Discussion Threads</h3>
                        <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm">
                          New Thread
                        </button>
                      </div>
                      
                      {filteredThreads.map((thread) => (
                        <div
                          key={thread.id}
                          className={`border rounded-lg p-4 cursor-pointer transition-all ${
                            selectedThread?.id === thread.id
                              ? 'border-blue-500 bg-blue-600/10'
                              : 'border-gray-700 hover:border-gray-600 hover:bg-gray-800/50'
                          }`}
                          onClick={() => setSelectedThread(thread)}
                        >
                          <div className="flex items-start justify-between mb-3">
                            <div className="flex items-center gap-3">
                              <div className="w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center text-xs font-medium">
                                {thread.avatar}
                              </div>
                              <div>
                                <h4 className="font-medium text-sm">{thread.title}</h4>
                                <div className="flex items-center gap-2 text-xs text-gray-400">
                                  <span>{thread.user}</span>
                                  <span>•</span>
                                  <span>{thread.time}</span>
                                </div>
                              </div>
                            </div>
                            <div className="flex items-center gap-2">
                              <span className={`px-2 py-1 rounded text-xs ${getPriorityColor(thread.priority)}`}>
                                {thread.priority}
                              </span>
                              <span className={`px-2 py-1 rounded text-xs ${getStatusColor(thread.status)}`}>
                                {thread.status.replace('_', ' ')}
                              </span>
                            </div>
                          </div>
                          
                          <div className="flex items-center gap-4 text-xs text-gray-400">
                            <div className="flex items-center gap-1">
                              <MessageSquare className="w-3 h-3" />
                              <span>{thread.messageCount} messages</span>
                            </div>
                            <div className="flex items-center gap-1">
                              <Users className="w-3 h-3" />
                              <span>{thread.participantCount} participants</span>
                            </div>
                          </div>
                          
                          {thread.tags.length > 0 && (
                            <div className="flex items-center gap-2 mt-2">
                              {thread.tags.map((tag, index) => (
                                <span key={index} className="px-2 py-1 bg-gray-700 text-gray-300 rounded text-xs">
                                  {tag}
                                </span>
                              ))}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  )}

                  {activeTab === 'metadata' && (
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                      {/* File Metadata */}
                      <div className="bg-gray-900 p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Hash className="w-5 h-5 text-cyan-400" />
                          File Metadata
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <span className="text-gray-400">File Name:</span>
                              <p className="text-gray-300 font-mono">{selectedFile.name}</p>
                            </div>
                            <div>
                              <span className="text-gray-400">File Size:</span>
                              <p className="text-gray-300">{selectedFile.size}</p>
                            </div>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Hash Values:</span>
                            <div className="mt-2 space-y-2">
                              <div className="bg-gray-800 p-2 rounded">
                                <div className="text-xs text-gray-400 mb-1">SHA256:</div>
                                <div className="text-gray-300 font-mono text-xs break-all">
                                  a1b2c3d4e5f6789abcdef1234567890abcdef1234567890abcdef1234567890ab
                                </div>
                              </div>
                              <div className="bg-gray-800 p-2 rounded">
                                <div className="text-xs text-gray-400 mb-1">MD5:</div>
                                <div className="text-gray-300 font-mono text-xs">
                                  x1y2z3a4b5c6def7890abcdef123456
                                </div>
                              </div>
                            </div>
                          </div>

                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <span className="text-gray-400">Created:</span>
                              <p className="text-gray-300">{new Date(selectedFile.created || '').toLocaleString()}</p>
                            </div>
                            <div>
                              <span className="text-gray-400">Modified:</span>
                              <p className="text-gray-300">{new Date(selectedFile.acquisitionDate).toLocaleString()}</p>
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Forensic Metadata */}
                      <div className="bg-gray-900 p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Shield className="w-5 h-5 text-amber-400" />
                          Forensic Metadata
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div>
                            <span className="text-gray-400">Evidence ID:</span>
                            <p className="text-gray-300 font-mono">EVD-{selectedFile.id.padStart(6, '0')}</p>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Acquisition Method:</span>
                            <p className="text-gray-300">Physical Image</p>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Source Device:</span>
                            <p className="text-gray-300">Workstation WS-0234</p>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Examiner:</span>
                            <p className="text-gray-300">{selectedFile.chainOfCustody[0]}</p>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Case Reference:</span>
                            <p className="text-gray-300">#CS-00579</p>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Legal Status:</span>
                            <div className="flex items-center gap-2 mt-1">
                              <CheckCircle className="w-4 h-4 text-green-400" />
                              <span className="text-green-400">Admissible</span>
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* System Information */}
                      <div className="bg-gray-900 p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Info className="w-5 h-5 text-indigo-400" />
                          System Information
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <span className="text-gray-400">OS Version:</span>
                              <p className="text-gray-300">Windows 11 Pro</p>
                            </div>
                            <div>
                              <span className="text-gray-400">Architecture:</span>
                              <p className="text-gray-300">x64</p>
                            </div>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Computer Name:</span>
                            <p className="text-gray-300">DESKTOP-WS0234</p>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Domain:</span>
                            <p className="text-gray-300">CORPORATE.LOCAL</p>
                          </div>
                          
                          <div>
                            <span className="text-gray-400">Last Boot:</span>
                            <p className="text-gray-300">2024-03-15 08:30:15 UTC</p>
                          </div>
                        </div>
                      </div>

                      {/* Analysis Tools */}
                      <div className="bg-gray-900 p-4 rounded-lg">
                        <h3 className="font-semibold mb-4 flex items-center gap-2">
                          <Settings className="w-5 h-5 text-purple-400" />
                          Analysis History
                        </h3>
                        <div className="space-y-3 text-sm">
                          <div className="border-l-2 border-blue-400 pl-3">
                            <div className="font-medium text-gray-300">Volatility Analysis</div>
                            <div className="text-gray-400 text-xs">Completed • 3 hours ago</div>
                            <div className="text-gray-400 text-xs">Tool: Volatility 3.2.0</div>
                          </div>
                          
                          <div className="border-l-2 border-green-400 pl-3">
                            <div className="font-medium text-gray-300">String Extraction</div>
                            <div className="text-gray-400 text-xs">Completed • 4 hours ago</div>
                            <div className="text-gray-400 text-xs">Tool: strings (GNU binutils)</div>
                          </div>
                          
                          <div className="border-l-2 border-yellow-400 pl-3">
                            <div className="font-medium text-gray-300">Malware Scan</div>
                            <div className="text-gray-400 text-xs">In Progress • Started 1 hour ago</div>
                            <div className="text-gray-400 text-xs">Tool: YARA Rules v4.3.2</div>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </>
            )}
          </div>

          {/* Right Sidebar - Thread Messages */}
          {selectedThread && (
            <div className="w-96 border-l border-gray-800 bg-black flex flex-col">
              {/* Thread Header */}
              <div className="p-4 border-b border-gray-800">
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-semibold text-sm leading-tight">{selectedThread.title}</h3>
                  <button className="p-1 text-gray-400 hover:text-white">
                    <MoreVertical className="w-4 h-4" />
                  </button>
                </div>
                
                <div className="flex items-center gap-2 mb-3">
                  <div className="w-6 h-6 bg-gray-600 rounded-full flex items-center justify-center text-xs">
                    {selectedThread.avatar}
                  </div>
                  <div className="text-sm">
                    <span className="text-gray-300">Created by </span>
                    <span className="font-medium">{selectedThread.user}</span>
                  </div>
                </div>
                
                <div className="flex items-center gap-3 text-xs text-gray-400">
                  <span>{selectedThread.time}</span>
                  <span>•</span>
                  <span>{selectedThread.messageCount} messages</span>
                  <span>•</span>
                  <span>{selectedThread.participantCount} participants</span>
                </div>
                
                {selectedThread.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-2">
                    {selectedThread.tags.map((tag, index) => (
                      <span key={index} className="px-2 py-1 bg-gray-700 text-gray-300 rounded text-xs">
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {/* Messages */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {threadMessages.map((message) => (
                  <div key={message.id} className="space-y-2">
                    <div className="flex gap-3">
                      <div className="w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center text-xs font-medium flex-shrink-0">
                        {message.avatar}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-medium text-sm">{message.user}</span>
                          <span className="text-xs text-gray-400">{message.time}</span>
                          {message.isApproved === false && (
                            <span className="px-2 py-0.5 bg-yellow-600/20 text-yellow-400 text-xs rounded">
                              Pending Approval
                            </span>
                          )}
                          {message.isApproved === true && (
                            <CheckCircle className="w-3 h-3 text-green-400" />
                          )}
                        </div>
                        <div className="text-sm text-gray-300 mb-2">{message.message}</div>
                        
                        {/* Reactions */}
                        {message.reactions.length > 0 && (
                          <div className="flex items-center gap-2 mb-2">
                            {message.reactions.map((reaction, index) => (
                              <button
                                key={index}
                                className="flex items-center gap-1 px-2 py-1 bg-gray-800 rounded-full text-xs hover:bg-gray-700"
                              >
                                <span>{reaction.type}</span>
                                <span className="text-gray-400">{reaction.count}</span>
                              </button>
                            ))}
                          </div>
                        )}
                        
                        {/* Action Buttons */}
                        <div className="flex items-center gap-3 text-xs">
                          <button className="flex items-center gap-1 text-gray-400 hover:text-white">
                            <Reply className="w-3 h-3" />
                            Reply
                          </button>
                          <button className="flex items-center gap-1 text-gray-400 hover:text-white">
                            <ThumbsUp className="w-3 h-3" />
                            React
                          </button>
                          {message.isApproved === false && (
                            <button className="text-green-400 hover:text-green-300">
                              Approve
                            </button>
                          )}
                        </div>
                      </div>
                    </div>
                    
                    {/* Replies */}
                    {message.replies && message.replies.map((reply) => (
                      <div key={reply.id} className="ml-8 pl-4 border-l-2 border-gray-700">
                        <div className="flex gap-3">
                          <div className="w-6 h-6 bg-gray-600 rounded-full flex items-center justify-center text-xs font-medium flex-shrink-0">
                            {reply.avatar}
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-1">
                              <span className="font-medium text-sm">{reply.user}</span>
                              <span className="text-xs text-gray-400">{reply.time}</span>
                              {reply.isApproved && (
                                <CheckCircle className="w-3 h-3 text-green-400" />
                              )}
                            </div>
                            <div className="text-sm text-gray-300">{reply.message}</div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ))}
              </div>

              {/* Message Input */}
              <div className="p-4 border-t border-gray-800">
                <div className="bg-gray-900 rounded-lg p-3">
                  <input
                    type="text"
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    placeholder="Type your message... (@mention users)"
                    className="w-full bg-transparent text-white placeholder-gray-400 text-sm focus:outline-none mb-2"
                    onKeyPress={(e) => e.key === 'Enter' && setNewMessage('')}
                  />
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-xs text-gray-400">
                      <span>@ to mention</span>
                      <span>•</span>
                      <span>Shift+Enter for new line</span>
                    </div>
                    <button
                      onClick={() => setNewMessage('')}
                      className="p-1.5 bg-blue-600 text-white rounded hover:bg-blue-700"
                    >
                      <Send className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
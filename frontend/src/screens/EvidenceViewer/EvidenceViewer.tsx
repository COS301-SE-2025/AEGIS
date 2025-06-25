
import {useEffect,  useState } from "react";
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
import { Link } from "react-router-dom";
import { SidebarToggleButton } from '../../context/SidebarToggleContext';

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

  const initialAnnotationThreads: AnnotationThread[] = [
    
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

const [annotationThreads, setAnnotationThreads] = useState<AnnotationThread[]>(() => {
  const saved = localStorage.getItem('annotationThreads');
  return saved ? JSON.parse(saved) : initialAnnotationThreads;
});

useEffect(() => {
  localStorage.setItem('annotationThreads', JSON.stringify(annotationThreads));
}, [annotationThreads]);

  const [newThreadTitle, setNewThreadTitle] = useState('');
  const [selectedFile, setSelectedFile] = useState<FileItem | null>(files[0]);
  const [selectedThread, setSelectedThread] = useState<AnnotationThread | null>(annotationThreads[0]);
  const [replyingToMessageId, setReplyingToMessageId] = useState<string | null>(null);
  const [replyText, setReplyText] = useState('');
  const [newMessage, setNewMessage] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [activeTab, setActiveTab] = useState<'overview' | 'threads' | 'metadata'>('overview');
  const [allThreadMessages, setAllThreadMessages] = useState<{ [threadId: string]: ThreadMessage[] }>(() => {
    const saved = localStorage.getItem('allThreadMessages');
    return saved ? JSON.parse(saved) : { '1': threadMessages };
  });

useEffect(() => {
  localStorage.setItem('allThreadMessages', JSON.stringify(allThreadMessages));
}, [allThreadMessages]);


  const handleSendMessage = () => {
  if (!newMessage.trim() || !selectedThread) return;

  const newMsg: ThreadMessage = {
    id: Date.now().toString(),
    user: 'Agent.User',
    avatar: 'AU',
    time: 'Just now',
    message: newMessage,
    isApproved: false,
    reactions: [],
    replies: []
  };

  setAllThreadMessages(prev => ({
    ...prev,
    [selectedThread.id]: [...(prev[selectedThread.id] || []), newMsg]
  }));
  setNewMessage('');
};


  
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

  function addNestedReply(messages: ThreadMessage[], parentId: string, reply: ThreadMessage): ThreadMessage[] {
    return messages.map(msg => {
      if (msg.id === parentId) {
        return {
          ...msg,
          replies: [...(msg.replies || []), reply]
        };
      } else if (msg.replies) {
        return {
          ...msg,
          replies: addNestedReply(msg.replies, parentId, reply)
        };
      }
      return msg;
    });
  }

  function updateReplyReaction(replies: ThreadMessage[], replyId: string, user: string): ThreadMessage[] {
  return replies.map(reply => {
    if (reply.id === replyId) {
      const existing = reply.reactions.find(r => r.type === '👍');
      if (existing) {
        if (existing.users.includes(user)) return reply;
        return {
          ...reply,
          reactions: reply.reactions.map(r =>
            r.type === '👍' ? { ...r, count: r.count + 1, users: [...r.users, user] } : r
          )
        };
      } else {
        return {
          ...reply,
          reactions: [...reply.reactions, { type: '👍', count: 1, users: [user] }]
        };
      }
    } else if (reply.replies) {
      return {
        ...reply,
        replies: updateReplyReaction(reply.replies, replyId, user)
      };
    }
    return reply;
  });
}

function updateReplyApproval(replies: ThreadMessage[], replyId: string): ThreadMessage[] {
  return replies.map(reply => {
    if (reply.id === replyId) {
      return { ...reply, isApproved: true };
    } else if (reply.replies) {
      return {
        ...reply,
        replies: updateReplyApproval(reply.replies, replyId)
      };
    }
    return reply;
  });
}


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
              <SidebarToggleButton/>
              <button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">
                Dashboard
              </button>
              <Link to="/case-management"><button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Case Management
              </button></Link>
              <Link to="/evidence-viewer"><button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Evidence Viewer
              </button></Link>
              <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                <Link to="/secure-chat">Secure Chat</Link>
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
                          <div className="flex items-center gap-2">
                            <div className="bg-gray-900 p-4 rounded-lg space-y-2">
                              <input
                                type="text"
                                placeholder="Thread title"
                                className="w-full px-3 py-2 bg-black border border-gray-700 rounded text-white text-sm"
                                value={newThreadTitle}
                                onChange={(e) => setNewThreadTitle(e.target.value)}
                              />
                              <button
                                className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm"
                                onClick={() => {
                                  if (!newThreadTitle.trim()) return;
                                  const newThread: AnnotationThread = {
                                    id: Date.now().toString(),
                                    title: newThreadTitle,
                                    user: 'Agent.User',
                                    avatar: 'AU',
                                    time: 'Just now',
                                    messageCount: 0,
                                    participantCount: 1,
                                    status: 'open',
                                    priority: 'low',
                                    tags: [],
                                    fileId: selectedFile?.id || '1'
                                  };
                                  setAnnotationThreads(prev => [...prev, newThread]);
                                  setSelectedThread(newThread);
                                  setAllThreadMessages(prev => ({ ...prev, [newThread.id]: [] }));
                                  setNewThreadTitle('');
                                }}
                              >
                                Create Thread
                              </button>
                            </div>
                          </div>
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
              {(allThreadMessages[selectedThread?.id || ""] || []).map((message) => (
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
                        <button
                          className="flex items-center gap-1 text-gray-400 hover:text-white"
                          onClick={() => setReplyingToMessageId(message.id)} 
                        >
                          <Reply className="w-3 h-3" />
                          Reply
                        </button>
                            {replyingToMessageId === message.id && (
                            <div className="mt-2 ml-1">
                              <input
                                type="text"
                                value={replyText}
                                onChange={(e) => setReplyText(e.target.value)}
                                placeholder="Type your reply..."
                                className="w-full bg-gray-800 text-white text-sm px-3 py-2 rounded border border-gray-600 focus:outline-none"
                              />
                              <button
                                className="mt-1 px-3 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700"
                                onClick={() => {
                                  if (!replyText.trim() || !selectedThread) return;

                                  const reply: ThreadMessage = {
                                    id: `${replyingToMessageId}-reply-${Date.now()}`,
                                    user: 'Agent.User',
                                    avatar: 'AU',
                                    time: 'Just now',
                                    message: replyText,
                                    isApproved: true,
                                    reactions: []
                                  };

                                  setAllThreadMessages(prev => ({
                                    ...prev,
                                    [selectedThread.id]: addNestedReply(prev[selectedThread.id], replyingToMessageId!, reply)
                                  }));

                                  setReplyText('');
                                  setReplyingToMessageId(null);
                                }}
                              >
                                Send Reply
                              </button>
                            </div>
                          )}

                          <button
                            className="flex items-center gap-1 text-gray-400 hover:text-white"
                            onClick={() => {
                              const currentUser = 'Agent.User';
                              setAllThreadMessages(prev => ({
                                ...prev,
                                [selectedThread.id]: prev[selectedThread.id].map(msg => {
                                  if (msg.id !== message.id) return msg;
                                  const existing = msg.reactions.find(r => r.type === '👍');
                                  if (existing) {
                                    if (existing.users.includes(currentUser)) return msg; // Already reacted
                                    return {
                                      ...msg,
                                      reactions: msg.reactions.map(r =>
                                        r.type === '👍'
                                          ? { ...r, count: r.count + 1, users: [...r.users, currentUser] }
                                          : r
                                      )
                                    };
                                  } else {
                                    return {
                                      ...msg,
                                      reactions: [...msg.reactions, { type: '👍', count: 1, users: [currentUser] }]
                                    };
                                  }
                                })
                              }));
                            }}
                          >
                          <ThumbsUp className="w-3 h-3" />
                            React
                          </button>
                          {message.isApproved === false && (
                          <button
                            className="text-green-400 hover:text-green-300"
                            onClick={() => {
                              setAllThreadMessages(prev => ({
                                ...prev,
                                [selectedThread.id]: prev[selectedThread.id].map(msg =>
                                  msg.id === message.id ? { ...msg, isApproved: true } : msg
                                )
                              }));
                            }}
                          >
                            Approve
                          </button>
                          )}
                        </div>
                      </div>
                    </div>
                    
{/* Replies - Recursive Component */}
                    {message.replies && message.replies.map((reply) => {
                      const renderReply = (replyItem: ThreadMessage, depth: number = 0) => {
                        return (
                          <div key={replyItem.id} className={`${depth === 0 ? 'ml-8' : 'ml-6'} pl-4 border-l-2 border-gray-700 space-y-2`}>
                            <div className="flex gap-3">
                              <div className="w-6 h-6 bg-gray-600 rounded-full flex items-center justify-center text-xs font-medium flex-shrink-0">
                                {replyItem.avatar}
                              </div>
                              <div className="flex-1">
                                <div className="flex items-center gap-2 mb-1">
                                  <span className="font-medium text-sm">{replyItem.user}</span>
                                  <span className="text-xs text-gray-400">{replyItem.time}</span>
                                  {replyItem.isApproved === false && (
                                    <span className="px-2 py-0.5 bg-yellow-600/20 text-yellow-400 text-xs rounded">
                                      Pending Approval
                                    </span>
                                  )}
                                  {replyItem.isApproved === true && (
                                    <CheckCircle className="w-3 h-3 text-green-400" />
                                  )}
                                </div>
                                <div className="text-sm text-gray-300 mb-2">{replyItem.message}</div>

                                {/* Reactions */}
                                {replyItem.reactions.length > 0 && (
                                  <div className="flex items-center gap-2 mb-2">
                                    {replyItem.reactions.map((reaction, index) => (
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
                                  <button
                                    className="flex items-center gap-1 text-gray-400 hover:text-white"
                                    onClick={() => setReplyingToMessageId(replyItem.id)}
                                  >
                                    <Reply className="w-3 h-3" />
                                    Reply
                                  </button>
                                  <button
                                    className="flex items-center gap-1 text-gray-400 hover:text-white"
                                    onClick={() => {
                                      const currentUser = 'Agent.User';
                                      setAllThreadMessages(prev => ({
                                        ...prev,
                                        [selectedThread.id]: prev[selectedThread.id].map(msg => ({
                                          ...msg,
                                          replies: updateReplyReaction(msg.replies || [], replyItem.id, currentUser)
                                        }))
                                      }));
                                    }}
                                  >
                                    <ThumbsUp className="w-3 h-3" />
                                    React
                                  </button>
                                  {replyItem.isApproved === false && (
                                    <button
                                      className="text-green-400 hover:text-green-300"
                                      onClick={() => {
                                        setAllThreadMessages(prev => ({
                                          ...prev,
                                          [selectedThread.id]: prev[selectedThread.id].map(msg => ({
                                            ...msg,
                                            replies: updateReplyApproval(msg.replies || [], replyItem.id)
                                          }))
                                        }));
                                      }}
                                    >
                                      Approve
                                    </button>
                                  )}
                                </div>

                                {/* Reply box for replying to this reply */}
                                {replyingToMessageId === replyItem.id && (
                                  <div className="mt-2 ml-1">
                                    <input
                                      type="text"
                                      value={replyText}
                                      onChange={(e) => setReplyText(e.target.value)}
                                      placeholder="Type your reply..."
                                      className="w-full bg-gray-800 text-white text-sm px-3 py-2 rounded border border-gray-600 focus:outline-none"
                                    />
                                    <button
                                      className="mt-1 px-3 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700"
                                      onClick={() => {
                                        if (!replyText.trim() || !selectedThread) return;

                                        const nestedReply: ThreadMessage = {
                                          id: `${replyItem.id}-reply-${Date.now()}`,
                                          user: 'Agent.User',
                                          avatar: 'AU',
                                          time: 'Just now',
                                          message: replyText,
                                          isApproved: true,
                                          reactions: [],
                                          replies: []
                                        };

                                        setAllThreadMessages(prev => ({
                                          ...prev,
                                          [selectedThread.id]: addNestedReply(prev[selectedThread.id], replyItem.id, nestedReply)
                                        }));

                                        setReplyText('');
                                        setReplyingToMessageId(null);
                                      }}
                                    >
                                      Send Reply
                                    </button>
                                  </div>
                                )}

                                {/* Nested Replies - Recursive Call */}
                                {replyItem.replies && replyItem.replies.length > 0 && (
                                  <div className="mt-2">
                                    {replyItem.replies.map((nestedReply) => 
                                      renderReply(nestedReply, depth + 1)
                                    )}
                                  </div>
                                )}
                              </div>
                            </div>
                          </div>
                        );
                      };

                      return renderReply(reply);
                    })}
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
                    onKeyPress={(e) => { 
                      if(e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault();
                        handleSendMessage();
                        }
                      }}
                  />
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-xs text-gray-400">
                      <span>@ to mention</span>
                      <span>•</span>
                      <span>Shift+Enter for new line</span>
                    </div>
                    <button
                      onClick={handleSendMessage}
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
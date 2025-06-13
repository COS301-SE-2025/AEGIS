import {
  Send,
  FileText,
  Folder,
  Home,
  MessageSquare,
  Menu,
  Plus,
  Search,
  MoreVertical,
  Users,
  Check,
  CheckCheck
} from "lucide-react";
// Navigation will be handled by parent component
import { useState, useEffect, useRef } from "react";

// Type definitions
interface Message {
  id: number;
  user: string;
  color: string;
  content: string;
  time: string;
  status: string;
  self?: boolean;
}

interface Group {
  id: number;
  name: string;
  lastMessage: string;
  lastMessageTime: string;
  unreadCount: number;
  members: string[];
  avatar: string;
}

type ChatMessages = Record<number, Message[]>;

export const SecureChatPage = (): JSX.Element => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [activeChat, setActiveChat] = useState<Group | null>(null);
  const [message, setMessage] = useState("");
  const [showNewGroupModal, setShowNewGroupModal] = useState(false);
  const [newGroupName, setNewGroupName] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const chatEndRef = useRef<HTMLDivElement>(null);

  // Mock data for groups and messages
  const [groups, setGroups] = useState<Group[]>([
    {
      id: 1,
      name: "Incident Response Team",
      lastMessage: "Uploading memory dump to sandbox for detonation and behavioral analysis.",
      lastMessageTime: "2 min ago",
      unreadCount: 3,
      members: ["IR Lead", "Threat Intel", "Forensics", "You"],
      avatar: "🔴"
    },
    {
      id: 2,
      name: "Malware Analysis Unit",
      lastMessage: "Binary shows signs of process hollowing. Investigating persistence.",
      lastMessageTime: "5 min ago",
      unreadCount: 0,
      members: ["Malware Analyst", "Reverse Engineer", "You"],
      avatar: "🦠"
    },
    {
      id: 3,
      name: "Legal & Compliance",
      lastMessage: "Preserve chain of custody logs for Host-29.",
      lastMessageTime: "1 hour ago",
      unreadCount: 1,
      members: ["Legal/Compliance", "CISO", "You"],
      avatar: "⚖️"
    },
    {
      id: 4,
      name: "Threat Intelligence",
      lastMessage: "IOC package updated. Includes hashes, domains, and mutex strings.",
      lastMessageTime: "3 hours ago",
      unreadCount: 0,
      members: ["Threat Intel", "Analyst", "You"],
      avatar: "🎯"
    }
  ]);

  const [chatMessages, setChatMessages] = useState<ChatMessages>({
    1: [
      { id: 1, user: "IR Lead", color: "text-red-400", content: "Initial triage complete. Malware sample isolated from Host-22.", time: "10:23 AM", status: "read" },
      { id: 2, user: "Threat Intel", color: "text-yellow-400", content: "YARA rule matched with UNC2452 variant. Likely APT activity.", time: "10:25 AM", status: "read" },
      { id: 3, user: "You", color: "text-green-400", self: true, content: "Confirmed lateral movement from Host-22 to Host-29 via SMB.", time: "10:27 AM", status: "read" },
      { id: 4, user: "Forensics", color: "text-purple-400", content: "Disk image acquisition started for Host-29. ETA: 15 minutes.", time: "10:30 AM", status: "read" },
      { id: 5, user: "You", color: "text-green-400", self: true, content: "Blocking C2 domain on perimeter firewall. DNS sinkhole active.", time: "10:32 AM", status: "delivered" },
      { id: 6, user: "Legal/Compliance", color: "text-pink-400", content: "Reminder: Preserve chain of custody logs for Host-29.", time: "10:35 AM", status: "unread" },
      { id: 7, user: "Malware Analyst", color: "text-blue-400", content: "Binary shows signs of process hollowing. Investigating persistence.", time: "10:38 AM", status: "unread" },
      { id: 8, user: "IR Lead", color: "text-red-400", content: "Prepare post-incident report template. Add TTP mapping to MITRE ATT&CK.", time: "10:40 AM", status: "unread" },
      { id: 9, user: "Threat Intel", color: "text-yellow-400", content: "IOC package updated. Includes hashes, domains, and mutex strings.", time: "10:42 AM", status: "unread" },
      { id: 10, user: "You", color: "text-green-400", self: true, content: "Uploading memory dump to sandbox for detonation and behavioral analysis.", time: "10:45 AM", status: "sent" },
    ],
    2: [
      { id: 1, user: "Malware Analyst", color: "text-blue-400", content: "Starting reverse engineering of the suspected payload.", time: "9:15 AM", status: "read" },
      { id: 2, user: "You", color: "text-green-400", self: true, content: "Sample uploaded to isolated environment. Proceed with analysis.", time: "9:20 AM", status: "read" },
    ],
    3: [
      { id: 1, user: "Legal/Compliance", color: "text-pink-400", content: "Need incident documentation for regulatory compliance.", time: "8:30 AM", status: "unread" },
    ],
    4: [
      { id: 1, user: "Threat Intel", color: "text-yellow-400", content: "New IOCs detected in the wild. Updating threat feeds.", time: "7:45 AM", status: "read" },
    ]
  });

  const filteredGroups = groups.filter(group =>
    group.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const scrollToBottom = () => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [chatMessages, activeChat]);

  const handleSendMessage = (e?: React.MouseEvent | React.KeyboardEvent) => {
    e?.preventDefault();
    if (!message.trim() || !activeChat) return;

    const newMessage: Message = {
      id: Date.now(),
      user: "You",
      color: "text-green-400",
      self: true,
      content: message,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      status: "sent"
    };

    setChatMessages(prev => ({
      ...prev,
      [activeChat.id]: [...(prev[activeChat.id] || []), newMessage]
    }));

    // Update last message in group
    setGroups(prev => prev.map(group =>
      group.id === activeChat.id
        ? { ...group, lastMessage: message, lastMessageTime: "now" }
        : group
    ));

    setMessage("");
  };

  const handleCreateGroup = (e?: React.MouseEvent | React.KeyboardEvent) => {
    e?.preventDefault();
    if (!newGroupName.trim()) return;

    const newGroup: Group = {
      id: Date.now(),
      name: newGroupName,
      lastMessage: "Group created",
      lastMessageTime: "now",
      unreadCount: 0,
      members: ["You"],
      avatar: "🔒"
    };

    setGroups(prev => [...prev, newGroup]);
    setChatMessages(prev => ({
      ...prev,
      [newGroup.id]: []
    }));

    setNewGroupName("");
    setShowNewGroupModal(false);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "sent":
        return <Check className="w-4 h-4 text-gray-400" />;
      case "delivered":
        return <CheckCheck className="w-4 h-4 text-gray-400" />;
      case "read":
        return <CheckCheck className="w-4 h-4 text-blue-400" />;
      default:
        return null;
    }
  };

  return (
    <div className="bg-black flex w-full h-screen text-white relative overflow-hidden">
      {/* Main Sidebar */}
      <div className={`fixed z-30 top-0 left-0 h-full w-72 bg-gray-900 transition-transform duration-300 ease-in-out ${sidebarOpen ? "translate-x-0" : "-translate-x-full"}`}>
        <div className="p-6">
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
          <nav className="space-y-2">
            <button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-gray-800 rounded-lg">
              <Home className="w-5 h-5" />
              Dashboard
            </button>
            <button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-gray-800 rounded-lg">
              <Folder className="w-5 h-5" />
              Case Management
            </button>
            <button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-gray-800 rounded-lg">
              <FileText className="w-5 h-5" />
              Evidence Viewer
            </button>
            <button className="w-full flex items-center gap-3 text-left px-4 py-2 bg-gray-800 hover:bg-gray-700 rounded-lg">
              <MessageSquare className="w-5 h-5" />
              Secure Chat
            </button>
          </nav>
        </div>
      </div>

      {/* Overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-20"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Chat Layout */}
      <div className="flex flex-1 h-screen">
        {/* Chat List Sidebar */}
        <div className="w-80 bg-gray-900 border-r border-gray-800 flex flex-col">
          {/* Chat Header */}
          <div className="p-4 border-b border-gray-800">
            <div className="flex items-center justify-between mb-4">
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="text-white hover:text-blue-400 mr-3"
              >
                <Menu className="w-6 h-6" />
              </button>
              <h2 className="text-xl font-bold flex-1">Chats</h2>
              <button
                onClick={() => setShowNewGroupModal(true)}
                className="text-white hover:text-blue-400"
                title="Create new group"
              >
                <Plus className="w-6 h-6" />
              </button>
            </div>
            
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                placeholder="Search chats..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-400"
              />
            </div>
          </div>

          {/* Chat List */}
          <div className="flex-1 overflow-y-auto">
            {filteredGroups.map((group) => (
              <div
                key={group.id}
                onClick={() => {
                  setActiveChat(group);
                  setGroups(prev =>
                    prev.map(g => g.id === group.id ? { ...g, unreadCount: 0 } : g)
                  );
                }}

                className={`p-4 border-b border-gray-800 cursor-pointer hover:bg-gray-800 transition-colors ${
                  activeChat?.id === group.id ? "bg-gray-700" : ""
                }`}
              >
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 bg-gray-700 rounded-full flex items-center justify-center text-xl">
                    {group.avatar}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold text-white truncate">{group.name}</h3>
                      <span className="text-xs text-gray-400">{group.lastMessageTime}</span>
                    </div>
                    <div className="flex items-center justify-between mt-1">
                      <p className="text-sm text-gray-400 truncate">{group.lastMessage}</p>
                      {group.unreadCount > 0 && (
                        <span className="bg-blue-500 text-white text-xs rounded-full px-2 py-1 min-w-5 text-center">
                          {group.unreadCount}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Active Chat Area */}
        <div className="flex-1 flex flex-col bg-gray-950">
          {activeChat ? (
            <>
              {/* Chat Header */}
              <div className="p-4 border-b border-gray-800 bg-gray-900">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-gray-700 rounded-full flex items-center justify-center text-lg">
                      {activeChat.avatar}
                    </div>
                    <div>
                      <h3 className="font-semibold text-white">{activeChat.name}</h3>
                      <p className="text-sm text-gray-400 flex items-center gap-1">
                        <Users className="w-4 h-4" />
                        {activeChat.members.length} members
                      </p>
                    </div>
                  </div>
                  <button className="text-gray-400 hover:text-white">
                    <MoreVertical className="w-5 h-5" />
                  </button>
                </div>
              </div>

              {/* Messages Area */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {(chatMessages[activeChat.id] || []).map((msg: Message) => (
                  <div
                    key={msg.id}
                    className={`flex ${msg.self ? "justify-end" : "justify-start"}`}
                  >
                    <div
                      className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                        msg.self
                          ? "bg-blue-600 text-white"
                          : "bg-gray-800 text-white"
                      }`}
                    >
                      {!msg.self && (
                        <p className={`text-xs font-bold ${msg.color} mb-1`}>
                          {msg.user}
                        </p>
                      )}
                      <p className="text-sm">{msg.content}</p>
                      <div className="flex items-center justify-end gap-1 mt-1">
                        <span className="text-xs text-gray-300">{msg.time}</span>
                        {msg.self && getStatusIcon(msg.status)}
                      </div>
                    </div>
                  </div>
                ))}
                <div ref={chatEndRef} />
              </div>

              {/* Message Input */}
              <div className="p-4 border-t border-gray-800 bg-gray-900">
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleSendMessage(e)}
                    placeholder="Type a secure message..."
                    className="flex-1 p-3 rounded-lg bg-gray-800 text-white border border-gray-700 placeholder-gray-400"
                  />
                  <button
                    onClick={handleSendMessage}
                    className="px-4 py-3 bg-blue-600 hover:bg-blue-500 rounded-lg flex items-center justify-center transition-colors"
                  >
                    <Send className="w-5 h-5" />
                  </button>
                </div>
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center">
              <div className="text-center text-gray-400">
                <MessageSquare className="w-16 h-16 mx-auto mb-4 opacity-50" />
                <h3 className="text-xl font-semibold mb-2">Welcome to Secure Chat</h3>
                <p>Select a group to start secure communication</p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* New Group Modal */}
      {showNewGroupModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="bg-gray-800 rounded-lg p-6 w-full max-w-md">
            <h3 className="text-xl font-bold mb-4">Create New Group</h3>
            <div>
              <input
                type="text"
                value={newGroupName}
                onChange={(e) => setNewGroupName(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleCreateGroup(e)}
                placeholder="Enter group name..."
                className="w-full p-3 rounded-lg bg-gray-700 text-white border border-gray-600 placeholder-gray-400 mb-4"
                autoFocus
              />
              <div className="flex justify-end gap-2">
                <button
                  onClick={() => setShowNewGroupModal(false)}
                  className="px-4 py-2 text-gray-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateGroup}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-white"
                >
                  Create
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
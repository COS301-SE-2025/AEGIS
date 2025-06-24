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
  CheckCheck,
  Paperclip,
  LogOut,
  X,
  Reply,
  Download,
  Eye
} from "lucide-react";
import {Link} from "react-router-dom";
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
  attachment?: {
    name: string;
    type: string;
    size: string;
    url?: string;
    isImage?: boolean;
  };
  replyTo?: {
    id: number;
    user: string;
    content: string;
    attachment?: {
      name: string;
      type: string;
    };
  };
}


interface Group {
  id: number;
  name: string;
  lastMessage: string;
  lastMessageTime: string;
  unreadCount: number;
  members: string[];
  avatar: string;
  hasStarted?: boolean;

}

type ChatMessages = Record<number, Message[]>;

export const SecureChatPage = (): JSX.Element => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [activeChat, setActiveChat] = useState<Group | null>(null);
  const [message, setMessage] = useState("");
  const [showNewGroupModal, setShowNewGroupModal] = useState(false);
  const [newGroupName, setNewGroupName] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [showMoreMenu, setShowMoreMenu] = useState(false);
  const [showChatSearch, setShowChatSearch] = useState(false);
  const [chatSearchQuery, setChatSearchQuery] = useState("");
  const [replyingTo, setReplyingTo] = useState<Message | null>(null);
  const [showAttachmentPreview, setShowAttachmentPreview] = useState(false);
  const [previewFile, setPreviewFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string>("");
  const [attachmentMessage, setAttachmentMessage] = useState("");
  const [showImageModal, setShowImageModal] = useState(false);
  const [modalImageUrl, setModalImageUrl] = useState("");
  const [previewFileData, setPreviewFileData] = useState<string>("");
  const [typingUsers, setTypingUsers] = useState<Record<number, string[]>>({});
  const [hasMounted, setHasMounted] = useState(false);

  const chatEndRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const moreMenuRef = useRef<HTMLDivElement>(null);
  //for adding member
  const [showAddMembersModal, setShowAddMembersModal] = useState(false);
  const [newMemberEmail, setNewMemberEmail] = useState("");
  const [availableUsers] = useState([
  "alex.morgan@company.com"
]);
  // Mock data for groups and messages
  const [groups, setGroups] = useState<Group[]>([]);

  const [chatMessages, setChatMessages] = useState<ChatMessages>({});

  // Add this function to simulate incoming messages
// Add this enhanced function to simulate realistic flowing conversations
const simulateIncomingMessage = (chatId: number, delay: number = 1500) => {
const teamMembers = [
  { name: "Alex Morgan", role: "Forensics Analyst", color: "text-blue-400" }
];

  // Get current conversation context
  const currentMessages = chatMessages[chatId] || [];
  const lastMessage = currentMessages[currentMessages.length - 1];
  
  // Conversation flow patterns based on last message content
  const getContextualResponse = (lastMsg: string, sender: string) => {
  const lowerMsg = lastMsg.toLowerCase();
  
  if (lowerMsg.includes('hello') || lowerMsg.includes('hi') || lowerMsg.includes('hey')) {
    return ["Hey! Ready to review that evidence file?", "Hi there! Got the forensic data ready."];
  }
  
  if (lowerMsg.includes('evidence') || lowerMsg.includes('file') || lowerMsg.includes('case')) {
    return [
      "Hash verified. Clean sample.",
      "Metadata extracted successfully.", 
      "Found deleted files in slack space.",
      "Timeline established. 3 access points.",
      "Registry analysis complete.",
      "Network logs show suspicious activity."
    ];
  }
  
  // Default short responses
  return [
    "Got it.",
    "Confirmed.", 
    "Checking now.",
    "Analysis complete.",
    "Roger that.",
    "On it."
  ];
};

  setTimeout(() => {
    // Choose appropriate team member based on context
    let availableMembers = teamMembers;
    const lastSender = lastMessage?.user;
    
    // Don't let the same person respond twice in a row
    if (lastSender && lastSender !== "You") {
      availableMembers = teamMembers.filter(member => 
        `${member.name} (${member.role})` !== lastSender
      );
    }
    
    const selectedMember = availableMembers[Math.floor(Math.random() * availableMembers.length)];
    const responses = getContextualResponse(lastMessage?.content || "", lastMessage?.user || "");
    const selectedResponse = responses[Math.floor(Math.random() * responses.length)];
    
    const newMessage: Message = {
      id: Date.now() + Math.random(),
      user: `${selectedMember.name} (${selectedMember.role})`,
      color: selectedMember.color,
      content: selectedResponse,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      status: "read"
    };

    setChatMessages(prev => ({
      ...prev,
      [chatId]: [...(prev[chatId] || []), newMessage]
    }));

    // Update group's last message
    setGroups(prev => prev.map(group =>
      group.id === chatId
        ? { 
            ...group, 
            lastMessage: selectedResponse, 
            lastMessageTime: "now", 
            unreadCount: group.id === activeChat?.id ? 0 : group.unreadCount + 1,
            hasStarted: true

          }
        : group
    ));
  }, delay);
};
const simulateTyping = (chatId: number, userName?: string) => {
  const user = "Alex Morgan (Forensics Analyst)";

  
  
   setTypingUsers(prev => ({
    ...prev,
    [chatId]: [user]
  }));

  setTimeout(() => {
    setTypingUsers(prev => ({
      ...prev,
      [chatId]: (prev[chatId] || []).filter(u => u !== user)
    }));
  },  15000 + Math.random() * 25000);

};


  const filteredGroups = groups.filter(group =>
    group.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Filter messages based on chat search
  const filteredMessages = activeChat && chatMessages[activeChat.id] 
    ? chatMessages[activeChat.id].filter(msg =>
        msg.content.toLowerCase().includes(chatSearchQuery.toLowerCase())
      )
    : [];

  const displayMessages = showChatSearch && chatSearchQuery ? filteredMessages : (activeChat ? chatMessages[activeChat.id] || [] : []);

  const scrollToBottom = () => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [chatMessages, activeChat]);

  // Close more menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (moreMenuRef.current && !moreMenuRef.current.contains(event.target as Node)) {
        setShowMoreMenu(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  useEffect(() => {
  const savedSidebar = localStorage.getItem("sidebarOpen");
  if (savedSidebar) {
    setSidebarOpen(savedSidebar === "true");
    }
  }, []);

  useEffect(() => {
    localStorage.setItem("sidebarOpen", sidebarOpen.toString());
  }, [sidebarOpen]);

  useEffect(() => {
  setHasMounted(true);
}, []);

  // Clean up preview URL when component unmounts or preview changes
  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

// Simulate random chat activity with better conversation flow
  useEffect(() => {
    const interval = setInterval(() => {
      if (groups.length === 0) return;
      
      // Randomly pick a chat to add activity to
      const randomChat = groups[Math.floor(Math.random() * groups.length)];
      const teamMembers = [
    { name: "Alex Morgan", role: "Forensics Analyst", color: "text-blue-400" }
  ];
  const randomUser = `${teamMembers[0].name} (${teamMembers[0].role})`;    
    // 30% chance of just typing, 70% chance of sending message
    if (Math.random() > 0.7) {
      simulateTyping(randomChat.id, randomUser);
    } else {
      simulateIncomingMessage(randomChat.id, 1000);
    }
  }, 15000 + Math.random() * 25000); // Every 15-40 seconds

  return () => clearInterval(interval);
}, [groups, chatMessages]);
// Save to localStorage
  useEffect(() => {
    const hasRealMessages = Object.values(chatMessages).some(msgs => msgs.length > 0);
    const activeGroups = groups.filter(g => g.hasStarted);

    if (activeGroups.length > 0 && hasRealMessages) {
      localStorage.setItem('chatGroups', JSON.stringify(activeGroups));
      localStorage.setItem('chatMessages', JSON.stringify(chatMessages));
    }
  }, [groups, chatMessages]);

  useEffect(() => {
    const savedGroups = localStorage.getItem('chatGroups');
    const savedMessages = localStorage.getItem('chatMessages');
    
    if (savedGroups) setGroups(JSON.parse(savedGroups));
    if (savedMessages) setChatMessages(JSON.parse(savedMessages));
  }, []);


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
    status: "sent",
    ...(replyingTo && {
      replyTo: {
        id: replyingTo.id,
        user: replyingTo.user,
        content: replyingTo.content,
        ...(replyingTo.attachment && {
          attachment: {
            name: replyingTo.attachment.name,
            type: replyingTo.attachment.type
          }
        })
      }
    })
  };

  setChatMessages(prev => ({
    ...prev,
    [activeChat.id]: [...(prev[activeChat.id] || []), newMessage]
  }));

  setGroups(prev => prev.map(group =>
    group.id === activeChat.id
      ? { ...group, lastMessage: message, lastMessageTime: "now", hasStarted: true }
      : group
  ));

  setMessage("");
  setReplyingTo(null);

  // Simulate status updates
  setTimeout(() => {
    setChatMessages(prev => ({
      ...prev,
      [activeChat.id]: prev[activeChat.id].map(msg =>
        msg.id === newMessage.id ? { ...msg, status: "delivered" } : msg
      )
    }));
  }, 1000);

  setTimeout(() => {
    setChatMessages(prev => ({
      ...prev,
      [activeChat.id]: prev[activeChat.id].map(msg =>
        msg.id === newMessage.id ? { ...msg, status: "read" } : msg
      )
    }));
  }, 2000);

  // Trigger auto-reply simulation
  simulateIncomingMessage(activeChat.id, 3000 + Math.random() * 5000);
};

  const handleFileSelection = async (event: React.ChangeEvent<HTMLInputElement>) => {
  const files = event.target.files;
  if (!files || files.length === 0) return;

  const file = files[0];
  
  // Convert file to base64 for persistent storage
  const fileData = await new Promise<string>((resolve) => {
    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target?.result as string);
    reader.readAsDataURL(file);
  });
  
  const url = URL.createObjectURL(file);
  
  setPreviewFile(file);
  setPreviewUrl(url);
  setShowAttachmentPreview(true);
  setAttachmentMessage("");

  // Store the base64 data for later use
  setPreviewFileData(fileData);

  // Reset file input
  if (fileInputRef.current) {
    fileInputRef.current.value = '';
  }
};
  const meaningfulGroups = groups.filter(g => g.hasStarted);
if (meaningfulGroups.length > 0) {
  localStorage.setItem('chatGroups', JSON.stringify(meaningfulGroups));
}


  const handleSendAttachment = () => {
  if (!previewFile || !activeChat || !previewFileData) return;

  const isImage = previewFile.type.startsWith('image/');
  
  const newMessage: Message = {
    id: Date.now(),
    user: "You",
    color: "text-green-400",
    self: true,
    content: attachmentMessage || `Shared ${isImage ? 'an image' : 'a file'}: ${previewFile.name}`,
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    status: "sent",
    attachment: {
      name: previewFile.name,
      type: previewFile.type,
      size: (previewFile.size / 1024).toFixed(1) + " KB",
      url: previewFileData, // Use base64 data instead of blob URL
      isImage
    },
      ...(replyingTo && {
        replyTo: {
          id: replyingTo.id,
          user: replyingTo.user,
          content: replyingTo.content,
          ...(replyingTo.attachment && {
            attachment: {
              name: replyingTo.attachment.name,
              type: replyingTo.attachment.type
            }
          })
        }
      })
    };

    setChatMessages(prev => ({
      ...prev,
      [activeChat.id]: [...(prev[activeChat.id] || []), newMessage]
    }));
      // Update last message in group
    const lastMessageText = attachmentMessage ? attachmentMessage : `📎 ${previewFile.name}`;
    setGroups(prev => prev.map(group =>
      group.id === activeChat.id
        ? { ...group, lastMessage: lastMessageText, lastMessageTime: "now" }
        : group
    ));
    // Reset states
    setShowAttachmentPreview(false);
    setPreviewFile(null);
    setPreviewUrl("");
    setAttachmentMessage("");
    setReplyingTo(null);
  };

  const handleCancelAttachment = () => {
    setShowAttachmentPreview(false);
    setPreviewFile(null);
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
    }
    setPreviewUrl("");
    setAttachmentMessage("");
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

  const handleExitGroup = () => {
    if (!activeChat) return;
    
    setGroups(prev => prev.filter(group => group.id !== activeChat.id));
    setChatMessages(prev => {
      const newMessages = { ...prev };
      delete newMessages[activeChat.id];
      return newMessages;
    });
    setActiveChat(null);
    setShowMoreMenu(false);
  };
  const handleReplyToMessage = (message: Message) => {
    setReplyingTo(message);
  };

  const handleImageClick = (url: string) => {
    setModalImageUrl(url);
    setShowImageModal(true);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "sent":
        return <Check className="w-4 h-4 text-muted-foreground" />;
      case "delivered":
        return <CheckCheck className="w-4 h-4 text-muted-foreground" />;
      case "read":
        return <CheckCheck className="w-4 h-4 text-blue-400" />;
      default:
        return null;
    }
  };
  const handleAddMember = (e?: React.MouseEvent | React.KeyboardEvent) => {
  e?.preventDefault();
  if (!newMemberEmail.trim() || !activeChat) return;

  // Check if user is already in the group
  if (activeChat.members.includes(newMemberEmail)) {
    alert("User is already in this group");
    return;
  }

  // Update the active chat members
  const updatedChat = {
    ...activeChat,
    members: [...activeChat.members, newMemberEmail]
  };

  // Update groups state
  setGroups(prev => prev.map(group =>
    group.id === activeChat.id ? updatedChat : group
  ));

  // Update active chat
  setActiveChat(updatedChat);

  // Add a system message about the new member
  const systemMessage: Message = {
    id: Date.now(),
    user: "System",
    color: "text-gray-400",
    content: `${newMemberEmail} was added to the group`,
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    status: "read"
  };

  setChatMessages(prev => ({
    ...prev,
    [activeChat.id]: [...(prev[activeChat.id] || []), systemMessage]
  }));

  setNewMemberEmail("");
  setShowAddMembersModal(false);
  setShowMoreMenu(false);
  

};
  const MessageComponent = ({ msg }: { msg: Message }) => (
    <div className={`flex ${msg.self ? "justify-end" : "justify-start"} group`}>
      <div
        className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg relative ${
          msg.self
            ? "bg-blue-600 text-white"
            : "bg-muted text-foreground"
        }`}
      >
        {/* Reply preview */}
        {msg.replyTo && (
          <div className={`mb-2 p-2 rounded border-l-4 ${
            msg.self ? 'border-white/30 bg-white/10' : 'border-blue-400 bg-blue-50 dark:bg-blue-900/20'
          }`}>
            <p className={`text-xs font-semibold ${msg.self ? 'text-white/80' : 'text-blue-600'}`}>
              {msg.replyTo.user}
            </p>
            <p className={`text-xs truncate ${msg.self ? 'text-white/70' : 'text-muted-foreground'}`}>
              {msg.replyTo.attachment 
                ? `📎 ${msg.replyTo.attachment.name}`
                : msg.replyTo.content
              }
            </p>
          </div>
        )}

        {!msg.self && (
          <p className={`text-xs font-bold ${msg.color} mb-1`}>
            {msg.user}
          </p>
        )}

        {/* Attachment preview */}
        {msg.attachment && (
          <div className="mb-2">
            {msg.attachment.isImage ? (
              <div className="relative">
                <img
                  src={msg.attachment.url}
                  alt={msg.attachment.name}
                  className="max-w-full h-auto rounded cursor-pointer hover:opacity-90 transition-opacity"
                  onClick={() => handleImageClick(msg.attachment!.url!)}
                />
                <button
                  onClick={() => handleImageClick(msg.attachment!.url!)}
                  className="absolute top-2 right-2 bg-black bg-opacity-50 text-white p-1 rounded-full hover:bg-opacity-70 transition-all"
                >
                  <Eye className="w-4 h-4" />
                </button>
              </div>
            ) : (
              <div className={`p-3 rounded border ${msg.self ? 'bg-black/20 border-white/20' : 'bg-accent border-border'}`}>
                <div className="flex items-center gap-2">
                  <FileText className="w-5 h-5" />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate text-sm">{msg.attachment.name}</p>
                    <p className="text-xs opacity-70">{msg.attachment.size}</p>
                  </div>
                  <button className="p-1 hover:bg-black/20 rounded">
                    <Download className="w-4 h-4" />
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        <p className="text-sm">{msg.content}</p>
        
        <div className="flex items-center justify-between mt-1">
          <span className="text-xs opacity-70">{msg.time}</span>
          <div className="flex items-center gap-1">
            {msg.self && getStatusIcon(msg.status)}
            {!msg.self && (
              <button
                onClick={() => handleReplyToMessage(msg)}
                className="opacity-0 group-hover:opacity-100 p-1 hover:bg-black/20 rounded transition-all"
                title="Reply"
              >
                <Reply className="w-3 h-3" />
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );


  return (
    <div className="bg-background flex w-full h-screen text-foreground relative">
      {/* Main Sidebar - Fixed positioning without overlay */}
      {hasMounted && (
      <div className={`fixed z-30 top-0 left-0 h-full w-72 bg-card border-r border-border transition-transform duration-300 ease-in-out ${sidebarOpen ? "translate-x-0" : "-translate-x-full"}`}>
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
            <span className="font-bold text-foreground text-xl">AEGIS</span>
          </div>

          {/* Navigation */}
          <nav className="space-y-2">
            <Link to="/dashboard"><button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-muted rounded-lg">
              <Home className="w-5 h-5" />
              Dashboard
            </button></Link>
            <Link to="/case-management"><button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-muted rounded-lg">
              <Folder className="w-5 h-5" />
              Case Management
            </button></Link>
            <Link to="/evidence-viewer"><button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-muted rounded-lg">
              <FileText className="w-5 h-5" />
              Evidence Viewer
            </button></Link>
            <button className="w-full flex items-center gap-3 text-left px-4 py-2 bg-muted hover:bg-accent rounded-lg">
              <MessageSquare className="w-5 h-5" />
              Secure Chat
            </button>
          </nav>
        </div>
      </div>)}
        {/* Overlay */}
        {sidebarOpen && (
          <div
            className="fixed inset-0 bg-black bg-opacity-50 z-20"
            onClick={() => setSidebarOpen(false)}
          />
        )}

      {/* Chat Layout - Adjusted margin for sidebar */}
      <div className={`flex flex-1 h-screen transition-all duration-300 ${sidebarOpen ? 'ml-72' : 'ml-0'}`}>
        {/* Chat List Sidebar */}
        <div className="w-80 min-w-80 max-w-80 bg-card border-r border-border flex flex-col">          {/* Chat Header */}
          <div className="p-4 border-b border">
            <div className="flex items-center justify-between mb-4">
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="text-foreground hover:text-blue-400 mr-3"
              >
                <Menu className="w-6 h-6" />
              </button>
              <h2 className="text-xl font-bold flex-1">Chats</h2>
              <button
                onClick={() => setShowNewGroupModal(true)}
                className="text-foreground hover:text-blue-400"
                title="Create new group"
              >
                <Plus className="w-6 h-6" />
              </button>
            </div>
            
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search chats..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-muted border border-border rounded-lg text-foreground placeholder-muted-foreground"
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
                  setShowChatSearch(false);
                  setChatSearchQuery("");
                  setReplyingTo(null);
                }}
                className={`p-4 border-b border-border cursor-pointer hover:bg-muted transition-colors ${
                  activeChat?.id === group.id ? "bg-accent" : ""
                }`}
              >
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 bg-accent rounded-full flex items-center justify-center text-xl">
                    {group.avatar}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold text-foreground truncate">{group.name}</h3>
                      <span className="text-xs text-muted-foreground">{group.lastMessageTime}</span>
                    </div>
                    <div className="flex items-center justify-between mt-1">
                      <p className="text-sm text-muted-foreground truncate">{group.lastMessage}</p>
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
        <div className="flex-1 flex flex-col bg-background">
          {activeChat ? (
            <>
              {/* Chat Header */}
              <div className="p-4 border-b border-border bg-card">
                {showChatSearch ? (
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => {
                        setShowChatSearch(false);
                        setChatSearchQuery("");
                      }}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <X className="w-5 h-5" />
                    </button>
                    <div className="flex-1 relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <input
                        type="text"
                        placeholder="Search messages..."
                        value={chatSearchQuery}
                        onChange={(e) => setChatSearchQuery(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 bg-muted border border-border rounded-lg text-foreground placeholder-muted-foreground"
                        autoFocus
                      />
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 bg-accent rounded-full flex items-center justify-center text-lg">
                        {activeChat.avatar}
                      </div>
                      <div>
                        <h3 className="font-semibold text-foreground">{activeChat.name}</h3>
                        <p className="text-sm text-muted-foreground flex items-center gap-1">
                          <Users className="w-4 h-4" />
                          {activeChat.members.length} members
                        </p>
                      </div>
                    </div>
                    <div className="relative" ref={moreMenuRef}>
                      <button 
                        onClick={() => setShowMoreMenu(!showMoreMenu)}
                        className="text-muted-foreground hover:text-foreground"
                      >
                        <MoreVertical className="w-5 h-5" />
                      </button>
                      
                      {/* More Menu Dropdown */}
                      {showMoreMenu && (
                      <div className="absolute right-0 top-8 bg-background border border-border rounded-lg shadow-lg py-2 w-48 z-50">
                          <button
                            onClick={() => {
                              setShowChatSearch(true);
                              setShowMoreMenu(false);
                            }}
                            className="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-muted"
                          >
                            <Search className="w-4 h-4" />
                            Search
                          </button>
                          <button
                            onClick={() => {
                              setShowAddMembersModal(true);
                              setShowMoreMenu(false);
                            }}
                            className="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-muted"
                          >
                            <Users className="w-4 h-4" />
                            Add Members
                          </button>
                          <button
                            onClick={handleExitGroup}
                            className="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-muted text-red-400"
                          >
                            <LogOut className="w-4 h-4" />
                            Exit Group
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>

              {/* Messages Area */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {displayMessages.map((msg: Message) => (
                  <MessageComponent key={msg.id} msg={msg} />
                ))}
                <div ref={chatEndRef} />
              </div>

              {/* Reply Preview */}
              {replyingTo && (
                <div className="px-4 py-2 bg-muted border-t border-border">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <p className="text-sm font-medium text-blue-600">
                        Replying to {replyingTo.user}
                      </p>
                      <p className="text-xs text-muted-foreground truncate">
                        {replyingTo.attachment 
                          ? `📎 ${replyingTo.attachment.name}`
                          : replyingTo.content
                        }
                      </p>
                    </div>
                    <button
                      onClick={() => setReplyingTo(null)}
                      className="p-1 hover:bg-accent rounded"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              )}

            {/* Typing Indicator */}
            {typingUsers[activeChat.id]?.length > 0 && (
              <div className="px-4 py-2 text-sm text-muted-foreground">
                {typingUsers[activeChat.id].join(", ")} {typingUsers[activeChat.id].length === 1 ? "is" : "are"} typing...
              </div>
            )}
              {/* Message Input */}
              <div className="p-4 border-t border-border bg-card">
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => fileInputRef.current?.click()}
                    className="p-3 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg transition-colors"
                    title="Attach file"
                  >
                    <Paperclip className="w-5 h-5" />
                  </button>
                  <input
                    ref={fileInputRef}
                    type="file"
                    onChange={handleFileSelection}
                    className="hidden"
                    accept="*/*"
                  />
                  <input
                    type="text"
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleSendMessage(e)}
                    placeholder="Type a secure message..."
                    className="flex-1 p-3 rounded-lg bg-muted text-foreground border border-border placeholder-muted-foreground"
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
              <div className="text-center text-muted-foreground">
                <MessageSquare className="w-16 h-16 mx-auto mb-4 opacity-50" />
                <h3 className="text-xl font-semibold mb-2">Welcome to Secure Chat</h3>
                <p>Select a group to start secure communication</p>
              </div>
            </div>
          )}
        </div>
      </div>
      {/* Attachment Preview Modal */}
      {showAttachmentPreview && previewFile && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
<div className="rounded-lg p-6 w-full max-w-md max-h-[90vh] overflow-y-auto border-[3px] border-border bg-background shadow-xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-bold text-foreground text-lg mb-4">Send Attachment</h3>
              <button
                onClick={handleCancelAttachment}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* File Preview */}
            <div className="mb-4">
              {previewFile.type.startsWith('image/') ? (
                 // Fixed size image preview container
            <div className="w-full h-64 overflow-hidden rounded-lg border border-border bg-muted flex items-center justify-center">
            <img
              src={previewUrl}
              alt={previewFile.name}
              className="max-w-full max-h-full object-contain"
            />
            </div>
              ) : (
               // Fixed size file preview
              <div className="w-full h-32 p-4 bg-muted rounded-lg border border-border flex items-center justify-center">
                <div className="flex items-center gap-3">
                  <FileText className="w-12 h-12 text-blue-500 flex-shrink-0" />
                  <div className="min-w-0">
                    <p className="font-medium truncate">{previewFile.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {(previewFile.size / 1024).toFixed(1)} KB
                    </p>
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Reply Preview in Attachment Modal */}
            {replyingTo && (
              <div className="mb-4 p-3 bg-muted rounded-lg border-l-4 border-blue-400">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-blue-600">
                      Replying to {replyingTo.user}
                    </p>
                    <p className="text-xs text-muted-foreground truncate">
                      {replyingTo.attachment 
                        ? `📎 ${replyingTo.attachment.name}`
                        : replyingTo.content
                      }
                    </p>
                  </div>
                  <button
                    onClick={() => setReplyingTo(null)}
                    className="p-1 hover:bg-accent rounded"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              </div>
            )}

            {/* Message Input */}
            <div className="mb-4">
              <input
                type="text"
                value={attachmentMessage}
                onChange={(e) => setAttachmentMessage(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSendAttachment()}
                placeholder="Add a message..."
                className="w-full p-3 rounded-lg bg-muted text-foreground border border-border placeholder-muted-foreground"
                autoFocus
              />
            </div>

            {/* Action Buttons */}
            <div className="flex justify-end gap-2">
              <button
                onClick={handleCancelAttachment}
                className="px-4 py-2 text-muted-foreground hover:text-foreground"
              >
                Cancel
              </button>
              <button
                onClick={handleSendAttachment}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-white flex items-center gap-2"
              >
                <Send className="w-4 h-4" />
                Send
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Image Viewer Modal */}
      {showImageModal && (
        <div className="fixed inset-0 bg-black bg-opacity-90 z-50 flex items-center justify-center p-4">
          <div className="relative max-w-4xl max-h-full">
            <button
              onClick={() => setShowImageModal(false)}
              className="absolute top-4 right-4 text-white hover:text-gray-300 bg-black bg-opacity-50 rounded-full p-2"
            >
              <X className="w-6 h-6" />
            </button>
            <img
              src={modalImageUrl}
              alt="Full size view"
              className="max-w-full max-h-full object-contain rounded-lg"
            />
          </div>
        </div>
      )}

      {/* New Group Modal */}
      {showNewGroupModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="rounded-lg p-6 w-full max-w-md max-h-[90vh] overflow-y-auto border-[3px] border-border bg-background shadow-xl">
            <h3 className="text-xl font-bold mb-4">Create New Group</h3>
            <div>
              <input
                type="text"
                value={newGroupName}
                onChange={(e) => setNewGroupName(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleCreateGroup(e)}
                placeholder="Enter group name..."
                className="w-full p-3 rounded-lg bg-muted text-foreground border border-border placeholder-muted-foreground mb-4"
                autoFocus
              />
              <div className="flex justify-end gap-2">
                <button
                  onClick={() => setShowNewGroupModal(false)}
                  className="px-4 py-2 text-muted-foreground hover:text-foreground"
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
      {/* Add Members Modal */}
      {showAddMembersModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
            <div className="rounded-lg p-6 w-full max-w-md max-h-[90vh] overflow-y-auto border-[3px] border-border bg-background shadow-xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold">Add Members</h3>
              <button
                onClick={() => setShowAddMembersModal(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Current Members */}
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-muted-foreground mb-2">Current Members</h4>
              <div className="space-y-1 max-h-32 overflow-y-auto">
                {activeChat?.members.map((member, index) => (
                  <div key={index} className="flex items-center gap-2 p-2 bg-muted rounded text-sm">
                    <Users className="w-4 h-4 text-muted-foreground" />
                    <span>{member}</span>
                    {member === "You" && (
                      <span className="text-xs bg-blue-500 text-white px-2 py-1 rounded">You</span>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {/* Add New Member */}
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-muted-foreground mb-2">Add New Member</h4>
              <div className="space-y-3">
                <input
                  type="email"
                  value={newMemberEmail}
                  onChange={(e) => setNewMemberEmail(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleAddMember(e)}
                  placeholder="Enter email address..."
                  className="w-full p-3 rounded-lg bg-muted text-foreground border border-border placeholder-muted-foreground"
                  autoFocus
                />
                
                {/* Quick Add Suggestions */}
                <div>
                  <p className="text-xs text-muted-foreground mb-2">Quick Add:</p>
                  <div className="grid grid-cols-1 gap-1 max-h-32 overflow-y-auto">
                    {availableUsers
                      .filter(user => !activeChat?.members.includes(user))
                      .map((user) => (
                      <button
                        key={user}
                        onClick={() => setNewMemberEmail(user)}
                        className="text-left p-2 hover:bg-muted rounded text-sm border border-border"
                      >
                        {user}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowAddMembersModal(false)}
                className="px-4 py-2 text-muted-foreground hover:text-foreground"
              >
                Cancel
              </button>
              <button
                onClick={handleAddMember}
                disabled={!newMemberEmail.trim()}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:bg-gray-400 disabled:cursor-not-allowed rounded-lg text-white flex items-center gap-2"
              >
                <Users className="w-4 h-4" />
                Add Member
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}
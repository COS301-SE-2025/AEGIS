import {
  Send,
  FileText,
  Folder,
  Home,
  MessageSquare,
  
} from "lucide-react";
import { Link } from "react-router-dom";

export  const SecureChatPage = (): JSX.Element => {
  const messages = [
  { user: "IR Lead", color: "text-red-400", bg: "bg-gray-800", content: "Initial triage complete. Malware sample isolated from Host-22." },
  { user: "Threat Intel", color: "text-yellow-400", bg: "bg-gray-700", content: "YARA rule matched with UNC2452 variant. Likely APT activity." },
  { user: "You", color: "text-green-400", bg: "bg-gray-700", self: true, content: "Confirmed lateral movement from Host-22 to Host-29 via SMB." },
  { user: "Forensics", color: "text-purple-400", bg: "bg-gray-800", content: "Disk image acquisition started for Host-29. ETA: 15 minutes." },
  { user: "You", color: "text-green-400", bg: "bg-gray-700", self: true, content: "Blocking C2 domain on perimeter firewall. DNS sinkhole active." },
  { user: "Legal/Compliance", color: "text-pink-400", bg: "bg-gray-800", content: "Reminder: Preserve chain of custody logs for Host-29." },
  { user: "Malware Analyst", color: "text-blue-400", bg: "bg-gray-700", content: "Binary shows signs of process hollowing. Investigating persistence." },
  { user: "IR Lead", color: "text-red-400", bg: "bg-gray-800", content: "Prepare post-incident report template. Add TTP mapping to MITRE ATT&CK." },
  { user: "Threat Intel", color: "text-yellow-400", bg: "bg-gray-700", content: "IOC package updated. Includes hashes, domains, and mutex strings." },
  { user: "You", color: "text-green-400", bg: "bg-gray-700", self: true, content: "Uploading memory dump to sandbox for detonation and behavioral analysis." },
];

  return (
    <div className="bg-black flex flex-row justify-center w-full min-h-screen text-white">
      {/* Sidebar */}
      <div className="w-72 bg-gray-900 p-6 flex flex-col justify-between">
        <div>
          {/* Logo */}
          <div className="flex items-center gap-3 mb-12">
            <div className="w-12 h-12 rounded-lg overflow-hidden">
              <img
                src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
                alt="AEGIS Logo"
                className="w-full h-full object-cover"
              />
            </div>
            <span className="font-bold text-white text-2xl">AEGIS</span>
          </div>

          {/* Navigation */}
          <nav className="space-y-4">
        <button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-gray-800 rounded-lg">
            <Home className="w-5 h-5" />
            Dashboard
        </button>
        <Link
            to="/case-management"
            className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-gray-800 rounded-lg"
            >
            <Folder className="w-5 h-5" />
            Case Management
            </Link>
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

      {/* Chat Area */}
      <div className="flex-1 flex flex-col bg-gray-950 p-8">
        <h1 className="text-3xl font-bold mb-6">Secure Chat</h1>

        {/* Chat Box */}
        <div className="flex-1 overflow-y-auto space-y-4 mb-4 pr-2">
          {messages.map((msg, index) => (
            <div
              key={index}
              className={`${msg.self ? "self-end ml-auto" : ""} max-w-md ${msg.bg} p-4 rounded-lg`}
            >
              <p>
                <span className={`font-bold ${msg.color}`}>{msg.user}:</span>{" "}
                {msg.content}
              </p>
            </div>
          ))}
        </div>

        {/* Input Area */}
        <form className="flex items-center gap-2 mt-auto">
          <input
            type="text"
            placeholder="Type a secure message..."
            className="flex-1 p-3 rounded-lg bg-gray-800 text-white border border-gray-700 placeholder-gray-400"
          />
          <button
            type="submit"
            className="px-4 py-3 bg-blue-600 hover:bg-blue-500 rounded-lg flex items-center justify-center"
          >
            <Send className="w-5 h-5" />
          </button>
        </form>
      </div>
    </div>
  );
}

import {
  Bell,
  FileText,
  Filter,
  Paperclip,
  Folder,
  Home, 
  MessageSquare,
  Search,
  Settings,
  Share2,
  Plus
} from "lucide-react";

import { Link } from "react-router-dom";
export const CaseManagementPage = () => {
  // Timeline event data
  const timelineEvents = [
    {
      date: "2025-05-25",
      time: "23:30",
      description: "Initial Access via Phishing Email",
    },
    {
      date: "2025-05-26",
      time: "00:05",
      description: "Lateral Movement Attempt detected",
    },
    {
      date: "2025-05-26",
      time: "01:10",
      description: "System Compromise (Server B)",
    },
    {
      date: "2025-05-26",
      time: "02:45",
      description: "Data Staging Identified",
    },
    {
      date: "2025-05-26",
      time: "09:00",
      description: "Data Exfiltration Commenced",
    },
    {
      date: "2025-05-26",
      time: "09:00",
      description: "Case Initiated- Operation ShadowBroker",
    },
  ];

  // User data for assigned team
  const teamMembers = [
    { id: 1, name: "Agent Benji", role: "Lead Analyst" },
    { id: 2, name: "Agent Tshepi", role: "Security Expert" },
    { id: 3, name: "Agent Lwando", role: "Forensics Specialist" },
    { id: 3, name: "Agent Thati", role: "Network log Specialist" },
    { id: 3, name: "Agent Tshire", role: "Malware Specialist" },
  ];

  // Evidence data
  const evidenceItems = [
    { name: "System logs (Shadow.exe...)", id: 1 },
    { name: "Malware Sample", id: 2 },
  ];

  return (
    <div className="min-h-screen bg-black">
      {/* Left Sidebar - Fixed */}
      <div className="fixed left-0 top-0 h-full w-80 bg-black border-r border-gray-800 p-6 flex flex-col z-10">
        {/* Logo */}
<div className="flex items-center gap-3 mb-8">
  <div className="w-14 h-14 rounded-lg overflow-hidden">
    <img
      src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
      alt="AEGIS Logo"
      className="w-full h-full object-cover"
    />
  </div>
  <span className="font-bold text-white text-2xl">AEGIS</span>
</div>


        {/* Navigation Menu */}
        <nav className="flex-1 space-y-2">
          <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-3 rounded-lg transition-colors cursor-pointer">
            <Home className="w-6 h-6" />
           <Link to="/dashboard"> <span className="text-lg">Dashboard</span></Link>
          </div>

          <div className="flex items-center gap-3 bg-blue-600 text-white p-3 rounded-lg">
            <FileText className="w-6 h-6" />
            <span className="text-lg font-semibold">Case Management</span>
          </div>

          <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-3 rounded-lg transition-colors cursor-pointer">
            <Folder className="w-6 h-6" />
            <Link to="/evidence-viewer"><span className="text-lg">Evidence Viewer</span></Link>
          </div>

      
          <div className="flex items-center gap-3 text-gray-400 hover:text-white hover:bg-gray-800 p-3 rounded-lg transition-colors cursor-pointer">
            <MessageSquare className="w-6 h-6" />
            <span className="text-lg"><Link to="/secure-chat"> Secure Chat</Link></span>
          </div>
        </nav>

        {/* User Profile */}
        <div className="border-t border-gray-700 pt-4">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-gray-600 rounded-full flex items-center justify-center">
             <Link to="/profile" > <span className="text-white font-medium">AT</span></Link>
            </div>
            <div>
              <p className="font-semibold text-white">Agent Tshire</p>
              <p className="text-gray-400 text-sm">Status: Ongoing</p>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="ml-80 min-h-screen bg-black">
        {/* Top Navigation Bar - Fixed */}
        <div className="sticky top-0 bg-black border-b border-gray-800 p-4 z-5">
          <div className="flex items-center justify-between">
            {/* Navigation Tabs */}
            <div className="flex items-center gap-6">
              <Link to="/dashboard"> <button className="text-gray-400 hover:text-white px-4 py-2 rounded-lg transition-colors">
                Dashboard
              </button></Link>
              <Link to="/evidence-viewer"><button className="text-gray-400 hover:text-white px-4 py-2 rounded-lg transition-colors">
                Evidence Viewer
              </button></Link>
              <button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">
                Case Management
              </button>
              
              <button className="text-gray-400 hover:text-white px-4 py-2 rounded-lg transition-colors">
               <Link to="/secure-chat"> Secure Chat</Link>
              </button>
            </div>

            {/* Right Side Controls */}
            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  className="w-80 h-12 bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 text-white placeholder-gray-400 focus:outline-none focus:border-blue-500"
                  placeholder="Search cases, evidence, users"
                />
              </div>
              <button className="p-2 text-gray-400 hover:text-white transition-colors">
                <Bell className="w-6 h-6" />
              </button>
              <button className="p-2 text-gray-400 hover:text-white transition-colors">
               <Link to="/settings" > <Settings className="w-6 h-6" /></Link>
              </button>
              <div className="w-10 h-10 bg-gray-600 rounded-full flex items-center justify-center">
                <Link to="/profile" ><span className="text-white font-medium text-sm">AT</span></Link>
              </div>
            </div>
          </div>
        </div>

        {/* Page Content */}
        <div className="p-6">
          {/* Page Header */}
          <div className="flex items-center justify-between mb-8">
            <h1 className="text-3xl font-bold text-white">Case Details & Timeline</h1>
            <div className="flex gap-4">
              <button className="flex items-center gap-2 px-4 py-2 bg-gray-800 text-white border border-gray-600 rounded-lg hover:bg-gray-700 transition-colors">
                <Plus className="w-4 h-4" />
                 Create Case
              </button>
              <button className="flex items-center gap-2 px-4 py-2 bg-gray-800 text-white border border-gray-600 rounded-lg hover:bg-gray-700 transition-colors">
                <Share2 className="w-4 h-4" />
                Share Case
              </button>
              <button className="flex items-center gap-2 px-4 py-2 bg-gray-800 text-white border border-gray-600 rounded-lg hover:bg-gray-700 transition-colors">
                <Filter className="w-4 h-4" />
                Filter Timeline
              </button>

            </div>
          </div>

          {/* Main Content Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Case Details Section */}
            <div className="lg:col-span-1">
              <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 mb-6">
                {/* Case Title and Threat Level */}
                <div className="flex justify-between items-start mb-6">
                  <h2 className="text-xl font-bold text-white">
                    Operation ShadowBroker<br />Malware Analysis
                  </h2>
                  <span className="bg-red-900/30 text-red-400 border border-red-400 rounded-full px-3 py-1 text-sm">
                    High Threat
                  </span>
                </div>

                {/* Status */}
                <div className="mb-6">
                  <h3 className="text-gray-400 mb-2">Status:</h3>
                  <p className="text-white">Ongoing</p>
                </div>

                {/* Assigned Team */}
                <div className="mb-6">
                  <h3 className="text-gray-400 mb-4">Assigned Team</h3>
                  <div className="space-y-3">
                    {teamMembers.map((member) => (
                      <div key={member.id} className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center">
                          <span className="text-white text-sm font-medium">
                            {member.name.split(' ').map(n => n[0]).join('')}
                          </span>
                        </div>
                        <div>
                          <span className="text-white">{member.name}</span>
                          <span className="text-gray-400 ml-2">({member.role})</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Timestamps */}
                <div className="mb-6">
                  <h3 className="text-gray-400 mb-2">Timestamps:</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-gray-400 text-sm">Created:</p>
                      <p className="text-white">2025-03-20</p>
                      <p className="text-white">8:09:00</p>
                    </div>
                    <div>
                      <p className="text-gray-400 text-sm">Last Updated:</p>
                      <p className="text-white">2025-05-20</p>
                      <p className="text-white">7:14:30</p>
                    </div>
                  </div>
                </div>

                {/* Associated Evidence */}
                <div>
                  <h3 className="text-gray-400 mb-4">Associated Evidence:</h3>
                  <div className="space-y-3">
                    {evidenceItems.map((item) => (
                      <div key={item.id} className="flex items-center gap-3">
                        <Paperclip className="w-5 h-5 text-blue-500" />
                        <span className="text-blue-500 hover:text-blue-400 cursor-pointer">{item.name}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Investigation Timeline Section */}
            <div className="lg:col-span-2">
              <div className="bg-gray-900 border border-gray-700 rounded-lg p-6">
                <h2 className="text-2xl font-semibold text-white mb-8">Investigation Timeline</h2>
                
                <div className="relative">
                  {/* Timeline events */}
                  {timelineEvents.map((event, index) => (
                    <div key={index} className="flex items-start mb-8 relative">
                      {/* Timeline line */}
                      {index < timelineEvents.length - 1 && (
                        <div className="absolute left-20 top-10 w-0.5 h-16 bg-gray-600"></div>
                      )}
                      
                      {/* Date and time */}
                      <div className="w-32 text-right pr-4">
                        <div className="text-gray-400 text-sm">
                          {event.date}
                        </div>
                        <div className="text-gray-400 text-sm">
                          {event.time}
                        </div>
                      </div>

                      {/* Timeline marker */}
                      <div className="w-8 h-8 bg-blue-600 rounded-full border-4 border-black flex items-center justify-center relative z-10">
                        <div className="w-2 h-2 bg-white rounded-full"></div>
                      </div>

                      {/* Event description */}
                      <div className="flex-1 ml-4">
                        <div className="bg-gray-800 border border-gray-600 rounded-lg p-4">
                          <p className="text-white">{event.description}</p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};


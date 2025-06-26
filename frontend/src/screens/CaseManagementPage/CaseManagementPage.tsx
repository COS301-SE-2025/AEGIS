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
  Plus,
  Calendar,  
  Clock
} from "lucide-react";
import { useState, useEffect  } from 'react';
import { Link } from "react-router-dom";
//thati added
import { SidebarToggleButton } from '../../context/SidebarToggleContext';
import {ShareButton} from "../ShareCasePage/sharecasebutton";
//
import { useParams } from 'react-router-dom';

export const CaseManagementPage = () => {
const storedUser = sessionStorage.getItem("user");
  const user = storedUser ? JSON.parse(storedUser) : null;
  const displayName = user?.name || user?.email?.split("@")[0] || "Agent User";
  const initials = displayName
    .split(" ")
    .map((part: string) => part[0])
    .join("")
    .toUpperCase();

const userRole = "admin"; // for now
const caseName = "Malware"; 
const { caseId } = useParams();
if (!caseId) {
  return <div className="p-6 text-red-500">No case selected.</div>;
}

<SidebarToggleButton />

  // Timeline event data
  const [timelineEvents, setTimelineEvents] = useState<{ date: string; time: string; description: string }[]>([]);
  const [hasLoaded, setHasLoaded] = useState(false);

// 1. Load timeline data when component mounts or caseId changes
useEffect(() => {
  if (caseId) {
    const saved = localStorage.getItem(`timeline-${caseId}`);
    if (saved) {
      try {
        const parsedEvents = JSON.parse(saved);
        setTimelineEvents(parsedEvents);
        console.log(`Loaded ${parsedEvents.length} events for case ${caseId}`);
      } catch (error) {
        console.warn(`Failed to parse timeline for case ${caseId}:`, error);
        setTimelineEvents([]);
      }
    } else {
      setTimelineEvents([]);
    }

    setHasLoaded(true); //  Only set after loading finishes
  }
}, [caseId]);


// 2. Save timeline data whenever it changes (but only if we have a valid caseId)
useEffect(() => {
  if (caseId && hasLoaded) {
    try {
      localStorage.setItem(`timeline-${caseId}`, JSON.stringify(timelineEvents));
      console.log(`Saved ${timelineEvents.length} events for case ${caseId}`);
    } catch (error) {
      console.error(`Failed to save timeline for case ${caseId}:`, error);
    }
  }
}, [timelineEvents, caseId, hasLoaded]);

  
  
  const [newEventDescription, setNewEventDescription] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);

  //state declaration for filtering the timeline
  const [filterKeyword, setFilterKeyword] = useState('');
  const [showFilterInput, setShowFilterInput] = useState(false);

  const [filterDate, setFilterDate] = useState('');


  // ADD THESE NEW FUNCTIONS
  const getCurrentTimestamp = () => {
    const now = new Date();
    const date = now.toISOString().split('T')[0]; // YYYY-MM-DD format
    const time = now.toTimeString().slice(0, 5); // HH:MM format
    return { date, time };
  };

    const addEvent = () => {
      if (newEventDescription.trim()) {
        const { date, time } = getCurrentTimestamp();
        const newEvent = { date, time, description: newEventDescription.trim() };
        
        setTimelineEvents(prevEvents => [...prevEvents, newEvent]); //  use previous events
        setNewEventDescription('');
        setShowAddForm(false);
        console.log("Adding event:", newEvent);
console.log("Current caseId:", caseId);

      }
      
    };


    const deleteEvent = (index: number) => {
    const updatedEvents = [...timelineEvents];
    updatedEvents.splice(index, 1);
    setTimelineEvents(updatedEvents);
  };


  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      addEvent();
    }
  };


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
    { name: "screenshot_evidence", id: 3 },

  ];

  return (
    <div className="min-h-screen bg-background">
      {/* Left Sidebar - Fixed */}
      <div className="fixed left-0 top-0 h-full w-80 bg-background border-r border p-6 flex flex-col z-10">
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


        {/* Navigation Menu */}
        <nav className="flex-1 space-y-2">
          <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors cursor-pointer">
            <Home className="w-6 h-6" />
           <Link to="/dashboard"> <span className="text-lg">Dashboard</span></Link>
          </div>

          <div className="flex items-center gap-3 bg-blue-600 text-white p-3 rounded-lg">
            <FileText className="w-6 h-6" />
            <span className="text-lg font">Case Management</span>
          </div>

          <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors cursor-pointer">
            <Folder className="w-6 h-6" />
            <Link to="/evidence-viewer"><span className="text-lg">Evidence Viewer</span></Link>
          </div>

      
          <div className="flex items-center gap-3 text-muted-foreground hover:text-foreground hover:bg-muted p-3 rounded-lg transition-colors cursor-pointer">
            <MessageSquare className="w-6 h-6" />
            <span className="text-lg"><Link to="/secure-chat"> Secure Chat</Link></span>
          </div>
        </nav>

        {/* User Profile */}
        <div className="border-t border-bg-accent pt-4">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-muted rounded-full flex items-center justify-center">
              <Link to="/profile">
                <span className="text-foreground font-medium">{initials}</span>
              </Link>
            </div>
            <div>
              <p className="font-semibold text-foreground">{displayName}</p>
              <p className="text-muted-foreground text-sm">{user?.email || "user@dfir.com"}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="ml-80 min-h-screen bg-background">
        {/* Top Navigation Bar - Fixed */}
        <div className="sticky top-0 bg-background border-b border p-4 z-5">
          <div className="flex items-center justify-between">
            {/* Navigation Tabs */}
            <div className="flex items-center gap-6">
              <SidebarToggleButton/>
              <Link to="/dashboard"> <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Dashboard
              </button></Link>
               <button className="text-blue-500 bg-blue-500/10 px-4 py-2 rounded-lg">
                Case Management
              </button>
              <Link to="/evidence-viewer"><button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
                Evidence Viewer
              </button></Link>
              <button className="text-muted-foreground hover:text-foreground px-4 py-2 rounded-lg transition-colors">
               <Link to="/secure-chat"> Secure Chat</Link>
              </button>
            </div>

            {/* Right Side Controls */}
             <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <input
                  className="w-80 h-12 bg-popover border rounded-lg pl-10 pr-4 text-foreground placeholder-muted-foreground focus:outline-none focus:border-blue-500"
                  placeholder="Search cases, evidence, users"
                />
              </div>
              <button className="p-2 text-muted-foreground hover:text-foreground transition-colors">
                <Bell className="w-6 h-6" />
              </button>
              <button className="p-2 text-muted-foreground hover:text-foreground transition-colors">
               <Link to="/settings" > <Settings className="w-6 h-6" /></Link>
              </button>
              <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center">
                <Link to="/profile" ><span className="text-foreground font-medium text-sm">{initials}</span></Link>
              </div>
            </div>
          </div>
        </div>

        {/* Page Content */}
        <div className="p-6">
          {/* Page Header */}
          <div className="flex items-center justify-between mb-8">
            <h1 className="text-3xl font-bold text-foreground">Case Details & Timeline</h1>
            <div className="flex gap-4">
              <Link to="/create-case"><button className="flex items-center gap-2 px-4 py-2 bg-popover border rounded-lg pl-10 pr-4 text-foreground placeholder-muted-foreground focus:outline-none focus:border-blue-500">
                <Plus className="w-4 h-4" />
                 Create Case
              </button></Link>
              <button className="flex items-center gap-2 px-4 py-2 bg-popover border rounded-lg pl-10 pr-4 text-foreground placeholder-muted-foreground focus:outline-none focus:border-blue-500">
                <Share2 className="w-4 h-4" />
                  {userRole === "admin" && (
                  <ShareButton caseId={caseId} caseName={caseName} />
                )}
              </button>
              <button
                onClick={() => setShowFilterInput(!showFilterInput)}
                className="flex items-center gap-2 px-4 py-2 bg-popover border rounded-lg pl-10 pr-4 text-foreground placeholder-muted-foreground focus:outline-none focus:border-blue-500"
              >
                <Filter className="w-4 h-4" />
                Filter Timeline
              </button>

          {showFilterInput && (
            <div className="mt-4 mb-6 flex flex-col gap-2 md:flex-row">
              {/* Keyword Input */}
              <input
                type="text"
                placeholder="Filter by keyword..."
                value={filterKeyword}
                onChange={(e) => setFilterKeyword(e.target.value)}
                className="flex-1 px-3 py-2 border border-gray-300 text-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />

              {/* Date Input */}
              <input
                type="date"
                value={filterDate}
                onChange={(e) => setFilterDate(e.target.value)}
                className="px-3 py-2 border border-gray-300 text-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />

              {/* Clear Button */}
              <button
                onClick={() => {
                  setFilterKeyword('');
                  setFilterDate('');
                }}
                className="px-4 py-2 bg-gray-500 text-foreground rounded-md hover:bg-gray-600 transition-colors"
              >
                Clear
              </button>
            </div>
          )}



            </div>
          </div>

          {/* Main Content Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Case Details Section */}
            <div className="lg:col-span-1">
              <div className="bg-card border border-bg-accent rounded-lg p-6 mb-6">
                {/* Case Title and Threat Level */}
                <div className="flex justify-between items-start mb-6">
                  <h2 className="text-xl font-bold text-foreground">
                    <br />Malware
                  </h2>
                  <span className="bg-red-900/30 text-red-400 border border-red-400 rounded-full px-3 py-1 text-sm">
                    Critical
                  </span>
                </div>

                {/* Status */}
                <div className="mb-6">
                  <h3 className="text-muted-foreground mb-2">Status:</h3>
                  <p className="text-foreground">Ongoing</p>
                </div>

                {/* Assigned Team */}
                <div className="mb-6">
                  <h3 className="text-muted-foreground mb-4">Assigned Team</h3>
                  <div className="space-y-3">
                    {teamMembers.map((member) => (
                      <div key={member.id} className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-muted rounded-full flex items-center justify-center">
                          <span className="text-foreground text-sm font-medium">
                            {member.name.split(' ').map(n => n[0]).join('')}
                          </span>
                        </div>
                        <div>
                          <span className="text-foreground">{member.name}</span>
                          <span className="text-muted-foreground ml-2">({member.role})</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Timestamps */}
                <div className="mb-6">
                  <h3 className="text-muted-foreground mb-2">Timestamps:</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-muted-foreground text-sm">Created:</p>
                      <p className="text-foreground">2025-03-20</p>
                      <p className="text-foreground">8:09:00</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground text-sm">Last Updated:</p>
                      <p className="text-foreground">2025-05-20</p>
                      <p className="text-foreground">7:14:30</p>
                    </div>
                  </div>
                </div>

                {/* Associated Evidence */}
                <div>
              <Link to="/evidence-viewer" className="block">
                <h3 className="text-muted-foreground mb-4 hover:text-gray-300 cursor-pointer transition-colors">
                  Associated Evidence:
                </h3>
              </Link>
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
              <div className="bg-card border border-bg-accent rounded-lg p-6">
                <div className="flex items-center justify-between mb-8">
                  <h2 className="text-2xl font-semibold text-foreground">Investigation Timeline</h2>
                  <button
                    onClick={() => setShowAddForm(!showAddForm)}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-foreground rounded-lg hover:bg-blue-700 transition-colors"
                  >
                    <Plus size={18} />
                    Add Event
                  </button>
                </div>

                {/* Add Event Form */}
                {showAddForm && (
                  <div className="mb-8 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                    <div className="flex items-center gap-2 mb-3">
                      <Calendar size={16} className="text-blue-600" />
                      <span className="text-sm text-blue-800">
                        Will be timestamped: {getCurrentTimestamp().date} at {getCurrentTimestamp().time}
                      </span>
                    </div>
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={newEventDescription}
                        onChange={(e) => setNewEventDescription(e.target.value)}
                        onKeyPress={handleKeyPress}
                        placeholder="Enter event description..."
                        className="flex-1 px-3 py-2 border border-gray-300 text-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        autoFocus
                      />
                      <button
                        onClick={addEvent}
                        disabled={!newEventDescription.trim()}
                        className="px-4 py-2 bg-green-600 text-foreground rounded-md hover:bg-green-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
                      >
                        Add
                      </button>
                      <button
                        onClick={() => {
                          setShowAddForm(false);
                          setNewEventDescription('');
                        }}
                        className="px-4 py-2 bg-gray-500 text-foreground rounded-md hover:bg-gray-600 transition-colors"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                )}
                
                <div className="relative">
                  {/* Timeline events */}
                  {timelineEvents  .filter(event => {
                    const matchesKeyword = event.description.toLowerCase().includes(filterKeyword.toLowerCase());
                    const matchesDate = !filterDate || event.date === filterDate;
                    return matchesKeyword && matchesDate;
                  })
                  .map((event, index) => (
                    <div key={index} className="flex items-start mb-8 relative">
                      {/* Timeline line */}
                      {index < timelineEvents.length - 1 && (
                        <div className="absolute left-20 top-10 w-0.5 h-16 bg-muted"></div>
                      )}
                      
                      {/* Date and time */}
                      <div className="w-32 text-right pr-4">
                        <div className="text-muted-foreground text-sm flex items-center justify-end gap-1">
                          <Calendar size={12} />
                          {event.date}
                        </div>
                        <div className="text-muted-foreground text-sm flex items-center justify-end gap-1">
                          <Clock size={12} />
                          {event.time}
                        </div>
                      </div>

                      {/* Timeline marker */}
                      <div className="w-8 h-8 bg-blue-600 rounded-full border-4 border-background flex items-center justify-center relative z-10">
                        <div className="w-2 h-2 bg-white rounded-full"></div>
                      </div>

                      {/* Event description */}
                      <div className="flex-1 ml-4">
                        <div className="bg-muted border border rounded-lg p-4 flex justify-between items-center">
                          <p className="text-foreground">{event.description}</p>
                          <button
                            onClick={() => deleteEvent(index)}
                            className="ml-4 px-2 py-1 text-xs text-red-600 hover:underline"
                          >
                            Delete
                          </button>
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


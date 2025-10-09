import React, { useState } from "react";
import {
  Shield,
  Lock,
  Eye,
  CheckCircle,
  Search,
  Users,
  FileText,
  MessageSquare,
  Database,
  Link,
  HardDrive,
  Network,
  Calendar,
  Cpu,
  BarChart2,
  X,
} from "lucide-react";
import { HelpMenu } from "../../components/ui/HelpMenu";

export const LandingPage: React.FC = () => {
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<string[]>([]);

  // Searchable content from the page
  const searchableContent = [
    { title: "Case Management", content: "Create and track cases with unique IDs, assign roles, and build visual timelines" },
    { title: "Real-time Collaboration", content: "Multiple users can work on cases simultaneously with real-time commenting" },
    { title: "Secure Communication", content: "End-to-end encrypted communication, secure file sharing" },
    { title: "Chain of Custody", content: "Maintain immutable chain of custody with automated logging" },
    { title: "Multi-Format Evidence", content: "Support for logs, images, packet captures, and disk images" },
    { title: "Access Controls", content: "Role-based access control with customizable permissions" },
    { title: "AI-Powered Analysis", content: "Automated metadata extraction, pattern recognition" },
    { title: "Relationship Mapping", content: "Interactive graph-based visualization of relationships" },
    { title: "Visual Timelines", content: "Generate comprehensive event timelines and sequence charts" },
    { title: "AES-256 Encryption", content: "Data encrypted at rest and in transit" },
    { title: "Audit Logging", content: "Comprehensive activity tracking" },
    { title: "Legal Compliance", content: "Maintains chain of custody for legal requirements" },
  ];

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    if (query.trim() === "") {
      setSearchResults([]);
      return;
    }

    const results = searchableContent
      .filter(item => 
        item.title.toLowerCase().includes(query.toLowerCase()) ||
        item.content.toLowerCase().includes(query.toLowerCase())
      )
      .map(item => item.title);
    
    setSearchResults(results);
  };

  const openSearch = () => {
    setIsSearchOpen(true);
  };

  const closeSearch = () => {
    setIsSearchOpen(false);
    setSearchQuery("");
    setSearchResults([]);
  };

  const scrollToSection = (sectionTitle: string) => {
    // Simple scroll to section based on title
    const element = document.querySelector(`h3:contains("${sectionTitle}")`);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
    closeSearch();
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      
      {/* Navigation */}
      <nav className="flex items-center justify-between px-6 py-4 bg-gray-900/95 backdrop-blur-sm">
        <div className="flex items-center space-x-8">
          <div className="flex items-center space-x-2">
            <img
              src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
              alt="AEGIS Logo"
              className="h-11 w-11 object-cover"
            />
            <span className="text-xl font-bold">AEGIS</span>
          </div>

          <div className="hidden md:flex space-x-6 text-sm">
            <a href="#" className="hover:text-blue-400">Demo</a>
            <a href="#" className="hover:text-blue-400">Company</a>
          </div>
        </div>
        <div className="flex items-center space-x-4">
          <button 
            onClick={openSearch}
            className="hover:text-blue-400 transition-colors"
            aria-label="Search"
          >
            <Search className="h-5 w-5 text-gray-400 hover:text-blue-400" />
          </button>
          <button className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg text-sm font-medium">
            <a href="/login">Get Started</a>
          </button>
        </div>
      </nav>

      {/* Search Modal */}
      {isSearchOpen && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-start justify-center pt-20">
          <div className="bg-gray-800 rounded-lg p-6 w-full max-w-2xl mx-4 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Search AEGIS Features</h3>
              <button
                onClick={closeSearch}
                className="text-gray-400 hover:text-white"
                aria-label="Close search"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            
            <div className="relative mb-4">
              <Search className="absolute left-3 top-3 h-5 w-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search for features, capabilities, or security..."
                value={searchQuery}
                onChange={(e) => handleSearch(e.target.value)}
                className="w-full pl-10 pr-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:border-blue-500"
                autoFocus
              />
            </div>

            {/* Search Results */}
            {searchQuery && (
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {searchResults.length > 0 ? (
                  searchResults.map((result, index) => (
                    <button
                      key={index}
                      onClick={() => scrollToSection(result)}
                      className="w-full text-left p-3 rounded hover:bg-gray-700 transition-colors"
                    >
                      <div className="font-medium text-blue-400">{result}</div>
                      <div className="text-sm text-gray-300 mt-1">
                        {searchableContent.find(item => item.title === result)?.content}
                      </div>
                    </button>
                  ))
                ) : (
                  <div className="text-gray-400 text-center py-8">
                    No results found for "{searchQuery}"
                  </div>
                )}
              </div>
            )}

            {/* Quick suggestions when no query */}
            {!searchQuery && (
              <div className="space-y-2">
                <div className="text-sm text-gray-400 mb-3">Popular searches:</div>
                {["Security", "Case Management", "Encryption", "AI Analysis"].map((suggestion) => (
                  <button
                    key={suggestion}
                    onClick={() => handleSearch(suggestion)}
                    className="inline-block mr-2 mb-2 px-3 py-1 bg-gray-700 rounded-full text-sm hover:bg-gray-600 transition-colors"
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
      
      {/* Hero Section with Enhanced Animated Background */}
      <section className="relative px-6 py-20 text-center overflow-hidden">
        {/* Multi-layer animated background */}
        <div className="absolute inset-0 bg-gradient-to-br from-purple-900/30 via-blue-900/30 to-pink-900/30"></div>
        
        {/* Floating particles with different sizes and speeds */}
        <div className="absolute inset-0">
          {/* Large floating orbs */}
          <div className="absolute top-20 left-20 w-4 h-4 bg-blue-400/60 rounded-full animate-bounce" style={{animationDelay: '0s', animationDuration: '3s'}}></div>
          <div className="absolute top-40 right-32 w-3 h-3 bg-purple-400/50 rounded-full animate-bounce" style={{animationDelay: '1s', animationDuration: '4s'}}></div>
          <div className="absolute bottom-40 left-40 w-5 h-5 bg-pink-400/70 rounded-full animate-bounce" style={{animationDelay: '2s', animationDuration: '2.5s'}}></div>
          <div className="absolute top-60 right-20 w-2 h-2 bg-blue-300/80 rounded-full animate-bounce" style={{animationDelay: '0.5s', animationDuration: '3.5s'}}></div>
          <div className="absolute bottom-20 right-40 w-3 h-3 bg-green-400/60 rounded-full animate-bounce" style={{animationDelay: '1.5s', animationDuration: '4.5s'}}></div>
          <div className="absolute top-32 left-60 w-2 h-2 bg-yellow-400/50 rounded-full animate-bounce" style={{animationDelay: '2.5s', animationDuration: '3s'}}></div>
          
          {/* Pulsing network nodes */}
          <div className="absolute top-1/4 left-1/4 w-6 h-6 bg-blue-500/30 rounded-full animate-pulse border-2 border-blue-400/50"></div>
          <div className="absolute top-3/4 right-1/4 w-8 h-8 bg-purple-500/20 rounded-full animate-pulse border-2 border-purple-400/40" style={{animationDelay: '1s'}}></div>
          <div className="absolute bottom-1/3 left-1/3 w-4 h-4 bg-pink-500/40 rounded-full animate-pulse border-2 border-pink-400/60" style={{animationDelay: '2s'}}></div>
        </div>

        {/* Animated connection lines */}
        <svg className="absolute inset-0 w-full h-full opacity-30">
          <g>
            {/* Animated lines with gradient strokes */}
            <line x1="10%" y1="20%" x2="80%" y2="30%" stroke="url(#gradient1)" strokeWidth="2">
              <animate attributeName="stroke-dasharray" values="0,100;50,50;100,0;0,100" dur="4s" repeatCount="indefinite"/>
            </line>
            <line x1="20%" y1="60%" x2="70%" y2="20%" stroke="url(#gradient2)" strokeWidth="2">
              <animate attributeName="stroke-dasharray" values="100,0;50,50;0,100;100,0" dur="3s" repeatCount="indefinite"/>
            </line>
            <line x1="80%" y1="70%" x2="30%" y2="40%" stroke="url(#gradient3)" strokeWidth="2">
              <animate attributeName="stroke-dasharray" values="0,100;25,75;75,25;100,0;0,100" dur="5s" repeatCount="indefinite"/>
            </line>
            <line x1="15%" y1="80%" x2="85%" y2="15%" stroke="url(#gradient4)" strokeWidth="1">
              <animate attributeName="stroke-dasharray" values="50,50;100,0;0,100;50,50" dur="3.5s" repeatCount="indefinite"/>
            </line>
            <line x1="70%" y1="60%" x2="25%" y2="80%" stroke="url(#gradient5)" strokeWidth="1">
              <animate attributeName="stroke-dasharray" values="0,100;100,0;0,100" dur="4.5s" repeatCount="indefinite"/>
            </line>
          </g>
          <defs>
            <linearGradient id="gradient1" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#3B82F6" stopOpacity="0" />
              <stop offset="50%" stopColor="#3B82F6" stopOpacity="0.8" />
              <stop offset="100%" stopColor="#3B82F6" stopOpacity="0" />
            </linearGradient>
            <linearGradient id="gradient2" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#8B5CF6" stopOpacity="0" />
              <stop offset="50%" stopColor="#8B5CF6" stopOpacity="0.8" />
              <stop offset="100%" stopColor="#8B5CF6" stopOpacity="0" />
            </linearGradient>
            <linearGradient id="gradient3" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#EC4899" stopOpacity="0" />
              <stop offset="50%" stopColor="#EC4899" stopOpacity="0.8" />
              <stop offset="100%" stopColor="#EC4899" stopOpacity="0" />
            </linearGradient>
            <linearGradient id="gradient4" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#10B981" stopOpacity="0" />
              <stop offset="50%" stopColor="#10B981" stopOpacity="0.6" />
              <stop offset="100%" stopColor="#10B981" stopOpacity="0" />
            </linearGradient>
            <linearGradient id="gradient5" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#F59E0B" stopOpacity="0" />
              <stop offset="50%" stopColor="#F59E0B" stopOpacity="0.6" />
              <stop offset="100%" stopColor="#F59E0B" stopOpacity="0" />
            </linearGradient>
          </defs>
        </svg>

        {/* Floating geometric shapes */}
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute top-1/3 left-10 w-0 h-0 border-l-[10px] border-r-[10px] border-b-[15px] border-l-transparent border-r-transparent border-b-blue-400/40 animate-spin" style={{animationDuration: '8s'}}></div>
          <div className="absolute bottom-1/3 right-10 w-4 h-4 border-2 border-purple-400/50 rotate-45 animate-spin" style={{animationDuration: '6s', animationDirection: 'reverse'}}></div>
          <div className="absolute top-1/2 right-1/4 w-3 h-3 bg-pink-400/30 transform rotate-45 animate-pulse"></div>
        </div>

        <div className="relative z-10 max-w-4xl mx-auto">
          <h1 className="text-5xl md:text-6xl font-bold mb-6">
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">
            Unrivaled Security for a<br />
            Connected World
            </span>
          </h1>
          <p className="text-xl text-gray-300 mb-8 max-w-2xl mx-auto">
            AEGIS provides advanced endpoint cybersecurity solutions to protect your
            enterprise from the most sophisticated threats.
          </p>
          <button className="bg-blue-600 hover:bg-blue-700 px-8 py-3 rounded-lg text-lg font-medium transition-all transform hover:scale-105">
            Request a Demo
          </button>
        </div>
      </section>

{/* Core Capabilities */}
      <section id="features" className="px-6 py-20 ">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            Core Capabilities for Digital Forensics Teams
          </h2>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-blue-500 transition-all">
              <FileText className="h-12 w-12 text-blue-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Case Management</h3>
              <p className="text-gray-300 text-sm">
                Create and track cases with unique IDs, assign roles, and build visual timelines to map incidents and correlate related events across investigations.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-purple-500 transition-all">
              <Users className="h-12 w-12 text-purple-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Real-time Collaboration</h3>
              <p className="text-gray-300 text-sm">
                Multiple users can work on cases simultaneously with real-time commenting, threaded discussions, and annotations directly on evidence pieces.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-green-500 transition-all">
              <Lock className="h-12 w-12 text-green-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Secure Communication</h3>
              <p className="text-gray-300 text-sm">
                End-to-end encrypted communication, secure file sharing, and encryption at rest for all stored evidence and case data.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-red-500 transition-all">
              <Shield className="h-12 w-12 text-red-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Chain of Custody</h3>
              <p className="text-gray-300 text-sm">
                Maintain immutable chain of custody with automated logging of all evidence modifications, ensuring legal compliance and audit trails.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-yellow-500 transition-all">
              <Database className="h-12 w-12 text-yellow-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Multi-Format Evidence</h3>
              <p className="text-gray-300 text-sm">
                Support for various evidence formats including logs, images, packet captures, and disk images with visualization tools for easy interpretation.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-pink-500 transition-all">
              <Eye className="h-12 w-12 text-pink-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Access Controls</h3>
              <p className="text-gray-300 text-sm">
                Role-based access control with customizable permissions and temporary access tokens for external collaborators on specific cases.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* DFIR Challenges & Solutions */}
      <section className="px-6 py-20">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            Solving Critical DFIR Challenges
          </h2>
          <div className="grid md:grid-cols-2 gap-12 items-center">
            <div>
              <h3 className="text-2xl font-semibold mb-8 text-red-400">The Challenges DFIR Teams Face:</h3>
              <div className="space-y-4">
                <div className="flex items-start space-x-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  <p className="text-gray-300">Distributed teams struggle with real-time communication and secure evidence sharing.</p>
                </div>
                <div className="flex items-start space-x-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  <p className="text-gray-300">Vast amounts of investigation data make evidence correlation and pattern detection difficult.</p>
                </div>
                <div className="flex items-start space-x-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  <p className="text-gray-300">Maintaining digital evidence integrity and chain of custody in collaborative environments.</p>
                </div>
                <div className="flex items-start space-x-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  <p className="text-gray-300">Delays and inefficiencies in investigation workflows impact critical incident response.</p>
                </div>
              </div>
            </div>
            <div className="bg-gradient-to-br from-red-900/20 to-orange-900/20 p-8 rounded-xl">
              <div className="space-y-4">
                <div className="flex items-center space-x-2">
                  <MessageSquare className="h-6 w-6 text-red-400" />
                  <div className="h-4 bg-red-500/60 rounded flex-1"></div>
                </div>
                <div className="flex items-center space-x-2">
                  <Network className="h-6 w-6 text-orange-400" />
                  <div className="h-4 bg-orange-400/40 rounded w-3/4"></div>
                </div>
                <div className="flex items-center space-x-2">
                  <HardDrive className="h-6 w-6 text-red-400" />
                  <div className="h-4 bg-red-600/70 rounded w-5/6"></div>
                </div>
                <div className="flex items-center space-x-2">
                  <Calendar className="h-6 w-6 text-yellow-400" />
                  <div className="h-4 bg-yellow-500/50 rounded w-1/2"></div>
                </div>
              </div>
            </div>
          </div>
          
          <div className="mt-16">
            <h3 className="text-2xl font-semibold mb-8 text-blue-400">How AEGIS Solves These Problems:</h3>
            <div className="grid md:grid-cols-2 gap-6">
              <div className="flex items-start space-x-3">
                <CheckCircle className="h-6 w-6 text-green-400 mt-1" />
                <p className="text-gray-300">Secure, real-time collaboration platform with encrypted communication and file sharing capabilities.</p>
              </div>
              <div className="flex items-start space-x-3">
                <CheckCircle className="h-6 w-6 text-green-400 mt-1" />
                <p className="text-gray-300">Automated data correlation and AI-powered pattern recognition to detect attack patterns and connections.</p>
              </div>
              <div className="flex items-start space-x-3">
                <CheckCircle className="h-6 w-6 text-green-400 mt-1" />
                <p className="text-gray-300">Immutable chain of custody with comprehensive audit logging and automated evidence tracking.</p>
              </div>
              <div className="flex items-start space-x-3">
                <CheckCircle className="h-6 w-6 text-green-400 mt-1" />
                <p className="text-gray-300">Streamlined investigation workflows with visual timelines, case management, and structured reporting.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

       {/* Advanced Features */}
      <section className="px-6 py-20">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            Advanced Features & AI Integration
          </h2>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-cyan-500 transition-all">
              <Cpu className="h-12 w-12 text-cyan-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">AI-Powered Analysis</h3>
              <p className="text-gray-300 text-sm">
                Automated metadata extraction, pattern recognition, and threat intelligence integration to enhance investigation efficiency.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-indigo-500 transition-all">
              <Link className="h-12 w-12 text-indigo-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Relationship Mapping</h3>
              <p className="text-gray-300 text-sm">
                Interactive graph-based visualization of relationships between evidence, including recurring IPs, tools used, and targets.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-teal-500 transition-all">
              <BarChart2 className="h-12 w-12 text-teal-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Visual Timelines</h3>
              <p className="text-gray-300 text-sm">
                Generate comprehensive event timelines and sequence charts for clearer incident reconstruction and reporting.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Interface Screenshots */}
      <section className="px-6 py-20 ">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            Intuitive Interface, Powerful Insights
          </h2>
          <div className="grid grid-cols-3 gap-6">
          {[
            "/Create_Case.png",
            "/Dashboard.png",
            "/Cases.png",
            "/Case_Management.png",
            "/Evidence_Viewer.png",
            "/Secure_Chat.png",
          ].map((src, i) => (
            <div
              key={i}
              className="bg-gray-900 rounded-lg overflow-hidden border border-gray-700 hover:border-blue-500 transition-all"
            >
              <div className="aspect-video bg-black flex items-center justify-center">
                <img
                  src={`${src}?w=400&h=225&c=7&r=0&o=5&dpr=1.3&pid=1.7`}
                  alt={`Screenshot ${i + 1}`}
                  className="object-cover w-full h-full"
                />
              </div>
            </div>
          ))}
        </div>

        </div>
      </section>

       {/* Security & Compliance */}
      <section className="px-6 py-20 ">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            Enterprise-Grade Security & Compliance
          </h2>
          <div className="grid md:grid-cols-4 gap-6">
            <div className="text-center">
              <div className="w-16 h-16 bg-blue-500/20 rounded-full flex items-center justify-center mx-auto mb-4">
                <Lock className="h-8 w-8 text-blue-400" />
              </div>
              <h3 className="font-semibold mb-2">AES-256 Encryption</h3>
              <p className="text-sm text-gray-400">Data encrypted at rest and in transit</p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-purple-500/20 rounded-full flex items-center justify-center mx-auto mb-4">
                <Shield className="h-8 w-8 text-purple-400" />
              </div>
              <h3 className="font-semibold mb-2">Role-Based Access</h3>
              <p className="text-sm text-gray-400">Granular permission controls</p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-4">
                <FileText className="h-8 w-8 text-green-400" />
              </div>
              <h3 className="font-semibold mb-2">Audit Logging</h3>
              <p className="text-sm text-gray-400">Comprehensive activity tracking</p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-red-500/20 rounded-full flex items-center justify-center mx-auto mb-4">
                <Database className="h-8 w-8 text-red-400" />
              </div>
              <h3 className="font-semibold mb-2">Legal Compliance</h3>
              <p className="text-sm text-gray-400">Maintains chain of custody</p>
            </div>
          </div>
        </div>
      </section>

      {/* Capabilities Section */}
      <section className="px-6 py-20 ">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            Explore AEGIS Capabilities
          </h2>
        
          <div className="text-center max-w-4xl mx-auto space-y-6">
            <p className="text-gray-300">
              Our advanced cybersecurity solutions are built to adapt to emerging threats,
              ensuring comprehensive coverage and seamless integration with your existing infrastructure.
            </p>
            <p className="text-gray-300">
              Get deeper insights with industry-leading threat intelligence and predictive analytics,
              detecting threats early and enabling automated security measures to protect your organization.
            </p>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="px-6 py-20 bg-gradient-to-r from-blue-600 to-purple-600">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-4xl font-bold mb-6">
            Ready to Transform Your Digital Forensics Workflow?
          </h2>
          <p className="text-xl mb-8 opacity-90">
            Join leading DFIR teams who trust AEGIS to secure their investigations, accelerate their analysis, and maintain evidence integrity.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <button className="bg-white text-blue-700 hover:bg-gray-100 px-8 py-3 rounded-lg text-lg font-medium transition-all transform hover:scale-105">
              Request a Demo
            </button>
            
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="px-6 py-8 bg-gray-900 border-t border-gray-800">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
         
          <div className="text-sm text-gray-500">
            © 2025 AEGIS Security Solutions. All rights reserved.
          </div>
        </div>
      </footer>


      <HelpMenu />
    </div>
  );
};
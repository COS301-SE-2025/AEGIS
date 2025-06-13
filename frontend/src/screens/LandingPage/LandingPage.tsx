
import React from "react";
import {
  Shield,
  Lock,
  Eye,
  Zap,
  CheckCircle,
  Play,
  Search,
  User,
} from "lucide-react";
import { HelpMenu } from "../../components/ui/HelpMenu";

export const LandingPage: React.FC = () => {
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
            {/*<a href="#" className="hover:text-blue-400">Products</a>*/}
            {/*<a href="#" className="hover:text-blue-400">Solutions</a>*/}
            <a href="#" className="hover:text-blue-400">Demo</a>
            <a href="#" className="hover:text-blue-400">Company</a>
            {/*<a href="#" className="hover:text-blue-400">Help</a>*/}
          </div>
        </div>
        <div className="flex items-center space-x-4">
          <Search className="h-5 w-5 text-gray-400" />
          <button className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg text-sm font-medium">
            <a href="/login">Get Started</a>
          </button>
          <User className="h-5 w-5 text-gray-400" />
        </div>
      </nav>
      {/* Hero Section */}
      <section className="relative px-6 py-20 text-center">
        <div className="absolute inset-0 bg-gradient-to-br from-purple-900/20 via-blue-900/20 to-pink-900/20"></div>
        <div className="absolute inset-0">
          {/* Animated network background */}
          <div className="absolute top-20 left-20 w-2 h-2 bg-blue-400 rounded-full opacity-60 animate-pulse"></div>
          <div className="absolute top-40 right-32 w-1 h-1 bg-purple-400 rounded-full opacity-40 animate-pulse"></div>
          <div className="absolute bottom-40 left-40 w-3 h-3 bg-pink-400 rounded-full opacity-50 animate-pulse"></div>
          <div className="absolute top-60 right-20 w-2 h-2 bg-blue-300 rounded-full opacity-70 animate-pulse"></div>
          {/* Connection lines */}
          <svg className="absolute inset-0 w-full h-full opacity-20">
            <line x1="10%" y1="20%" x2="80%" y2="30%" stroke="url(#gradient1)" strokeWidth="1" />
            <line x1="20%" y1="60%" x2="70%" y2="20%" stroke="url(#gradient2)" strokeWidth="1" />
            <line x1="80%" y1="70%" x2="30%" y2="40%" stroke="url(#gradient3)" strokeWidth="1" />
            <defs>
              <linearGradient id="gradient1" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#3B82F6" stopOpacity="0" />
                <stop offset="50%" stopColor="#3B82F6" stopOpacity="0.5" />
                <stop offset="100%" stopColor="#3B82F6" stopOpacity="0" />
              </linearGradient>
              <linearGradient id="gradient2" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#8B5CF6" stopOpacity="0" />
                <stop offset="50%" stopColor="#8B5CF6" stopOpacity="0.5" />
                <stop offset="100%" stopColor="#8B5CF6" stopOpacity="0" />
              </linearGradient>
              <linearGradient id="gradient3" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#EC4899" stopOpacity="0" />
                <stop offset="50%" stopColor="#EC4899" stopOpacity="0.5" />
                <stop offset="100%" stopColor="#EC4899" stopOpacity="0" />
              </linearGradient>
            </defs>
          </svg>
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
      <section className="px-6 py-20 ">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            Core Capabilities That Set AEGIS Apart
          </h2>
          <div className="grid md:grid-cols-4 gap-8">
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-blue-500 transition-all">
              <Eye className="h-12 w-12 text-blue-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Advanced Threat Detection</h3>
              <p className="text-gray-300 text-sm">
                Our sophisticated AI engine detects and stops threats before they can damage your infrastructure and data.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-purple-500 transition-all">
              <Lock className="h-12 w-12 text-purple-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Real-time Data Encryption</h3>
              <p className="text-gray-300 text-sm">
                Enterprise-grade end-to-end encryption at all your business transactions and communications.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-green-500 transition-all">
              <Zap className="h-12 w-12 text-green-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Automated Incident Response</h3>
              <p className="text-gray-300 text-sm">
                Immediate response to security incidents, automated threat containment and remediation.
              </p>
            </div>
            <div className="bg-gray-900/50 p-6 rounded-xl border border-gray-700 hover:border-red-500 transition-all">
              <Shield className="h-12 w-12 text-red-400 mb-4" />
              <h3 className="text-xl font-semibold mb-3">Proactive Vulnerability Assessment</h3>
              <p className="text-gray-300 text-sm">
                Continuous monitoring and assessment of your security posture to prevent attacks.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Challenges & Solutions */}
      <section className="px-6 py-20">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            Solving Your Toughest Challenges
          </h2>
          <div className="grid md:grid-cols-2 gap-12 items-center">
            <div>
              <h3 className="text-2xl font-semibold mb-8 text-red-400">The Challenges You Face:</h3>
              <div className="space-y-4">
                <div className="flex items-start space-x-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  <p className="text-gray-300">Growing cybersecurity threats and increasingly complex attack vectors.</p>
                </div>
                <div className="flex items-start space-x-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  <p className="text-gray-300">Manual incident response leads to delays and increased damage.</p>
                </div>
                <div className="flex items-start space-x-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  <p className="text-gray-300">Lack of unified visibility across diverse IT environments.</p>
                </div>
                <div className="flex items-start space-x-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  <p className="text-gray-300">Compliance burdens and audit complexities.</p>
                </div>
              </div>
            </div>
            <div className="bg-gradient-to-br from-red-900/20 to-orange-900/20 p-8 rounded-xl">
              <div className="space-y-4">
                <div className="h-6 bg-red-500/60 rounded"></div>
                <div className="h-4 bg-red-400/40 rounded w-3/4"></div>
                <div className="h-4 bg-orange-500/50 rounded w-1/2"></div>
                <div className="h-6 bg-red-600/70 rounded w-5/6"></div>
              </div>
            </div>
          </div>
          
          <div className="mt-16">
            <h3 className="text-2xl font-semibold mb-8 text-blue-400">AEGIS: The Solution:</h3>
            <div className="grid md:grid-cols-2 gap-6">
              <div className="flex items-start space-x-3">
                <CheckCircle className="h-6 w-6 text-green-400 mt-1" />
                <p className="text-gray-300">Unified advanced defense platform to guard all layers from endpoint threats.</p>
              </div>
              <div className="flex items-start space-x-3">
                <CheckCircle className="h-6 w-6 text-green-400 mt-1" />
                <p className="text-gray-300">Advanced protocols and real-time alerts during imminent threat conditions.</p>
              </div>
              <div className="flex items-start space-x-3">
                <CheckCircle className="h-6 w-6 text-green-400 mt-1" />
                <p className="text-gray-300">A centralized dashboard provides a comprehensive, unified view of all systems.</p>
              </div>
              <div className="flex items-start space-x-3">
                <CheckCircle className="h-6 w-6 text-green-400 mt-1" />
                <p className="text-gray-300">Built-in compliance frameworks and automated reporting simplify audits and regulatory adherence.</p>
              </div>
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
            "/login.png",
            "/dashboard.png",
            "/cases.png",
            "/caseManagement.png",
            "/evidence.png",
            "/chat.png",
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

      {/* Client Testimonials */}
      <section className="px-6 py-20">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-4xl font-bold text-center mb-16">
            See AEGIS in Action & Hear From Our Clients
          </h2>
          <div className="grid md:grid-cols-2 gap-8">
            <div className="bg-gray-900/50 rounded-xl overflow-hidden border border-gray-700">
              <div className="aspect-video bg-gradient-to-br from-gray-800 to-gray-700 flex items-center justify-center">
                <div className="w-16 h-16 bg-white/10 rounded-full flex items-center justify-center hover:bg-white/20 transition-all cursor-pointer">
                  <Play className="h-8 w-8 text-white ml-1" />
                </div>
              </div>
              <div className="p-6">
                <p className="text-gray-300 mb-4">
                  "AEGIS transformed our security operations. The automated threat response and predictive protection saved our organization from a critical security breach."
                </p>
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 bg-blue-500 rounded-full flex items-center justify-center">
                    <User className="h-6 w-6 text-white" />
                  </div>
                  <div>
                    <p className="font-semibold">Sarah Chen</p>
                    <p className="text-sm text-gray-400">CISO, TechCorp Solutions</p>
                  </div>
                </div>
              </div>
            </div>
            
            <div className="bg-gray-900/50 rounded-xl overflow-hidden border border-gray-700">
              <div className="aspect-video bg-gradient-to-br from-purple-800 to-pink-800 flex items-center justify-center">
                <div className="w-16 h-16 bg-white/10 rounded-full flex items-center justify-center hover:bg-white/20 transition-all cursor-pointer">
                  <Play className="h-8 w-8 text-white ml-1" />
                </div>
              </div>
              <div className="p-6">
                <p className="text-gray-300 mb-4">
                  "The integration was seamless, and the proactive threat monitoring gave us the confidence to expand our digital operations securely."
                </p>
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 bg-purple-500 rounded-full flex items-center justify-center">
                    <User className="h-6 w-6 text-white" />
                  </div>
                  <div>
                    <p className="font-semibold">Michael Rodriguez</p>
                    <p className="text-sm text-gray-400">IT Director, Global Finance Inc</p>
                  </div>
                </div>
              </div>
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
          <div className="flex flex-wrap justify-center gap-4 mb-12">
            <button className="bg-blue-600 hover:bg-blue-700 px-6 py-2 rounded-lg text-sm font-medium">
              Request a Demo
            </button>
            <button className="bg-gray-700 hover:bg-gray-600 px-6 py-2 rounded-lg text-sm font-medium">
              Product Tour
            </button>
            <button className="bg-gray-700 hover:bg-gray-600 px-6 py-2 rounded-lg text-sm font-medium">
              Identity & Access
            </button>
          </div>
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
      <section className="px-6 py-20 bg-gradient-to-r from-pink-600 to-purple-600">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-4xl font-bold mb-6">
            Ready to Strengthen Your Security Posture?
          </h2>
          <p className="text-xl mb-8 opacity-90">
            Join thousands of organizations that trust AEGIS to protect their digital assets and maintain business continuity.
          </p>
          <button className="bg-white text-purple-700 hover:bg-gray-100 px-8 py-3 rounded-lg text-lg font-medium transition-all transform hover:scale-105">
            Get Started with AEGIS
          </button>
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

import React from 'react';
import './EvidenceViewer.css';
import { FaTachometerAlt, FaFolderOpen, FaComments, FaUser, FaCog, FaSignOutAlt, FaBell, FaDownload, FaShare, FaExpand, FaFile, FaCode, FaImage, FaVideo, FaComment } from 'react-icons/fa';
import {
  Bell,
  FileText,
  Filter,
  Folder,
  Home,
  Link,
  MessageSquare,
  Search,
  Settings,
  Share2,
  Plus,
  House,
  SlidersHorizontal,
  ArrowUpDown
} from "lucide-react";

import logo from './images/logo.png';

function EvidenceViewer() {
  return (
    <div className="evidence-viewer">
      <aside className="sidebar">
        <nav className="nav-links">
          <div className="nav-item">
            <House className="icon" />
            Dashboard
          </div>
          <div className="nav-item">
            <FaFolderOpen className="icon" />
            Case management
          </div>
          <div className="nav-item active">
            <FaFolderOpen className="icon" />
            Evidence Viewer
          </div>
          <div className="nav-item">
            <FaComments className="icon" />
            Secure chat
          </div>
        </nav>
        <div className="profile-section">
          <div className="user-info">
            <FaUser className="icon" />
            <span>Agent User</span>
          </div>
          <div className="settings-logout">
            <div><FaCog className="icon" /> settings</div>
            <div><FaSignOutAlt className="icon" /> Logout</div>
          </div>
        </div>
      </aside>

      <div className="content-area">
        <header className="topbar">
          <div className="logo-section">
            <img src={logo} alt="AEGIS Logo" className="logo" />
            <h2>AEGIS</h2>
          </div>
          <nav className="topnav">
            <span className="topnav-item">Dashboard</span>
            <span className="topnav-item active">Evidence Viewer</span>
            <span className="topnav-item">case management</span>
            <span className="topnav-item">Secure chat</span>
          </nav>
          <div className="topnav-right">
            <input type="text" placeholder="Search cases, evidence, users" />
            <FaBell className="icon" />
            <FaCog className="icon" />
            <FaUser className="icon" />
          </div>
        </header>

        <main className="content">
          <div className="case-files-section">
            <div className="case-number">IOS-0273</div>
            <div className="case-files-panel">
              <div className="case-files-header">
                <h3>Case files</h3>
              </div>
              <div className="case-files-controls">
                <div className="search-container">
                  <Search className="search-icon" size={16} />
                  <input 
                    type="text" 
                    placeholder="Search Evidence" 
                    className="evidence-search"
                  />
                </div>
                <div className="control-buttons">
                  <button className="control-btn">
                    <SlidersHorizontal size={16} />
                    filter
                  </button>
                  <button className="control-btn">
                    <ArrowUpDown size={16} />
                    sort
                  </button>
                </div>
              </div>
              <div className="case-file-item">
                <FaFile className="icon" />
                <span>system_logs.exe</span>
              </div>
              <div className="case-file-item">
                <FaFile className="icon" />
                <span>malware_sample.exe</span>
              </div>
            </div>
          </div>

          <div className="main-viewer">
            <div className="viewer-placeholder">
              Select a file to view
            </div>
          </div>

          <div className="right-panel">
            <div className="annotation-tools">
              <h4>Annotation tools</h4>
              <div className="tool-grid">
                <button className="tool-btn"><FaCode /></button>
                <button className="tool-btn"><FaImage /></button>
                <button className="tool-btn"><FaComment /></button>
                <button className="tool-btn"><FaVideo /></button>
              </div>
            </div>

            <div className="indicators-panel">
              <h4>Indicators of compromise (IOCs)</h4>
              <div className="indicator-item">
                <div className="ip-address">IP Address: 192.168.1.100</div>
                <div>CVE Reference: https://cve...</div>
              </div>
              <div className="indicator-item">
                <div className="hash-value">Hash (MD5):</div>
                <div>a1b2c3d4e5f67890</div>
                <div style={{color: '#ff6b6b'}}>Confidence: High</div>
              </div>
            </div>

            <div className="comments-panel">
              <h4>Comments & Collaboration</h4>
              <input 
                type="text" 
                placeholder="Add a new comment..." 
                className="comment-input"
              />
              <button className="add-comment-btn">Add comment</button>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

export default EvidenceViewer;
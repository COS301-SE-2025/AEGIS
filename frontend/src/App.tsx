import { Routes, Route } from "react-router-dom";
import { Navigate } from "react-router-dom";
//PAGES
import { LoginPage } from "./screens/LoginPage";
import { RegistrationPage } from "./screens/RegistrationPage";
import {ResetPasswordPage} from "./screens/ResetPasswordPage";
import {ForgotPasswordPage} from "./screens/ForgotPasswordPage";
import { CaseManagementPage } from "./screens/CaseManagementPage";
import { SecureChatPage } from "./screens/SecureChatPage";
import {SettingsPage} from "./screens/SettingsPage";
import { ProfilePage } from "./screens/ProfilePage";
import { DashBoardPage } from "./screens/DashboardPage";
import { LandingPage } from "./screens/LandingPage";
import {VerifyEmailPage} from "./screens/VerifyEmailPage/VerifyEmailPage";
import {TermsAndConditionsPage} from "./screens/TermsAndConditionsPage/TermsAndConditionsPage";
import {FAQ} from "./screens/FAQ"
import {About} from "./screens/About"
import { TutorialsPage } from "./screens/TutorialsPage";  
import NextStepsPage from "./screens/NextStepsPage/NextStepsPage";
import { NotificationsPage } from "./screens/NotificationsPage";
import {TenantsPage} from "./screens/TenantsPage";
import { TenantRegistrationPage } from "./screens/TenantRegistrationPage/TenantRegistrationPage";
import { TeamRegistrationPage } from "./screens/TeamRegistrationPage";
import { SystemAdminDashboard } from "./screens/SystemAdminDashboard";
import { TenantAdminDashboard } from "./screens/TenantAdminDashboard";
import { TeamsPage } from "./screens/TeamsPage";

import { ReportDashboard } from "./screens/ReportDashboard/ReportDashboard";
import {ReportEditor} from "./screens/ReportEditor/ReportEditor";
import { IOCPage } from "./screens/IOCPage";
import { ChainOfCustody } from "./screens/ChainOfCustody";

//FORMS
import {CreateCaseForm} from "./screens/CreateCasePage/CreateCasePage";
import { UploadEvidenceForm } from "./screens/UploadEvidencePage/UploadEvidencePage";
import {AssignCaseMembersForm} from "./screens/AssignCaseMembersPage/AssignCaseMembersPage";
import { EvidenceViewer } from "./screens/EvidenceViewer";
import { ShareCaseForm } from "./screens/ShareCasePage/ShareCasePage";

//sidebar toggle
import { SidebarProvider } from './context/SidebarToggleContext';
import { Toaster } from 'react-hot-toast';
import { useEffect, useState } from "react"; // make sure this import exists

//Notification WS
import NotificationsWSProvider from "./components/NotificationWSProvider";
import {DFIRAuditLogsPage} from "./screens/DFIRAuditLogsPage/DFIRAuditLogsPage";

function decodeRoleFromToken(): string {
  try {
    const token = sessionStorage.getItem("authToken");
    if (!token) return "";
    const base64 = token.split(".")[1];
    const payload = JSON.parse(
      atob(base64.replace(/-/g, "+").replace(/_/g, "/"))
    );
    return payload?.role ?? "";
  } catch {
    return "";
  }
}

function readRoleFromSession(): string {
  try {
    const raw = sessionStorage.getItem("user");
    if (!raw) return decodeRoleFromToken();
    const parsed = JSON.parse(raw);
    return parsed?.role ?? decodeRoleFromToken();
  } catch {
    return decodeRoleFromToken();
  }
}


export default function App() {
    const [role, setRole] = useState<string>(readRoleFromSession());
  const isDFIRAdmin = role === "DFIR Admin";

  // keep in sync if sessionStorage is updated elsewhere (e.g., after profile fetch)
  useEffect(() => {
    const onStorage = (e: StorageEvent) => {
      if (e.key === "user") setRole(readRoleFromSession());
    };
    window.addEventListener("storage", onStorage);

    // optional: listen for a custom event you can dispatch after updating user
    const onUserUpdated = () => setRole(readRoleFromSession());
    window.addEventListener("user-updated", onUserUpdated);

    return () => {
      window.removeEventListener("storage", onStorage);
      window.removeEventListener("user-updated", onUserUpdated);
    };
  }, []);
  return (
  <SidebarProvider>
        {/* Mount globally so WS is active across all routes */}
    <NotificationsWSProvider />
  <Toaster position="top-right" reverseOrder={false} />
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegistrationPage />} />
      <Route path="/reset-password" element={<ResetPasswordPage />} />
      <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      <Route path="/secure-chat" element={<SecureChatPage />} />
      <Route path="/settings" element={<SettingsPage />} />
      <Route path="/profile" element={<ProfilePage />} />
      <Route path="/dashboard" element={<DashBoardPage />} />
      <Route path="/create-case" element={<CreateCaseForm />} />
      <Route path ="/upload-evidence/:caseId" element={<UploadEvidenceForm />} />
      <Route path = "/assign-case-members/:caseId" element={<AssignCaseMembersForm />} />
      <Route path="/evidence-viewer/:caseId" element={<EvidenceViewer />} />
      <Route path="/case-management/:caseId" element={<CaseManagementPage />} />
      <Route path="/case-management" element={<CaseManagementPage />} />
      <Route path="/evidence-viewer" element={<EvidenceViewer />} />

      <Route path="/case/:caseId/next-steps" element={<NextStepsPage />} />
        <Route
          path="/report-dashboard"
          element={isDFIRAdmin ? <ReportDashboard /> : <Navigate to="/" replace />}
        />


      <Route path="/report-editor/:reportId" element={<ReportEditor />} />
      <Route path="*" element={<div style={{padding:16,color:'#fff'}}>Not Found</div>} />
      
      {/* Additional routes */}
      <Route path="/landing-page" element={<LandingPage />} />
      <Route path="/cases/:caseId/share" element={<ShareCaseForm />} />
      <Route path="/verify-email" element={<VerifyEmailPage />} />
      <Route path="/terms" element={<TermsAndConditionsPage />} />
      <Route path="/faq" element={<FAQ />} />
      <Route path="/about" element={<About />} />
      <Route path="/tutorials" element={<TutorialsPage />} />
      <Route path="/notifications" element={<NotificationsPage />} />
      <Route path="/tenants" element={<TenantsPage />} />
      <Route path="/tenant-registration" element={<TenantRegistrationPage />} />
      <Route path="/team-registration" element={<TeamRegistrationPage />} />
      <Route path="/system-admin-dashboard" element={<SystemAdminDashboard />} />
      <Route path="/tenant-admin-dashboard" element={<TenantAdminDashboard />} />
      <Route path="/teams" element={<TeamsPage />} />
      <Route path="/cases/:case_id/iocs" element={<IOCPage />} />
      <Route path="/chain-of-custody/:caseId" element={<ChainOfCustody />} />
      <Route path="/dfir-audit-logs" element={isDFIRAdmin ? <DFIRAuditLogsPage /> : <Navigate to="/" replace />} />

      {/* Fallback route */}
    </Routes>
    </SidebarProvider>
  );
}

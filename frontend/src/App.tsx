import { Routes, Route } from "react-router-dom";

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
import {ThemeProvider} from "./context/ThemeContext"
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
import { IOCPage } from "./screens/IOCPage";
//FORMS
import {CreateCaseForm} from "./screens/CreateCasePage/CreateCasePage";
import { UploadEvidenceForm } from "./screens/UploadEvidencePage/UploadEvidencePage";
import {AssignCaseMembersForm} from "./screens/AssignCaseMembersPage/AssignCaseMembersPage";
import { EvidenceViewer } from "./screens/EvidenceViewer";
import { ShareCaseForm } from "./screens/ShareCasePage/ShareCasePage";

//sidebar toggle
import { SidebarProvider } from './context/SidebarToggleContext';
import { Toaster } from 'react-hot-toast';


export default function App() {
  return (
  <SidebarProvider>
  <ThemeProvider>
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
      <Route path ="/upload-evidence" element={<UploadEvidenceForm />} />
      <Route path = "/assign-case-members" element={<AssignCaseMembersForm />} />
      <Route path="/evidence-viewer/:caseId" element={<EvidenceViewer />} />
      <Route path="/case-management/:caseId" element={<CaseManagementPage />} />
      <Route path="/case-management" element={<CaseManagementPage />} />
      <Route path="/evidence-viewer" element={<EvidenceViewer />} />

      <Route path="/case/:caseId/next-steps" element={<NextStepsPage />} />

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
      
      {/* Fallback route */}
    </Routes>
    </ThemeProvider>
    </SidebarProvider>
  );
}

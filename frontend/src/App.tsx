import { Routes, Route } from "react-router-dom";
import { LoginPage } from "./screens/LoginPage";
import { RegistrationPage } from "./screens/RegistrationPage";
import {ResetPasswordPage} from "./screens/ResetPasswordPage";
import { CaseManagementPage } from "./screens/CaseManagementPage";
import { SecureChatPage } from "./screens/SecureChatPage";
import {SettingsPage} from "./screens/SettingsPage";
import { ProfilePage } from "./screens/ProfilePage";
import { DashBoardPage } from "./screens/DashboardPage";

import {CreateCaseForm} from "./screens/CreateCasePage/CreateCasePage";
import { UploadEvidenceForm } from "./screens/UploadEvidencePage/UploadEvidencePage";
import {AssignCaseMembersForm} from "./screens/AssignCaseMembersPage/AssignCaseMembersPage";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LoginPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegistrationPage />} />
      <Route path="/reset-password" element={<ResetPasswordPage />} />
      <Route path="/case-management" element={<CaseManagementPage />} />
      <Route path="/secure-chat" element={<SecureChatPage />} />
      <Route path="/settings" element={<SettingsPage />} />
      <Route path="/profile" element={<ProfilePage />} />
      <Route path="/dashboard" element={<DashBoardPage />} />
      <Route path="/create-case" element={<CreateCaseForm />} />
      <Route path ="/upload-evidence" element={<UploadEvidenceForm />} />
      <Route path = "/assign-case-members" element={<AssignCaseMembersForm />} />
    </Routes>
  );
}

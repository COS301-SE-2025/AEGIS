import { Routes, Route } from "react-router-dom";
import { LoginPage } from "./screens/LoginPage";
import { RegistrationPage } from "./screens/RegistrationPage";
import ResetPasswordPage from "./screens/ResetPasswordPage";
import { CaseManagementPage } from "./screens/CaseManagementPage";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LoginPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegistrationPage />} />
      <Route path="/reset-password" element={<ResetPasswordPage />} />
      <Route path="/case-management" element={<CaseManagementPage />} />

    </Routes>
  );
}

import { Routes, Route } from "react-router-dom";
import { LoginPage } from "./screens/LoginPage";
import { RegistrationPage } from "./screens/RegistrationPage";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LoginPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegistrationPage />} />
      
    </Routes>
  );
}

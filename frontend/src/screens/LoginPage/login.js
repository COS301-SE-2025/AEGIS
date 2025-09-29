import { useState } from "react";
import { useNavigate } from "react-router-dom";

const useLoginForm = () => {
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  });
  const [errors, setErrors] = useState({});
  const navigate = useNavigate();

  const validate = () => {
    const newErrors = {};

    if (!formData.email.trim()) {
      newErrors.email = "Email is required";
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = "Email is invalid";
    }

    if (!formData.password.trim()) {
      newErrors.password = "Password is required";
    } else if (formData.password.length < 6) {
      newErrors.password = "Password must be at least 6 characters";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange = (e) => {
    const { id, value } = e.target;
    setFormData((prev) => ({ ...prev, [id]: value }));
  };
const handleSubmit = async (e) => {
  e.preventDefault();
  if (!validate()) return;

  try {
    const res = await fetch("http://localhost:8080/api/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(formData),
    });
    const payload = await res.json();

    if (res.ok && payload.success && payload.data?.token) {
      const token = payload.data.token;

      // 🔓 Decode JWT
      const base64Payload = token.split(".")[1];
      const decodedPayload = JSON.parse(atob(base64Payload));

      const userData = {
        id: decodedPayload.user_id,
        email: decodedPayload.email,
        name: decodedPayload.full_name,
        role: decodedPayload.role,
        tenantId: decodedPayload.tenant_id,
        teamId: decodedPayload.team_id,
      };

      // Store token & user info
      sessionStorage.setItem("authToken", token);
      sessionStorage.setItem("tenantId", userData.tenantId);
      sessionStorage.setItem("teamId", userData.teamId);
      sessionStorage.setItem("user", JSON.stringify(userData));
      window.dispatchEvent(new Event("auth:updated"));
      // Audit log
      const loginAuditEntry = {
        timestamp: new Date().toISOString(),
        user: userData.email,
        action: "User logged in",
        userId: userData.id,
      };
      const previousLogs = JSON.parse(localStorage.getItem("caseActivities") || "[]");
      localStorage.setItem("caseActivities", JSON.stringify([loginAuditEntry, ...previousLogs]));

      // Role-based redirect
      switch (userData.role) {
        case "System Admin":
          navigate("/system-admin-dashboard");
          break;
        case "Tenant Admin":
          navigate("/tenant-admin-dashboard");
          break;
        case "DFIR Admin":
        default:
          navigate("/dashboard");
          break;
      }
    } else {
      setErrors({ general: payload.message || "Login failed" });
    }
  } catch (err) {
    setErrors({ general: err.message || "Network error" });
  }
};

  return { formData, handleChange, handleSubmit, errors };
};

export default useLoginForm;

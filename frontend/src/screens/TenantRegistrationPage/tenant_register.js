import { useState } from "react";
import { useNavigate } from "react-router-dom";

const useRegistrationForm = () => {
  const [formData, setFormData] = useState({
    full_name: "",
    organization_name: "",
    domain_name: "",
    email: "",
    password: "",
    role: "",
  });

  const [errors, setErrors] = useState({});
  const navigate = useNavigate();

  const validate = () => {
    const newErrors = {};

    if (!formData.full_name.trim()) {
      newErrors.full_name = "Full name is required";
    }
    if (!formData.organization_name.trim()) {
      newErrors.organization_name = "Organization name is required";
    }
    if (!formData.domain_name.trim()) {
      newErrors.domain_name = "Domain name is required";
    }
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

    if (!formData.role) {
      newErrors.role = "Role must be selected";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

const handleChange = (e) => {
  const { id, value } = e.target;

  const newFormData = { ...formData, [id]: value };

  // Auto-generate password: random 8-character alphanumeric string
  if (id === "full_name") {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let randomPassword = '';
    for (let i = 0; i < 8; i++) {
      randomPassword += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    newFormData.password = randomPassword;
  }

  setFormData(newFormData);
};


  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validate()) return;

    try {
      const res = await fetch("http://localhost:8080/api/v1/register/tenant", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(formData),
      });

      const payload = await res.json();

      if (res.ok && payload.success) {
        navigate("/tenants");
      } else {
        setErrors({ general: payload.message || "Registration failed" });
      }
    } catch (err) {
      setErrors({ general: err.message || "Network error" });
    }
  };

  return { formData, handleChange, handleSubmit, errors };
};

export default useRegistrationForm;

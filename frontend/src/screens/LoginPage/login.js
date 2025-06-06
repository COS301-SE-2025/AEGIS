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
        // Store the real token
        sessionStorage.setItem("authToken", payload.data.token);
        sessionStorage.setItem(
          "user",
          JSON.stringify({ email: payload.data.email, id: payload.data.id })
        );
        navigate("/dashboard");
      } else {
        // Use general so your UI reads errors.general
        setErrors({ general: payload.message || "Login failed" });
      }
    } catch (err) {
      setErrors({ general: err.message || "Network error" });
    }
  };

  return { formData, handleChange, handleSubmit, errors };
};

export default useLoginForm;

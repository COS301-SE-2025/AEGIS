import { useState } from "react";
import { useNavigate } from "react-router-dom";

const useRegistrationForm = () => {
  const [formData, setFormData] = useState({
    full_name: "",
    team_name: "",
    email: "",
    password: "",
    role: "",
  });

  const [errors, setErrors] = useState({});
  const navigate = useNavigate();

  const validate = () => {
    const newErrors = {};

    if (!formData.team_name.trim()) {
      newErrors.team_name = "Team name is required";
    }
    if (!formData.full_name.trim()) {
      newErrors.full_name = "Full name is required";
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

  // Auto-generate password from full name
  if (id === "full_name") {
    const firstName = value.trim().split(" ")[0];
    const randomNum = Math.floor(1000 + Math.random() * 9000);
    newFormData.password = `${firstName}${randomNum}`;
  }

  setFormData(newFormData);
};

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validate()) return;

    try {
      const res = await fetch("http://localhost:8080/api/v1/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(formData),
      });

      const payload = await res.json();

      if (res.ok && payload.success) {
        navigate("/login");
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

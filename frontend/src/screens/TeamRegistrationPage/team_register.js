import { useState, useEffect } from "react";
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
  const [showPopup, setShowPopup] = useState(false); 
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
      const token = sessionStorage.getItem("authToken");
      if (!token) {
        setErrors({ general: "No auth token found, please login again" });
        return;
      }
      const res = await fetch("http://localhost:8080/api/v1/register/team", {
         method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`, 
        },
        body: JSON.stringify(formData),
      });

      const payload = await res.json();

      if (res.ok && payload.success) {
        setShowPopup(true);
      } else {
        setErrors({ general: payload.message || "Registration failed" });
      }
    } catch (err) {
      setErrors({ general: err.message || "Network error" });
    }
  };
  
    // auto-close popup after 3s with fade out
  useEffect(() => {
    if (showPopup) {
      const timer = setTimeout(() => {
        setShowPopup(false);
        navigate("/teams");
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [showPopup, navigate]);

  return { formData, handleChange, handleSubmit, errors,showPopup };
};

export default useRegistrationForm;

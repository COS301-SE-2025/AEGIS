import { useState } from "react";
import { useNavigate } from "react-router-dom";

const useRegistrationForm = () => {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    password: "",
    role: "",
  });

  const [errors, setErrors] = useState({});
  const navigate = useNavigate();

    const validate = () => {
    const newErrors = {};

    if (!formData.name.trim()) newErrors.name = "Name is required";

    if (!formData.email.trim()) {
        newErrors.email = "Email is required";
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
        newErrors.email = "Email is invalid";
    }

    if (!formData.password.trim()) {
        newErrors.password = "Password is required";
    } else if (formData.password.length < 6) {
        newErrors.password = "Password must be at least 6 characters";
    } else if (!/\d/.test(formData.password)) {
        newErrors.password = "Password must include at least one number";
    } else if (!/[!@#$%^&*(),.?":{}|<>]/.test(formData.password)) {
        newErrors.password = "Password must include at least one special character";
    }

    if (!formData.role) newErrors.role = "Role must be selected";

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
      const response = await fetch("http://localhost:8080/api/v1/admin/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        navigate("/login");
      } else {
        console.error("Registration failed");
      }
    } catch (error) {
      console.error("Error submitting form:", error);
    }
  };

  return { formData, handleChange, handleSubmit, errors };
};

export default useRegistrationForm;

import { JSX } from "react";
import { Button } from "../../components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Link } from "react-router-dom";
// @ts-ignore
import useRegistrationForm from "./register";

export const RegistrationPage = (): JSX.Element => {
  const { formData, errors, handleChange, handleSubmit } = useRegistrationForm();

  return (
    <div className="min-h-screen bg-gradient-to-br from-[#0f172a] via-[#1e293b] to-[#0f172a] flex items-center justify-center px-4">
      <div className="w-full max-w-xl">
        <Card className="rounded-2xl shadow-xl border border-white/10 bg-white/5 backdrop-blur-md text-white">
          <CardHeader className="text-center pt-8 pb-4">
            <img
              src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
              alt="AEGIS Logo"
              className="w-16 h-16 mx-auto mb-4"
            />
            <CardTitle className="text-2xl font-semibold drop-shadow-sm">
              Create your AEGIS account
            </CardTitle>
            <p className="text-muted-foreground text-sm">
              Enter your details to get started
            </p>
          </CardHeader>

          <CardContent className="px-6 py-8">
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Name */}
              <div>
                <Label htmlFor="full_name" className="text-sm font-medium">
                  Full Name
                </Label>
                <Input
                  id="full_name"
                  placeholder="Jane Doe"
                  value={formData.full_name}
                  onChange={handleChange}
                  className="h-11 bg-white/10 border border-white/20 placeholder-white/70 text-white"
                />
                {errors.full_name && <p className="text-red-300 text-xs mt-1">{errors.fullName}</p>}
              </div>

              {/* Email */}
              <div>
                <Label htmlFor="email" className="text-sm font-medium">
                  Email Address
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@aegis.com"
                  value={formData.email}
                  onChange={handleChange}
                  className="h-11 bg-white/10 border border-white/20 placeholder-white/70 text-white"
                />
                {errors.email && <p className="text-red-300 text-xs mt-1">{errors.email}</p>}
              </div>

              {/* Password */}
              <div>
                <Label htmlFor="password" className="text-sm font-medium">
                  Password
                </Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="Create a strong password"
                  value={formData.password}
                  onChange={handleChange}
                  className="h-11 bg-white/10 border border-white/20 placeholder-white/70 text-white"
                />
                {errors.password && <p className="text-red-300 text-xs mt-1">{errors.password}</p>}
              </div>

              {/* Role */}
              <div>
                <Label htmlFor="role" className="text-sm font-medium">
                  Role
                </Label>
                <select
                  id="role"
                  value={formData.role}
                  onChange={handleChange}
                  className="h-11 bg-white/10 border border-white/20 text-white placeholder-white/70 w-full rounded-lg px-4 focus:outline-none"
                >
                  <option value="" disabled hidden>Select your role</option>
                  {[
                    "Admin",
                    "Audit Reviewer",
                    "Cloud Forensics Specialist",
                    "Compliance Officer",
                    "Detection Engineer",
                    "DFIR Manager",
                    "Digital Evidence Technician",
                    "Disk Forensics Analyst",
                    "Endpoint Forensics Analyst",
                    "Evidence Archivist",
                    "Forensic Analyst",
                    "Forensics Analyst",
                    "Generic user",
                    "Image Forensics Analyst",
                    "Incident Commander",
                    "Incident Responder",
                    "IT Infrastructure Liaison",
                    "Legal Counsel",
                    "Legal/Compliance Liaison",
                    "Log Analyst",
                    "Malware Analyst",
                    "Memory Forensics Analyst",
                    "Mobile Device Analyst",
                    "Network Evidence Analyst",
                    "Packet Analyst",
                    "Policy Analyst",
                    "Reverse Engineer",
                    "SIEM Analyst",
                    "SOC Analyst",
                    "Threat Hunter",
                    "Threat Intelligence Analyst",
                    "Triage Analyst",
                    "Training Coordinator",
                    "Vulnerability Analyst"
                  ].map((role) => (
                    <option key={role} value={role} className="text-black">
                      {role}
                    </option>
                  ))}
                </select>
                {errors.role && <p className="text-red-300 text-xs mt-1">{errors.role}</p>}
              </div>

              {errors.general && (
                <p className="text-center text-red-400 text-sm">{errors.general}</p>
              )}

              {/* Submit */}
              <Button
                type="submit"
                className="w-full h-12 text-white font-semibold text-base bg-blue-600 hover:bg-blue-500 transition"
              >
                Sign Up
              </Button>

              {/* Redirect to login */}
              <p className="text-center text-white/80 text-sm pt-2">
                Already have an account?{" "}
                <Link to="/" className="text-blue-300 hover:underline font-medium">
                  Sign in
                </Link>
              </p>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

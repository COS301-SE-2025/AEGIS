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
import Confetti from "react-confetti";
// @ts-ignore
import useRegistrationForm from "./team_register";

export const TeamRegistrationPage = (): JSX.Element => {
  const { formData, errors, handleChange, handleSubmit,showPopup } = useRegistrationForm();

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
              Create a DFIR Team Account
            </CardTitle>
            <p className="text-muted-foreground text-sm">
              Enter team details to get started
            </p>
          </CardHeader>

          <CardContent className="px-6 py-8">
            <form onSubmit={handleSubmit} className="space-y-6">
              
              {/* Team */}
              <div>
                <Label htmlFor="team_name" className="text-sm font-medium">
                  Team Name
                </Label>
                <Input
                  id="team_name"
                  placeholder="Incident Intel Team"
                  value={formData.team_name}
                  onChange={handleChange}
                  className="h-11 bg-white/10 border border-white/20 placeholder-white/70 text-white"
                />
                {errors.team_name && <p className="text-red-300 text-xs mt-1">{errors.teamName}</p>}
              </div>

              {/* Name */}
              <div>
                <Label htmlFor="full_name" className="text-sm font-medium">
                  Full Name (DFIR Admin)
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
                  type="text"
                  placeholder="Password will be auto-generated"
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
                    "DFIR Admin",
                    
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
        {/* 🎉 AEGIS Popup */}
        {showPopup && (
          <div className="fixed inset-0 flex items-center justify-center z-50 pointer-events-none">
            <Confetti
              colors={["#1E3A8A", "#FFFFFF", "#000000"]}
              numberOfPieces={300}
              recycle={false}
            />
            <div className="bg-black text-white shadow-2xl rounded-3xl p-10 text-center max-w-lg border border-blue-500 animate-float-drift animate-fade-out">
              <h2 className="text-3xl font-extrabold text-blue-400 mb-4">🎉 You’ve been registered!</h2>
              <p className="text-gray-200 text-lg">
                An email has been sent to verify your email.
                <br />
                Welcome to the <span className="font-semibold">AEGIS</span> family, may your logs be clear and your cases epic, fellow <em>Aegy</em>!
              </p>
            </div>
          </div>
        )}
    </div>
  );
};

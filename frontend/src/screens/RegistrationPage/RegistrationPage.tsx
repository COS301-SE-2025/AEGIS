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
//@ts-ignore
import useRegistrationForm from "./register";

export const RegistrationPage = (): JSX.Element => {
  return (
    <div className="relative min-h-screen w-full overflow-hidden">
      {/* Background Grid of 6 Image Tiles */}
      <div className="absolute inset-0 grid grid-cols-3 grid-rows-2 gap-0 z-0">
        {Array.from({ length: 6 }).map((_, index) => (
          <div
            key={index}
            className="bg-cover bg-center"
            style={{
              backgroundImage:
                "url('https://img.freepik.com/premium-photo/data-schemas-computer-data-technologies-data-protection-generative-ai_655310-724.jpg')",
              filter: "brightness(1.0) saturate(2.5)",
            }}
          />
        ))}
      </div>

      {/* Dark overlay for better contrast */}
      <div className="absolute inset-0 bg-black/40 z-10" />

      {/* Registration Form Container */}
      <div className="relative z-20 flex items-center justify-center min-h-screen px-4">
        <Card className="w-full max-w-[615px] rounded-2xl bg-white/30 backdrop-blur-lg border border-white/20 shadow-2xl p-6 text-white">
          <CardHeader className="p-0">
            <div className="flex flex-col items-center">
              <img
                className="w-[100px] h-[90px] mb-6"
                alt="Logo"
                src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
              />
              <CardTitle className="text-3xl font-bold text-center mb-2 drop-shadow-md">
                Create your AEGIS account
              </CardTitle>
              <p className="font-light text-lg mb-8 text-center drop-shadow-sm">
                Enter your information to register
              </p>
            </div>
          </CardHeader>

          <CardContent className="p-0 space-y-6">
            <form className="w-full space-y-6">
              <div className="space-y-2">
                <Label htmlFor="name" className="text-lg font-medium">
                  Name
                </Label>
                <Input
                  id="name"
                  placeholder="Enter your Full Name"
                  className="h-[50px] rounded-[10px] border-white/30 bg-white/50 text-white placeholder-white/80"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="email" className="text-lg font-medium">
                  Email
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="Enter your email"
                  className="h-[50px] rounded-[10px] border-white/30 bg-white/50 text-white placeholder-white/80"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="password" className="text-lg font-medium">
                  Password
                </Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="Create a password"
                  className="h-[50px] rounded-[10px] border-white/30 bg-white/50 text-white placeholder-white/80"
                />
              </div>

              <div className="space-y-2">
              <Label htmlFor="role" className="text-lg font-medium mb-2 block">
                Role
              </Label>
              <select
                id="role"
                className="h-[50px] rounded-[10px] border-white/30 bg-white/50 text-white placeholder-white/80 px-4 focus:outline-none"
                defaultValue=""
              >
                <option value="" disabled hidden className="text-gray-400">
                  Select your role
                </option>
                <option className="text-black">Incident Responder</option>
                <option className="text-black">Forensics Analyst</option>
                <option className="text-black">Malware Analyst</option>
                <option className="text-black">Threat Intelligence Analyst</option>
                <option className="text-black">DFIR Manager</option>
                <option className="text-black">Legal/Compliance Liaison</option>
              </select>
            </div>



              {/* Sign Up Button */}
              <Button
                type="submit"
                className="w-full h-[60px] text-[22px] font-medium bg-[#1018ff] text-white hover:bg-[#0b13cc] transition"
              >
                Sign Up
              </Button>

              {/* Already have an account? */}
              <div className="text-center pt-2">
                <p className="text-base">
                  Already have an account?{" "}
                  <Button
                    variant="link"
                    className="text-[#a0c9ff] text-lg font-light p-0 h-auto align-baseline hover:underline"
                    asChild
                  >
                    <Link to="/">Sign in</Link>
                  </Button>
                </p>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Link } from "react-router-dom";
import { Button } from "../../components/ui/button";

export function ResetPasswordPage(): JSX.Element {
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
      {/* Reset Password Form Container */}
      <div className="relative z-20 flex items-center justify-center min-h-screen px-4">
      <Card className="w-full max-w-[615px] rounded-2xl bg-white/30 backdrop-blur-lg border border-white/20 shadow-2xl p-6 text-white">
        <CardHeader className="p-0">
          <div className="flex flex-col items-center">
            {/* Logo */}
            <img
              className="w-[122px] h-[111px] -mt-14 mb-6"
              alt="Logo"
              src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
            />
            {/* Headings */}
            <CardTitle className="text-3xl font-bold text-center mb-2 drop-shadow-md">
              Reset Your Password
            </CardTitle>
            <p className="font-extralight text-lg mb-8 text-center drop-shadow-sm">
              Enter your email address and we'll send you a reset link
            </p>
          </div>
        </CardHeader>

        <CardContent className="p-0 space-y-6">
          <form className="w-full space-y-6">
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

            <Button
                type="submit"
                className="w-full h-[60px] text-[22px] font-medium bg-[#1018ff] text-white hover:bg-[#0b13cc] transition"
              >
                Send Reset Link
              </Button>

            <div className="text-center mt-4">
              <p className="text-base">
                Remember your password?{" "}
                <Button
                  variant="link"
                  className="text-[#1018ff] text-xl font-light p-0 h-auto"
                  asChild
                >
                  <Link to="/login">Sign in</Link>
                </Button>
              </p>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
    </div>
  );
}

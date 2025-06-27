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

export const ResetPasswordPage = (): JSX.Element => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-[#0f172a] via-[#1e293b] to-[#0f172a] flex items-center justify-center px-4">
      <div className="w-full max-w-xl">
        <Card className="rounded-2xl shadow-xl border border-white/10 bg-white/5 backdrop-blur-md text-white">
          <CardHeader className="text-center pt-10 pb-4">
            <img
              src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
              alt="AEGIS Logo"
              className="w-16 h-16 mx-auto mb-4"
            />
            <CardTitle className="text-2xl font-semibold drop-shadow-sm">
              Reset Your Password
            </CardTitle>
            <p className="text-muted-foreground text-sm">
              Enter your email and we'll send you a reset link.
            </p>
          </CardHeader>

          <CardContent className="px-6 py-8">
            <form className="space-y-6">
              <div>
                <Label htmlFor="email" className="text-sm font-medium">
                  Email Address
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@aegis.com"
                  className="h-11 bg-white/10 border border-white/20 placeholder-white/70 text-white"
                />
              </div>

              <Button
                type="submit"
                className="w-full h-12 text-white font-semibold text-base bg-blue-600 hover:bg-blue-500 transition"
              >
                Send Reset Link
              </Button>

              <p className="text-center text-white/80 text-sm pt-4">
                Remember your password?{" "}
                <Link to="/login" className="text-blue-300 hover:underline font-medium">
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

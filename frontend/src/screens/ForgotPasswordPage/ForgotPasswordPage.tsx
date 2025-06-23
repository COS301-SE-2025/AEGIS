import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Button } from "../../components/ui/button";
import { Link } from "react-router-dom";

export const ForgotPasswordPage = (): JSX.Element => {
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
              Set New Password
            </CardTitle>
            <p className="text-muted-foreground text-sm">
              Enter your new password below to reset it
            </p>
          </CardHeader>

          <CardContent className="px-6 py-8">
            <form className="space-y-6">
              <div>
                <Label htmlFor="newPassword" className="text-sm font-medium">
                  New Password
                </Label>
                <Input
                  id="newPassword"
                  type="password"
                  placeholder="Enter new password"
                  className="h-11 bg-white/10 border border-white/20 placeholder-white/70 text-white"
                />
              </div>

              <div>
                <Label htmlFor="confirmPassword" className="text-sm font-medium">
                  Confirm New Password
                </Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder="Re-enter new password"
                  className="h-11 bg-white/10 border border-white/20 placeholder-white/70 text-white"
                />
              </div>

              <Button
                type="submit"
                className="w-full h-12 text-white font-semibold text-base bg-blue-600 hover:bg-blue-500 transition"
              >
                Reset Password
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

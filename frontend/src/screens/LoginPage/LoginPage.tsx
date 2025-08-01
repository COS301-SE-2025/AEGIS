import { JSX } from "react";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { Link } from "react-router-dom";
// @ts-ignore
import useLoginForm from "./login";

export const LoginPage = (): JSX.Element => {
  const { formData, handleChange, handleSubmit, errors } = useLoginForm();

  const formFields = [
    {
      id: "email",
      label: "Email address",
      placeholder: "you@aegis.com",
      type: "email",
    },
    {
      id: "password",
      label: "Password",
      placeholder: "Enter your password",
      type: "password",
    },
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-[#0f172a] via-[#1e293b] to-[#0f172a] flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <Card className="rounded-2xl shadow-xl border border-white/10 bg-white/5 backdrop-blur-md text-white">
          <CardHeader className="text-center pt-8 pb-2">
            <img
              src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
              alt="AEGIS Logo"
              className="w-16 h-16 mx-auto mb-4"
            />
            <CardTitle className="text-2xl font-semibold drop-shadow-sm">
              Welcome to AEGIS
            </CardTitle>
            <p className="text-muted-foreground text-sm">Sign in to continue</p>
          </CardHeader>
          <CardContent className="px-6 py-8">
            <form onSubmit={handleSubmit} className="space-y-6">
              {formFields.map((field) => (
                <div key={field.id}>
                  <label
                    htmlFor={field.id}
                    className="block text-sm font-medium mb-1 text-white"
                  >
                    {field.label}
                  </label>
                  <Input
                    id={field.id}
                    type={field.type}
                    placeholder={field.placeholder}
                    value={formData[field.id]}
                    onChange={handleChange}
                    className="h-11 bg-white/10 border border-white/20 placeholder-white/70 text-white focus:border-blue-500"
                  />
                  {errors[field.id] && (
                    <p className="text-red-400 text-xs mt-1">
                      {errors[field.id]}
                    </p>
                  )}
                </div>
              ))}

              {errors.general && (
                <p className="text-red-400 text-sm text-center">{errors.general}</p>
              )}

              <div className="flex justify-between items-center text-sm">
                <Link to="/reset-password" className="text-blue-300 hover:underline">
                  Forgot password?
                </Link>
              </div>

              <Button
                type="submit"
                className="w-full h-12 text-white text-base font-semibold bg-blue-600 hover:bg-blue-500 transition"
              >
                Sign In
              </Button>
            </form>

          </CardContent>
        </Card>
      </div>
    </div>
  );
};

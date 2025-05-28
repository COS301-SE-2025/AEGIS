import { JSX } from "react";
import { Button } from "../../components/ui/button";
import { Card, CardContent } from "../../components/ui/card";
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
      placeholder: "Enter your email",
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
    <div className="relative min-h-screen w-full overflow-hidden">
      {/* Background: Grid of repeated images */}
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

      {/* Subtle overlay to unify background appearance */}
      <div className="absolute inset-0 bg-black/40 z-10" />

      {/* Login Form */}
      <div className="relative z-20 flex items-center justify-center min-h-screen px-4">
        <div className="w-full max-w-[615px]">
          <Card className="rounded-2xl bg-white/30 backdrop-blur-lg border border-white/20 shadow-2xl transition hover:shadow-[0_0_30px_rgba(0,0,0,0.2)]">
            <CardContent className="p-8">
              <div className="flex flex-col items-center text-white">
                {/* Logo */}
                <img
                  className="w-[100px] h-[90px] mb-6"
                  alt="Logo"
                  src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
                />

                {/* Headings */}
                <h1 className="font-bold text-3xl mb-2 text-center drop-shadow-md">
                  Welcome to AEGIS
                </h1>
                <p className="font-light text-lg mb-8 text-center drop-shadow-sm">
                  Log in to your AEGIS account
                </p>

                {/* Form */}
                <form className="w-full space-y-6"  onSubmit={handleSubmit}>
                  {formFields.map((field) => (
                    <div key={field.id} className="space-y-2 text-white">
                      <label
                        htmlFor={field.id}
                        className="font-medium text-lg block drop-shadow-sm"
                      >
                        {field.label}
                      </label>
                      <Input
                        id={field.id}
                        type={field.type}
                        placeholder={field.placeholder}
                        value={formData[field.id]}
                        onChange={handleChange}
                        className="h-[50px] rounded-[10px] border-white/40 bg-white/50 text-white placeholder-white/80"
                      />
                      {errors[field.id] && (
                        <p className="text-red-400 text-sm">{errors[field.id]}</p>)}
                    </div>
                  ))}
                  {errors.general && (
                    <p className="text-red-400 text-sm text-center">{errors.general}</p>
                  )}

                  {/* Forgot password */}
                  <div className="flex justify-end">
                    <Link to="/reset-password">
                      <span className="font-light text-lg text-[#a0c9ff] hover:underline">
                        Forgot password?
                      </span>
                    </Link>
                  </div>

                  {/* Login Button */}
                  <Button
                    type="submit"
                    className="w-full h-[60px] text-[22px] font-medium bg-[#1018ff] text-white hover:bg-[#0b13cc] transition"
                  >
                    Login
                  </Button>

                  {/* Sign up */}
                  <div className="text-center pt-4 text-white">
                    <p className="text-base">
                      Don&apos;t have an account?{" "}
                      <Button
                        variant="link"
                        className="text-[#a0c9ff] text-lg font-light p-0 h-auto align-baseline hover:underline"
                        asChild
                      >
                        <Link to="/register">Sign up</Link>
                      </Button>
                    </p>
                  </div>
                </form>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
};

import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { LoaderPinwheelIcon } from "lucide-react";

export const VerifyEmailPage = () => {
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get("token");
    if (!token) {
      setStatus("error");
      return;
    }

    fetch("/api/auth/verify-email", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ token }),
    })
      .then((response) => {
        if (response.ok) {
          setStatus("success");
        } else {
          setStatus("error");
        }
      })
      .catch(() => {
        setStatus("error");
      });
  }, [searchParams]);

  useEffect(() => {
    if (status === "success") {
      const timer = setTimeout(() => {
        navigate("/login");
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [status, navigate]);

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col items-center justify-center font-mono p-4">
      {status === "loading" && (
        <>
          <LoaderPinwheelIcon className="animate-spin text-muted-foreground w-8 h-8" />
          <p className="mt-4 text-muted-foreground">Verifying your email...</p>
        </>
      )}

      {status === "success" && (
        <p className="text-primary text-lg animate-pulse">✅ Email verified. Redirecting...</p>
      )}

      {status === "error" && (
        <p className="text-destructive text-lg">❌ Invalid or expired token.</p>
      )}
    </div>
  );
};
